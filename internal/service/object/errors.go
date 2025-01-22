package object

import "errors"

var (
	ErrRepositoryExists   = errors.New("repository with this id already exists")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrNoAvailableBuckets = errors.New("no available buckets")
)
