package request

type CreateURLRequest struct {
	URL            string `json:"url" binding:"required,url"`
	CustomSlug     string `json:"customSlug" binding:"omitempty,alphanum,min=3,max=8"`
	ExpirationDate string `json:"expirationDate" binding:"omitempty,datetime=2006-01-02"`
}

type ValidateSlugRequest struct {
	CustomSlug string `json:"customSlug" binding:"required,alphanum,min=3,max=8"`
}

type UpdateExpirationRequest struct {
	ExpirationDate string `json:"expirationDate" binding:"required,datetime=2006-01-02"`
}
