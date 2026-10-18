package main

import (
	"fmt"
	"os"
	"sis-text-analyzer/task"
)

func run() error {
	service := task.NewService()
	for _, text := range []string{"Hello Go world", "Сәлем Go 2026"} {
		created, err := service.Create(text)
		if err != nil {
			return err
		}
		fmt.Printf("Created: id=%d status=%s input=%q\n", created.ID, created.Status, created.Input)
		completed, err := service.Process(created.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Completed: id=%d status=%s result=%+v\n", completed.ID, completed.Status, completed.Result)
	}
	_, err := service.Create("   ")
	fmt.Println("Expected invalid input:", err)
	_, err = service.Get(999)
	fmt.Println("Expected missing task:", err)
	_, err = service.Process(1)
	fmt.Println("Expected repeated processing:", err)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
