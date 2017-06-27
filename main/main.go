package main

func main() {
	a := App{}
	a.Initialize("john", "new_sub_test_db")

	a.Run(":8080")
}
