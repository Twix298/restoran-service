package app

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"main/internal/db"
	"main/types"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	store        db.Store
	jwtSecretKey []byte
}

// Конструктор для Handler
func NewHandler(store *db.Store) *Handler {
	return &Handler{store: *store,
		jwtSecretKey: []byte("very-secret-key")}
}

// Структура HTTP-запроса на вход в аккаунт
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Структура HTTP-ответа на вход в аккаунт
// В ответе содержится JWT-токен авторизованного пользователя
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func showError(w http.ResponseWriter, message string, code int) {
	http.Error(w, message, code)
}

func showErrorPage(w *http.ResponseWriter, pageQuery string) {
	errorMessage := fmt.Sprintf("Invalid 'page' value: %s", pageQuery)
	http.Error(*w, errorMessage, http.StatusBadRequest)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	payload := jwt.MapClaims{
		"sub": "email",
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(h.jwtSecretKey)
	if err != nil {
		fmt.Println("JWT token signing, ", err)
		w.WriteHeader(http.StatusInternalServerError)
		r.Write(w)
	}

	resp, err := json.Marshal(LoginResponse{AccessToken: t})
	if err != nil {
		fmt.Println("error marshal struct, ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	lat, err := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, err := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err != nil {
		showError(w, "Bad request", http.StatusBadRequest)
		return
	}
	places, _, err := h.store.GetPlacesByLocation(lat, lon)
	if err != nil {
		showError(w, "Error get location", http.StatusBadRequest)
	}
	if err != nil {
		log.Fatal("error while getting places: ", err)
	}

	jsonStruct := struct {
		Places []types.Plase `json:"places"`
	}{
		Places: places,
	}
	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.MarshalIndent(jsonStruct, "", "	")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageQuery := r.URL.Query().Get("page")
	fmt.Println("pageQuery = ", pageQuery)
	if pageQuery != "" {
		page, _ = strconv.Atoi(pageQuery)
	} else {
		showErrorPage(&w, pageQuery)
		return
	}
	if page < 1 {
		showErrorPage(&w, pageQuery)
		return
	}
	limit := 10
	offset := ((page - 1) * limit)
	values, total, err := h.store.GetPlaces(limit, offset)
	if err != nil {
		log.Fatal("error while getting places: ", err)
	}
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	execDir := filepath.Dir(execPath) // Директория, где лежит исполняемый файл
	filePath := filepath.Join(execDir, "..", "template", "index.html")

	fmt.Println(filePath)
	var tpl = template.Must(template.ParseFiles(filePath))
	htmlStruct := struct {
		Places []types.Plase
		Total  int
		Prev   int
		Next   int
		Last   int
	}{
		Places: values,
		Total:  total,
		Prev:   page - 1,
		Next:   page + 1,
		Last:   (total / limit),
	}
	if err := tpl.Execute(w, htmlStruct); err != nil {
		fmt.Println("error code")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) IndexJSONHandler(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageQuery := r.URL.Query().Get("page")
	fmt.Println("pageQuery = ", pageQuery)
	if pageQuery != "" {
		page, _ = strconv.Atoi(pageQuery)
	} else {
		showErrorPage(&w, pageQuery)
		return
	}
	if page < 1 {
		showErrorPage(&w, pageQuery)
		return
	}
	limit := 10
	offset := ((page - 1) * limit)
	values, total, err := h.store.GetPlaces(limit, offset)
	if err != nil {
		log.Fatal("error while getting places: ", err)
	}

	jsonStruct := struct {
		Places []types.Plase `json:"places"`
		Total  int           `json:"total"`
		Prev   int           `json:"prev"`
		Next   int           `json:"next"`
		Last   int           `json:"last"`
	}{
		Places: values,
		Total:  total,
		Prev:   page - 1,
		Next:   page + 1,
		Last:   (total / limit) - 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jsonStruct)
}

func (h *Handler) JwtMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			showError(w, "Missing JWT", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return h.jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			showError(w, "Invalid or expired JWT", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
