package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Fungsi untuk natural sort
func naturalLess(a, b string) bool {
	aRunes := []rune(a)
	bRunes := []rune(b)

	var aNum, bNum int


	i, j := 0, 0

	for i < len(aRunes) && j < len(bRunes) {
		aCh, bCh := aRunes[i], bRunes[j]

		// Jika kedua karakter adalah digit
		if unicode.IsDigit(aCh) && unicode.IsDigit(bCh) {
			// Ekstrak angka penuh dari a
			aNum = 0
			for i < len(aRunes) && unicode.IsDigit(aRunes[i]) {
				aNum = aNum*10 + int(aRunes[i]-'0')
				i++
			}

			// Ekstrak angka penuh dari b
			bNum = 0
			for j < len(bRunes) && unicode.IsDigit(bRunes[j]) {
				bNum = bNum*10 + int(bRunes[j]-'0')
				j++
			}

			if aNum != bNum {
				return aNum < bNum
			}
		} else {
			if aCh != bCh {
				return aCh < bCh
			}
			i++
			j++
		}
	}

	return len(aRunes) < len(bRunes)
}

func convertImagesToPdf(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}
	files := form.File["images"]

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada file yang diunggah"})
		return
	}

	// Urutkan file dengan natural sort
	sort.Slice(files, func(i, j int) bool {
		return naturalLess(files[i].Filename, files[j].Filename)
	})

	// Buat folder jika belum ada
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat folder uploads"})
		return
	}

	// Buat nama PDF unik
	pdfName := fmt.Sprintf("output_%d.pdf", time.Now().Unix())
	pdfPath := "output/" + pdfName

	if err := os.MkdirAll("output", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat folder output"})
		return
	}

	var imagePaths []string
	for _, file := range files {
		// Validasi ukuran file
		if file.Size > 10<<20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File terlalu besar (maks 10MB)"})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya mendukung file JPG, JPEG, atau PNG"})
			return
		}

		// Simpan file sementara dengan nama asli
		filePath := "uploads/" + file.Filename
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file"})
			return
		}

		// Konversi ke JPG jika perlu
		jpgPath := filePath
		if ext == ".png" {
			jpgPath = strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".jpg"
			if err := convertToJpg(filePath, jpgPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengonversi PNG ke JPG"})
				return
			}
			os.Remove(filePath)
		} else if ext == ".jpeg" {
			newPath := strings.TrimSuffix(filePath, ".jpeg") + ".jpg"
			if err := os.Rename(filePath, newPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengganti nama file"})
				return
			}
			jpgPath = newPath
		}

		imagePaths = append(imagePaths, jpgPath)
	}

	// Buat PDF dengan gambar terurut
	if err := api.ImportImagesFile(imagePaths, pdfPath, nil, nil); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat PDF: " + err.Error()})
		return
	}

	// Bersihkan file gambar
	for _, path := range imagePaths {
		os.Remove(path)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "PDF berhasil dibuat",
		"pdf":     pdfName,
		"path":    "/pdfs/" + pdfName,
	})
}

func convertToJpg(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return err
	}

	if format == "png" {
		rgba := image.NewRGBA(img.Bounds())
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				rgba.Set(x, y, img.At(x, y))
			}
		}
		img = rgba
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	os.MkdirAll("uploads", os.ModePerm)
	os.MkdirAll("output", os.ModePerm)

	r := gin.Default()
	r.Use(CORSMiddleware())

	r.Static("/static", "./static")
	r.Static("/pdfs", "./output")
	
	r.POST("/convert", convertImagesToPdf)
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	fmt.Println("Server berjalan di http://localhost:8080")
	r.Run(":8080")
}