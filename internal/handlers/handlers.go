// Package handlers contains JOBHOO's HTTP handlers. Handlers are thin: they
// parse the request, call a repository or the AI provider, and render a
// template. Business rules that matter beyond a single request belong in
// internal/models or a dedicated service, not here.
package handlers

import (
	"github.com/jobhoo/jobhoo/internal/ai"
	"github.com/jobhoo/jobhoo/internal/database"
)

type Handlers struct {
	Render       *Renderer
	Jobs         *database.JobsRepo
	Users        *database.UsersRepo
	Sessions     *database.SessionsRepo
	Companies    *database.CompaniesRepo
	Applications *database.ApplicationsRepo
	SavedJobs    *database.SavedJobsRepo
	Profiles     *database.CandidateProfilesRepo
	AI           ai.Provider
}

func New(
	render *Renderer,
	jobs *database.JobsRepo,
	users *database.UsersRepo,
	sessions *database.SessionsRepo,
	companies *database.CompaniesRepo,
	applications *database.ApplicationsRepo,
	savedJobs *database.SavedJobsRepo,
	profiles *database.CandidateProfilesRepo,
	aiProvider ai.Provider,
) *Handlers {
	return &Handlers{
		Render: render, Jobs: jobs, Users: users, Sessions: sessions,
		Companies: companies, Applications: applications, SavedJobs: savedJobs,
		Profiles: profiles, AI: aiProvider,
	}
}
