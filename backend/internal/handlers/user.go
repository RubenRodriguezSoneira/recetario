package handlers

import (
	"encoding/json"
	"net/http"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
)

type UserHandler struct {
	users UserStore
}

type ProfileUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
}

func NewUserHandler(users UserStore) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	user, err := h.users.GetUserByID(userID)
	if err != nil {
		log.Info("Profile lookup failed", "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toPublicUser(user))
}

func (h *UserHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())

	userID, ok := appmiddleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	var req ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetUserByID(userID)
	if err != nil {
		log.Info("Profile lookup failed", "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.AvatarURL = req.AvatarURL

	if err := h.users.UpdateUser(user); err != nil {
		log.Error("Failed to update profile", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toPublicUser(user))
}
