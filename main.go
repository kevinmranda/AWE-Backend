package main

import (
	initializers "github.com/kevinmranda/AWE-Backend/Initializers"
	migrations "github.com/kevinmranda/AWE-Backend/Migrations"
	routes "github.com/kevinmranda/AWE-Backend/Routes"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	migrations.SyncDatabase()
	routes.Routes()
}

func main() {
}
