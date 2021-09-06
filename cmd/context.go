package main

import "os"

func isDevelopment() bool {
	return os.Getenv("FILE_SERVER_ENVIRONMENT") == "DEV"
}
