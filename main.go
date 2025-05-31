package main

import (
    "log"
    "net/http"
    
    "github.com/jeepinbird/stampkeeper/internal/config"
    "github.com/jeepinbird/stampkeeper/internal/database"
    "github.com/jeepinbird/stampkeeper/internal/router"
)

func main() {
    cfg := config.Load()
    
    db, err := database.Connect(cfg.DatabasePath)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    if err := database.Migrate(db); err != nil {
        log.Fatal("Failed to run migrations:", err)
    }
    
    if err := database.Seed(db); err != nil {
        log.Println("Warning: Failed to seed sample data:", err)
    }
    
    r := router.Setup(db)
    
    log.Printf("StampKeeper server starting on :%s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}