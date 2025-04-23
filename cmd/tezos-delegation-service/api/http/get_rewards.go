package http

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tezos-delegation-service/internal/model"
	"github.com/tezos-delegation-service/internal/usecase"
)

// GetRewardsHandler handles operations API requests.
type GetRewardsHandler struct {
	getRewardsFunc  usecase.GetRewardsFunc
	paginationLimit uint16
}

// NewGetRewardsHandler creates a new delegation handler.
func NewGetRewardsHandler(paginationLimit uint16, getRewardsFunc usecase.GetRewardsFunc) *GetRewardsHandler {
	return &GetRewardsHandler{
		getRewardsFunc:  getRewardsFunc,
		paginationLimit: paginationLimit,
	}
}

// GetRewards handles GET /xtz/delegations requests.
func (h *GetRewardsHandler) GetRewards(c *gin.Context) {
	ctx := c.Request.Context()

	page, limit, err := h.validateRequestParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.getRewardsFunc(ctx, strconv.Itoa(page), strconv.Itoa(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.setRequestIDHeader(c)
	h.setETagHeader(c, response)

	if c.GetHeader("If-None-Match") == c.Writer.Header().Get("ETag") {
		c.Status(http.StatusNotModified)
		return
	}

	c.JSON(http.StatusOK, response)
}

// validateRequestParams validates and parses request parameters.
func (h *GetRewardsHandler) validateRequestParams(c *gin.Context) (int, int, error) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		return 0, 0, errors.New("invalid page number")
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", fmt.Sprintf("%d", h.paginationLimit)))
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, fmt.Errorf("limit must be between 1 and 100, got %d", limit)
	}

	return page, limit, nil
}

// setPaginationHeaders sets pagination headers for the response.
func (h *GetRewardsHandler) setPaginationHeaders(c *gin.Context, pInfo model.PaginationInfo) {
	c.Header("X-Page-Current", strconv.Itoa(pInfo.CurrentPage))
	c.Header("X-Page-Per-Page", strconv.Itoa(pInfo.PerPage))

	if pInfo.HasPrevPage {
		c.Header("X-Page-Prev", strconv.Itoa(pInfo.PrevPage))
	}
	if pInfo.HasNextPage {
		c.Header("X-Page-Next", strconv.Itoa(pInfo.NextPage))
	}
}

// setRequestIDHeader sets the X-Request-ID header for the response.
func (h *GetRewardsHandler) setRequestIDHeader(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
		c.Header("X-Request-ID", requestID)
	}
}

// setCacheHeaders sets the cache headers for the response.
func (h *GetRewardsHandler) setCacheHeaders(c *gin.Context, year string) {
	if year != "" {
		c.Header("Cache-Control", "public, max-age=3600") // 1h cache
	} else {
		c.Header("Cache-Control", "public, max-age=300") // 5m cache
	}
}

// setETagHeader sets the ETag header for the response.
func (h *GetRewardsHandler) setETagHeader(c *gin.Context, response *model.RewardsResponse) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		etag := `"` + strconv.FormatInt(time.Now().UnixNano(), 36) + `"`
		c.Header("ETag", etag)
		return
	}

	hasher := sha256.New()
	hasher.Write(jsonData)

	page := c.Query("page")
	if page == "" {
		page = "1"
	}
	hasher.Write([]byte("page:" + page))

	limit := c.Query("limit")
	if limit == "" {
		limit = fmt.Sprintf("%d", h.paginationLimit)
	}
	hasher.Write([]byte("limit:" + limit))

	hashBytes := hasher.Sum(nil)
	etag := `"` + hex.EncodeToString(hashBytes) + `"`
	c.Header("ETag", etag)
}
