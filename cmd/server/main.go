package main

import (
	"github.com/inview-team/gorynych/internal/application"
	"github.com/inview-team/gorynych/pkg/provider/yandex"
)

func main() {
	objectRepo := yandex.New()
	app := application.New(objectRepo)
}
