package main

import (
	"context"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
	todo "todolistBackend"
	"todolistBackend/pkg/handler"
	"todolistBackend/pkg/logging"
	"todolistBackend/pkg/repository"
	"todolistBackend/pkg/service"
)

const (
	configPath = "configs"
	configName = "config"
)

func main() {
	logger := logging.GetLogger()
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing configs: %s", err.Error())
	}
	if err := godotenv.Load(); err != nil {
		logger.Fatalf("error loading env variables: %s", err.Error())
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.name"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("db.password"),
	})
	if err != nil {
		logger.Fatalf("failed to initialize db due to error: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(todo.Server)
	go func() {
		if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
			logger.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logger.Printf("Todoapp started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Printf("Todoapp shutting down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logger.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logger.Errorf("error occured on db connection closing: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	return viper.ReadInConfig()
}
