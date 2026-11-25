package product

type CreateProductRequest struct {
	Name        string `form:"name" binding:"required"`
	CategoryID  string `form:"category_id" binding:"required"`
	Description string `form:"description"`
	Price       int64  `form:"price" binding:"required,gt=0"`
}

type UpdateProductRequest struct {
	Name        string `form:"name"`
	CategoryID  string `form:"category_id"`
	Description string `form:"description"`
	Price       int64  `form:"price" binding:"gt=0"`
}
