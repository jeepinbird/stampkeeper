package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/services"
)

type ViewHandler struct {
	db           *sql.DB
	templates    *template.Template
	stampService *services.StampService
	boxService   *services.BoxService
}

func NewViewHandler(db *sql.DB, templates *template.Template) *ViewHandler {
	return &ViewHandler{
		db:           db,
		templates:    templates,
		stampService: services.NewStampService(db),
		boxService:   services.NewBoxService(db),
	}
}

func (h *ViewHandler) GetStampsView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	view := vars["view"]

	stamps, err := h.stampService.GetStamps(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateName := view + "-view.html"
	err = h.templates.ExecuteTemplate(w, templateName, stamps)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		return
	}
}

func (h *ViewHandler) GetStampsMoreView(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	view := vars["view"]

	// Ensure we have an offset parameter for "more" requests
	if r.URL.Query().Get("offset") == "" {
		r.URL.Query().Set("offset", "50") // Default offset for more requests
	}

	stamps, err := h.stampService.GetStamps(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templateName := view + "-more.html"
	err = h.templates.ExecuteTemplate(w, templateName, stamps)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		return
	}
}

func (h *ViewHandler) GetStampDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	stamp, err := h.stampService.GetStampByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Stamp not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = h.templates.ExecuteTemplate(w, "stamp-detail.html", stamp)
	if err != nil {
		fmt.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}

func (h *ViewHandler) GetBoxesView(w http.ResponseWriter, r *http.Request) {
	boxes, err := h.boxService.GetBoxes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.templates.ExecuteTemplate(w, "box-list.html", boxes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}