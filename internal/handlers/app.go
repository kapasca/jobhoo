package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"jobhoo/internal/middleware"
	"jobhoo/internal/repository"
	"jobhoo/internal/services/aimatching"
	authsvc "jobhoo/internal/services/auth"
)

// App wires together everything a handler needs: repositories, services,
// and the template renderer. Handlers are methods on *App.
type App struct {
	DB *sql.DB

	Users        *repository.UserRepo
	Sessions     *repository.SessionRepo
	Jobs         *repository.JobRepo
	Applications *repository.ApplicationRepo

	Auth       *authsvc.Service
	AIMatching aimatching.Provider

	Render *Renderer

	UploadDir string // absolute path to the uploads folder on disk
}

// PageData is a small helper for building template data maps consistently.
type PageData map[string]interface{}

func (a *App) newPageData(r *http.Request, title string) PageData {
	return PageData{
		"Title": title,
		"User":  middleware.CurrentUser(r),
	}
}

func (a *App) renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.WriteHeader(status)
	data := a.newPageData(r, "Error")
	data["Error"] = message
	a.Render.Render(w, "404.html", data)
}

// saveUpload streams a multipart file to disk under a subfolder (resumes/documents)
// and returns the public URL path plus the original filename.
func (a *App) saveUpload(fh *multipart.FileHeader, subfolder string) (publicPath, filename string, err error) {
	src, err := fh.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	ext := filepath.Ext(fh.Filename)
	if strings.ToLower(ext) != ".pdf" {
		return "", "", fmt.Errorf("hanya file PDF yang diperbolehkan")
	}

	safeName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destDir := filepath.Join(a.UploadDir, subfolder)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return "", "", err
	}
	destPath := filepath.Join(destDir, safeName)

	dst, err := os.Create(destPath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", "", err
	}

	return "/uploads/" + subfolder + "/" + safeName, fh.Filename, nil
}
