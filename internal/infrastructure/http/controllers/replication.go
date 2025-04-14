package controllers

type ReplicateInput struct {
	SourceStorage Storage `json:"source_storage"`
	TargetStorage Storage `json:"target_storage"`
}

type Storage struct {
	ProviderID string `json:"provider_id"`
	Bucket     string `json:"bucket"`
}
