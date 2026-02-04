package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"io"
	"github.com/febrd/maungdb/engine/auth"
	"github.com/febrd/maungdb/engine/executor"
	"github.com/febrd/maungdb/engine/parser"
	"github.com/febrd/maungdb/engine/schema"
	"github.com/febrd/maungdb/engine/storage"
)

// ===========================
// Request & Response Structs
// ===========================

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UseRequest struct {
	Database string `json:"database"`
}

type CreateDBRequest struct {
	Name string `json:"name"`
}

type SchemaRequest struct {
	Table  string   `json:"table"`
	Fields []string `json:"fields"`
	Read   []string `json:"read,omitempty"`
	Write  []string `json:"write,omitempty"`
}

type QueryRequest struct {
	Query string `json:"query"`
}

type APIResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message,omitempty"`
	Data    *executor.ExecutionResult `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
}

type CreateSchemaRequest struct {
	Table  string   `json:"table"`
	Fields []string `json:"fields"` 
}

func startServer(port string) {
	if err := storage.Init(); err != nil {
		panic(err)
	}

	http.HandleFunc("/auth/login", handleLogin)
	http.HandleFunc("/auth/logout", handleLogout)
	http.HandleFunc("/auth/whoami", handleWhoami)

	http.HandleFunc("/db/create", handleCreateDB)
	http.HandleFunc("/db/use", handleUse)

	http.HandleFunc("/schema/create", handleSchemaCreate)
	http.HandleFunc("/query", handleQuery)

	http.HandleFunc("/db/export", handleExport)
	http.HandleFunc("/db/import", handleImport)

	serveWebUI()

	fmt.Println("🐯 MaungDB Server running")
	fmt.Println("🌐 Web UI  : http://localhost:" + port)
	fmt.Println("🔌 API     : http://localhost:" + port + "/query")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("❌ Server error:", err)
	}
}


func setupHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
}

func sendError(w http.ResponseWriter, msg string) {
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   msg,
	})
}

func sendSuccess(w http.ResponseWriter, msg string, data *executor.ExecutionResult) {
	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: msg,
		Data:    data,
	})
}


func handleLogin(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if r.Method != http.MethodPost {
		sendError(w, "Method kudu POST")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "JSON Error")
		return
	}

	if err := auth.Login(req.Username, req.Password); err != nil {
		sendError(w, "Gagal Login: "+err.Error())
		return
	}

	user, _ := auth.CurrentUser()
	sendSuccess(
		w,
		fmt.Sprintf("✅ Login sukses salaku %s (%s)", user.Username, user.Role),
		nil,
	)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if r.Method != http.MethodPost {
		sendError(w, "Method kudu POST")
		return
	}

	if err := auth.Logout(); err != nil {
		sendError(w, err.Error())
		return
	}

	sendSuccess(w, "✅ Logout hasil", nil)
}

func handleWhoami(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)

	user, err := auth.CurrentUser()
	if err != nil {
		sendError(w, err.Error())
		return
	}

	sendSuccess(
		w,
		"OK",
		&executor.ExecutionResult{
			Message: fmt.Sprintf(
				"%s (%s) | db: %s",
				user.Username,
				user.Role,
				user.Database,
			),
		},
	)
}


func handleCreateDB(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if r.Method != http.MethodPost {
		sendError(w, "Method kudu POST")
		return
	}

	if err := auth.RequireRole("supermaung"); err != nil {
		sendError(w, err.Error())
		return
	}

	var req CreateDBRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "JSON Error")
		return
	}

	if err := storage.CreateDatabase(req.Name); err != nil {
		sendError(w, err.Error())
		return
	}

	sendSuccess(w, "✅ Database dijieun", nil)
}

func handleUse(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if r.Method != http.MethodPost {
		sendError(w, "Method kudu POST")
		return
	}

	if _, err := auth.CurrentUser(); err != nil {
		sendError(w, "❌ Anjeun kedah login heula")
		return
	}

	var req UseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "JSON Error")
		return
	}

	if err := auth.SetDatabase(req.Database); err != nil {
		sendError(w, err.Error())
		return
	}

	sendSuccess(w, "✅ Ayeuna ngangge database: "+req.Database, nil)
}

// ===========================
// SCHEMA HANDLER
// ===========================



// 2. Tambahkeun Handler na
func handleSchemaCreate(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if r.Method != http.MethodPost {
		sendError(w, "Method kudu POST")
		return
	}

	// Cek Login & Role
	if err := auth.RequireRole("admin"); err != nil {
		sendError(w, "Akses ditolak: "+err.Error())
		return
	}

	user, _ := auth.CurrentUser()
	if user.Database == "" {
		sendError(w, "Pilih database heula (use)")
		return
	}

	// Decode JSON Body
	var req CreateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Format JSON Salah: "+err.Error())
		return
	}

	// Validasi input
	if req.Table == "" || len(req.Fields) == 0 {
		sendError(w, "Table sareng Fields teu kenging kosong")
		return
	}


	perms := map[string][]string{
		"read":  {"user", "admin", "supermaung"},
		"write": {"admin", "supermaung"},
	}


	if err := schema.Create(user.Database, req.Table, req.Fields, perms); err != nil {
		sendError(w, "Gagal nyieun schema: "+err.Error())
		return
	}

	sendSuccess(w, fmt.Sprintf("✅ Schema tabel '%s' parantos didamel!", req.Table), nil)
}


func handleQuery(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if r.Method != http.MethodPost {
		sendError(w, "Method kudu POST")
		return
	}

	user, err := auth.CurrentUser()
	if err != nil {
		sendError(w, "❌ Anjeun kedah login heula")
		return
	}

	if user.Database == "" {
		sendError(w, "❌ Database can dipilih (POST /db/use)")
		return
	}

	if err := auth.RequireRole("user"); err != nil {
		sendError(w, err.Error())
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "JSON Error")
		return
	}

	cmd, err := parser.Parse(req.Query)
	if err != nil {
		sendError(w, "Syntax Error: "+err.Error())
		return
	}

	result, err := executor.Execute(cmd)
	if err != nil {
		sendError(w, "Execution Error: "+err.Error())
		return
	}

	sendSuccess(w, "Query Berhasil", result)
}

// HANDLER EXPORT (Download CSV)
func handleExport(w http.ResponseWriter, r *http.Request) {
    // CORS Setup (Penting buat browser)
    w.Header().Set("Access-Control-Allow-Origin", "*")
    if r.Method == "OPTIONS" { return }

    // Ambil parameter tabel dari URL ?table=nama_tabel
    table := r.URL.Query().Get("table")
    if table == "" {
        http.Error(w, "Parameter 'table' wajib diisi", http.StatusBadRequest)
        return
    }

    // Panggil fungsi storage
    filePath, err := storage.ExportCSV(table)
    if err != nil {
        http.Error(w, "Gagal export: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Set Header supaya browser download file
    w.Header().Set("Content-Disposition", "attachment; filename="+table+".csv")
    w.Header().Set("Content-Type", "text/csv")
    http.ServeFile(w, r, filePath)
    
    // Opsional: Hapus file CSV temp saenggeus dikirim (bersih-bersih)
    // os.Remove(filePath) 
}

// HANDLER IMPORT (Upload CSV)
func handleImport(w http.ResponseWriter, r *http.Request) {
    setupHeader(w)
    if r.Method != "POST" { return }

    // 1. Parse Multipart Form (Max 10MB)
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        sendError(w, "File terlalu besar")
        return
    }

    // 2. Ambil file dari form
    file, _, err := r.FormFile("csv_file")
    if err != nil {
        sendError(w, "Gagal maca file")
        return
    }
    defer file.Close()

    tableName := r.FormValue("table")
    if tableName == "" {
        sendError(w, "Ngaran tabel kosong")
        return
    }

    // 3. Simpen file upload ka temp
    tempFile, err := os.CreateTemp("", "upload-*.csv")
    if err != nil {
        sendError(w, "Gagal nyieun temp file")
        return
    }
    defer os.Remove(tempFile.Name()) // Hapus mun beres

    // Copy eusi file
    if _, err := io.Copy(tempFile, file); err != nil {
        sendError(w, "Gagal nyalin file")
        return
    }

    // 4. Proses Import ka Storage
    count, err := storage.ImportCSV(tableName, tempFile.Name())
    if err != nil {
        sendError(w, "Gagal import: "+err.Error())
        return
    }

    sendSuccess(w, fmt.Sprintf("✅ Suksés import %d baris data ka tabel '%s'", count, tableName), nil)
}