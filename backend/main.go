package main

import (
	auth "go-forum-project/auth-service/cmd/app"
	forum "go-forum-project/forum-service/cmd/app"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		auth.RunAuthApp()
	}()

	go func() {
		defer wg.Done()
		forum.RunForumApp()
	}()

	wg.Wait()
}
