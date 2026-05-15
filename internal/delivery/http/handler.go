package httpHandler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"subscriptions/internal/models"
	"subscriptions/internal/repository"
	"time"

	"github.com/gorilla/mux"
)

const (
	LAYOUT              = "01-2006"
	KEY                 = "Content-Type"
	CONTENT_TYPE        = "application/json"
	SERVER_ERROR        = "internal server error"
	DATA_INVALID_FORMAT = "invalid date format. Use MM-YYYY"
	HANDLER_ERROR       = "Handler Error: %v"
)

type SubscriptionHandler struct {
	Repo *repository.SubscriptionRepo
}

func NewSubscriptionHandler(repo *repository.SubscriptionRepo) *SubscriptionHandler {
	return &SubscriptionHandler{Repo: repo}
}

type createInput struct {
	ServiceName string `json:"service_name" example:"Netflix"`
	Price       int    `json:"price" example:"15"`
	UserID      string `json:"user_id" example:"user-uuid"`
	StartDate   string `json:"start_date" example:"05-2026"`
	EndDate     string `json:"end_date" example:"05-2027"`
}

// CreateSubscription creates a new subscription
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body createInput true "Subscription Data"
// @Success 201 {object} models.Subscription
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "failed to create subscription"
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	log.Println("Handler: Processing CreateSubscription request")
	var input createInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf(HANDLER_ERROR, err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if input.Price <= 0 {
		http.Error(w, "subscription price must be greater than zero", http.StatusBadRequest)
		return
	}
	startDate, err := time.Parse(LAYOUT, input.StartDate)
	if err != nil {
		http.Error(w, DATA_INVALID_FORMAT, http.StatusBadRequest)
		return
	}

	sub := models.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		UserID:      input.UserID,
		StartDate:   startDate,
	}

	if input.EndDate != "" {
		end, err := time.Parse(LAYOUT, input.EndDate)
		if err == nil {
			sub.EndDate = &end
		}
	}

	if err := h.Repo.Create(r.Context(), &sub); err != nil {
		log.Printf(HANDLER_ERROR, err)
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}

	log.Printf("Handler: Subscription created for user %s", sub.UserID)
	w.Header().Set(KEY, CONTENT_TYPE)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sub)
}

// GetStats returns spending statistics for a specific service
// @Summary Get statistics
// @Tags subscriptions
// @Param user_id path string true "User ID"
// @Param from path string true "Start Date (MM-YYYY)"
// @Param to path string true "End Date (MM-YYYY)"
// @Param sub_name path string true "Service Name"
// @Success 200 {object} map[string]int
// @Failure 400 {string} string "invalid date format"
// @Router /subscriptions/stats/{user_id}/{from}/{to}/{sub_name} [get]
func (h *SubscriptionHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uID := vars["user_id"]
	from := vars["from"]
	to := vars["to"]
	sName := vars["sub_name"]

	log.Printf("Handler: Fetching stats for user %s, service %s", uID, sName)

	startDate, err := time.Parse(LAYOUT, from)
	if err != nil {
		http.Error(w, DATA_INVALID_FORMAT, http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse(LAYOUT, to)
	if err != nil {
		http.Error(w, DATA_INVALID_FORMAT, http.StatusBadRequest)
		return
	}

	stats, err := h.Repo.GetTotalStats(r.Context(), uID, sName, startDate, endDate)
	if err != nil {
		log.Printf(HANDLER_ERROR, err)
		http.Error(w, "failed to get statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set(KEY, CONTENT_TYPE)
	json.NewEncoder(w).Encode(map[string]int{sName: stats})
}

// ReadSubByID returns a single subscription by ID
// @Summary Get subscription by ID
// @Tags subscriptions
// @Param user_id path string true "User ID"
// @Param sub_id path string true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 404 {string} string "subscription not found"
// @Router /subscriptions/by-id/{user_id}/{sub_id} [get]
func (h *SubscriptionHandler) ReadSubByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uID := vars["user_id"]
	sID := vars["sub_id"]

	log.Printf("Handler: Fetching subscription %s", sID)

	sub, err := h.Repo.GetSubByID(r.Context(), uID, sID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		log.Printf(HANDLER_ERROR, err)
		http.Error(w, SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	w.Header().Set(KEY, CONTENT_TYPE)
	json.NewEncoder(w).Encode(sub)
}

// GetSubList returns all subscriptions for a user
// @Summary List subscriptions
// @Tags subscriptions
// @Param user_id path string true "User ID"
// @Success 200 {array} models.Subscription
// @Router /subscriptions/list/{user_id} [get]
func (h *SubscriptionHandler) GetSubList(w http.ResponseWriter, r *http.Request) {
	uID := mux.Vars(r)["user_id"]
	log.Printf("Handler: Fetching subscription list for user %s", uID)

	subs, err := h.Repo.GetSubList(r.Context(), uID)
	if err != nil {
		log.Printf(HANDLER_ERROR, err)
		http.Error(w, SERVER_ERROR, http.StatusInternalServerError)
		return
	}

	w.Header().Set(KEY, CONTENT_TYPE)
	json.NewEncoder(w).Encode(subs)
}

// DeleteSubByID removes a subscription record
// @Summary Delete subscription
// @Tags subscriptions
// @Param user_id path string true "User ID"
// @Param sub_id path string true "Subscription ID"
// @Success 200 {string} string "subscription deleted successfully"
// @Router /subscriptions/{user_id}/{sub_id} [delete]
func (h *SubscriptionHandler) DeleteSubByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uID := vars["user_id"]
	sID := vars["sub_id"]

	log.Printf("Handler: Deleting subscription %s", sID)

	err := h.Repo.DeleteSubByID(r.Context(), uID, sID)
	if err != nil {
		log.Printf(HANDLER_ERROR, err)
		http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("subscription deleted successfully"))
}

// UpdateSubByID updates subscription details
// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param sub_id path string true "Subscription ID"
// @Param input body createInput true "Update Data"
// @Success 200 {object} models.Subscription
// @Router /subscriptions/{user_id}/{sub_id} [put]
func (h *SubscriptionHandler) UpdateSubByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uID := vars["user_id"]
	sID := vars["sub_id"]

	log.Printf("Handler: Updating subscription %s", sID)

	var input createInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	startDate, _ := time.Parse(LAYOUT, input.StartDate)
	sub := models.Subscription{
		ServiceName: input.ServiceName,
		Price:       input.Price,
		StartDate:   startDate,
	}

	if input.EndDate != "" {
		end, _ := time.Parse(LAYOUT, input.EndDate)
		sub.EndDate = &end
	}

	if err := h.Repo.UpdateSubByID(r.Context(), uID, sID, &sub); err != nil {
		log.Printf(HANDLER_ERROR, err)
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}
	w.Header().Set(KEY, CONTENT_TYPE)
	json.NewEncoder(w).Encode(sub)
}
