package service

import "errors"

var (
	ErrRepositoryExists   = errors.New("repository with this id already exists")
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrNoAvailableBuckets = errors.New("no available buckets")

	ErrUploadNotFound = errors.New("upload not found")
	ErrUploadBig      = errors.New("upload is too big")
	ErrWrongOffset    = errors.New("wrong offset")
)
