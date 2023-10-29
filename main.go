package main

import (
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/gateway"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	logger.Log.Infoln("Is New Version Deployed")

	gateway.Bootstrap()
}
