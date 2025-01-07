package main

import (
	"context"
	"fmt"

	"github.com/inview-team/gorynych/pkg/provider/yandex"
)

func main() {
	ctx := context.Background()
	yandexStorage, _ := yandex.New(ctx, yandex.Credentials{
		AccessKeyID:     "...",
		SecretAccessKey: "...",
	})

	buckets, err := yandexStorage.Bucket.List(ctx)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(buckets)
}
