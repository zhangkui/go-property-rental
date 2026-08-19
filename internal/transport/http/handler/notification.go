package handler

import (
	"net/http"
	"time"

	"go-property-rental/internal/domain/entity"
)

func (h Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	items, err := h.Notifications.List(r.Context(), actor(r), page, size, r.URL.Query().Get("status"), r.URL.Query().Get("category"))
	if err != nil {
		fail(w, 500, err)
		return
	}
	unread, _ := h.Notifications.UnreadCount(r.Context(), actor(r))
	write(w, 200, map[string]any{"items": items, "unread": unread, "page": page, "page_size": size})
}
func (h Handler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	if err := h.Notifications.MarkRead(r.Context(), actor(r), r.PathValue("id")); err != nil {
		fail(w, 400, err)
		return
	}
	write(w, 200, map[string]string{"status": "read"})
}
func (h Handler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if err := h.Notifications.MarkAllRead(r.Context(), actor(r)); err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]string{"status": "read"})
}
func (h Handler) ListReminderRules(w http.ResponseWriter, r *http.Request) {
	items, err := h.Notifications.Rules(r.Context())
	if err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]any{"items": items})
}
func (h Handler) SaveReminderRule(w http.ResponseWriter, r *http.Request) {
	var input entity.ReminderRule
	if err := decode(r, &input); err != nil {
		fail(w, 400, err)
		return
	}
	if err := h.Notifications.SaveRule(r.Context(), input); err != nil {
		fail(w, 400, err)
		return
	}
	write(w, 200, input)
}
func (h Handler) EnableReminderRule(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Enabled bool `json:"enabled"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, 400, err)
		return
	}
	if err := h.Notifications.EnableRule(r.Context(), r.PathValue("id"), input.Enabled); err != nil {
		fail(w, 400, err)
		return
	}
	write(w, 200, input)
}
func (h Handler) RunReminders(w http.ResponseWriter, r *http.Request) {
	if err := h.Notifications.RunReminders(r.Context(), time.Now().UTC()); err != nil {
		fail(w, 500, err)
		return
	}
	write(w, 200, map[string]string{"status": "completed"})
}
