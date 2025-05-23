package services

import (
	"andorralee/internal/config"
	"context"
	"github.com/docker/docker/api/types/image"
	"io"
)

// PullDockerImage 拉取 Docker 镜像
func PullDockerImage(imageName string) error {
	reader, err := config.DockerCli.ImagePull(context.Background(), imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// 等待拉取完成
	_, _ = io.Copy(io.Discard, reader)
	return nil
}

// ListDockerImages 列出本地镜像
func ListDockerImages() ([]image.Summary, error) {
	return config.DockerCli.ImageList(context.Background(), image.ListOptions{})
}
