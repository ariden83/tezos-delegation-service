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

// GetDelegationHandler handles delegation API requests.
type GetDelegationHandler struct {
	getDelegationsFunc usecase.GetDelegationsFunc
}

// NewGetDelegationHandler creates a new delegation handler.
func NewGetDelegationHandler(getDelegationsFunc usecase.GetDelegationsFunc) *GetDelegationHandler {
	return &GetDelegationHandler{
		getDelegationsFunc: getDelegationsFunc,
	}
}

// GetDelegations handles GET /xtz/delegations requests.
func (h *GetDelegationHandler) GetDelegations(c *gin.Context) {
	ctx := c.Request.Context()

	page, limit, year, err := h.validateRequestParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxDelegationID := h.extractMaxDelegationID(c)

	response, err := h.getDelegationsFunc(ctx, strconv.Itoa(page), strconv.Itoa(limit), year, maxDelegationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.setPaginationHeaders(c, response.Pagination)
	h.setRequestIDHeader(c)
	h.setCacheHeaders(c, year)
	h.setMaxDelegationIDHeader(c, response.MaxDelegationID)
	h.setETagHeader(c, response)

	if c.GetHeader("If-None-Match") == c.Writer.Header().Get("ETag") {
		c.Status(http.StatusNotModified)
		return
	}

	c.JSON(http.StatusOK, response)
}

// extractMaxDelegationID extracts and parses the X-Max-Delegation-ID header.
func (h *GetDelegationHandler) extractMaxDelegationID(c *gin.Context) int64 {
	maxDelegationID := int64(0)
	maxDelegationIDStr := c.GetHeader("X-Max-Delegation-ID")
	if maxDelegationIDStr != "" {
		parsedID, err := strconv.ParseInt(maxDelegationIDStr, 10, 64)
		if err == nil && parsedID > 0 {
			maxDelegationID = parsedID
		}
	}
	return maxDelegationID
}

// validateRequestParams validates and parses request parameters.
func (h *GetDelegationHandler) validateRequestParams(c *gin.Context) (int, int, string, error) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		return 0, 0, "", errors.New("invalid page number")
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		return 0, 0, "", fmt.Errorf("limit must be between 1 and 100, got %d", limit)
	}

	year := c.DefaultQuery("year", "")
	return page, limit, year, nil
}

// setPaginationHeaders sets pagination headers for the response.
func (h *GetDelegationHandler) setPaginationHeaders(c *gin.Context, pInfo model.PaginationInfo) {
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
func (h *GetDelegationHandler) setRequestIDHeader(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
		c.Header("X-Request-ID", requestID)
	}
}

// setCacheHeaders sets the cache headers for the response.
func (h *GetDelegationHandler) setCacheHeaders(c *gin.Context, year string) {
	if year != "" {
		c.Header("Cache-Control", "public, max-age=3600") // 1h cache
	} else {
		c.Header("Cache-Control", "public, max-age=300") // 5m cache
	}
}

// setETagHeader sets the ETag header for the response.
func (h *GetDelegationHandler) setETagHeader(c *gin.Context, response *model.DelegationResponse) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		etag := `"` + strconv.FormatInt(time.Now().UnixNano(), 36) + `"`
		c.Header("ETag", etag)
		return
	}
	hash := sha256.Sum256(jsonData)
	etag := `"` + hex.EncodeToString(hash[:]) + `"`
	c.Header("ETag", etag)
}

// setMaxDelegationIDHeader sets the X-Max-Delegation-ID header with the highest delegation ID.
func (h *GetDelegationHandler) setMaxDelegationIDHeader(c *gin.Context, maxDelegationID int64) {
	if maxDelegationID > 0 {
		c.Header("X-Max-Delegation-ID", strconv.FormatInt(maxDelegationID, 10))
	}
}
