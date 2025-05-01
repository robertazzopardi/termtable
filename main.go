package main

func main() {
	app := NewApp()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
