package blog

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func PrintContent(content string) string {
	fmt.Println(content)
	return content
}

type PostListItem struct {
	UUID       string
	Title      string
	UpdateDate string
	Published  bool
}

type Token struct {
	Secret string
	Time   time.Time
}

type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var gToken Token

func AuthMiddware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := ctx.Request.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusTemporaryRedirect, "/editor/login")
		} else {
			decodedValue, err := url.QueryUnescape(token.Value)
			if err != nil {
				ctx.Redirect(http.StatusTemporaryRedirect, "/editor/login")
			}
			if token.Value == "" || decodedValue != gToken.Secret {
				ctx.Redirect(http.StatusTemporaryRedirect, "/editor/login")
			} else {
				ctx.SetCookie(token.Name, decodedValue, 3600*24, token.Path, token.Domain, token.Secure, token.HttpOnly)
				ctx.Next()
			}
		}
	}
}

func LoginViewHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "editor/login.html", gin.H{})
}

func LoginViewAPIHandle(c *gin.Context) {
	// generate the secret
	gToken.Time = time.Now().Add(time.Hour * 24)
	var loginInfo LoginInfo
	err := c.BindJSON(&loginInfo)
	if err != nil {
		c.String(http.StatusForbidden, "")
	}

	username := os.Getenv("BLOG_ADMIN_USERNAME")
	password := os.Getenv("BLOG_ADMIN_PASSWORD")
	if loginInfo.Username == username && loginInfo.Password == password {
		string_secret := loginInfo.Username + loginInfo.Password + time.Now().Format("2006-01-02 11:21:00")
		buffer := []byte(string_secret)
		sum := sha256.Sum256(buffer)
		gToken.Secret = base64.StdEncoding.EncodeToString(sum[:])
		c.SetCookie("token", gToken.Secret, 3600*24, "/", c.Request.Host, false, true)
		c.String(http.StatusOK, gToken.Secret)
	} else {
		c.String(http.StatusForbidden, "")
	}
}

func PostListViewHandle(c *gin.Context) {
	db := GetDBConnection()
	var posts []Post
	if err := db.Where("1 = 1").
		Order("created_at").
		Find(&posts).Error; err != nil {
		c.HTML(http.StatusOK, "editor/index.html",
			gin.H{
				"posts": []PostListItem{},
			})
	} else {
		var postListItems []PostListItem = make([]PostListItem, len(posts))
		for i := 0; i < len(posts); i++ {
			postListItems[i].UUID = posts[i].UUID.String()
			postListItems[i].Title = posts[i].Title
			postListItems[i].Published = posts[i].PublishState
			postListItems[i].UpdateDate = posts[i].UpdatedAt.Format("2006-01-02 15:04:05")
		}
		c.HTML(http.StatusOK, "editor/index.html",
			gin.H{
				"posts": postListItems,
			})
	}
}

func EditorViewHandle(c *gin.Context) {
	_uuid := c.Param("uuid")
	_, err := uuid.Parse(_uuid)
	if err != nil {
		c.HTML(http.StatusOK, "editor/compose.html",
			gin.H{
				"error": true,
			})
	}
	var query PostQuery
	db := GetDBConnection()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("uuid = ?", _uuid).First(&query.Post).Error; err != nil {
			return err
		}
		if err := tx.Where("uuid = ?", _uuid).First(&query.PostContent).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.HTML(http.StatusOK, "editor/compose.html",
			gin.H{
				"error": true,
			})
	} else {
		c.HTML(http.StatusOK, "editor/compose.html",
			gin.H{
				"error":   false,
				"title":   query.Post.Title,
				"content": query.PostContent.Content,
			},
		)
	}
}

func PreviewAPIHandle(c *gin.Context) {
	content, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("Error")
		c.String(http.StatusBadRequest, "Error Request.")
	}
	c.String(http.StatusOK, RenderToHtmlStr(content))
}

func SavePostAPIHandle(c *gin.Context) {
	_uuid := c.Param("uuid")
	_, err := uuid.Parse(_uuid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       400,
			"description": "post data error",
		})
	}

	query := PostQuery{}
	err = c.BindJSON(&query)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       400,
			"description": "post data error",
		})
	} else {
		// update publish state
		query.Post.Link = fmt.Sprintf("/post/%s", _uuid)
		content, err := base64.StdEncoding.DecodeString(query.PostContent.Content)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":       400,
				"description": err.Error(),
			})
		} else {
			query.Post.Summary, _ = SplitSummary(string(content))
			db := GetDBConnection()
			err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Model(&Post{}).Where("uuid = ?", _uuid).Updates(query.Post).Error; err != nil {
					return err
				}

				if err := tx.Model(&PostContent{}).Where("uuid = ?", _uuid).Updates(query.PostContent).Error; err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":       400,
					"description": err.Error(),
				})
			} else {
				c.JSON(http.StatusOK, gin.H{})
			}
		}

	}
}

func PublishPostAPIHandle(c *gin.Context) {
	_uuid := c.Param("uuid")
	_, err := uuid.Parse(_uuid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       400,
			"description": "post data error",
		})
	}

	query := PostQuery{}
	err = c.BindJSON(&query)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       400,
			"description": "post data error",
		})
	} else {
		// update publish state
		query.Post.PublishState = true
		query.Post.Link = fmt.Sprintf("/post/%s", _uuid)
		content, err := base64.StdEncoding.DecodeString(query.PostContent.Content)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":       400,
				"description": err.Error(),
			})
		} else {
			query.Post.Summary, _ = SplitSummary(string(content))
			db := GetDBConnection()
			err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Model(&Post{}).Where("uuid = ?", _uuid).Updates(query.Post).Error; err != nil {
					return err
				}

				if err := tx.Model(&PostContent{}).Where("uuid = ?", _uuid).Updates(query.PostContent).Error; err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":       400,
					"description": err.Error(),
				})
			} else {
				c.JSON(http.StatusOK, gin.H{})
			}
		}

	}
}

func CreateNewPostAPIHandle(c *gin.Context) {
	_uuid := uuid.New()
	post := Post{
		UUID: _uuid,
	}
	post_content := PostContent{
		UUID: _uuid,
	}

	db := GetDBConnection()
	err := db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&post).Error; err != nil {
			return err
		}

		if err := tx.Create(&post_content).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.String(http.StatusForbidden, "")
	} else {
		c.String(http.StatusOK, post.UUID.String())
	}
}

func DeletePostAPIHandle(c *gin.Context) {
	param := c.Param("uuid")
	_, err := uuid.Parse(param)
	if err != nil {
		c.String(http.StatusBadRequest, "error request")
	}

	db := GetDBConnection()
	err = db.Transaction(func(tx *gorm.DB) error {
		var _tx *gorm.DB
		if _tx = tx.Where("uuid = ?", param).Delete(&Post{}); _tx.Error != nil {
			return _tx.Error
		}
		if _tx = tx.Where("uuid = ?", param).Delete(&PostContent{}); _tx.Error != nil {
			return _tx.Error
		}
		return nil
	})

	if err != nil {
		c.String(http.StatusForbidden, err.Error())
	} else {
		c.String(http.StatusOK, param)
	}
}
