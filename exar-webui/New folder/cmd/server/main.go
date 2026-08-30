package main // sets the Go package name to main

import ( // starts the import list for required packages
	"fmt" // executes this step of the current logic
	"net/http" // executes this step of the current logic

	"expenses/internal/config" // executes this step of the current logic
	"expenses/internal/web" // executes this step of the current logic
) // ends the current grouped declaration

func main() { // defines main, which handles this unit of behavior
	host := config.EnvOrDefault(config.HostEnv, "127.0.0.1") // initializes host with the result of this expression
	port := config.EnvOrDefault(config.PortEnv, "8000") // initializes port with the result of this expression
	configPath := config.EnvOrDefault(config.ConfigFileEnv, "configs.toml") // initializes configPath with the result of this expression

	if _, err := config.RefreshAPIToken(configPath); err != nil { // checks this condition before continuing
		panic(err) // executes this step of the current logic
	} // closes the current block scope

	mux := http.NewServeMux() // initializes mux with the result of this expression
	web.RegisterRoutes(mux) // executes this step of the current logic

	addr := host + ":" + port // initializes addr with the result of this expression
	fmt.Printf("Serving %s at http://%s\n", config.PageTitle, addr)
	fmt.Printf("Wrote API token to %s\n", configPath) // formats text for output or errors
	if err := http.ListenAndServe(addr, mux); err != nil { // checks this condition before continuing
		panic(err) // executes this step of the current logic
	} // closes the current block scope
} // closes the current block scope
