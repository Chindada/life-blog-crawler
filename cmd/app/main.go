package main

import "lbc/internal/app"

func main() {
	app.Run()
	app.Clear()
	app.SetMobile()
	app.Run()

	select {}
}
