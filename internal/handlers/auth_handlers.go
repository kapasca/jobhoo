package handlers

import (
	"net/http"
	"strings"
	"time"

	"jobhoo/internal/models"
	"jobhoo/internal/repository"
	authsvc "jobhoo/internal/services/auth"
)

func (a *App) HomePage(w http.ResponseWriter, r *http.Request) {
	jobs, err := a.Jobs.ListOpen()
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal memuat lowongan")
		return
	}
	if len(jobs) > 6 {
		jobs = jobs[:6]
	}
	data := a.newPageData(r, "")
	data["Jobs"] = jobs
	a.Render.Render(w, "home.html", data)
}

func (a *App) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := a.newPageData(r, "Masuk")
	a.Render.Render(w, "login.html", data)
}

func (a *App) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	user, err := a.Users.GetByEmail(email)
	if err != nil || !authsvc.CheckPassword(user.PasswordHash, password) {
		data := a.newPageData(r, "Masuk")
		data["Error"] = "Email atau password salah."
		a.Render.Render(w, "login.html", data)
		return
	}

	sessionID, expiresAt, err := a.Auth.CreateSession(user.ID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal membuat sesi login")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     authsvc.SessionCookieName,
		Value:    sessionID,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	switch user.Role {
	case models.RoleCandidate:
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	case models.RoleRecruiter:
		rp, err := a.Users.GetRecruiterProfile(user.ID)
		if err == nil && rp.Status == models.RecruiterApproved {
			http.Redirect(w, r, "/recruiter/dashboard", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/recruiter/pending-approval", http.StatusSeeOther)
		}
	case models.RoleSuperAdmin:
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(authsvc.SessionCookieName); err == nil {
		_ = a.Auth.DestroySession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     authsvc.SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) RegisterChoicePage(w http.ResponseWriter, r *http.Request) {
	data := a.newPageData(r, "Daftar")
	a.Render.Render(w, "register_choose.html", data)
}

func (a *App) RegisterCandidatePage(w http.ResponseWriter, r *http.Request) {
	data := a.newPageData(r, "Daftar Candidate")
	a.Render.Render(w, "register_candidate.html", data)
}

func (a *App) RegisterCandidateSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		a.registerCandidateError(w, r, "Form tidak valid atau file terlalu besar.")
		return
	}

	fullName := strings.TrimSpace(r.FormValue("full_name"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	if fullName == "" || email == "" || len(password) < 8 {
		a.registerCandidateError(w, r, "Lengkapi semua field. Password minimal 8 karakter.")
		return
	}

	if _, err := a.Users.GetByEmail(email); err == nil {
		a.registerCandidateError(w, r, "Email sudah terdaftar.")
		return
	}

	file, header, err := r.FormFile("resume")
	if err != nil {
		a.registerCandidateError(w, r, "Resume PDF wajib diupload.")
		return
	}
	file.Close()

	resumePath, resumeFilename, err := a.saveUpload(header, "resumes")
	if err != nil {
		a.registerCandidateError(w, r, "Gagal mengunggah resume: "+err.Error())
		return
	}

	hash, err := authsvc.HashPassword(password)
	if err != nil {
		a.registerCandidateError(w, r, "Gagal memproses password.")
		return
	}

	userID, err := a.Users.CreateUser(email, hash, models.RoleCandidate)
	if err != nil {
		a.registerCandidateError(w, r, "Gagal membuat akun.")
		return
	}

	if err := a.Users.CreateCandidateProfile(userID, fullName, resumePath, resumeFilename); err != nil {
		a.registerCandidateError(w, r, "Gagal menyimpan profil.")
		return
	}

	sessionID, expiresAt, err := a.Auth.CreateSession(userID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal membuat sesi login")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: authsvc.SessionCookieName, Value: sessionID, Expires: expiresAt,
		Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (a *App) registerCandidateError(w http.ResponseWriter, r *http.Request, msg string) {
	data := a.newPageData(r, "Daftar Candidate")
	data["Error"] = msg
	a.Render.Render(w, "register_candidate.html", data)
}

func (a *App) RegisterRecruiterPage(w http.ResponseWriter, r *http.Request) {
	data := a.newPageData(r, "Daftar Recruiter")
	a.Render.Render(w, "register_recruiter.html", data)
}

func (a *App) RegisterRecruiterSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		a.registerRecruiterError(w, r, "Form tidak valid atau file terlalu besar.")
		return
	}

	companyName := strings.TrimSpace(r.FormValue("company_name"))
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	if companyName == "" || email == "" || len(password) < 8 {
		a.registerRecruiterError(w, r, "Lengkapi semua field. Password minimal 8 karakter.")
		return
	}

	if _, err := a.Users.GetByEmail(email); err == nil {
		a.registerRecruiterError(w, r, "Email sudah terdaftar.")
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		a.registerRecruiterError(w, r, "Dokumen pendukung PDF wajib diupload.")
		return
	}
	file.Close()

	docPath, docFilename, err := a.saveUpload(header, "documents")
	if err != nil {
		a.registerRecruiterError(w, r, "Gagal mengunggah dokumen: "+err.Error())
		return
	}

	hash, err := authsvc.HashPassword(password)
	if err != nil {
		a.registerRecruiterError(w, r, "Gagal memproses password.")
		return
	}

	userID, err := a.Users.CreateUser(email, hash, models.RoleRecruiter)
	if err != nil {
		a.registerRecruiterError(w, r, "Gagal membuat akun.")
		return
	}

	if err := a.Users.CreateRecruiterProfile(userID, companyName, docPath, docFilename); err != nil {
		a.registerRecruiterError(w, r, "Gagal menyimpan profil.")
		return
	}

	sessionID, expiresAt, err := a.Auth.CreateSession(userID)
	if err != nil {
		a.renderError(w, r, http.StatusInternalServerError, "Gagal membuat sesi login")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: authsvc.SessionCookieName, Value: sessionID, Expires: expiresAt,
		Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/recruiter/pending-approval", http.StatusSeeOther)
}

func (a *App) registerRecruiterError(w http.ResponseWriter, r *http.Request, msg string) {
	data := a.newPageData(r, "Daftar Recruiter")
	data["Error"] = msg
	a.Render.Render(w, "register_recruiter.html", data)
}

func (a *App) RecruiterPendingApprovalPage(w http.ResponseWriter, r *http.Request) {
	data := a.newPageData(r, "Menunggu Persetujuan")
	a.Render.Render(w, "pending_approval.html", data)
}

var _ = repository.ErrNotFound
