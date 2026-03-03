package host

import (
	"database/sql"
	"fmt"
	"time"

	"GoT0Emergency/internal/infra/db"
)

type Host struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	IP              string `json:"ip"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	AuthType        string `json:"auth_type"` // "password" or "key"
	Password        string `json:"password"`
	KeyPath         string `json:"key_path"`
	LastConnectedAt string `json:"last_connected_at"`
	CreatedAt       string `json:"created_at"`
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateHost(h *Host) error {
	query := `
		INSERT INTO hosts (name, ip, port, username, auth_type, password, key_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	res, err := db.DB.Exec(query, h.Name, h.IP, h.Port, h.Username, h.AuthType, h.Password, h.KeyPath)
	if err != nil {
		return fmt.Errorf("failed to create host: %w", err)
	}
	id, _ := res.LastInsertId()
	h.ID = id
	return nil
}

func (s *Service) GetHosts() ([]Host, error) {
	query := `SELECT id, name, ip, port, username, auth_type, password, key_path, last_connected_at, created_at FROM hosts`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query hosts: %w", err)
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

func (s *Service) GetHost(id int64) (*Host, error) {
	query := `SELECT id, name, ip, port, username, auth_type, password, key_path, last_connected_at, created_at FROM hosts WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	var h Host
	var lastConn sql.NullTime
	var created sql.NullTime

	err := row.Scan(&h.ID, &h.Name, &h.IP, &h.Port, &h.Username, &h.AuthType, &h.Password, &h.KeyPath, &lastConn, &created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("host not found")
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

func (s *Service) UpdateLastConnected(id int64) error {
	_, err := db.DB.Exec("UPDATE hosts SET last_connected_at = ? WHERE id = ?", time.Now(), id)
	return err
}

func (s *Service) DeleteHost(id int64) error {
	_, err := db.DB.Exec("DELETE FROM hosts WHERE id = ?", id)
	return err
}
