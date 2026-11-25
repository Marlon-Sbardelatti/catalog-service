package product

import (
	apierrors "catalog-service/internal/errors"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductHandler struct {
	service *ProductService
}

func NewHandler(service *ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) RegisterRoutes(router *gin.RouterGroup) {
	products := router.Group("/products")

	products.GET("", h.FindAll)
	products.GET("/:id", h.FindByID)
	products.POST("", h.Create)
	products.PUT("/:id", h.Update)
	products.DELETE("/:id", h.Delete)
}

func (h *ProductHandler) FindAll(c *gin.Context) {
	products, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		apierrors.Write(c, apierrors.Internal(), err)
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) FindByID(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid product id", nil), err)
		return
	}
	product, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest

	if err := c.ShouldBind(&req); err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid request body", nil), err)
		return
	}

	categoryID, err := bson.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid category id", nil), err)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("image is required", nil), err)
		return
	}

	product, err := h.service.Create(c.Request.Context(), req, categoryID, file)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid product id", nil), err)
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid product id", nil), err)
		return
	}

	var req UpdateProductRequest

	if err := c.ShouldBind(&req); err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid request body", nil), err)
		return
	}

	var categoryID bson.ObjectID

	if req.CategoryID != "" {
		categoryID, err = bson.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			apierrors.Write(c, apierrors.BadRequest("invalid category id", nil), err)
			return
		}
	}

	file, err := c.FormFile("image")

	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		apierrors.Write(c, apierrors.BadRequest("failed to read image", nil), err)
		return
	}

	product, err := h.service.Update(c.Request.Context(), id, req, categoryID, file)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.JSON(http.StatusOK, product)
}
