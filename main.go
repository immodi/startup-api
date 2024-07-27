package main

import "immodi/startup/handlers"

func main() {
	r := handlers.MakeGinEngine()
	println("Currently Listening on http://localhost:8080....")

	r.Run()
}
