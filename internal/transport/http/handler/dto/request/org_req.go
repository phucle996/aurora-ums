package reqdto

type CreateTenantRequest struct {
	Name   string `json:"name" binding:"required"`
	Domain string `json:"domain" binding:"required"`
}

type CreateWorkspaceRequest struct {
	Name string  `json:"name" binding:"required"`
	Slug *string `json:"slug"`
}
