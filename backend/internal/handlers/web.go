package handlers

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"

	"recipe-app/internal/appmiddleware"
	"recipe-app/internal/logger"
)

type WebHandler struct {
	templates *template.Template
	users     UserStore
}

type PageData struct {
	Title    string
	User     *User
	RecipeID string
}

func NewWebHandler(users UserStore) *WebHandler {
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		// Templates not found, create empty template for tests
		templates = template.New("")
	}

	return &WebHandler{
		templates: templates,
		users:     users,
	}
}

func (h *WebHandler) renderTemplate(w http.ResponseWriter, templateName string, data PageData) {
	// Simple approach: create a new template set each time
	templates := template.Must(template.ParseFiles(
		"web/templates/layout.html",
		"web/templates/header.html",
		"web/templates/footer.html",
		"web/templates/"+templateName,
	))

	err := templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *WebHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "RecipeApp - Discover, Create & Share Recipes",
		User:  h.getUserFromContext(r),
	}

	h.renderTemplate(w, "index.html", data)
}

func (h *WebHandler) HandleRecipes(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "All Recipes - RecipeApp",
		User:  h.getUserFromContext(r),
	}

	h.renderTemplate(w, "recipes.html", data)
}

func (h *WebHandler) HandleNewRecipe(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title: "Create New Recipe - RecipeApp",
		User:  user,
	}

	h.renderTemplate(w, "new-recipe.html", data)
}

func (h *WebHandler) HandleRecipeDetail(w http.ResponseWriter, r *http.Request) {
	recipeID := chi.URLParam(r, "id")
	data := PageData{
		Title:    "Recipe Detail - RecipeApp",
		User:     h.getUserFromContext(r),
		RecipeID: recipeID,
	}

	h.renderTemplate(w, "recipe-detail.html", data)
}

func (h *WebHandler) HandleEditRecipe(w http.ResponseWriter, r *http.Request) {
	user := h.getUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	recipeID := chi.URLParam(r, "id")
	data := PageData{
		Title:    "Edit Recipe - RecipeApp",
		User:     user,
		RecipeID: recipeID,
	}

	h.renderTemplate(w, "edit-recipe.html", data)
}

func (h *WebHandler) getUserFromContext(r *http.Request) *User {
	claims, ok := appmiddleware.GetUserClaims(r.Context())
	if !ok || claims == nil || claims.UserID == "" || h.users == nil {
		return nil
	}

	u, err := h.users.GetUserByID(claims.UserID)
	if err != nil {
		logger.FromContext(r.Context()).Error("Failed to load user from context", "error", err)
		return nil
	}

	return &User{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}
