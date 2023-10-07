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
