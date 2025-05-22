package main

import (
	app "gonesem/internal/application"
	"log"
)

func main() {
	options := app.NewOptions()
	options.Parse()

	app, err := app.NewApplication(options)

	if err != nil {
		log.Fatalln(err)
	}

	app.Run()
}
