package main

import (
	"fmt"
	"regexp"
)

func main2() {
	var validPath = regexp.MustCompile("/((edit|save|view|toc|delete|upload)/(([a-zA-Z0-9]+)(/([0-9]))*)*)*$")
	m := validPath.FindStringSubmatch("http://localhost:8099/view/TestPage")
	fmt.Println(m[0])
	fmt.Println(m[1])
	fmt.Println(m[2])
	fmt.Println(m[3])
	fmt.Println(m[4])
	fmt.Println(m[5])
	fmt.Println(m[6])
}
