package main

import (
	"github.com/gin-gonic/gin"
	"sort"
	"strconv"
	"time"
)

type TODO struct {
	Username string        `json:"username"`
	Index    int           `json:"index"`
	Content  string        `json:"content"`
	Done     bool          `json:"done"`
	Deadline UnixTimestamp `json:"deadline"` // 使用 int64 类型
}
type UnixTimestamp int64
type USER struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type TODOWithOriginalIndex struct {
	Todo  TODO
	Index int // 添加一个新的字段来保存原始索引
}

var todosFile = "todos.json" //todo文件储存地址
var usersFile = "users.json" //用户文件储存地址 // 使用 map 来跟踪每个用户删除的待办事项索引
var currentUser string

func main() {
	r := gin.Default()

	authGroup := r.Group("/")
	authGroup.POST("/todo", TodoCreation)          //增
	authGroup.DELETE("/todo/:index", TodoDeletion) //删(不改动序号)
	authGroup.PUT("/todo/:index", TodoUpdate)      //改
	authGroup.GET("/todo", ListTodos)              //查(使用条件筛选)
	authGroup.GET("/todo/:index", GetTodo)         //获取单个todo信息

	r.POST("/register", useregister) //注册
	r.POST("/login", userlogin)      //登录
	r.Run(":8100")                   //运行在8100端口
}

func TodoCreation(c *gin.Context) {
	currentUser := currentUser
	if currentUser == "" {
		c.JSON(401, ErrUser)
		return
	}

	var todo TODO
	if err := c.BindJSON(&todo); err != nil {
		c.JSON(400, ErrInvalidTODOFormat)
		return
	}

	if todo.Deadline == 0 {
		defaultDeadline := time.Now().Add(time.Hour * 24 * 7)
		todo.Deadline = UnixTimestamp(defaultDeadline.Unix())
	} else if int64(todo.Deadline) < time.Now().Unix() {
		c.JSON(400, ErrInvalidDeadline)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	// 计算用户的索引
	userIndex := 1
	for _, t := range existingTodos {
		if t.Username == currentUser {
			userIndex++
		}
	}

	// 为新的 Todo 分配 index 序号
	todo.Index = userIndex
	todo.Username = currentUser

	existingTodos = append(existingTodos, todo)

	err = saveTodosToFile(existingTodos)
	if err != nil {
		c.JSON(500, ErrSaveTODOData)
		return
	}
	c.JSON(200, TodoSubmitSuccess)
}

func TodoDeletion(c *gin.Context) {
	currentUser := currentUser
	if currentUser == "" {
		c.JSON(401, gin.H{"status": "用户未登录或无效的用户"})
		return
	}
	indexToDelete, err := strconv.Atoi(c.Param("index"))
	if err != nil || indexToDelete < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrTODONotFound)
		return
	}

	var deletedContent string // 用于保存被删除的待办事项内容

	// 遍历待办事项列表，找到与当前用户匹配的待办事项并匹配索引
	for index, todo := range existingTodos {
		if todo.Username == currentUser && index == indexToDelete {
			// 检查是否已经在已删除的待办事项列表中
			if todo.Content == "此Todo已被删除" {
				c.JSON(400, TodoDeleteSuccess)
				return
			}
			// 保存被删除的待办事项内容
			deletedContent = todo.Content

			// 标记待办事项为已删除
			todo.Content = "此Todo已被删除"
			todo.Done = true

			// 更新待办事项回到列表
			existingTodos[index] = todo
			err = saveTodosToFile(existingTodos)
			if err != nil {
				c.JSON(500, ErrSaveTODOData)
				return
			}

			// 在 JSON 响应中包括被删除的内容
			c.JSON(200, gin.H{"status": "删除成功", "被删除的内容是": deletedContent})
			return
		}
	}

	// 如果没有匹配的待办事项，返回错误
	c.JSON(404, ErrTODOIndexNotExist)
}

func TodoUpdate(c *gin.Context) {
	currentUser := currentUser
	if currentUser == "" {
		c.JSON(401, ErrUser)
		return
	}
	indexToUpdate, err := strconv.Atoi(c.Param("index"))
	if err != nil || indexToUpdate < 0 {
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

	// 遍历待办事项列表，找到与当前用户匹配的待办事项并匹配索引
	for index, existingTodo := range existingTodos {
		if existingTodo.Username == currentUser && index == indexToUpdate {
			// 更新待办事项内容
			existingTodo.Content = todo.Content
			existingTodo.Done = todo.Done
			existingTodo.Deadline = todo.Deadline

			// 更新待办事项回到列表
			existingTodos[index] = existingTodo

			err = saveTodosToFile(existingTodos)
			if err != nil {
				c.JSON(500, ErrSaveTODOData)
				return
			}

			c.JSON(200, gin.H{"status": "修改成功"})
			return
		}
	}

	// 如果没有匹配的待办事项，返回错误
	c.JSON(404, ErrTODOIndexNotExist)
}

func ListTodos(c *gin.Context) {
	// 从请求上下文中获取当前用户的用户名
	currentUser := currentUser
	if currentUser == "" {
		c.JSON(401, ErrUser)
		return
	}

	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}

	// 获取查询参数
	deadline := c.DefaultQuery("deadline", "0") // 默认值设置为 "0"，表示不进行筛选
	reverse := c.DefaultQuery("reverse", "false")
	finished := c.DefaultQuery("finished", "")

	// 转换 reverse 字符串为布尔值
	reverseSort := (reverse == "true")

	// 根据 finished 参数过滤待办事项
	filteredTodos := []TODOWithOriginalIndex{} // 使用新的结构体保存待办事项和原始索引
	for index, todo := range existingTodos {
		// 检查索引是否在 deletedTodoIndexes 中，如果在就跳过
		if todo.Content == "此Todo已被删除" {
			continue
		}

		// 只返回属于当前用户的待办事项
		if todo.Username != currentUser {
			continue
		}

		// 直接将查询参数转换为整数
		queryDeadline, err := strconv.ParseInt(deadline, 10, 64)
		if err != nil {
			c.JSON(400, ErrInvalidDeadline)
			return
		}

		// 检查截止日期是否符合筛选条件
		if (finished == "true" && todo.Done) || (finished == "false" && !todo.Done) || finished == "" {
			if queryDeadline == 0 || (queryDeadline != 0 && int64(todo.Deadline) <= queryDeadline) {
				// 保存待办事项和原始索引
				filteredTodos = append(filteredTodos, TODOWithOriginalIndex{todo, index})
			}
		}
	}

	// 根据 reverseSort 参数排序
	if reverseSort {
		sort.Slice(filteredTodos, func(i, j int) bool {
			return int64(filteredTodos[i].Todo.Deadline) < int64(filteredTodos[j].Todo.Deadline)
		})
	} else {
		sort.Slice(filteredTodos, func(i, j int) bool {
			return int64(filteredTodos[i].Todo.Deadline) > int64(filteredTodos[j].Todo.Deadline)
		})
	}

	// 返回结果
	todosWithIndex := []map[string]interface{}{}

	for _, todo := range filteredTodos {
		todoWithIndex := map[string]interface{}{
			"index":    todo.Index,
			"content":  todo.Todo.Content,
			"done":     todo.Todo.Done,
			"deadline": todo.Todo.Deadline, // 已经是 int64 格式
		}
		todosWithIndex = append(todosWithIndex, todoWithIndex)
	}

	c.JSON(200, todosWithIndex)
}

func GetTodo(c *gin.Context) {
	currentUser := currentUser
	if currentUser == "" {
		c.JSON(401, gin.H{"status": "用户未登录或无效的用户"})
		return
	}
	indexToGet, err := strconv.Atoi(c.Param("index"))
	if err != nil || indexToGet < 0 {
		c.JSON(404, ErrTODOIndexNotExist)
		return
	}
	filteredTodo := []TODO{}
	existingTodos, err := loadTodosFromFile()
	if err != nil {
		c.JSON(500, ErrReadTODOData)
		return
	}
	for index, todo := range existingTodos {
		// 检查索引是否在 deletedTodoIndexes 中，如果在就跳过
		if todo.Content == "此Todo已被删除" {
			continue
		}

		// 只返回属于当前用户的待办事项
		if todo.Username != currentUser {
			continue
		}

		if todo.Username == currentUser && index == indexToGet {
			filteredTodo = append(filteredTodo, todo)
			c.JSON(200, filteredTodo)
			return
		}

	}
	c.JSON(404, ErrTODONotFound)
}

func useregister(c *gin.Context) {
	var user USER
	if err := c.BindJSON(&user); err != nil {
		c.JSON(400, ErrInvalidUSERFormat)
		return
	}

	if len(user.Password) <= 6 { //密码长度过短提示重新设置
		c.JSON(400, ErrInvalidPassword)
		return
	}

	existingUsers, err := loadUsersFromFile()
	if err != nil {
		c.JSON(500, ErrReadUserData)
		return
	}

	// 检查是否已经存在相同的用户名
	for _, existingUser := range existingUsers {
		if existingUser.Username == user.Username {
			c.JSON(400, ErrRegister)
			return
		}
	}

	existingUsers = append(existingUsers, user)
	err = saveUsersToFile(existingUsers)
	if err != nil {
		c.JSON(500, ErrSaveUserData)
		return
	}

	c.JSON(200, UserRegisterSuccess)
}

func userlogin(c *gin.Context) {
	var user USER
	if err := c.BindJSON(&user); err != nil {
		c.JSON(400, ErrInvalidUSERFormat)
		return
	}

	existingUsers, err := loadUsersFromFile()
	if err != nil {
		c.JSON(500, ErrReadUserData)
		return
	}

	var foundUser USER
	for _, existingUser := range existingUsers {
		if existingUser.Username == user.Username && existingUser.Password == user.Password {
			foundUser = existingUser
			break
		}
	}

	if foundUser.Username != "" {
		currentUser = foundUser.Username // 设置全局变量为当前用户的用户名
		c.JSON(200, gin.H{"status": "用户登录成功", "username": foundUser.Username})
	} else {
		c.JSON(404, ErrUserlogin)
	}
}
