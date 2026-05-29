package main

import "fmt"

func CreateLemonProject() {
	if !ProjectFolderIsEmpty(".") {
		PrintError("Folder is not empty to initialize new lemon project!")
	}

	//create lemon project structure
	CreateFolder("assets")
	CreateFile("LEMON_README.md", LEMON_README)
	CreateFolder("logic")
	CreateFolder("markup")
	CreateFile("index.lm", LEMON_HELLO_PAGE_TEMPLATE)
	CreateFolder("styles")

	fmt.Println("Initialized new lemon project with:")
	fmt.Println("")
	fmt.Println("  assets/")
	fmt.Println("  LEMON_README.md")
	fmt.Println("  logic/")
	fmt.Println("  markup/")
	fmt.Println("  index.lm")
	fmt.Println("  styles/")
	fmt.Println("")
	fmt.Println("4 folders, 2 files")
	fmt.Println()

}

const LEMON_PROJECT_STRUCTURE = `Project Structure:
  assets/    
  LEMON_README.md
  logic/
  markup/
  index.lm
  styles/

4 folders, 2 files
`
const LEMON_README = "This is a project created by " + LEMON_VERSION + "\n\n\n" +
	LEMON_HELP + "\n\n\n" +
	LEMON_PROJECT_STRUCTURE + "\n"

const LEMON_HELLO_PAGE_TEMPLATE = `<!DOCTYPE html>
<html>
<body>
    <Card Title="Hello World!"></Card>
</body>
</html>

<template name="Card">
    <div class="card">
        <h1>{{ Title }}</h1>
    </div>
</template>`
