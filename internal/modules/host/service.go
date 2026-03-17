// Package host provides host management functionality
// Handles CRUD operations for remote hosts stored in SQLite database
package host

import (
	"database/sql"
	"time"

	"GoT0Emergency/internal/infra/db"
	"GoT0Emergency/internal/pkg/log"
)

// Host represents a remote host configuration
type Host struct {
	ID              int64  `json:"id"`                // Unique identifier
	Name            string `json:"name"`              // Display name
	IP              string `json:"ip"`                // IP address or hostname
	Port            int    `json:"port"`              // SSH port (usually 22)
	Username        string `json:"username"`          // SSH username
	AuthType        string `json:"auth_type"`         // Authentication type: "password" or "key"
	Password        string `json:"password"`          // SSH password (for password auth)
	KeyPath         string `json:"key_path"`          // Path to private key file (for key auth)
	LastConnectedAt string `json:"last_connected_at"` // Last successful connection time
	CreatedAt       string `json:"created_at"`        // Record creation time
}

// Service provides host management operations
type Service struct{}

// NewService creates a new host service instance
func NewService() *Service {
	return &Service{}
}

// CreateHost creates a new host record in the database
// h: pointer to Host struct with host information
// Returns: error if database operation fails
func (s *Service) CreateHost(h *Host) error {
	query := `
		INSERT INTO hosts (name, ip, port, username, auth_type, password, key_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	res, err := db.DB.Exec(query, h.Name, h.IP, h.Port, h.Username, h.AuthType, h.Password, h.KeyPath)
	if err != nil {
		return log.Errorf("failed to create host: %w", err)
	}
	id, _ := res.LastInsertId()
	h.ID = id
	return nil
}

// GetHosts retrieves all hosts from the database
// Returns: slice of Host structs and error if query fails
func (s *Service) GetHosts() ([]Host, error) {
	query := `SELECT id, name, ip, port, username, auth_type, password, key_path, last_connected_at, created_at FROM hosts`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, log.Errorf("failed to query hosts: %w", err)
	}
	defer rows.Close()

	var hosts []Host
	for rows.Next() {
		var h Host
		var lastConn sql.NullTime
		var created sql.NullTime

		err := rows.Scan(&h.ID, &h.Name, &h.IP, &h.Port, &h.Username, &h.AuthType, &h.Password, &h.KeyPath, &lastConn, &created)
		if err != nil {
			return nil, err
		}
		if lastConn.Valid {
			h.LastConnectedAt = lastConn.Time.Format(time.RFC3339)
		}
		if created.Valid {
			h.CreatedAt = created.Time.Format(time.RFC3339)
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// GetHost retrieves a single host by ID
// id: the host unique identifier
// Returns: pointer to Host struct and error if not found or query fails
func (s *Service) GetHost(id int64) (*Host, error) {
	query := `SELECT id, name, ip, port, username, auth_type, password, key_path, last_connected_at, created_at FROM hosts WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	var h Host
	var lastConn sql.NullTime
	var created sql.NullTime

	err := row.Scan(&h.ID, &h.Name, &h.IP, &h.Port, &h.Username, &h.AuthType, &h.Password, &h.KeyPath, &lastConn, &created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, log.Errorf("host not found")
		}
		return nil, err
	}
	if lastConn.Valid {
		h.LastConnectedAt = lastConn.Time.Format(time.RFC3339)
	}
	if created.Valid {
		h.CreatedAt = created.Time.Format(time.RFC3339)
	}
	return &h, nil
}

// UpdateLastConnected updates the last connection timestamp for a host
// id: the host unique identifier
// Returns: error if database operation fails
func (s *Service) UpdateLastConnected(id int64) error {
	_, err := db.DB.Exec("UPDATE hosts SET last_connected_at = ? WHERE id = ?", time.Now(), id)
	return err
}

// DeleteHost removes a host record from the database
// id: the host unique identifier
// Returns: error if database operation fails
func (s *Service) DeleteHost(id int64) error {
	_, err := db.DB.Exec("DELETE FROM hosts WHERE id = ?", id)
	return err
}

// UpdateHost updates an existing host record in the database
// h: pointer to Host struct with updated information
// Returns: error if database operation fails
func (s *Service) UpdateHost(h *Host) error {
	query := `
		UPDATE hosts
		SET name = ?, ip = ?, port = ?, username = ?, auth_type = ?, password = ?, key_path = ?
		WHERE id = ?
	`
	_, err := db.DB.Exec(query, h.Name, h.IP, h.Port, h.Username, h.AuthType, h.Password, h.KeyPath, h.ID)
	if err != nil {
		return log.Errorf("failed to update host: %w", err)
	}
	return nil
}
