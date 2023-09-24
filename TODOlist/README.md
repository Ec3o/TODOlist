# 一个简单的To-Do list Web应用程序

## 1.提交todo功能

**对应路由**:[localhost:8100/add_todo]()

**期望实现功能简述**:用户在访问该路由时,返回add_todo.html界面;使用页面表单提交数据时,向数据库todolist.db中提交content、deadline(以‘2023-09-23‘格式提交)、done三个参数。后端会检查用户的输入,对deadline在明天之前的提交进行报错以及过滤；当用户的输入有效时，提交数据至数据库，并向用户返回提交成功的HTML界面。

## 2.删除todo功能

**对应路由**:[localhost:8100/delect_todo]()

**期望实现功能简述**:用户在访问该路由时,返回delect_todo.html界面；在该页面列出所有todo数据,按deadline由近到远进行排序;使用复选框勾选要删除的todo,html将要删除的数据提交至数据库

## 3.列出todo功能

**对应路由**:[localhost:8100/list_todo]()

**期望实现功能简述**:用户在访问该路由时,返回list_todo.html界面；在该页面列出所有todo数据以及序号,提供多个选择，可按照deadline进行排序列出或按照已完成/未完成进行分类

## 4.修改todo功能

**对应路由**:[localhost:8100/update_todo]()

**期望实现功能简述**:用户在访问该路由时,返回update_todo.html界面；在该页面按照deadline由近到远列出所有todo数据以及序号

## 5.查询todo功能

**对应路由**:[localhost:8100/search_todo]()

**期望实现功能简述**:用户在访问该路由时,返回search_todo.html界面；在该页面按照表单提交deadline参数,即可查询该deadline之前的所有todo事项。