package main

import (
	"fmt"
	"os"

	"sis-text-analyzer/task"
)

func run() error {
	service := task.NewService()
	pool, err := task.NewPool(service, 2, 3)
	if err != nil {
		return err
	}
	defer pool.Shutdown()

	texts := []string{"Hello Go world", "Сәлем Go 2026", "Hello 🌍"}
	ids := make([]int, 0, len(texts))
	for _, text := range texts {
		created, err := service.Create(text)
		if err != nil {
			return err
		}
		fmt.Printf("Created: id=%d status=%s input=%q\n", created.ID, created.Status, created.Input)
		if err := pool.Submit(created.ID); err != nil {
			return err
		}
		ids = append(ids, created.ID)
		fmt.Printf("Submitted: id=%d\n", created.ID)
	}

	pool.Shutdown()
	for _, id := range ids {
		completed, err := service.Get(id)
		if err != nil {
			return err
		}
		fmt.Printf("Completed: id=%d status=%s result=%+v\n", completed.ID, completed.Status, completed.Result)
	}
	_, err = service.Create("   ")
	fmt.Println("Expected invalid input:", err)
	_, err = service.Get(999)
	fmt.Println("Expected missing task:", err)
	fmt.Println("Expected submission after shutdown:", pool.Submit(ids[0]))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
