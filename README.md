# Web后端开发第一课

# 杭电助手简单后端Go项目开发--ToDo List

## 项目初始化

打开**GoLand**，创建项目文件夹，打开终端，输入以下代码并回车：

```go
go mod init <文件夹名称>
```

新建一个go文件，命名为 `main.go` ,之后用到的主程序代码都写在里面啦 ♡>𖥦<)!!

在`main.go`文件里面，我们开始正式构建go项目

## Link Start！

#### 一些库导入以及前期准备

一个基本的go文件框架大概长得像以下这个样子：

```go
package main
import(
    //一些必要的库导入
)
//一些必要的初始化
func main(){
    //实现一些必要的功能
}
```

本项目使用**gin**框架进行简单的后端开发

要使用**gin**框架，请在终端中输入如下命令：

```go
go get -u github.com/gin-gonic/gin
```

并在用到**gin**框架时，在**import**部分将其导入

```go
import(
    "github.com/gin-gonic/gin"
)
```

必要的准备工作完成。

#### 路由及其功能实现

接下来是main函数部分

```go
r := gin.Default() 
```

 这导入了Gin框架并使用默认的配置创建了一个Gin的路由引擎

接下来给出一个示例:

```go
r.GET("/todo",func(c *gin.Context){
    //要在路由中实现的功能
})
```

上面的实例中,“**GET**”部分是访问路由时使用的方法，“**/todo**”是希望访问的路径,"**func(c *gin.Context)**"用来传递一些必要的参数。有了一定的了解之后，我们可以利用自己对编程语言的理解完成各个功能的实现，下面是各个功能实现过程的解析：

在所有功能开始做之前，我进行了一些必要的初始化，例如定义结构体内容、文件变量等。下面给出具体代码

**定义主包**

```go
package main
```

**导入必要的库**

```go
import (
	"encoding/json"#用来读写json文件
	"github.com/gin-gonic/gin"#gin框架
	"io/ioutil"#用于读取和写入文件、目录
	"sort"#用于排序
	"strconv"#用于数据类型转换
	"time"#用于处理时间相关操作
)
```

**定义结构体**

```go
type TODO struct {
	Content  string `json:"content"`
	Done     bool   `json:"done"`
	Deadline string `json:"deadline"`
}
```

这段定义定义了三个变量，分别是**字符型**` content`、**布尔型**`done`和**字符型**`deadline`用来进行数据处理

**定义文件**

```go
var todosFile = "todos.json"
```

**处理文件读写**

```go
package main

import (
	"encoding/json"
	"io/ioutil"
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

```

为了代码的易维护性,我们将其封装在`fileio.go`文件中,并将其放置于`main.go`同一目录下,它可以实现文件的读写功能o(*￣▽￣*)ブ

#### status ok,link start!qwq

## 添加功能

#### Todo上传

```go
r.POST("/todo", func(c *gin.Context) {
		var todo TODO  #将TODO实例化
		if err := c.BindJSON(&todo); err != nil {
           #这个函数尝试将上传的数据按照json格式解析，如果失败则返回错误信息，说明用户上传的信息不合法
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
            #time.Parse函数用来尝试将上传的deadline解析
			if err != nil || parsedDeadline.Before(time.Now()) {
				// 如果解析失败或时间早于当前时间，返回错误
                #deadline.Before函数用来比较两个时间值并返回一个布尔型结果
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

		c.JSON(200, gin.H{"status": "数据提交成功"})
	})
```

`nil`代表了空值,类似于`null`,常用于判断是否产生了错误,经常用于错误处理

我们进行了几次判断:

1.判断用户上传是否有效,否则报错

2.判断是否设置了`deadline`,否则设置默认`deadline`为七天后

3.解析上传的`deadline`(如果有的话),判断是否为有效时间,否则报错

4.尝试读取数据,失败报错

5.尝试上传数据,失败报错

6.数据上传成功,返回`200`状态码和提示语

#### Todo删除

```go
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

		c.JSON(200, gin.H{"status": "删除成功", "被删除的数据是": deletedTodo})
	})
```

我们进行了几次判断:

1. 判断用户上传参数是否有效,否则报错.
2. 尝试读取数据,失败报错.
3. 尝试上传删除后的数据,失败报错.
4. 数据删除成功,返回200状态码和被删除的数据.

一些补充:

```go
index, err := strconv.Atoi(c.Param("index"))
```

这个函数的意思是,尝试将获得的`index`参数(ASCII)转换成整型.`Atoi` 代表 "ASCII to Integer".

Go的`append`函数用法似乎与python的不同:

在python中,列表`list`添加新元素使用`append`函数:

```python
list.append(element)
```

在go中,通常进行重新赋值

```go
list=append(list,element)
```

删除操作即相当于把原列表(切片)取list[index:]部分和list[:index]部分重新拼接即可.

#### Todo修改

```go
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

		c.JSON(200, gin.H{"status": "修改成功"})
	})
```

我们进行了几次判断:

1. 判断用户上传参数是否有效,否则报错.
2. 尝试读取数据,失败报错.
3. 尝试上传修改后的数据,失败报错.
4. 数据删除成功,返回`200`状态码和提示语

#### 汇总Todo

```go
r.GET("/todo", func(c *gin.Context) {
		// 读取已有的 TODO 数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取 TODO 数据"})
			return
		}

		// 创建一个新的切片来存储包含序号的 TODO 数据
		todosWithIndex := []map[string]interface{}{}

		// 遍历现有的 TODO 数据，为每个 TODO 添加序号
		for index, todo := range existingTodos {
			todoWithIndex := map[string]interface{}{
				"index":    index,
				"content":  todo.Content,
				"done":     todo.Done,
				"deadline": todo.Deadline,
			}
			todosWithIndex = append(todosWithIndex, todoWithIndex)
		}

		c.JSON(200, todosWithIndex)
	})
```

其中

```go
todosWithIndex := []map[string]interface{}{}
```

这段代码的意思是创建一个元素为`string`-`任意类型`的键值对元素空切片用来存储数据.这段代码遍历了所有`todo`并为其增加了索引,最后输出`todo`.实现了我们所需要的功能

#### 查询Todo

```go
r.GET("/todo/:index", func(c *gin.Context) {
		index, err := strconv.Atoi(c.Param("index"))
		if err != nil || index < 0 {
			c.JSON(404, gin.H{"error": "抱歉，您访问的 ToDo 目前不存在，请先创建"})
			return
		}

		// 读取已有的 TODO 数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取 TODO 数据"})
			return
		}

		// 检查索引是否超出范围
		if index >= len(existingTodos) {
			c.JSON(404, gin.H{"error": "抱歉，您访问的 ToDo 目前不存在，请先创建"})
			return
		}

		// 获取单个 TODO，并添加序号
		todoWithIndex := map[string]interface{}{
			"index":    index,
			"content":  existingTodos[index].Content,
			"done":     existingTodos[index].Done,
			"deadline": existingTodos[index].Deadline,
		}

		c.JSON(200, todoWithIndex)
	})
```

类似查询所有todo的代码,不过我们的输出是单个的而已

#### 今日Todo

```go
r.GET("/list_today", func(c *gin.Context) {
		// 获取今天的日期
		today := time.Now().Format("2006-01-02")#获得今天的日期信息并按格式处理

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
			if todo.Deadline == today #简单的{
				todayTodos = append(todayTodos, todo)
			}
		}

		if len(todayTodos) == 0 {
			c.JSON(404, gin.H{"message": "今天没有截止的TODO"})
			return
		}

		c.JSON(200, todayTodos)
	})
```

获取今日todo并存储输出,就是这么简单~

```go
for _, todo := range existingTodos {
			if todo.Deadline == today #简单的{
				todayTodos = append(todayTodos, todo)
			}
		}
```

`_`代表了匿名变量,就是我们不会用到的变量可以用它来代替,它代表了序列中的index(序号),但这里我们暂时用不上.

#### 本周Todo

```go
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
```

后面几个模块功能其实差不多,都是基本的文件读写+条件判断,这里判断是在本周一之后,下周一之前的`Todo`会被列入我们的清单

#### 待办Todo

```go
r.GET("/unfinished_todo", func(c *gin.Context) {
		// 读取已有的TODO数据
		existingTodos, err := loadTodosFromFile()
		if err != nil {
			c.JSON(500, gin.H{"error": "无法读取TODO数据"})
			return
		}

		// 创建一个新的切片来存储未完成的 TODO 数据，包括序号
		unfinishedTodos := []map[string]interface{}{}

		// 遍历现有的 TODO 数据，找到未完成的 TODO
		for index, todo := range existingTodos {
			if !todo.Done {
				// 创建包含序号的 map
				todoWithIndex := map[string]interface{}{
					"index":    index,
					"content":  todo.Content,
					"done":     todo.Done,
					"deadline": todo.Deadline,
				}
				unfinishedTodos = append(unfinishedTodos, todoWithIndex)
			}
		}

		// 对未完成的 TODO 按照截止时间排序（由近到远）
		sort.Slice(unfinishedTodos, func(i, j int) bool {
			deadlineI, _ := time.Parse("2006-01-02", unfinishedTodos[i]["deadline"].(string))
			deadlineJ, _ := time.Parse("2006-01-02", unfinishedTodos[j]["deadline"].(string))
			return deadlineI.Before(deadlineJ)
		})

		if len(unfinishedTodos) == 0 {
			c.JSON(404, gin.H{"message": "没有未完成的 TODO"})
			return
		}

		c.JSON(200, unfinishedTodos)
	})
```

与之前的代码类似,但是增加了排序功能,更加实用.

```go
sort.Slice(unfinishedTodos, func(i, j int) bool {
			deadlineI, _ := time.Parse("2006-01-02", unfinishedTodos[i]["deadline"].(string))
			deadlineJ, _ := time.Parse("2006-01-02", unfinishedTodos[j]["deadline"].(string))
			return deadlineI.Before(deadlineJ)
		})
```

`sort.Slice` 是Go语言标准库中的一个函数，用于对切片进行排序。

`func(i, j int) bool { ... }` 是比较函数，它定义了如何比较切片中的两个元素（任务）以确定它们的顺序。该函数接受两个整数参数 `i` 和 `j`，它们表示切片中两个要比较的元素的索引。

Go语言会将元素按照func(i,j)为true的模式进行排序.代码比较两个任务的截止日期，使用 `deadlineI.Before(deadlineJ)` 来确定哪个任务的截止日期在前。如果 `deadlineI` 在 `deadlineJ` 之前，比较函数返回 `true`，表示任务 `i` 应该排在任务 `j` 之前，从而实现了按截止日期升序排序。

#### 一步之遥

```go
r.Run(":8100")
```

电脑8080端口被占用了,所以临时换了端口.

## Congratulations!

至此,程序可以正常运行,实现了以下功能:

- **数据持久化**
- **错误输入处理**
- **加入了截止日期**
- **在增删查改之外,添加了每日/每周/代办事项的提示**
- **数据排序**
- **功能函数模块化,易于维护.**

#### **可以尝试的改进:**

1. 尝试使用gorm数据库代替文件存储:考虑将数据存储到数据库中，以便更可靠地管理和检索TODO数据。
2. 用户身份验证和授权:添加多个用户，可以实现用户身份验证和授权系统，以确保只有授权用户可以执行操作。
3. 用户界面（Web界面或移动应用程序）:可以构建一个用户友好的前端界面
4. 任务分类和标签:允许用户为TODO任务添加标签或将它们分为不同的类别，以便更好地组织和过滤任务。
5. 提醒和通知:实现任务截止日期的提醒功能，以及通过电子邮件、短信或推送通知用户。
6. 任务优先级:允许用户为任务设置优先级，并按照优先级对任务进行排序。
7. 搜索和过滤:添加搜索功能，使用户可以根据关键字、日期范围或其他条件来查找和过滤任务。
8. 安全性增强：确保应用程序具有适当的安全性措施，如防止跨站脚本攻击（XSS）和SQL注入。

> Created By Ec3o.All right reserved.
>
> 项目地址详见[TODOlist/TODOlist at main · Ec3o/TODOlist · GitHub](https://github.com/Ec3o/TODOlist/tree/main/TODOlist)
>
> 先写这么多吧,相信不久后就会有第二版说明文档的
