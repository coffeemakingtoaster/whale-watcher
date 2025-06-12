package display

import "fmt"

func printfln(message string, args ...any) {
	fmt.Printf(message+"\n", args...)
}
