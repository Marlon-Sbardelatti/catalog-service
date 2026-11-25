package category

import (
	apierrors "catalog-service/internal/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CategoryHandler struct {
	service *CategoryService
}

func NewCategoryHandler(service *CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	categories := router.Group("/categories")

	categories.GET("", h.FindAll)
	categories.GET("/:id", h.FindByID)
	categories.POST("", h.Create)
	categories.PUT("/:id", h.Update)
	categories.DELETE("/:id", h.Delete)
}

func (h *CategoryHandler) FindAll(c *gin.Context) {
	categories, err := h.service.FindAll(c.Request.Context())
	if err != nil {
		apierrors.Write(c, apierrors.Internal(), err)
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) FindByID(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid category id", nil), err)
		return
	}
	category, err := h.service.FindByID(c.Request.Context(), id)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req CreateCategoryRequest

	if err := c.ShouldBind(&req); err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid request body", nil), err)
		return
	}

	category, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid category id", nil), err)
		return
	}

	var req UpdateCategoryRequest

	if err := c.ShouldBind(&req); err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid request body", nil), err)
		return
	}

	category, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		apierrors.Write(c, apierrors.BadRequest("invalid category id", nil), err)
		return
	}

	err = h.service.Delete(c.Request.Context(), id)
	if err != nil {
		apierrors.Write(c, toAPIError(err), err)
		return
	}

	c.Status(http.StatusNoContent)
}
