package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/register", useregister) // 注册
	r.POST("/login", userlogin)      // 登录

	// 使用 JWT 验证中间件
	authGroup := r.Group("/")
	authGroup.Use(JWTMiddleware())
	authGroup.POST("/todo", TodoCreation)          // 增
	authGroup.DELETE("/todo/:index", TodoDeletion) // 删(不改动序号)
	authGroup.PUT("/todo/:index", TodoUpdate)      // 改
	authGroup.GET("/todo", ListTodos)              // 查(使用条件筛选)
	authGroup.GET("/todo/:index", GetTodo)         // 获取单个 todo 信息

	r.Run(":8100") // 运行在 8100 端口
}
