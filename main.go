package main

import "github.com/kiRiLL3311/Go-multi-chat/initializers"

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
	initializers.SyncDatabase()
}

func main() {

}
