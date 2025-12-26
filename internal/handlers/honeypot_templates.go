package handlers

import (
	"andorralee/pkg/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HoneypotTemplate 蜜罐模板定义
type HoneypotTemplate struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Protocol     string            `json:"protocol"`
	ImageName    string            `json:"image_name"`
	DefaultPort  int               `json:"default_port"`
	Description  string            `json:"description"`
	Environment  map[string]string `json:"environment"`
	PortMappings map[string]string `json:"port_mappings"`
}

// 预定义蜜罐模板
var honeypotTemplates = []HoneypotTemplate{
	{
		ID:          "ssh-cowrie",
		Name:        "SSH蜜罐 (Cowrie)",
		Protocol:    "ssh",
		ImageName:   "andorralee/cowrie:v0.1",
		DefaultPort: 22,
		Description: "基于Cowrie的SSH蜜罐，模拟SSH服务器",
		Environment: map[string]string{
			"COWRIE_HOSTNAME":  "server",
			"COWRIE_LOG_LEVEL": "INFO",
		},
		PortMappings: map[string]string{
			"22":   "2222",
			"2222": "2223",
		},
	},
	{
		ID:          "http-dionaea",
		Name:        "HTTP蜜罐 (Dionaea)",
		Protocol:    "http",
		ImageName:   "andorralee/protocollure-r2:v1.0",
		DefaultPort: 80,
		Description: "基于Dionaea的HTTP蜜罐，模拟Web服务器",
		Environment: map[string]string{
			"DIONAEA_LOG_LEVEL": "info",
		},
		PortMappings: map[string]string{
			"80":  "8080",
			"443": "8443",
		},
	},
	{
		ID:          "ftp-dionaea",
		Name:        "FTP蜜罐 (Dionaea)",
		Protocol:    "ftp",
		ImageName:   "andorralee/protocollure-r2:v1.0",
		DefaultPort: 21,
		Description: "基于Dionaea的FTP蜜罐，模拟FTP服务器",
		Environment: map[string]string{
			"DIONAEA_LOG_LEVEL": "info",
		},
		PortMappings: map[string]string{
			"21": "2121",
		},
	},
	{
		ID:          "telnet-cowrie",
		Name:        "Telnet蜜罐 (Cowrie)",
		Protocol:    "telnet",
		ImageName:   "andorralee/cowrie:v0.1",
		DefaultPort: 23,
		Description: "基于Cowrie的Telnet蜜罐，模拟Telnet服务器",
		Environment: map[string]string{
			"COWRIE_HOSTNAME":  "server",
			"COWRIE_LOG_LEVEL": "INFO",
			"COWRIE_TELNET":    "true",
		},
		PortMappings: map[string]string{
			"23":   "2323",
			"2323": "2324",
		},
	},
	{
		ID:          "mysql-honeypot",
		Name:        "MySQL蜜罐",
		Protocol:    "mysql",
		ImageName:   "andorralee/dbdefende-r2:v1.0",
		DefaultPort: 3306,
		Description: "数据库蜜罐，模拟MySQL服务器",
		Environment: map[string]string{
			"HONEYPOT_TYPE": "mysql",
			"LOG_LEVEL":     "info",
		},
		PortMappings: map[string]string{
			"3306": "13306",
		},
	},
}

// GetHoneypotTemplates 获取所有蜜罐模板
func GetHoneypotTemplates(c *gin.Context) {
	utils.ResponseSuccess(c, honeypotTemplates)
}

// GetHoneypotTemplateByID 根据ID获取蜜罐模板
func GetHoneypotTemplateByID(c *gin.Context) {
	id := c.Param("id")

	for _, template := range honeypotTemplates {
		if template.ID == id {
			utils.ResponseSuccess(c, template)
			return
		}
	}

	utils.ResponseError(c, http.StatusNotFound, "蜜罐模板不存在")
}

// DeployHoneypotFromTemplate 从模板部署蜜罐
func DeployHoneypotFromTemplate(c *gin.Context) {
	templateID := c.Param("id")

	// 查找模板
	var template *HoneypotTemplate
	for _, t := range honeypotTemplates {
		if t.ID == templateID {
			template = &t
			break
		}
	}

	if template == nil {
		utils.ResponseError(c, http.StatusNotFound, "蜜罐模板不存在")
		return
	}

	// 获取部署参数
	var deployReq struct {
		Name        string            `json:"name" binding:"required"`
		CustomPorts map[string]string `json:"custom_ports,omitempty"`
		CustomEnv   map[string]string `json:"custom_env,omitempty"`
		AutoStart   bool              `json:"auto_start"`
	}

	if err := c.ShouldBindJSON(&deployReq); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 构建容器实例请求
	portMappings := template.PortMappings
	if deployReq.CustomPorts != nil {
		portMappings = deployReq.CustomPorts
	}

	environment := template.Environment
	if deployReq.CustomEnv != nil {
		// 合并环境变量
		for k, v := range deployReq.CustomEnv {
			environment[k] = v
		}
	}

	instanceReq := CreateContainerInstanceRequest{
		Name:          deployReq.Name,
		HoneypotName:  template.ID + "-" + deployReq.Name,
		ImageName:     template.ImageName,
		Protocol:      template.Protocol,
		InterfaceType: "network",
		PortMappings:  portMappings,
		Environment:   environment,
		Description:   "从模板 " + template.Name + " 部署",
		AutoStart:     deployReq.AutoStart,
	}

	// 手动创建内存容器实例
	instanceMutex.Lock()
	instance := &MemoryContainerInstance{
		ID:            nextID,
		Name:          instanceReq.Name,
		HoneypotName:  instanceReq.HoneypotName,
		ContainerName: instanceReq.HoneypotName,
		ContainerID:   "",
		IP:            "0.0.0.0",
		HoneypotIP:    "",
		Port:          template.DefaultPort,
		Protocol:      instanceReq.Protocol,
		InterfaceType: instanceReq.InterfaceType,
		Status:        "created",
		ImageName:     instanceReq.ImageName,
		ImageID:       "",
		PortMappings:  instanceReq.PortMappings,
		Environment:   instanceReq.Environment,
		CreateTime:    time.Now(),
		UpdateTime:    time.Now(),
		Description:   instanceReq.Description,
	}
	memoryInstances[nextID] = instance
	nextID++
	instanceMutex.Unlock()

	// 如果设置了自动启动，则尝试创建并启动实际的Docker容器
	if instanceReq.AutoStart {
		// 这里可以调用实际的容器创建逻辑
		// 为了简化，我们只更新状态为"auto_start_requested"
		instanceMutex.Lock()
		instance.Status = "auto_start_requested"
		instanceMutex.Unlock()

		// 在实际项目中，这里应该调用Docker API创建和启动容器
		// 例如：CreateAndStartDockerContainer(instanceReq)
		fmt.Printf("模板部署: AutoStart已启用，容器将自动启动 - %s\n", instance.Name)
	}

	utils.ResponseSuccess(c, map[string]interface{}{
		"message":       "从模板部署蜜罐成功",
		"template_id":   templateID,
		"template_name": template.Name,
		"instance_id":   instance.ID,
		"instance_name": instance.Name,
		"protocol":      instance.Protocol,
		"image_name":    instance.ImageName,
		"port_mappings": instance.PortMappings,
		"environment":   instance.Environment,
		"create_time":   instance.CreateTime,
	})
}

// GetSupportedProtocols 获取支持的协议列表
func GetSupportedProtocols(c *gin.Context) {
	protocols := make(map[string][]HoneypotTemplate)

	for _, template := range honeypotTemplates {
		protocols[template.Protocol] = append(protocols[template.Protocol], template)
	}

	utils.ResponseSuccess(c, protocols)
}
