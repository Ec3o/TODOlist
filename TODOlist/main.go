package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"strconv"
	"time"
)

type TODO struct {
	Content  string `json:"content"`
	Done     bool   `json:"done"`
	Deadline string `json:"deadline"`
}

var todosFile = "todos.json"

func main() {
	r := gin.Default() //导入了Gin框架并使用默认的配置创建了一个Gin的路由引擎

	// 添加 TODO
	r.POST("/todo", func(c *gin.Context) {
		var todo TODO
		if err := c.BindJSON(&todo); err != nil {
			c.JSON(400, gin.H{"error": "抱歉，您提供的TODO数据格式不正确"})
			return
		}

		// 检查是否传入截止时间字段
		if todo.Deadline == "" {
			// 如果没有传入，设置默认值为今天日期往后七天
			defaultDeadline := time.Now().Add(time.Hour * 24 * 7)
			todo.Deadline = defaultDeadline.Format("2006-01-02")
		} else {
			// 尝试解析截止时间字段
			parsedDeadline, err := time.Parse("2006-01-02", todo.Deadline)
			if err != nil || parsedDeadline.Before(time.Now()) {
				// 如果解析失败或时间早于当前时间，返回错误
				c.JSON(400, gin.H{"error": "无效的截止时间，格式应为 '2006-01-02' 并且不能是过去的时间"})
				return
			}
		}

		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		// 添加新的TODO
		existingTodos = append(existingTodos, todo)

		// 保存更新后的TODO数据
		err = saveTodosToFile(existingTodos)
		if err != nil {
			c.JSON(500, gin.H{"error": "无法保存TODO数据"})
			return
		}

		c.JSON(200, gin.H{"status": "ok"})
	})

	// 删除 TODO，并返回被删除的 TODO 的详细信息
	r.DELETE("/todo/:index", func(c *gin.Context) {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil || index < 0 {
			c.JSON(404, gin.H{"error": "抱歉，您要删除的ToDo目前不存在，请先创建"})
			return
		}

		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		// 检查索引是否超出范围
		if index >= len(existingTodos) {
			c.JSON(404, gin.H{"error": "抱歉，您要删除的ToDo目前不存在，请先创建"})
			return
		}

		// 删除指定索引的 TODO 项
		deletedTodo := existingTodos[index]
		existingTodos = append(existingTodos[:index], existingTodos[index+1:]...)

		// 保存更新后的TODO数据
		err = saveTodosToFile(existingTodos)
		if err != nil {
			c.JSON(500, gin.H{"error": "无法保存TODO数据"})
			return
		}

		c.JSON(200, gin.H{"status": "ok", "deletedTodo": deletedTodo})
	})

	// 修改 TODO
	r.PUT("/todo/:index", func(c *gin.Context) {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil || index < 0 {
			c.JSON(404, gin.H{"error": "抱歉，您要修改的ToDo目前不存在，请先创建"})
			return
		}

		var todo TODO
		if err := c.BindJSON(&todo); err != nil {
			c.JSON(400, gin.H{"error": "抱歉，您提供的TODO数据格式不正确"})
			return
		}

		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		// 检查索引是否超出范围
		if index >= len(existingTodos) {
			c.JSON(404, gin.H{"error": "抱歉，您要修改的ToDo目前不存在，请先创建"})
			return
		}

		// 更新指定索引的 TODO 项
		existingTodos[index] = todo

		// 保存更新后的TODO数据
		err = saveTodosToFile(existingTodos)
		if err != nil {
			c.JSON(500, gin.H{"error": "无法保存TODO数据"})
			return
		}

		c.JSON(200, gin.H{"status": "ok"})
	})

	// 列出 TODO
	r.GET("/todo", func(c *gin.Context) {
		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		c.JSON(200, existingTodos)
	})

	// 查询 TODO
	// 查询 TODO
	r.GET("/todo/:index", func(c *gin.Context) {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil || index < 0 {
			c.JSON(404, gin.H{"error": "抱歉，您访问的ToDo目前不存在，请先创建"})
			return
		}

		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		// 检查索引是否超出范围
		if index >= len(existingTodos) {
			c.JSON(404, gin.H{"error": "抱歉，您访问的ToDo目前不存在，请先创建"})
			return
		}

		c.JSON(200, existingTodos[index])
	})

	// 今日 TODO
	// 列出今天的 TODO
	r.GET("/list_today", func(c *gin.Context) {
		// 获取今天的日期
		today := time.Now().Format("2006-01-02")

		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		// 创建一个新的切片来存储符合条件的 TODO 数据
		todayTodos := []TODO{}

		// 遍历现有的 TODO 数据，找到截止日期为今天的 TODO
		for _, todo := range existingTodos {
			if todo.Deadline == today {
				todayTodos = append(todayTodos, todo)
			}
		}

		if len(todayTodos) == 0 {
			c.JSON(404, gin.H{"message": "今天没有截止的TODO"})
			return
		}

		c.JSON(200, todayTodos)
	})

	//本周 TODO
	// 列出本周内的 TODO
	// 查询本周的 TODO
	r.GET("/list_week", func(c *gin.Context) {
		// 获取本周一的日期
		today := time.Now()
		weekday := today.Weekday()
		diff := int(weekday - time.Monday)
		if diff < 0 {
			diff += 7
		}
		startOfWeek := today.AddDate(0, 0, -diff)

		// 获取下周一的日期
		endOfWeek := startOfWeek.AddDate(0, 0, 7)

		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		// 创建一个新的切片来存储符合条件的 TODO 数据
		thisWeekTodos := []TODO{}

		// 遍历现有的 TODO 数据，找到截止日期在本周内的 TODO
		for _, todo := range existingTodos {
			deadline, err := time.Parse("2006-01-02", todo.Deadline)
			if err != nil {
				continue // 如果日期格式不正确，跳过
			}

			// 检查是否截止日期在本周内
			if deadline.After(startOfWeek) && deadline.Before(endOfWeek) {
				thisWeekTodos = append(thisWeekTodos, todo)
			}
		}

		if len(thisWeekTodos) == 0 {
			c.JSON(404, gin.H{"message": "本周没有截止的 TODO"})
			return
		}

		c.JSON(200, thisWeekTodos)
	})

	// 运行在端口8100
	r.Run(":8100")
}
func loadTodosFromFile() ([]TODO, error) {
	data, err := ioutil.ReadFile(todosFile)
	if err != nil {
		return nil, err
	}

	var todos []TODO
	err = json.Unmarshal(data, &todos)
	if err != nil {
		return nil, err
	}

	return todos, nil
}
func saveTodosToFile(todos []TODO) error {
	data, err := json.Marshal(todos)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(todosFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
