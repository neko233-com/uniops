package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/neko233/uniops/internal/model"
	"github.com/neko233/uniops/internal/store"
)

type Config struct {
	ServiceName  string `json:"service_name"`
	BinaryURL    string `json:"binary_url"`
	AppPort      int    `json:"app_port"`
	Domain       string `json:"domain"`
	NginxPort    int    `json:"nginx_port"`
	HealthPath   string `json:"health_path"`
}

type LogFunc func(string)

type Service struct {
	db *store.DB
}

func NewService(db *store.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Deploy(deploymentID uint, logFn LogFunc) error {
	deployment, err := s.db.GetDeployment(deploymentID)
	if err != nil {
		return fmt.Errorf("deployment not found: %w", err)
	}

	server, err := s.db.GetServer(deployment.ServerID)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	var cfg Config
	if deployment.Config != "" {
		if err := json.Unmarshal([]byte(deployment.Config), &cfg); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "uniops"
	}
	if cfg.AppPort == 0 {
		cfg.AppPort = 6020
	}
	if cfg.NginxPort == 0 {
		cfg.NginxPort = 80
	}
	if cfg.HealthPath == "" {
		cfg.HealthPath = "/api/health"
	}

	// Update status to running
	now := time.Now()
	deployment.Status = "running"
	deployment.UpdatedAt = now
	s.db.UpdateDeployment(deployment)

	// Connect via SSH
	logFn(fmt.Sprintf("[deploy] Connecting to %s:%d...", server.Host, server.Port))

	client, err := s.dialSSH(server)
	if err != nil {
		return s.failDeployment(deployment, logFn, fmt.Sprintf("SSH connection failed: %v", err))
	}
	defer client.Close()
	logFn("[deploy] SSH connected")

	switch deployment.Type {
	case "nginx":
		err = s.deployNginx(client, cfg, logFn)
	case "backend":
		err = s.deployBackend(client, cfg, logFn)
	case "full":
		logFn("[deploy] === Deploying backend ===")
		if err = s.deployBackend(client, cfg, logFn); err != nil {
			break
		}
		logFn("[deploy] === Deploying nginx reverse proxy ===")
		err = s.deployNginx(client, cfg, logFn)
	default:
		err = fmt.Errorf("unknown deployment type: %s", deployment.Type)
	}

	if err != nil {
		return s.failDeployment(deployment, logFn, err.Error())
	}

	// Mark completed
	completedAt := time.Now()
	deployment.Status = "completed"
	deployment.CompletedAt = &completedAt
	deployment.UpdatedAt = completedAt
	s.db.UpdateDeployment(deployment)

	logFn("[deploy] Deployment completed successfully")
	return nil
}

func (s *Service) dialSSH(server *model.Server) (*ssh.Client, error) {
	var auth ssh.AuthMethod
	if server.AuthType == "key" {
		signer, err := ssh.ParsePrivateKey([]byte(server.AuthData))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		auth = ssh.PublicKeys(signer)
	} else {
		auth = ssh.Password(server.AuthData)
	}

	config := &ssh.ClientConfig{
		User:            server.Username,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	return ssh.Dial("tcp", addr, config)
}

func (s *Service) deployBackend(client *ssh.Client, cfg Config, logFn LogFunc) error {
	commands := []struct {
		desc string
		cmd  string
	}{
		{"Create app directory", "mkdir -p /opt/" + cfg.ServiceName},
		{"Download binary", fmt.Sprintf("curl -fsSL -o /opt/%s/%s %s && chmod +x /opt/%s/%s",
			cfg.ServiceName, cfg.ServiceName, cfg.BinaryURL, cfg.ServiceName, cfg.ServiceName)},
		{"Create systemd service", fmt.Sprintf(`cat > /etc/systemd/system/%s.service << 'UNIT'
[Unit]
Description=%s backend service
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/%s
ExecStart=/opt/%s/%s
Restart=always
RestartSec=5
Environment=PORT=%d

[Install]
WantedBy=multi-user.target
UNIT`, cfg.ServiceName, cfg.ServiceName, cfg.ServiceName, cfg.ServiceName, cfg.ServiceName, cfg.AppPort)},
		{"Reload systemd", "systemctl daemon-reload"},
		{"Enable service", fmt.Sprintf("systemctl enable %s", cfg.ServiceName)},
		{"Restart service", fmt.Sprintf("systemctl restart %s", cfg.ServiceName)},
		{"Check status", fmt.Sprintf("systemctl status %s --no-pager -l || true", cfg.ServiceName)},
	}

	for _, step := range commands {
		logFn(fmt.Sprintf("[backend] %s", step.desc))
		output, err := s.runSSH(client, step.cmd)
		if output != "" {
			logFn(output)
		}
		if err != nil {
			return fmt.Errorf("%s failed: %w", step.desc, err)
		}
	}

	// Wait for health check
	if cfg.BinaryURL != "" {
		logFn(fmt.Sprintf("[backend] Waiting for health check on :%d%s ...", cfg.AppPort, cfg.HealthPath))
		for i := 0; i < 10; i++ {
			output, err := s.runSSH(client, fmt.Sprintf("curl -sf http://localhost:%d%s 2>/dev/null || echo 'FAIL'", cfg.AppPort, cfg.HealthPath))
			if err == nil && !strings.Contains(output, "FAIL") {
				logFn("[backend] Health check passed")
				return nil
			}
			time.Sleep(3 * time.Second)
		}
		logFn("[backend] Health check warning: service may still be starting")
	}

	return nil
}

func (s *Service) deployNginx(client *ssh.Client, cfg Config, logFn LogFunc) error {
	domain := cfg.Domain
	if domain == "" {
		domain = "_"
	}

	upstreamPort := cfg.AppPort
	if upstreamPort == 0 {
		upstreamPort = 6020
	}

	commands := []struct {
		desc string
		cmd  string
	}{
		{"Install nginx", "apt-get update -qq && apt-get install -y -qq nginx curl 2>&1 | tail -5"},
		{"Write nginx config", fmt.Sprintf(`cat > /etc/nginx/sites-available/%s << 'NGINX'
upstream %s_backend {
    server 127.0.0.1:%d;
    keepalive 32;
}

server {
    listen %d;
    server_name %s;

    client_max_body_size 100M;

    # WebSocket support
    location /ws/ {
        proxy_pass http://%s_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 86400s;
    }

    # API and static
    location / {
        proxy_pass http://%s_backend;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
    }
}
NGINX`, cfg.ServiceName, cfg.ServiceName, upstreamPort, cfg.NginxPort, domain, cfg.ServiceName, cfg.ServiceName)},
		{"Enable site", fmt.Sprintf("ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/%s && rm -f /etc/nginx/sites-enabled/default", cfg.ServiceName, cfg.ServiceName)},
		{"Test config", "nginx -t 2>&1"},
		{"Restart nginx", "systemctl restart nginx && systemctl enable nginx"},
		{"Nginx status", "systemctl status nginx --no-pager -l || true"},
	}

	for _, step := range commands {
		logFn(fmt.Sprintf("[nginx] %s", step.desc))
		output, err := s.runSSH(client, step.cmd)
		if output != "" {
			logFn(output)
		}
		if err != nil {
			return fmt.Errorf("%s failed: %w", step.desc, err)
		}
	}

	// Verify nginx is responding
	logFn(fmt.Sprintf("[nginx] Verifying response on :%d ...", cfg.NginxPort))
	output, err := s.runSSH(client, fmt.Sprintf("curl -sf -o /dev/null -w '%%{http_code}' http://localhost:%d/ 2>/dev/null || echo '000'", cfg.NginxPort))
	if err == nil {
		logFn(fmt.Sprintf("[nginx] HTTP response: %s", strings.TrimSpace(output)))
	}

	return nil
}

func (s *Service) runSSH(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	output := strings.TrimSpace(stdout.String())
	errOutput := strings.TrimSpace(stderr.String())

	if errOutput != "" {
		if output != "" {
			output += "\n"
		}
		output += errOutput
	}

	if err != nil {
		log.Printf("[ssh] Command failed: %s\nOutput: %s", cmd, output)
		return output, err
	}

	return output, nil
}

// ExecCommand runs an arbitrary command on a server via SSH (for agent use)
func (s *Service) ExecCommand(server *model.Server, command string) (string, error) {
	client, err := s.dialSSH(server)
	if err != nil {
		return "", fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	return s.runSSH(client, command)
}

// ExecCommandStream runs a command and streams output via logFn
func (s *Service) ExecCommandStream(server *model.Server, command string, logFn LogFunc) error {
	client, err := s.dialSSH(server)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return err
	}

	if err := session.Start(command); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				logFn(string(buf[:n]))
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
		}
	}()

	errBuf := make([]byte, 1024)
	for {
		n, err := stderrPipe.Read(errBuf)
		if n > 0 {
			logFn(string(errBuf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
	}

	<-done
	return session.Wait()
}

func (s *Service) failDeployment(deployment *model.Deployment, logFn LogFunc, errMsg string) error {
	logFn(fmt.Sprintf("[deploy] FAILED: %s", errMsg))
	completedAt := time.Now()
	deployment.Status = "failed"
	deployment.CompletedAt = &completedAt
	deployment.UpdatedAt = completedAt
	s.db.UpdateDeployment(deployment)
	return fmt.Errorf("%s", errMsg)
}

// DeployNginxStandalone deploys nginx on a server without creating a deployment record
func (s *Service) DeployNginxStandalone(server *model.Server, cfg Config, logFn LogFunc) error {
	logFn(fmt.Sprintf("[deploy] Connecting to %s:%d...", server.Host, server.Port))
	client, err := s.dialSSH(server)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()
	logFn("[deploy] SSH connected")
	return s.deployNginx(client, cfg, logFn)
}

// DeployBackendStandalone deploys backend on a server without creating a deployment record
func (s *Service) DeployBackendStandalone(server *model.Server, cfg Config, logFn LogFunc) error {
	logFn(fmt.Sprintf("[deploy] Connecting to %s:%d...", server.Host, server.Port))
	client, err := s.dialSSH(server)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()
	logFn("[deploy] SSH connected")
	return s.deployBackend(client, cfg, logFn)
}
