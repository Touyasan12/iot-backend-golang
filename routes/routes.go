package routes

import (
	"iot-backend-cursor/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// Documentation routes (serve before API routes)
	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("./swagger-ui.html")
	})
	r.GET("/openapi.yaml", handlers.ServeOpenAPIYAML)
	r.GET("/openapi.json", handlers.ServeOpenAPIJSON)
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/docs")
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Dashboard
		api.GET("/dashboard", handlers.GetDashboard)

		// Feeder routes
		feeder := api.Group("/feeder")
		{
			feeder.GET("/schedules", handlers.GetFeederSchedules)
			feeder.POST("/schedules", handlers.CreateFeederSchedule)
			feeder.PUT("/schedules/:id", handlers.UpdateFeederSchedule)
			feeder.DELETE("/schedules/:id", handlers.DeleteFeederSchedule)
			feeder.POST("/manual", handlers.ManualFeed)
			feeder.GET("/last-feed", handlers.GetLastFeedInfo)
		}

		// UV routes
		uv := api.Group("/uv")
		{
			uv.GET("/schedules", handlers.GetUVSchedules)
			uv.POST("/schedules", handlers.CreateUVSchedule)
			uv.PUT("/schedules/:id", handlers.UpdateUVSchedule)
			uv.DELETE("/schedules/:id", handlers.DeleteUVSchedule)
			uv.POST("/manual", handlers.ManualUV)
			uv.POST("/manual/stop", handlers.StopManualUV)
			uv.GET("/status", handlers.GetUVStatus)
		}

		// History routes
		api.GET("/history", handlers.GetHistory)

		// Stock routes
		api.GET("/stock", handlers.GetStock)
		api.PUT("/stock", handlers.UpdateStock)

		// Sensor routes
		sensors := api.Group("/sensors")
		{
			sensors.GET("/current", handlers.GetCurrentSensor)
			sensors.GET("/history", handlers.GetSensorHistory)
			sensors.POST("/inject", handlers.InjectSensorData) // For testing/demo
		}

		// Demo routes
		api.POST("/demo/seed", handlers.SeedDemoData)
		api.POST("/demo/clear", handlers.ClearDemoData)
	}

	return r
}
