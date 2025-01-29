package routes

import "fmt"

const (
	uploadID = "upload_id"
	objectID = "object_id"
)

const regexp = "[a-f\\d]{32}"

var patternUploadID = fmt.Sprintf("{%s:%s}", uploadID, regexp)
var patternObjectID = fmt.Sprintf("{%s:%s}", objectID, regexp)
