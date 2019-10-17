package main

import (
	"fmt"

	"github.com/project-flogo/core/api"
	"github.com/project-flogo/core/engine"
)

func main() {

	app := StreamTest()

	e, err := api.NewEngine(app)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	engine.RunEngine(e)
}
