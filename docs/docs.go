package docs

import (
	"github.com/swaggo/swag"
)

var (
	// SwaggerInfo holds exported Swagger Info so clients can modify it
	SwaggerInfo = &swag.Spec{
		Version:     "1.0",
		Title:       "Andorralee Docker API",
		Description: "管理 Docker 镜像和数据库的接口",
		Host:        "localhost:8080",
		BasePath:    "/api/v1",
		Schemes:     []string{"http"},
	}
)

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
