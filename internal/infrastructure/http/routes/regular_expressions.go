package routes

import "fmt"

const (
	uploadID = "upload_id"
)

const regexp = "[a-f\\d]{32}"

var patternUploadID = fmt.Sprintf("{%s:%s}", uploadID, regexp)
