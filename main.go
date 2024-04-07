//

package main

import (
	"os"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	apiEndpoint = "https://api.openai.com/v1/chat/completions"
)

type BodyOfStocks struct {
	Movements []Transaction `json:"movements"`
}

func (b BodyOfStocks) StringES() string {
	return fmt.Sprintf("Movimientos de acciones: %v", b.Movements)

}

type Transaction struct {
	StockName   string  `json:"stockName"`
	StockPrice  float64 `json:"stockPrice"`
	StockAction string  `json:"stockAction"` // Buy or Sell
}

func getProfileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Get body
		decoder := json.NewDecoder(r.Body)
		var movements BodyOfStocks

		err := decoder.Decode(&movements)
		if err != nil {
			http.Error(w, "Error reading the request body", http.StatusBadRequest)
			return
		}

		movementsEs := movements.StringES()
		// Use your API KEY here
		apiKey := os.Getenv("API_KEY")
		fmt.Println(apiKey)
		client := resty.New()

		promptFormatted := fmt.Sprintf("%s%s%s", `Conservador Moderado Agresivo Teniendo en cuenta estos 3 perfiles. si compro en:`, movementsEs, `, que tipo de perfil tendria?

	Respondeme unicamente con el tipo de inversor, con una sola palabra, en español.`)
		response, err := client.R().
			SetAuthToken(apiKey).
			SetHeader("Content-Type", "application/json").
			SetBody(map[string]interface{}{
				"model": "gpt-3.5-turbo",
				"messages": []interface{}{map[string]interface{}{"role": "system",
					"content": promptFormatted}},
				"max_tokens": 50,
			}).
			Post(apiEndpoint)

		if err != nil {
			log.Fatalf("Error while sending send the request: %v", err)
		}

		body := response.Body()

		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("Error while decoding JSON response:", err)
			return
		}

		// Extract the content from the JSON response
		content := data["choices"].([]interface{})[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)

		//fmt.Println([]byte(content))

		// Send the response back to the client
		w.Write([]byte(content))
	})
}

func getEnvHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, c *http.Request) {
			fmt.Fprintln(w, "Hello, ENV!" + os.Getenv("TEST_ENV"))
	})
}

func main() {

	router := http.NewServeMux()

	// API
	router.Handle("POST /profile", getProfileHandler())
	router.Handle("GET /env", getEnvHandler())

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	// Iniciar el servidor HTTP en el puerto 8080
	fmt.Println(http.ListenAndServe("0.0.0.0:" + port, router))
}