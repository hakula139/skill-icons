package handler

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

//go:embed icons.json
var iconsJSON []byte

type catalog struct {
	icons        map[string]string
	iconNameList []string
	themedIcons  []string
}

var shortNames = map[string]string{
	"js":                "javascript",
	"ts":                "typescript",
	"py":                "python",
	"tailwind":          "tailwindcss",
	"vue":               "vuejs",
	"nuxt":              "nuxtjs",
	"go":                "golang",
	"cf":                "cloudflare",
	"wasm":              "webassembly",
	"postgres":          "postgresql",
	"k8s":               "kubernetes",
	"next":              "nextjs",
	"mongo":             "mongodb",
	"md":                "markdown",
	"ps":                "photoshop",
	"ai":                "illustrator",
	"pr":                "premiere",
	"ae":                "aftereffects",
	"scss":              "sass",
	"sc":                "scala",
	"net":               "dotnet",
	"gatsbyjs":          "gatsby",
	"gql":               "graphql",
	"vlang":             "v",
	"amazonwebservices": "aws",
	"bots":              "discordbots",
	"express":           "expressjs",
	"googlecloud":       "gcp",
	"mui":               "materialui",
	"windi":             "windicss",
	"unreal":            "unrealengine",
	"nest":              "nestjs",
	"ktorio":            "ktor",
	"pwsh":              "powershell",
	"au":                "audition",
	"rollup":            "rollupjs",
	"rxjs":              "reactivex",
	"rxjava":            "reactivex",
	"ghactions":         "githubactions",
	"sklearn":           "scikitlearn",
	"ml5":               "ml5js",
	"vb":                "visualbasic",
	"an":                "animate",
	"ca":                "capture",
	"cc":                "creativecloud",
	"ch":                "characteranimator",
	"me":                "mediaencoder",
	"pl":                "prelude",
	"ru":                "premiererush",
	"fs":                "fuse",
	"id":                "indesign",
	"ic":                "incopy",
	"sp":                "adobespark",
	"dw":                "dreamweaver",
	"dn":                "dimension",
	"ar":                "aero",
	"psc":               "photoshopclassic",
	"psx":               "photoshopexpress",
	"lr":                "lightroom",
	"lrc":               "lightroomclassic",
	"fr":                "fresco",
	"pf":                "portfolio",
	"st":                "stock",
	"be":                "behance",
	"br":                "bridge",
	"million":           "millionjs",
	"asm":               "assembly",
	"pop":               "popos",
	"nix":               "nixos",
	"hc":                "holyc",
	"yml":               "yaml",
	"twitter":           "x",
	"arc":               "arcbrowser",
	"hf":                "huggingface",
	"sqla":              "sqlalchemy",
	"notepad++":         "notepadpp",
	"jq":                "jqlang",
}

var app = sync.OnceValue(func() *gin.Engine { return newApp(iconsJSON) })

const (
	ICONS_PER_LINE = 15
	ONE_ICON       = 48
	SCALE          = float64(ONE_ICON) / float64(300-44)
)

func (c *catalog) generateSvg(iconNames []string, perLine int, hasTitlesEnabled bool, align string) string {
	iconSvgList := make([]string, len(iconNames))

	for i, name := range iconNames {
		iconSvgList[i] = c.icons[name]
	}

	length := int(math.Min(float64(perLine*300), float64(len(iconNames)*300))) - 44
	height := int(math.Ceil(float64(len(iconSvgList))/float64(perLine)))*300 - 44
	scaledHeight := int(float64(height) * SCALE)
	scaledWidth := int(float64(length) * SCALE)

	var svg string
	if align == "center" {
		svg = fmt.Sprintf(`
            <svg width="100%%" height="%d" viewBox="0 0 %d %d" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1">
        `, scaledHeight, length, height)
	} else if align == "right" {
		svg = fmt.Sprintf(`
            <svg width="calc(200%% - %dpx)" height="%d" viewBox="0 0 %d %d" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1">
        `, scaledWidth, scaledHeight, length, height)
	} else {
		svg = fmt.Sprintf(`
            <svg width="%d" height="%d" viewBox="0 0 %d %d" fill="none" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1">
        `, scaledWidth, scaledHeight, length, height)
	}

	for index, i := range iconSvgList {
		var title string
		if hasTitlesEnabled {
			title = fmt.Sprintf("<title>%s</title>", iconNames[index])
		}

		x := (index % perLine) * 300
		y := (index / perLine) * 300
		svg += fmt.Sprintf(`
		<g transform="translate(%d, %d)">
            %s
			%s
		</g>
		`, x, y, title, i)
	}

	svg += "</svg>"
	return svg
}

func (c *catalog) parseShortNames(names []string, theme string) []string {
	result := make([]string, len(names))

	for i, name := range names {
		if contains(c.iconNameList, name) {
			if contains(c.themedIcons, name) {
				result[i] = name + "-" + theme
			} else {
				result[i] = name
			}
		} else if val, ok := shortNames[name]; ok {
			if contains(c.themedIcons, val) {
				result[i] = val + "-" + theme
			} else {
				result[i] = val
			}
		}
	}
	return result
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func (c *catalog) iconRoute(r *gin.RouterGroup) {
	r.GET("/icons", func(ctx *gin.Context) {
		ctx.Request.ParseForm()
		iconParam := ctx.Request.Form.Get("i")

		theme := ctx.Request.Form.Get("theme")
		if theme == "" {
			theme = "auto"
		}

		perLineStr := ctx.Request.Form.Get("perline")
		if perLineStr == "" {
			perLineStr = "15"
		}

		hasTitlesEnabled := ctx.Request.Form.Get("titles") != ""

		align := ctx.Request.Form.Get("align")
		if align != "left" && align != "right" && align != "center" {
			align = "left"
		}

		if iconParam == "" {
			ctx.String(http.StatusBadRequest, "You didn't specify any icons!")
			return
		}

		if theme != "dark" && theme != "light" && theme != "auto" {
			ctx.String(http.StatusBadRequest, "Theme must be either 'light', 'dark' or 'auto'")
			return
		}

		perLine, err := strconv.Atoi(perLineStr)
		if err != nil || perLine < 1 || perLine > 50 {
			ctx.String(http.StatusBadRequest, "Icons per line must be a number between 1 and 50")
			return
		}

		var iconShortNames []string
		if iconParam == "all" {
			iconShortNames = c.iconNameList
		} else {
			iconShortNames = strings.Split(iconParam, ",")
		}

		iconNames := c.parseShortNames(iconShortNames, theme)
		if iconNames == nil {
			ctx.String(http.StatusBadRequest, "You didn't format the icons param correctly!")
			return
		}

		svg := c.generateSvg(iconNames, perLine, hasTitlesEnabled, align)

		ctx.Header("Content-Type", "image/svg+xml")
		ctx.String(http.StatusOK, svg)
	})
}

func newApp(data []byte) *gin.Engine {
	c := &catalog{icons: make(map[string]string)}
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&c.icons); err != nil {
		panic(err)
	}

	for key := range c.icons {
		c.iconNameList = append(c.iconNameList, strings.Split(key, "-")[0])
		if strings.Contains(key, "-light") || strings.Contains(key, "-dark") || strings.Contains(key, "-auto") {
			c.themedIcons = append(c.themedIcons, strings.Split(key, "-")[0])
		}
	}

	router := gin.New()
	c.iconRoute(router.Group("/api"))
	return router
}

func Handler(w http.ResponseWriter, r *http.Request) {
	app().ServeHTTP(w, r)
}

func Run() error {
	gin.SetMode(gin.ReleaseMode)
	return app().Run()
}
