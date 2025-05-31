package handlers

import (
	"database/sql"
	"html/template"
	"net/http"

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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