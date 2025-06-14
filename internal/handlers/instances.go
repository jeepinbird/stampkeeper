package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jeepinbird/stampkeeper/internal/models"
	"github.com/jeepinbird/stampkeeper/internal/services"
)

type InstanceHandler struct {
	db        *sql.DB
	templates *template.Template
	service   *services.InstanceService
}

func NewInstanceHandler(db *sql.DB, templates *template.Template) *InstanceHandler {
	return &InstanceHandler{
		db:        db,
		templates: templates,
		service:   services.NewInstanceService(db),
	}
}

func (h *InstanceHandler) CreateStampInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stampID := vars["stamp_id"]

	logPrefix := "handlers.instances.CreateStampInstance:"

	var instance models.StampInstance
	if err := json.NewDecoder(r.Body).Decode(&instance); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set required fields
	instance.ID = uuid.New().String()
	instance.StampID = stampID
	instance.DateAdded = time.Now()
	instance.DateModified = time.Now()
	
	if instance.Quantity <= 0 {
		instance.Quantity = 1
	}

	log.Printf("%s Creating Stamp Instance: %+v", logPrefix, instance)

	_, err := h.service.CreateStampInstance(&instance)
	if err != nil {
		log.Printf("%s Error creating stamp instance: %v", logPrefix, err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "An instance with this condition and box already exists", http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// After creating, fetch the full instance data to get BoxName etc.
	fullInstance, err := h.service.GetStampInstance(instance.ID)
	if err != nil {
		// This is not ideal, but we return the created ID anyway.
		// The client may have to do a refresh in this edge case.
		log.Printf("%s CRITICAL: Instance %s was created but could not be fetched for response: %v", logPrefix, instance.ID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(instance)
		return
	}

	log.SetPrefix("")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fullInstance) // Encode the full object with BoxName
}

func (h *InstanceHandler) UpdateStampInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]

	// Get the existing instance
	existingInstance, err := h.service.GetStampInstance(instanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Instance not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Parse updates
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Apply updates
	if condition, ok := updates["condition"]; ok {
		if condition == nil || condition == "" {
			existingInstance.Condition = nil
		} else if conditionStr, ok := condition.(string); ok {
			existingInstance.Condition = &conditionStr
		}
	}

	if boxID, ok := updates["box_id"]; ok {
		if boxID == nil || boxID == "" {
			existingInstance.BoxID = nil
		} else if boxIDStr, ok := boxID.(string); ok {
			existingInstance.BoxID = &boxIDStr
		}
	}

	if quantity, ok := updates["quantity"]; ok {
		if quantityFloat, ok := quantity.(float64); ok {
			existingInstance.Quantity = int(quantityFloat)
		}
	}

	existingInstance.DateModified = time.Now()

	// If quantity is 0, delete the instance
	if existingInstance.Quantity == 0 {
		if err := h.service.DeleteStampInstance(instanceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	updatedInstance, err := h.service.UpdateStampInstance(existingInstance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedInstance)
}

func (h *InstanceHandler) DeleteStampInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]

	if err := h.service.DeleteStampInstance(instanceID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *InstanceHandler) GetStampInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instance_id"]

	instance, err := h.service.GetStampInstance(instanceID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Instance not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}