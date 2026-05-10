package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

// Estrutura do usuário
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	BornDate string `json:"born_date"`
}

var db *sql.DB

func main() {
	var err error

	// 🔴 CONEXÃO COM O BANCO - AJUSTE AQUI!
	connStr := "host=localhost port=5432 user=postgres password=123456 dbname=postgres sslmode=disable"

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Erro ao conectar no banco:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("❌ Erro ao testar conexão:", err)
	}
	fmt.Println("✅ Conectado ao PostgreSQL!")

	// ============ ROTAS DA API ============
	http.HandleFunc("/api/usuarios", listarUsuarios)
	http.HandleFunc("/api/usuarios/", buscarUsuario)

	// ============ ARQUIVOS ESTÁTICOS ============
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// ============ INICIAR SERVIDOR ============
	porta := ":8080"
	fmt.Println("🚀 Servidor rodando em http://localhost" + porta)
	fmt.Println("📌 API de usuários: http://localhost" + porta + "/api/usuarios")
	fmt.Println("   Pressione Ctrl+C para parar.")

	err = http.ListenAndServe(porta, nil)
	if err != nil {
		log.Fatal("Erro ao iniciar o servidor: ", err)
	}
}

// Listar todos os usuários (GET /api/usuarios)
func listarUsuarios(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query("SELECT id, username, email, born_date FROM users ORDER BY id")
	if err != nil {
		http.Error(w, `{"erro":"Erro ao consultar banco"}`, http.StatusInternalServerError)
		log.Println("Erro na consulta:", err)
		return
	}
	defer rows.Close()

	var usuarios []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.BornDate)
		if err != nil {
			log.Println("Erro ao ler registro:", err)
			continue
		}
		usuarios = append(usuarios, u)
	}

	json.NewEncoder(w).Encode(usuarios)
}

// Buscar um usuário por ID (GET /api/usuarios/{id})
func buscarUsuario(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extrair ID da URL (ex: /api/usuarios/1)
	var id string
	fmt.Sscanf(r.URL.Path, "/api/usuarios/%s", &id)

	if id == "" {
		http.Error(w, `{"erro":"ID não informado"}`, http.StatusBadRequest)
		return
	}

	var u User
	err := db.QueryRow(
		"SELECT id, username, email, born_date FROM users WHERE id = $1", id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.BornDate)

	if err == sql.ErrNoRows {
		http.Error(w, `{"erro":"Usuário não encontrado"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"erro":"Erro ao buscar usuário"}`, http.StatusInternalServerError)
		log.Println("Erro ao buscar:", err)
		return
	}

	json.NewEncoder(w).Encode(u)
}
