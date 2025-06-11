package main

import "fmt"

func main() {
var ptr *string
fmt.Println(*ptr) // 潜在的空指针解引用
}
