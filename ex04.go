package main

import (
	"errors"
	"fmt"
	"main/types"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	jwt "github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey = []byte("very-secret-key")

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

var (
	errBadCredentials = errors.New("email or password is incorrect")
)

var data types.Plase

// func showError(w *http.ResponseWriter) {
// 	// errorMessage := fmt.Sprintf("Invalid 'page' value: %s", pageQuery)
// 	http.Error(*w, "Bad request", http.StatusBadRequest)
// }

func Search(c fiber.Ctx) error {

	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid 'lat' value")
	}
	lon, err := strconv.ParseFloat(c.Query("lon"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid 'lon' value")
	}

	// Assuming data.GetPlacesByLocation is a function that fetches places based on location
	places, _, err := data.GetPlacesByLocation(lat, lon)
	if err != nil {
		fmt.Println("Error while getting places: ", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	jsonStruct := struct {
		Places []types.Plase `json:"places"`
	}{
		Places: places,
	}

	return c.Status(fiber.StatusOK).JSON(jsonStruct)
}

func Login(c fiber.Ctx) error {
	// Create a mock JWT token for example purposes
	payload := jwt.MapClaims{
		"sub": "email",
		"exp": time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		fmt.Println("JWT token signing, ", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.Status(fiber.StatusOK).JSON(LoginResponse{AccessToken: t})
}

func jwtMiddleware(c fiber.Ctx) error {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing JWT. Peper, Please!")
	}

	tokenString := authHeader[len("Bearer "):]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired JWT")
	}

	return c.Next()
}

func main() {
	app := fiber.New()
	// port := 8888
	// mux := http.NewServeMux()

	publicGroup := app.Group("/api")
	publicGroup.Get("/get_token", Login)
	authGroup := app.Group("/api")
	authGroup.Use(jwtMiddleware)
	authGroup.Get("/recommend", Search)

	// authGroup.Use(keyauth.Config{
	// 	SuccessHandler: func(c fiber.Ctx) error {
	// 		return c.Next()
	// 	},
	// 	ErrorHandler: func(c fiber.Ctx, err error) error {
	// 		if err == keyauth.ErrMissingOrMalformedAPIKey {
	// 			return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	// 		}
	// 		return c.Status(fiber.StatusUnauthorized).SendString("Invalid or expired API Key")
	// 	},
	// 	KeyLookup:  "header:" + fiber.HeaderAuthorization,
	// 	AuthScheme: "Bearer",
	// })

	app.Listen("127.0.0.1:8888")
	// mux.HandleFunc("/api/recommend", searchHandler)

	// http.ListenAndServe(":"+strconv.Itoa(port), mux)
}
