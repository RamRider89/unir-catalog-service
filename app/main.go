package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// Definiciones de estructuras para manejar los datos JSON
type Author struct {
	ID        int    `json:"id"`
	FirstName string `json:"nombre"`
	LastName  string `json:"apellido"`
	Biography string `json:"biografia"`
}

type Book struct {
	ID             int    `json:"id"`
	Titulo         string `json:"titulo"`
	ISBN           string `json:"isbn"`
	AnoPublicacion int    `json:"ano_publicacion"`
	AutorID        int    `json:"autor_id"`
}

type CatalogItem struct {
	Book
	AuthorName string `json:"author_name"`
}

// Handler para el endpoint principal
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Catalog Service is up and running!")
}

// Handler para obtener el catálogo completo
func catalogHandler(w http.ResponseWriter, r *http.Request) {
	// Definir URLs de los servicios usando variables de entorno
	authorsServiceURL := os.Getenv("AUTHORS_SERVICE_URL")
	booksServiceURL := os.Getenv("BOOKS_SERVICE_URL")

	if authorsServiceURL == "" || booksServiceURL == "" {
		http.Error(w, "Service URLs not configured. Please set AUTHORS_SERVICE_URL and BOOKS_SERVICE_URL environment variables.", http.StatusInternalServerError)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second} // Cliente HTTP con timeout

	// 1. Obtener todos los autores
	authorsResp, err := client.Get(authorsServiceURL + "/authors")
	if err != nil {
		log.Printf("Error al obtener autores: %v", err)
		http.Error(w, "Error al obtener autores", http.StatusBadGateway)
		return
	}
	defer authorsResp.Body.Close()

	authorsBody, err := ioutil.ReadAll(authorsResp.Body)
	if err != nil {
		log.Printf("Error al leer respuesta de autores: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	var authors []Author
	if err := json.Unmarshal(authorsBody, &authors); err != nil {
		log.Printf("Error al decodificar autores: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	authorsMap := make(map[int]Author)
	for _, author := range authors {
		authorsMap[author.ID] = author
	}

	// 2. Obtener todos los libros
	booksResp, err := client.Get(booksServiceURL + "/books")
	if err != nil {
		log.Printf("Error al obtener libros: %v", err)
		http.Error(w, "Error al obtener libros", http.StatusBadGateway)
		return
	}
	defer booksResp.Body.Close()

	booksBody, err := ioutil.ReadAll(booksResp.Body)
	if err != nil {
		log.Printf("Error al leer respuesta de libros: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	var books []Book
	if err := json.Unmarshal(booksBody, &books); err != nil {
		log.Printf("Error al decodificar libros: %v", err)
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	// 3. Consolidar información
	var catalog []CatalogItem
	for _, book := range books {
		author, ok := authorsMap[book.AutorID]
		authorName := "Desconocido"
		if ok {
			authorName = fmt.Sprintf("%s %s", author.FirstName, author.LastName)
		}
		catalog = append(catalog, CatalogItem{
			Book:       book,
			AuthorName: authorName,
		})
	}

	// Responder con el catálogo consolidado
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Puerto por defecto para el servicio de catálogo
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/catalog", catalogHandler)

	log.Printf("Catalog Service listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}