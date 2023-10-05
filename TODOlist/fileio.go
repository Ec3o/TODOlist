package main

//此文件用于预定义文件读取保存函数
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

// 辅助函数：检查索引是否在切片中
func containsIndex(index int, indexList []int) bool {
	for _, i := range indexList {
		if i == index {
			return true
		}
	}
	return false
}
