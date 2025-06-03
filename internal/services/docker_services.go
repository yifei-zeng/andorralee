package services

import (
	"andorralee/internal/config"
	"andorralee/internal/repositories"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
)

// DockerImageService Docker镜像服务
type DockerImageService struct {
	Repo    repositories.DockerImageRepository
	LogRepo repositories.DockerImageLogRepository
}

// NewDockerImageService 创建Docker镜像服务
func NewDockerImageService() (*DockerImageService, error) {
	if config.MySQLDB == nil {
		return nil, fmt.Errorf("MySQL数据库未初始化")
	}

	return &DockerImageService{
		Repo:    repositories.NewMySQLDockerImageRepo(config.MySQLDB),
		LogRepo: repositories.NewMySQLDockerImageLogRepo(config.MySQLDB),
	}, nil
}

// IsDockerAvailable 检查Docker客户端是否可用
func IsDockerAvailable() bool {
	return config.DockerCli != nil
}

// PullDockerImage 拉取 Docker 镜像并记录到数据库
func PullDockerImage(imageName string) error {
	// 拉取镜像
	reader, err := config.DockerCli.ImagePull(context.Background(), imageName, image.PullOptions{})
	if err != nil {
		// 记录失败日志
		logPullImageOperation(imageName, "", "failed", err.Error())
		return err
	}
	defer reader.Close()

	// 等待拉取完成
	_, _ = io.Copy(io.Discard, reader)

	// 获取镜像信息并同步到数据库
	images, err := config.DockerCli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		logPullImageOperation(imageName, "", "failed", "镜像拉取成功但获取信息失败: "+err.Error())
		return nil
	}

	// 查找刚拉取的镜像
	var targetImage *image.Summary
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			if repoTag == imageName || (strings.HasPrefix(repoTag, imageName+":") && !strings.Contains(imageName, ":")) {
				targetImage = &img
				break
			}
		}
		if targetImage != nil {
			break
		}
	}

	if targetImage != nil {
		// 同步到数据库
		imageService, err := NewDockerImageService()
		if err != nil {
			logPullImageOperation(imageName, targetImage.ID, "success", "镜像拉取成功但数据库同步失败: "+err.Error())
			return nil
		}

		// 解析仓库和标签
		repo, tag := parseRepoTag(imageName)

		// 检查镜像是否已存在
		existingImage, _ := imageService.Repo.GetByImageID(targetImage.ID)
		if existingImage != nil {
			// 更新现有记录
			existingImage.Repository = repo
			existingImage.Tag = tag
			existingImage.Size = targetImage.Size
			existingImage.Digest = targetImage.ID
			existingImage.UpdatedAt = time.Now()
			err = imageService.Repo.Update(existingImage)
		} else {
			// 创建新记录
			newImage := &repositories.DockerImage{
				ImageID:    targetImage.ID,
				Repository: repo,
				Tag:        tag,
				Digest:     targetImage.ID,
				Size:       targetImage.Size,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			err = imageService.Repo.Create(newImage)
		}

		if err != nil {
			logPullImageOperation(imageName, targetImage.ID, "success", "镜像拉取成功但数据库同步失败: "+err.Error())
		} else {
			logPullImageOperation(imageName, targetImage.ID, "success", "镜像拉取成功并已同步到数据库")
		}
	} else {
		logPullImageOperation(imageName, "", "failed", "镜像拉取可能成功但未找到镜像信息")
	}

	return nil
}

// ListDockerImages 从数据库中列出所有镜像
func ListDockerImages() ([]image.Summary, error) {
	// 首先从Docker获取最新镜像列表
	images, err := config.DockerCli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 同步到数据库
	SyncDockerImagesToDB(images)

	return images, nil
}

// GetDockerImageByID 根据ID获取镜像详情
func GetDockerImageByID(imageID string) (*types.ImageInspect, error) {
	inspect, _, err := config.DockerCli.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		return nil, fmt.Errorf("获取镜像详情失败: %v", err)
	}

	// 记录查询日志
	logImageOperation(imageID, "inspect", "success", "查询镜像详情成功")

	return &inspect, nil
}

// DeleteDockerImage 删除本地Docker镜像并从数据库中移除
func DeleteDockerImage(imageID string) error {
	// 从Docker中删除镜像
	_, err := config.DockerCli.ImageRemove(context.Background(), imageID, image.RemoveOptions{Force: true})
	if err != nil {
		// 记录失败日志
		logImageOperation(imageID, "delete", "failed", "删除Docker镜像失败: "+err.Error())
		return fmt.Errorf("删除Docker镜像失败: %v", err)
	}
	return nil
}

// TagDockerImage 更新镜像标签（打标签）
func TagDockerImage(imageID, newRepo, newTag string) error {
	// 给镜像打标签
	err := config.DockerCli.ImageTag(context.Background(), imageID, newRepo+":"+newTag)
	if err != nil {
		logImageOperation(imageID, "tag", "failed", "镜像打标签失败: "+err.Error())
		return fmt.Errorf("镜像打标签失败: %v", err)
	}

	logImageOperation(imageID, "tag", "success", fmt.Sprintf("镜像打标签成功: %s:%s", newRepo, newTag))
	return nil
}

// SyncDockerImagesToDB 同步Docker镜像到数据库
func SyncDockerImagesToDB(images []image.Summary) {
	if config.MySQLDB == nil {
		return
	}

	imageService, err := NewDockerImageService()
	if err != nil {
		fmt.Printf("创建Docker镜像服务失败: %v\n", err)
		return
	}

	for _, img := range images {
		if len(img.RepoTags) == 0 {
			continue
		}

		for _, repoTag := range img.RepoTags {
			if repoTag == "<none>:<none>" {
				continue
			}

			repo, tag := parseRepoTag(repoTag)

			// 检查镜像是否已存在
			existingImage, _ := imageService.Repo.GetByImageID(img.ID)
			if existingImage != nil {
				// 更新现有记录
				existingImage.Repository = repo
				existingImage.Tag = tag
				existingImage.Size = img.Size
				existingImage.UpdatedAt = time.Now()
				imageService.Repo.Update(existingImage)
			} else {
				// 创建新记录
				newImage := &repositories.DockerImage{
					ImageID:    img.ID,
					Repository: repo,
					Tag:        tag,
					Digest:     img.ID,
					Size:       img.Size,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				imageService.Repo.Create(newImage)
			}
		}
	}
}

// parseRepoTag 解析仓库名和标签
func parseRepoTag(repoTag string) (string, string) {
	parts := strings.Split(repoTag, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return repoTag, "latest"
}

// logPullImageOperation 记录镜像拉取操作日志
func logPullImageOperation(imageName, imageID, status, message string) {
	// 记录到控制台
	fmt.Printf("[%s] 镜像操作: %s, 镜像ID: %s, 状态: %s, 消息: %s\n",
		time.Now().Format("2006-01-02 15:04:05"), imageName, imageID, status, message)

	// 记录到数据库
	if config.MySQLDB != nil {
		logRepo := repositories.NewMySQLDockerImageLogRepo(config.MySQLDB)
		log := &repositories.DockerImageLog{
			ImageID:   imageID,
			ImageName: imageName,
			Operation: "pull",
			Status:    status,
			Message:   message,
			Details:   fmt.Sprintf("镜像拉取操作: %s", imageName),
		}
		if err := logRepo.Create(log); err != nil {
			fmt.Printf("保存镜像操作日志失败: %v\n", err)
		}
	}
}

// logImageOperation 记录镜像操作日志
func logImageOperation(imageID, operation, status, message string) {
	// 记录到控制台
	fmt.Printf("[%s] 镜像操作: %s, 镜像ID: %s, 状态: %s, 消息: %s\n",
		time.Now().Format("2006-01-02 15:04:05"), operation, imageID, status, message)

	// 记录到数据库
	if config.MySQLDB != nil {
		logRepo := repositories.NewMySQLDockerImageLogRepo(config.MySQLDB)
		log := &repositories.DockerImageLog{
			ImageID:   imageID,
			Operation: operation,
			Status:    status,
			Message:   message,
			Details:   fmt.Sprintf("镜像%s操作", operation),
		}
		if err := logRepo.Create(log); err != nil {
			fmt.Printf("保存镜像操作日志失败: %v\n", err)
		}
	}
}
