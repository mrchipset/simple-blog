package blog

import (
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const PAGE_SIZE = 5

func MainViewHandle(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	db := GetDBConnection()
	var posts []Post
	var pagenation Pagenation
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("publish_state = true").
			Limit(PAGE_SIZE).
			Offset((page - 1) * PAGE_SIZE).
			Order("created_at desc").
			Find(&posts).Error; err != nil {
			return err
		}

		var count int64

		if err := tx.Model(&Post{}).Where("publish_state = true").Count(&count).Error; err != nil {
			return err
		}

		pagenation.Prev = int(math.Max(1, float64(page-1)))
		pagenation.Next = int(math.Min(math.Max(1, math.Ceil(float64(count)/PAGE_SIZE)), float64(page+1)))
		pagenation.Pages = make([]int, int(math.Ceil(float64(count)/PAGE_SIZE)))
		for i := 0; i < len(pagenation.Pages); i++ {
			pagenation.Pages[i] = i + 1
		}
		return nil

	})

	if err != nil {
		c.HTML(http.StatusOK, "view/index.html",
			gin.H{
				"title":      "Simple Blog",
				"posts":      []Post{},
				"pagenation": pagenation,
			})
	} else {
		c.HTML(http.StatusOK, "view/index.html",
			gin.H{
				"title":      "Simple Blog",
				"posts":      posts,
				"pagenation": pagenation,
			})
	}
}

func PostViewHandle(c *gin.Context) {
	_uuid := c.Param("uuid")
	_, err := uuid.Parse(_uuid)
	if err != nil {
		c.HTML(http.StatusBadRequest, "view/post.html", gin.H{
			"content": "load post content error",
		})
	}

	db := GetDBConnection()
	var post_content PostContent
	var post Post
	tx1 := db.Model(&Post{}).Where("uuid = ?", _uuid).First(&post)
	tx2 := db.Model(&PostContent{}).Where("uuid = ?", _uuid).First(&post_content)
	if tx1.RowsAffected == 0 || tx2.RowsAffected == 0 {
		c.HTML(http.StatusBadRequest, "view/post.html", gin.H{
			"content": "content not found",
		})
	} else {
		content, err := base64.StdEncoding.DecodeString(post_content.Content)
		if err != nil {
			c.HTML(http.StatusBadRequest, "view/post.html", gin.H{
				"content": err.Error,
			})
		}
		c.HTML(http.StatusBadRequest, "view/post.html", gin.H{
			"content": RenderToHtml(content),
			"title":   post.Title,
		})
	}
}

func PageViewHandle(c *gin.Context) {
	uuid := c.Param("uuid")
	c.String(http.StatusOK, fmt.Sprintf("page: %s", uuid))
}
