package main

import (
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
	"time"
)

type TODO struct {
	Content  string `json:"content"`
	Done     bool   `json:"done"`
	Deadline string `json:"deadline"`
}

var todosFile = "todos.json"

var deletedTodoIndexes []int

func main() {
	r := gin.Default()

	r.POST("/todo", TodoCreation)
	r.DELETE("/todo/:index", TodoDeletion)
	r.PUT("/todo/:index", TodoUpdate)
	r.GET("/todo", ListTodos)
	r.GET("/todo/:index", GetTodo)
	r.GET("/unfinished_todo", ListUnfinishedTodos)
	r.Run(":8100")
}

func TodoCreation(c *gin.Context) {
	var todo TODO
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(400, ErrInvalidTODOFormat)
		return
	}

	if todo.Deadline == "" {
		defaultDeadline := time.Now().Add(time.Hour * 24 * 7)
		todo.Deadline = defaultDeadline.Format("2006-01-02")
	} else {
		parsedDeadline, err := time.Parse("2006-01-02", todo.Deadline)
		if err != nil || parsedDeadline.Before(time.Now()) {
			c.JSON(400, ErrInvalidDeadline)
			return
		}
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	existingTodos = append(existingTodos, todo)

	err = saveTodosToFile(existingTodos)
	if err != nil {
		c.JSON(500, ErrSaveTODOData)
		return
	}

	c.JSON(200, gin.H{"status": "数据提交成功"})
}

func TodoDeletion(c *gin.Context) {
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrTODONotFound)
		return
	}

	if index >= len(existingTodos) {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	// 将索引添加到已删除的待办事项列表中
	deletedTodoIndexes = append(deletedTodoIndexes, index)

	// 更新已删除的待办事项回到列表
	existingTodos[index].Content = "此Todo已被删除"
	existingTodos[index].Done = true

	err = saveTodosToFile(existingTodos)
	if err != nil {
		c.JSON(500, ErrSaveTODOData)
		return
	}

	c.JSON(200, gin.H{"status": "删除成功", "被删除的数据是": existingTodos[index]})
}

func TodoUpdate(c *gin.Context) {
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	var todo TODO
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(400, ErrInvalidTODOFormat)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	if index >= len(existingTodos) {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	existingTodos[index] = todo

	err = saveTodosToFile(existingTodos)
	if err != nil {
		c.JSON(500, ErrSaveTODOData)
		return
	}

	c.JSON(200, gin.H{"status": "修改成功"})
}

func ListTodos(c *gin.Context) {
	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	// 获取查询参数
	deadline := c.DefaultQuery("deadline", "")
	reverse := c.DefaultQuery("reverse", "false")
	finished := c.DefaultQuery("finished", "")

	// 转换 reverse 字符串为布尔值
	reverseSort := (reverse == "true")

	// 根据 finished 参数过滤待办事项
	filteredTodos := []TODO{}
	for index, todo := range existingTodos {
		// 检查索引是否在 deletedTodoIndexes 中，如果在就跳过
		if containsIndex(index, deletedTodoIndexes) {
			continue
		}

		if (finished == "true" && todo.Done) || (finished == "false" && !todo.Done) || finished == "" {
			if deadline == "" || (deadline != "" && todo.Deadline <= deadline) {
				filteredTodos = append(filteredTodos, todo)
			}
		}
	}

	// 根据 reverseSort 参数排序
	if reverseSort {
		sort.Slice(filteredTodos, func(i, j int) bool {
			return filteredTodos[i].Deadline > filteredTodos[j].Deadline
		})
	} else {
		sort.Slice(filteredTodos, func(i, j int) bool {
			return filteredTodos[i].Deadline < filteredTodos[j].Deadline
		})
	}

	// 返回结果
	todosWithIndex := []map[string]interface{}{}

	for index, todo := range filteredTodos {
		todoWithIndex := map[string]interface{}{
			"index":    index,
			"content":  todo.Content,
			"done":     todo.Done,
			"deadline": todo.Deadline,
		}
		todosWithIndex = append(todosWithIndex, todoWithIndex)
	}

	c.JSON(200, todosWithIndex)
}

func GetTodo(c *gin.Context) {
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	if index >= len(existingTodos) {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	todoWithIndex := map[string]interface{}{
		"index":    index,
		"content":  existingTodos[index].Content,
		"done":     existingTodos[index].Done,
		"deadline": existingTodos[index].Deadline,
	}

	c.JSON(200, todoWithIndex)
}

func ListUnfinishedTodos(c *gin.Context) {
	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	unfinishedTodos := []map[string]interface{}{}

	for index, todo := range existingTodos {
		if !todo.Done {
			todoWithIndex := map[string]interface{}{
				"index":    index,
				"content":  todo.Content,
				"done":     todo.Done,
				"deadline": todo.Deadline,
			}
			unfinishedTodos = append(unfinishedTodos, todoWithIndex)
		}
	}

	sort.Slice(unfinishedTodos, func(i, j int) bool {
		deadlineI, _ := time.Parse("2006-01-02", unfinishedTodos[i]["deadline"].(string))
		deadlineJ, _ := time.Parse("2006-01-02", unfinishedTodos[j]["deadline"].(string))
		return deadlineI.Before(deadlineJ)
	})

	if len(unfinishedTodos) == 0 {
		c.JSON(404, ErrNoUnfinishedTODO)
		return
	}

	c.JSON(200, unfinishedTodos)
}
