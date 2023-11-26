package policies

// PolicyRule struct is used to create or delete the policies rule
type PolicyRule struct {
	Ptype    string `json:"ptype" binding:"required" validate:"required"`
	RoleName string `json:"role_name" binding:"required" validate:"required"`
	Path     string `json:"path" binding:"required" validate:"required"`
	Method   string `json:"method" binding:"required" validate:"required"`
}

// PolicyModel includes PolicyRule and an automatically generated id
type PolicyModel struct {
	ID int `json:"id"`
	PolicyRule
}

// Single return structure file
type Single struct {
	RoleName string `json:"role_name"`
	Path     string `json:"path"`
	Method   string `json:"method"`
}
