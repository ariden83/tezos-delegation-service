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

	fromDate, toDate, wallet, backer, err := h.validateRequestParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := usecase.GetRewardsInput{
		FromDate: fromDate,
		ToDate:   toDate,
		Wallet:   wallet,
		Backer:   backer,
	}
	response, err := h.getRewardsFunc(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.setRequestIDHeader(c)
	/* h.setETagHeader(c, response)

	if c.GetHeader("If-None-Match") == c.Writer.Header().Get("ETag") {
		c.Status(http.StatusNotModified)
		return
	}*/

	c.JSON(http.StatusOK, response)
}

// validateRequestParams validates and parses request parameters.
func (h *GetRewardsHandler) validateRequestParams(c *gin.Context) (fromDate, toDate *time.Time, wallet, backer model.WalletAddress, err error) {

	if fromStr := c.DefaultQuery("from", ""); fromStr != "" {
		t, errParsing := time.Parse("2006-01-02", fromStr)
		if errParsing != nil {
			err = errors.New("invalid 'from' date format. Use YYYY-MM-DD")
			return
		}
		fromDate = &t
	}

	if toStr := c.DefaultQuery("to", ""); toStr != "" {
		t, errParsing := time.Parse("2006-01-02", toStr)
		if errParsing != nil {
			err = errors.New("invalid 'to' date format. Use YYYY-MM-DD")
			return
		}
		toDate = &t
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
