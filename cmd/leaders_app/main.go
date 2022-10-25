package main

import (
	"fmt"
	"html/template"
	"net/http"

	"leaders_app/pkg/handlers"
	"leaders_app/pkg/items"
	"leaders_app/pkg/middleware"
	"leaders_app/pkg/session"
	"leaders_app/pkg/user"

	"github.com/gorilla/mux"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go.uber.org/zap"
)

func main() {
	templates := template.Must(template.ParseGlob("./templates/*"))

	dsn := "host=localhost user=postgres password=3546"
	dsn += " dbname=gusev port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err) // TODO
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err) // TODO
	}
	err = sqlDB.Ping()
	if err != nil {
		panic(err) // TODO
	}

	type testOrm struct {
		Di          int
		Description int `sql:"null"`
	}

	testItem := testOrm{}
	db.Table("test_table").Take(&testItem)
	fmt.Println(testItem)

	toCreate := testOrm{10500, 3}
	db.Table("test_table").Create(&toCreate)
	db.Table("test_table").Take(&testItem)
	fmt.Println(testItem)

	sm := session.NewSessionsManager()
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync() // flushes buffer, if any
	logger := zapLogger.Sugar()

	userRepo := user.NewMemoryRepo()
	itemsRepo := items.NewMemoryRepo()

	userHandler := &handlers.UserHandler{
		Tmpl:     templates,
		UserRepo: userRepo,
		Logger:   logger,
		Sessions: sm,
	}

	handlers := &handlers.ItemsHandler{
		Tmpl:      templates,
		Logger:    logger,
		ItemsRepo: itemsRepo,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", userHandler.Index).Methods("GET")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/logout", userHandler.Logout).Methods("POST")

	r.HandleFunc("/items", handlers.List).Methods("GET")
	r.HandleFunc("/items/new", handlers.AddForm).Methods("GET")
	r.HandleFunc("/items/new", handlers.Add).Methods("POST")
	r.HandleFunc("/items/{id}", handlers.Edit).Methods("GET")
	r.HandleFunc("/items/{id}", handlers.Update).Methods("POST")
	r.HandleFunc("/items/{id}", handlers.Delete).Methods("DELETE")

	mux := middleware.Auth(sm, r)
	mux = middleware.AccessLog(logger, mux)
	mux = middleware.Panic(mux)

	addr := ":8080"
	logger.Infow("starting server",
		"type", "START",
		"addr", addr,
	)
	http.ListenAndServe(addr, mux)
}
