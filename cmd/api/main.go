package main

import (
	"log"
	"net/http"

	"web-page-analyzer/internal/presentation/handlers"
	"web-page-analyzer/internal/presentation/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	analysisHandler := &handlers.AnalysisHandler{}

	routes.SetupRoutes(router, analysisHandler)

	log.Println("Starting Web Page Analyzer API server on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
