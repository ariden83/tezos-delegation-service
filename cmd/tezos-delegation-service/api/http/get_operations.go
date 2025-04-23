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

// GetOperationsHandler handles operations API requests.
type GetOperationsHandler struct {
	getOperationsFunc usecase.GetOperationsFunc
	paginationLimit   uint16
}

// NewGetOperationsHandler creates a new operation's handler.
func NewGetOperationsHandler(paginationLimit uint16, getOperationsFunc usecase.GetOperationsFunc) *GetOperationsHandler {
	return &GetOperationsHandler{
		getOperationsFunc: getOperationsFunc,
		paginationLimit:   paginationLimit,
	}
}

// GetOperations handles GET /xtz/operations requests for retrieving all on-chain
// transactions related to staking (delegate/undelegate, stake/unstake, rewards payments).
func (h *GetOperationsHandler) GetOperations(c *gin.Context) {
	ctx := c.Request.Context()

	page, limit, operationType, wallet, baker, err := h.validateRequestParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.getOperationsFunc(ctx, strconv.Itoa(page), strconv.Itoa(limit), operationType, wallet, baker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.setPaginationHeaders(c, response.Pagination)
	h.setRequestIDHeader(c)
	h.setETagHeader(c, response)

	if c.GetHeader("If-None-Match") == c.Writer.Header().Get("ETag") {
		c.Status(http.StatusNotModified)
		return
	}

	c.JSON(http.StatusOK, response)
}

// validateRequestParams validates and parses request parameters.
func (h *GetOperationsHandler) validateRequestParams(c *gin.Context) (page int, limit int, operationType model.OperationType, wallet model.WalletAddress, backer model.WalletAddress, err error) {
	page, err = strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		err = errors.New("invalid page number")
		return
	}

	limit, err = strconv.Atoi(c.DefaultQuery("limit", fmt.Sprintf("%d", h.paginationLimit)))
	if err != nil || limit < 1 || limit > 100 {
		err = fmt.Errorf("limit must be between 1 and 100, got %d", limit)
		return
	}

	year := c.DefaultQuery("year", "")
	if year != "" {
		_, err := strconv.Atoi(year)
		if err != nil {
			err = errors.New("invalid year format")
		}
	}

	operationTypeStr := c.DefaultQuery("type", "")
	operationType = model.OperationType(operationTypeStr)

	if operationType != "" && operationType.IsValid() == false {
		err = fmt.Errorf("invalid operation type: %s", operationType.String())
		return
	}

	walletStr := c.DefaultQuery("wallet", "")
	wallet = model.WalletAddress(walletStr)
	if wallet == "" {
		err = errors.New("missing wallet address")
		return
	} else if wallet.IsValid() == false {
		err = fmt.Errorf("invalid wallet address: %s", wallet.String())
	}

	backerStr := c.DefaultQuery("backer", "")
	backer = model.WalletAddress(backerStr)
	if backer == "" {
		err = errors.New("missing backer address")
	} else if backer.IsValid() == false {
		err = fmt.Errorf("invalid backer address: %s", backer.String())
	}

	return
}

// setPaginationHeaders sets pagination headers for the response.
func (h *GetOperationsHandler) setPaginationHeaders(c *gin.Context, pInfo model.PaginationInfo) {
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
func (h *GetOperationsHandler) setRequestIDHeader(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
		c.Header("X-Request-ID", requestID)
	}
}

// setCacheHeaders sets the cache headers for the response.
func (h *GetOperationsHandler) setCacheHeaders(c *gin.Context, year string) {
	if year != "" {
		c.Header("Cache-Control", "public, max-age=3600") // 1h cache
	} else {
		c.Header("Cache-Control", "public, max-age=300") // 5m cache
	}
}

// setETagHeader sets the ETag header for the response.
func (h *GetOperationsHandler) setETagHeader(c *gin.Context, response *model.OperationsResponse) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		etag := `"` + strconv.FormatInt(time.Now().UnixNano(), 36) + `"`
		c.Header("ETag", etag)
		return
	}

	hasher := sha256.New()
	hasher.Write(jsonData)

	maxDelegationIDHeader := c.GetHeader("X-Max-Delegation-ID")
	hasher.Write([]byte("X-Max-Delegation-ID:" + maxDelegationIDHeader))

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

	year := c.Query("year")
	if year != "" {
		hasher.Write([]byte("year:" + year))
	}

	opType := c.Query("type")
	if opType != "" {
		hasher.Write([]byte("type:" + opType))
	}

	hashBytes := hasher.Sum(nil)
	etag := `"` + hex.EncodeToString(hashBytes) + `"`
	c.Header("ETag", etag)
}
