package handler

import (
	"net/http"
	"strings"
)

func (h Handler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	items, err := h.Approvals.List(r.Context(), page, size, r.URL.Query().Get("status"), r.URL.Query().Get("request_type"), actor(r))
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	write(w, http.StatusOK, map[string]any{"items": items, "page": page, "page_size": size})
}

func (h Handler) CreateApproval(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RequestType  string   `json:"request_type"`
		ResourceType string   `json:"resource_type"`
		ResourceID   string   `json:"resource_id"`
		Title        string   `json:"title"`
		Summary      string   `json:"summary"`
		Reviewers    []string `json:"reviewers"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	x, err := h.Approvals.Create(r.Context(), input.RequestType, input.ResourceType, input.ResourceID, input.Title, input.Summary, actor(r), input.Reviewers)
	if err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	h.recordAudit(r, "approval.create", "approval", x.ID, x)
	write(w, http.StatusCreated, x)
}

func (h Handler) GetApproval(w http.ResponseWriter, r *http.Request) {
	x, steps, history, err := h.Approvals.Detail(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, http.StatusNotFound, err)
		return
	}
	write(w, http.StatusOK, map[string]any{"approval": x, "steps": steps, "history": history})
}

func (h Handler) DecideApproval(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Decision string `json:"decision"`
		Comment  string `json:"comment"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.Approvals.Decide(r.Context(), r.PathValue("id"), actor(r), strings.TrimSpace(input.Decision), input.Comment); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h Handler) CancelApproval(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Comment string `json:"comment"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.Approvals.Cancel(r.Context(), r.PathValue("id"), actor(r), input.Comment); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, map[string]string{"status": "cancelled"})
}
