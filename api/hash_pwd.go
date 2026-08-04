package main
import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)
func main() {
	bytes, _ := bcrypt.GenerateFromPassword([]byte("123456"), 14)
	fmt.Println(string(bytes))
}
