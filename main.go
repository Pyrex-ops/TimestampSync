package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/charmbracelet/log"
	"github.com/go-co-op/gocron/v2"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"lukechampine.com/blake3"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"upload_time/db/database"
)

//go:embed db/schema.sql
var ddl string
var db *sql.DB

var envUser, envPass string

// Basic Auth Middleware
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != envUser || pass != envPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Warn("Failed login attempt", "username", user, "ip", r.RemoteAddr)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Timestamp struct to hold data
type Timestamp struct {
	Name    string `json:"name"`
	Seconds int    `json:"seconds"`
}

type Message struct {
	Message string `json:"message"`
}

// Get all timestamps
func getAllTimestamps(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	queries := database.New(db)
	timestamps, err := queries.GetAllTimestamps(ctx)
	if err != nil {
		log.Error("Error in getting all timestamps", "error", err.Error(), "url", r.URL)
		return
	}
	log.Info("Requested all timestamps", "url", r.URL)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(timestamps)
	if err != nil {
		log.Error("Error in encoding all timestamps", "error", err.Error(), "url", r.URL)
		return
	}
}

func getTimestampByName(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	queries := database.New(db)
	h := blake3.New(8, nil)
	_, _ = h.Write([]byte(strings.ReplaceAll(r.PathValue("name"), "+", " ")))
	row, err := queries.GetTimestampById(ctx, int64(binary.BigEndian.Uint64(h.Sum(nil))))
	if err == nil {
		log.Info("Requested timestamp", "name", strings.ReplaceAll(r.PathValue("name"), "+", " "), "url", r.URL)
		if r.Header.Get("Content-Type") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(row)
			if err != nil {
				log.Error("Error in encoding timestamp", "error", err.Error(), "url", r.URL)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(Message{Message: "Error in encoding timestamp"})
				if err != nil {
					log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
					return
				}
				return
			}
		} else {
			w.Header().Set("Content-Type", "text/html")
			err = json.NewEncoder(w).Encode(int64(row.Seconds))
			if err != nil {
				log.Error("Error in encoding timestamp", "error", err.Error(), "url", r.URL)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				err = json.NewEncoder(w).Encode(Message{Message: "Error in encoding timestamp"})
				if err != nil {
					log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
					return
				}
				return
			}
		}
	} else {
		log.Info("Timestamp not found", "name", strings.ReplaceAll(r.PathValue("name"), "+", " "), "url", r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		err = json.NewEncoder(w).Encode(Message{Message: "Timestamp not found"})
		if err != nil {
			log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
			return
		}
	}
}

func insertTimestamp(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	queries := database.New(db)
	jd := json.NewDecoder(r.Body)
	var ts Timestamp
	type tempTimestamp struct {
		Name    string  `json:"name"`
		Seconds float32 `json:"seconds"`
	}
	var tts tempTimestamp //needed for float unmarshalling
	err := jd.Decode(&tts)
	ts.Name = tts.Name
	ts.Seconds = int(tts.Seconds)
	if err != nil {
		log.Error("Error in decoding timestamp", "error", err.Error(), "url", r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(Message{Message: "Error in decoding timestamp"})
		if err != nil {
			log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
			return
		}
		return
	}
	h := blake3.New(8, nil)
	_, _ = h.Write([]byte(ts.Name))
	row, err := queries.GetTimestampById(ctx, int64(binary.BigEndian.Uint64(h.Sum(nil))))
	if err == nil {
		log.Info("Timestamp found, updated", "name", ts.Name, "url", r.URL)
		if ts.Seconds < 0 {
			ts.Seconds = 0
		}
		err := queries.UpdateTimestampById(ctx, database.UpdateTimestampByIdParams{Timestamp: time.Now().Round(time.Second), Seconds: int64(ts.Seconds), Name: ts.Name, ID: row.ID})
		if err != nil {
			log.Error("Error in updating timestamp", "error", err.Error(), "url", r.URL)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(Message{Message: "Error in updating timestamp"})
			if err != nil {
				log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
				return
			}
			return
		}
	} else {
		log.Info("Timestamp not found in DB, added", "name", ts.Name, "url", r.URL)
		if ts.Seconds < 0 {
			ts.Seconds = 0
		}
		err = queries.InsertTimestamp(ctx, database.InsertTimestampParams{ID: int64(binary.BigEndian.Uint64(h.Sum(nil))), Name: ts.Name, Seconds: int64(ts.Seconds), Timestamp: time.Now().Round(time.Second)})
		if err != nil {
			log.Error("Error in inserting timestamp", "error", err.Error(), "url", r.URL)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			err = json.NewEncoder(w).Encode(Message{Message: "Error in inserting timestamp"})
			if err != nil {
				log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
				return
			}
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(Message{Message: "Timestamp saved successfully"})
	if err != nil {
		log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
		return
	}
}

func deleteTimestamp(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	queries := database.New(db)
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		log.Error("Error in decoding timestamp", "error", err.Error(), "url", r.URL)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(Message{Message: "Error in decoding timestamp"})
		if err != nil {
			log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
			return
		}
		return
	}
	err = queries.DeleteTimestampById(ctx, id)
	if err != nil {
		log.Error("Error in deleting timestamp", "error", err.Error(), "url", r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(Message{Message: "Error in deleting timestamp"})
		if err != nil {
			log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
			return
		}
		return
	} else {
		log.Info("Deleted timestamp", "id", id, "url", r.URL)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(Message{Message: "Timestamp deleted successfully"})
		if err != nil {
			log.Error("Error in encoding message", "error", err.Error(), "url", r.URL)
			return
		}
		return
	}
}

func cleanup() {
	ctx := context.Background()
	queries := database.New(db)
	err := queries.DeleteOldTimestamps(ctx)
	if err != nil {
		log.Error("Error in deleting old timestamps", "error", err.Error())
	}
	log.Info("Cleanup done")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file:", "err", err.Error())
	}
	log.Info("Scheduling cleaning task...")
	scheduler, err := gocron.NewScheduler()
	defer func() { _ = scheduler.Shutdown() }()
	if err != nil {
		log.Fatal("Unable to create a new scheduler", "error", err.Error())
	}
	GcSchedule := os.Getenv("GC_SCHEDULE")
	if GcSchedule == "" {
		GcSchedule = "0 4 */15 * 1"
		log.Info("No GC_SCHEDULE set, using default", "GcSchedule", GcSchedule)
	}
	if scheduler != nil {
		j, err := scheduler.NewJob(
			gocron.CronJob(GcSchedule, false),
			gocron.NewTask(cleanup),
		)
		if err != nil {
			log.Fatal("Unable to create a new job", "error", err.Error())
		}
		scheduler.Start()
		log.Info("Scheduler started")
		if j != nil {
			run, _ := j.NextRun()
			log.Info("Next run", "run", run)
		}
	}

	log.Info("Starting web server...")
	numbPtr := flag.Int("port", 8080, "Port")
	flag.Parse()

	// Check if port is still at default value
	if *numbPtr == 8080 {
		// Look for PORT environment variable
		portStr := os.Getenv("PORT")
		if portStr != "" {
			// Convert environment variable to integer
			portNum, err := strconv.Atoi(portStr)
			if err != nil {
				log.Warn("Invalid PORT environment variable value, using default port 8080", "invalidPort", portStr)
			} else {
				*numbPtr = portNum
				log.Info("Using port from environment variable", "portNum", portNum)
			}
		}
	}

	ctx := context.Background()

	// Get database path from environment variable
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Fallback to relative path if env var not set
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Unable to determine current working directory", "error", err.Error())
		}
		dbPath = filepath.Join(pwd, "data", "database.db")
	}

	// Create parent directories if they don't exist
	dirPath := filepath.Dir(dbPath)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatal("Unable to create directory", "error", err.Error())
	}

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	// create tables
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		log.Fatal(err.Error())
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err.Error())
			return
		}
	}(db)

	// Load credentials from environment variables
	envUser = os.Getenv("BASIC_AUTH_USER")
	envPass = os.Getenv("BASIC_AUTH_PASS")

	// Default credentials (should be changed)
	defaultUser := "admin"
	defaultPass := "password"
	// Warn if default credentials are being used
	if envUser == "" || envPass == "" || envUser == defaultUser || envPass == defaultPass {
		log.Warn("Using default credentials! Set BASIC_AUTH_USER and BASIC_AUTH_PASS environment variables.")
		envUser = defaultUser
		envPass = defaultPass
	}
	mux := http.NewServeMux()
	// Routes
	mux.HandleFunc("GET /timestamps", basicAuth(getAllTimestamps))          // GET all
	mux.HandleFunc("GET /timestamps/{name}", basicAuth(getTimestampByName)) // GET by name
	mux.HandleFunc("POST /timestamps", basicAuth(insertTimestamp))          // POST (Insert)
	mux.HandleFunc("DELETE /timestamps/{id}", basicAuth(deleteTimestamp))   // DELETE

	log.Infof("%s %d", "Listening on port", *numbPtr)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *numbPtr), mux))

}
