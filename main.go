package main

import (
    "fmt"
    "html/template"
    "math/rand"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/net/html"
)

func getImageURLsFromPage(url string) ([]string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    tokenizer := html.NewTokenizer(resp.Body)
    var imageURLs []string

    for {
        tokenType := tokenizer.Next()
        if tokenType == html.ErrorToken {
            break
        }

        token := tokenizer.Token()
        if token.Data == "img" {
            for _, attr := range token.Attr {
                if attr.Key == "src" && strings.HasPrefix(attr.Val, "http") {
                    imageURLs = append(imageURLs, attr.Val)
                }
            }
        }
    }
    return imageURLs, nil
}

func main() {
    rand.Seed(time.Now().UnixNano())
    r := gin.Default()

    // Load HTML template
    r.SetHTMLTemplate(template.Must(template.ParseFiles("templates/index.html")))

    // Redirect / to /web-image
    r.GET("/", func(c *gin.Context) {
        keyword := c.DefaultQuery("search", "nature")
        c.Redirect(http.StatusFound, fmt.Sprintf("/web-image?search=%s&page=1", keyword))
    })

    // /web-image endpoint (HTML or JSON response)
    r.GET("/web-image", func(c *gin.Context) {
        keyword := c.DefaultQuery("search", "nature")
        page := c.DefaultQuery("page", "1")
        nsfw := c.DefaultQuery("NSFW", "off")

        searchURL := fmt.Sprintf("https://wallhere.com/en/wallpapers?q=%s&page=%s", keyword, page)
        if nsfw == "on" {
            searchURL += "&NSFW=on"
        } else {
            searchURL += "&NSFW=off"
        }

        images, err := getImageURLsFromPage(searchURL)
        if err != nil || len(images) == 0 {
            c.String(500, "Failed to load images for keyword: %s", keyword)
            return
        }

        rand.Shuffle(len(images), func(i, j int) {
            images[i], images[j] = images[j], images[i]
        })

        if len(images) > 24 {
            images = images[:24]
        }

        // Support AJAX (fetch) request: return JSON if header is present
        if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
            c.JSON(http.StatusOK, images)
            return
        }

        // Otherwise, render full HTML
        pageNum, _ := strconv.Atoi(page)
        c.HTML(http.StatusOK, "index.html", gin.H{
            "images":   images,
            "keyword":  keyword,
            "nsfw":     nsfw,
            "nextPage": pageNum + 1,
        })
    })

    r.Run(":8080")
}
