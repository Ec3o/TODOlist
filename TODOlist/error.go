package main

// TODOError 此文件用于预定义错误类型
type TODOError struct {
	Message string `json:"error"`
}

var (
	ErrInvalidTODOFormat = TODOError{"抱歉，您提供的TODO数据格式不正确"}
	ErrInvalidDeadline   = TODOError{"无效的截止时间，格式应为 '2006-01-02' 并且不能是过去的时间"}
	ErrReadTODOData      = TODOError{"无法读取TODO数据"}
	ErrSaveTODOData      = TODOError{"无法保存TODO数据"}
	ErrTODOIndexNotExist = TODOError{"抱歉，您访问的ToDo目前不存在，请先创建"}
	ErrTODONotFound      = TODOError{"抱歉，您要删除的ToDo目前不存在，请先创建"}
	ErrNoTodayTODO       = TODOError{"今天没有截止的TODO"}
	ErrNoWeekTODO        = TODOError{"本周没有截止的TODO"}
	ErrNoUnfinishedTODO  = TODOError{"没有未完成的TODO"}
)
