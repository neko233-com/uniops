package store

import (
	"testing"

	"github.com/neko233/uniops/internal/model"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	return db
}

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)

	user := &model.User{
		Username: "testuser",
		Password: "hashedpassword",
		Role:     "operator",
	}

	if err := db.CreateUser(user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	got, err := db.GetUserByUsername("testuser")
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if got.Username != user.Username {
		t.Errorf("Username = %v, want %v", got.Username, user.Username)
	}
	if got.Role != user.Role {
		t.Errorf("Role = %v, want %v", got.Role, user.Role)
	}
}

func TestCreateAndGetServer(t *testing.T) {
	db := setupTestDB(t)

	server := &model.Server{
		Name:     "Test Server",
		Host:     "192.168.1.100",
		Port:     22,
		Username: "root",
		AuthType: "password",
		AuthData: "secret",
	}

	if err := db.CreateServer(server); err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	got, err := db.GetServer(server.ID)
	if err != nil {
		t.Fatalf("Failed to get server: %v", err)
	}

	if got.Name != server.Name {
		t.Errorf("Name = %v, want %v", got.Name, server.Name)
	}
	if got.Host != server.Host {
		t.Errorf("Host = %v, want %v", got.Host, server.Host)
	}
	if got.Port != server.Port {
		t.Errorf("Port = %d, want %d", got.Port, server.Port)
	}
}

func TestListServers(t *testing.T) {
	db := setupTestDB(t)

	for i := 0; i < 3; i++ {
		server := &model.Server{
			Name: "Server",
			Host: "192.168.1.100",
			Port: 22,
		}
		db.CreateServer(server)
	}

	servers, err := db.GetServers()
	if err != nil {
		t.Fatalf("Failed to list servers: %v", err)
	}

	if len(servers) != 3 {
		t.Errorf("Got %d servers, want 3", len(servers))
	}
}

func TestDeleteServer(t *testing.T) {
	db := setupTestDB(t)

	server := &model.Server{
		Name: "To Delete",
		Host: "192.168.1.100",
		Port: 22,
	}
	db.CreateServer(server)

	if err := db.DeleteServer(server.ID); err != nil {
		t.Fatalf("Failed to delete server: %v", err)
	}

	_, err := db.GetServer(server.ID)
	if err == nil {
		t.Error("Expected error getting deleted server")
	}
}

func TestInitAdmin(t *testing.T) {
	db := setupTestDB(t)

	if err := db.InitAdmin(); err != nil {
		t.Fatalf("Failed to init admin: %v", err)
	}

	admin, err := db.GetUserByUsername("root")
	if err != nil {
		t.Fatalf("Failed to get admin: %v", err)
	}

	if admin.Role != "admin" {
		t.Errorf("Admin role = %v, want admin", admin.Role)
	}
}

func TestInitAdminIdempotent(t *testing.T) {
	db := setupTestDB(t)

	if err := db.InitAdmin(); err != nil {
		t.Fatalf("Failed to init admin: %v", err)
	}

	if err := db.InitAdmin(); err != nil {
		t.Fatalf("Second InitAdmin should not fail: %v", err)
	}
}

func TestCreateAndGetAgent(t *testing.T) {
	db := setupTestDB(t)

	agent := &model.Agent{
		Name:     "Test Agent",
		Type:     "claude",
		Endpoint: "https://api.anthropic.com",
		APIKey:   "test-key",
	}

	if err := db.CreateAgent(agent); err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	got, err := db.GetAgent(agent.ID)
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if got.Name != agent.Name {
		t.Errorf("Name = %v, want %v", got.Name, agent.Name)
	}
	if got.Type != agent.Type {
		t.Errorf("Type = %v, want %v", got.Type, agent.Type)
	}
}

func TestDeleteAgent(t *testing.T) {
	db := setupTestDB(t)

	agent := &model.Agent{
		Name: "To Delete",
		Type: "openai",
	}
	db.CreateAgent(agent)

	if err := db.DeleteAgent(agent.ID); err != nil {
		t.Fatalf("Failed to delete agent: %v", err)
	}

	_, err := db.GetAgent(agent.ID)
	if err == nil {
		t.Error("Expected error getting deleted agent")
	}
}

func TestCreateSSHKey(t *testing.T) {
	db := setupTestDB(t)

	key := &model.SSHKey{
		Name:        "Test Key",
		PublicKey:   "ssh-rsa AAAA...",
		PrivateKey:  "-----BEGIN RSA PRIVATE KEY-----\n...",
		Fingerprint: "SHA256:abc123",
		ServerID:    1,
	}

	if err := db.CreateSSHKey(key); err != nil {
		t.Fatalf("Failed to create SSH key: %v", err)
	}

	got, err := db.GetSSHKey(key.ID)
	if err != nil {
		t.Fatalf("Failed to get SSH key: %v", err)
	}

	if got.Name != key.Name {
		t.Errorf("Name = %v, want %v", got.Name, key.Name)
	}
	if got.Fingerprint != key.Fingerprint {
		t.Errorf("Fingerprint = %v, want %v", got.Fingerprint, key.Fingerprint)
	}
}
