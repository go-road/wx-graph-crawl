package utils

import (
	"fmt"
)

func WordRed(format string, v ...interface{}) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 0, 31, 49, fmt.Sprintf(format, v...), 0x1B)
}

func WordGreen(format string, v ...interface{}) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 0, 32, 49, fmt.Sprintf(format, v...), 0x1B)
}

func WordYellow(format string, v ...interface{}) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 0, 33, 49, fmt.Sprintf(format, v...), 0x1B)
}

func WordBlue(format string, v ...interface{}) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 0, 34, 49, fmt.Sprintf(format, v...), 0x1B)
}

func WordCyan(format string, v ...interface{}) string {
	return fmt.Sprintf("%c[%d;%d;%dm%s%c[0m\n", 0x1B, 0, 36, 49, fmt.Sprintf(format, v...), 0x1B)
}

func ConsoleRed(content string) {
	fmt.Println(WordRed(content))
}

func ConsoleGreen(content string) {
	fmt.Println(WordGreen(content))
}

func ConsoleYellow(content string) {
	fmt.Println(WordYellow(content))
}

func ConsoleBlue(content string) {
	fmt.Println(WordBlue(content))
}

func ConsoleCyan(content string) {
	fmt.Println(WordCyan(content))
}
