package main

import (
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/gateway"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	gateway.Bootstrap()
}
