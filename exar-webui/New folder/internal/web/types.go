package web // sets the Go package name to web

type pageData struct { // declares the pageData struct type
	Errors  []string // executes this step of the current logic
	Success string // executes this step of the current logic
	Values  map[string]string // executes this step of the current logic
} // closes the current block scope
