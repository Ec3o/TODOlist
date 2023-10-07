# 版本大更新

**尝试了jwt鉴权和用户系统.jwt鉴权比较复杂,做了一个半成品失败了,不知道哪里出了问题无法鉴权.半成品地址在[Ec3o/TODOLIST_JWT (github.com)](https://github.com/Ec3o/TODOLIST_JWT)**

所以重点做了下用户系统,用比较简单的思路来模拟了一个登录(x)

### 用户系统

#### **定义用户结构体**

```go
type USER struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
```

#### **用户文件地址**

```go
var usersFile = "users.json"
```

#### **注册登录路由**

```go
r.POST("/register", useregister) //注册
r.POST("/login", userlogin)      //登录
```

#### **用户数据读取存储**

```go
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
```

#### **简单的用户注册函数**

```go
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
```

#### 简单的用户登录函数

```go
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
```

#### **实现用户登录态保持的原理**

很简单,定义了一个全局变量`currentUser`用于存储当前用户信息.在jwt中,用户信息更为安全,奈何技术力不足,暂时做不出来.原理类似,jwt是存在header中的用户信息,而currentuser是存在服务器端的用户信息.

### **错误类型定义**

**error.go**

```go
package main

//此文件用于预定义错误类型

// TODOError TODO错误类型
type TODOError struct {
    Message string `json:"error"`
}

// USERError 用户错误类型
type USERError struct {
    Message string `json:"error"`
}

// Successes 成功操作类型
type Successes struct {
    Message string `json:"status"`
}

// TODOError
var (
    ErrInvalidTODOFormat = TODOError{"抱歉，您提供的TODO数据格式不正确"}
    ErrInvalidDeadline   = TODOError{"无效的截止时间"}
    ErrReadTODOData      = TODOError{"无法读取TODO数据"}
    ErrSaveTODOData      = TODOError{"无法保存TODO数据"}
    ErrTODOIndexNotExist = TODOError{"抱歉，您访问的ToDo目前不存在，请先创建"}
    ErrTODONotFound      = TODOError{"抱歉，您要删除的ToDo目前不存在，请先创建"}
)

// USERError
var (
    ErrInvalidUSERFormat = USERError{"抱歉，您提供的用户数据格式不正确"}
    ErrInvalidPassword   = USERError{"密码不能为空或密码长度过短"}
    ErrReadUserData      = USERError{"无法读取用户数据"}
    ErrSaveUserData      = USERError{"无法保存保存数据"}
    ErrUserlogin         = USERError{"用户未注册或密码错误"}
    ErrUser              = USERError{"用户未登录或无效用户"}
    ErrRegister          = USERError{"用户名已经被注册"}
)

// Successes
var (
    TodoSubmitSuccess   = Successes{"数据提交成功"}
    UserRegisterSuccess = Successes{"用户注册成功"}
    TodoDeleteSuccess   = Successes{"数据删除成功"}
)
```

现在我们拥有更多更规范的提示类型了**๐•ᴗ•๐**

### **数据组织形式切换**

#### 数据上传

我们使用了用户名制,那么对应的数据组织形式就要发生改变.原来的todo序号从1-len(todo)连续.

使用了用户后,我将为每个用户独立分配从1开始连续的todo序号,便于各个用户的todo分开独立管理.

```go
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
```

我们首先获取用户名信息,再对数据进行处理后绑定json数据,并计算本次todo序号的值,最终提交成功.

#### 数据删除

原本使用的删除列表不再使用,转而判断todo内容来判断todo是否被删除.

```go
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
```

首先获取用户名信息,然后寻找用户名对应的序号的todo,最终对其进行置空操作.

#### 数据更新

```go
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
```

类似删除操作,只不过最后一步并非置空,而是修改.

#### 数据列出

```go
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
```

该函数部分是修改最大的.鉴于之前的'今日todo'和'本周todo'功能略显鸡肋,此处添加了数据筛选功能(参数可选,不提交参数即不筛选),包含`deadline`(**int64**类型**unix**时间戳),`reverse`(**bool**类型 截止时间正序/倒序)以及`done`(**bool**类型 完成情况)三个参数.**deadline**默认值为0,如不为0则筛选出deadline之前的所有**todo**;**reverse**为**true**代表升序,**false**代表降序;**done**代表完成情况**,true**为完成,**false**为未完成.只提供当前用户的todo.

#### **数据查找**

```go
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
```

和删除,更新没什么大区别,略过.

至此,项目基本成型,还有一些点可以在后期完善.

- 使用数据库来组织数据,而不是文件.
- 使用json web token来管理用户登录与注销的需求,而不是简单使用全局变量来解决.
- 可以尝试接入一些简单的HTML前端界面,让用户交互更加友好.
- 注意一些安全问题,例如抓包获取用户名密码等网络渗透攻击.