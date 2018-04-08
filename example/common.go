// Package example provides a small application that demonstrates the GOM usage
package example

//go:generate go-bindata -pkg $GOPACKAGE -o database.go database/migration/ database/script/
