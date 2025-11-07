package main

import (
	"fmt"
	"license/internal/hwid"
)

func main() {
	fmt.Println(hwid.GetFingerprint())
}
