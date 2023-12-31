package main

// 此文件用于预定义错误类型
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
var todosFile = "todos.json" //todo文件储存地址
var usersFile = "users.json" //用户文件储存地址 // 使用 map 来跟踪每个用户删除的待办事项索引
var currentUser string
var jwtKey = []byte("Ec3o")
