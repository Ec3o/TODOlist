package main

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 函数功能：从文件中读取数据
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

// 函数功能：将数据保存至文件中
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

// 函数功能:读取用户数据
func loadUsersFromFile() ([]USER, error) {
	data, err := ioutil.ReadFile(usersFile)
	if err != nil {
		return nil, err
	}

	var users []USER
	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// 函数功能:保存用户数据
func saveUsersToFile(users []USER) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(usersFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// 函数功能:创建token
func createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})
	tokenString, err := token.SignedString(jwtKey) // 替换成你自己的密钥
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "未提供令牌"})
			c.Abort()
			return
		}

		// 去掉 "Bearer " 部分
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil // 替换成你的密钥
		})

		if err != nil {
			c.JSON(401, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(401, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(401, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		currentUser = username
		c.Next()
	}
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
	indexToUpdate -= 1
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
			"index":    todo.Index + 1,
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
	indexToGet -= 1
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
		token, err := createToken(foundUser.Username)
		if err != nil {
			c.JSON(500, ErrUserlogin)
			return
		}

		// 返回用户名和令牌
		c.JSON(200, gin.H{"status": "用户登录成功", "username": foundUser.Username, "token": token})
	} else {
		c.JSON(404, ErrUserlogin)
	}
}
