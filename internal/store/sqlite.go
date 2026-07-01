package store

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"

	"github.com/neko233/uniops/internal/model"
)

type DB struct {
	*gorm.DB
}

func New(dbPath string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.Server{},
		&model.Agent{},
		&model.Session{},
		&model.Command{},
		&model.SSHKey{},
	)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) InitAdmin() error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("root"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &model.User{
		Username: "root",
		Password: string(hash),
		Role:     "admin",
	}
	return db.Create(admin).Error
}

func (db *DB) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (db *DB) CreateUser(user *model.User) error {
	return db.Create(user).Error
}

func (db *DB) GetServers() ([]model.Server, error) {
	var servers []model.Server
	err := db.Find(&servers).Error
	return servers, err
}

func (db *DB) GetServer(id uint) (*model.Server, error) {
	var server model.Server
	err := db.First(&server, id).Error
	return &server, err
}

func (db *DB) CreateServer(server *model.Server) error {
	return db.Create(server).Error
}

func (db *DB) UpdateServer(server *model.Server) error {
	return db.Save(server).Error
}

func (db *DB) DeleteServer(id uint) error {
	return db.Delete(&model.Server{}, id).Error
}

func (db *DB) GetAgents() ([]model.Agent, error) {
	var agents []model.Agent
	err := db.Find(&agents).Error
	return agents, err
}

func (db *DB) CreateAgent(agent *model.Agent) error {
	return db.Create(agent).Error
}

func (db *DB) GetAgent(id uint) (*model.Agent, error) {
	var agent model.Agent
	err := db.First(&agent, id).Error
	return &agent, err
}

func (db *DB) UpdateAgent(agent *model.Agent) error {
	return db.Save(agent).Error
}

func (db *DB) DeleteAgent(id uint) error {
	return db.Delete(&model.Agent{}, id).Error
}

func (db *DB) GetSessions() ([]model.Session, error) {
	var sessions []model.Session
	err := db.Find(&sessions).Error
	return sessions, err
}

func (db *DB) CreateSession(session *model.Session) error {
	return db.Create(session).Error
}

func (db *DB) GetSessionByID(id uint) (*model.Session, error) {
	var session model.Session
	err := db.First(&session, id).Error
	return &session, err
}

func (db *DB) UpdateSession(session *model.Session) error {
	return db.Save(session).Error
}

func (db *DB) CreateCommand(cmd *model.Command) error {
	return db.Create(cmd).Error
}

func (db *DB) GetCommandsBySession(sessionID uint) ([]model.Command, error) {
	var commands []model.Command
	err := db.Where("session_id = ?", sessionID).Find(&commands).Error
	return commands, err
}

func (db *DB) GetSSHKeysByServer(serverID uint) ([]model.SSHKey, error) {
	var keys []model.SSHKey
	err := db.Where("server_id = ?", serverID).Find(&keys).Error
	return keys, err
}

func (db *DB) CreateSSHKey(key *model.SSHKey) error {
	return db.Create(key).Error
}

func (db *DB) GetSSHKeyByFingerprint(fingerprint string) (*model.SSHKey, error) {
	var key model.SSHKey
	err := db.Where("fingerprint = ?", fingerprint).First(&key).Error
	return &key, err
}

func (db *DB) UpdateSSHKey(key *model.SSHKey) error {
	return db.Save(key).Error
}

func (db *DB) GetSSHKey(id uint) (*model.SSHKey, error) {
	var key model.SSHKey
	err := db.First(&key, id).Error
	return &key, err
}
