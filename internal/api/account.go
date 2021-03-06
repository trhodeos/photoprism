package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

// GET /api/v1/accounts
func GetAccounts(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/accounts", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.AccountSearch

		q := query.New(conf.Db())
		err := c.MustBindWith(&f, binding.Form)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		result, err := q.Accounts(f)

		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		c.Header("X-Result-Count", strconv.Itoa(f.Count))
		c.Header("X-Result-Offset", strconv.Itoa(f.Offset))

		c.JSON(http.StatusOK, result)
	})
}

// GET /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func GetAccount(router *gin.RouterGroup, conf *config.Config) {
	router.GET("/accounts/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		q := query.New(conf.Db())
		id := ParseUint(c.Param("id"))

		if m, err := q.AccountByID(id); err == nil {
			c.JSON(http.StatusOK, m)
		} else {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
		}
	})
}

// POST /api/v1/accounts
func CreateAccount(router *gin.RouterGroup, conf *config.Config) {
	router.POST("/accounts", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		var f form.Account

		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		if err := f.ServiceDiscovery(); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		m, err := entity.CreateAccount(f, conf.Db())

		log.Debugf("create account: %+v %+v", f, m)

		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		event.Success("account created")

		c.JSON(http.StatusOK, m)
	})
}

// PUT /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func UpdateAccount(router *gin.RouterGroup, conf *config.Config) {
	router.PUT("/accounts/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := ParseUint(c.Param("id"))

		q := query.New(conf.Db())

		m, err := q.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrPhotoNotFound)
			return
		}

		// 1) Init form with model values
		f, err := form.NewAccount(m)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		// 2) Update form with values from request
		if err := c.BindJSON(&f); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		// 3) Save model with values from form
		if err := m.Save(f, conf.Db()); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": txt.UcFirst(err.Error())})
			return
		}

		event.Success("account saved")

		m, err = q.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}

// DELETE /api/v1/accounts/:id
//
// Parameters:
//   id: string Account ID as returned by the API
func DeleteAccount(router *gin.RouterGroup, conf *config.Config) {
	router.DELETE("/accounts/:id", func(c *gin.Context) {
		if Unauthorized(c, conf) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		id := ParseUint(c.Param("id"))
		q := query.New(conf.Db())

		m, err := q.AccountByID(id)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, ErrAccountNotFound)
			return
		}

		if err := m.Delete(conf.Db()); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		event.Success("account deleted")

		c.JSON(http.StatusOK, m)
	})
}
