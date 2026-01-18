package auth

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/ssh"

	"github.com/nim-sam/gitport/pkg/logger"
)

type User struct {
	Name string `json:"name"`
	Perm string `json:"perm"`
}

var Data map[string]User

func InitUsers() error {
	file, err := os.Open(filepath.Join(logger.WorkDir, logger.Users))
	if err != nil {
		if os.IsNotExist(err) {
			logger.Logger.Warn("File not found, creating empty user data", "file", logger.Users)
			Data = make(map[string]User)
			return nil
		}
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &Data)
	if err != nil {
		return err
	}

	return nil
}

// EnsureHostAdmin checks if any admin users exist, and if not, adds the host's SSH key as admin
func EnsureHostAdmin() error {
	// Check if any admin exists
	hasAdmin := false
	for _, user := range Data {
		if user.Perm == "admin" {
			hasAdmin = true
			break
		}
	}

	if hasAdmin {
		return nil
	}

	// Try to find host's SSH public key
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Check common SSH key locations
	keyPaths := []string{
		filepath.Join(homeDir, ".ssh", "id_ed25519.pub"),
		filepath.Join(homeDir, ".ssh", "id_rsa.pub"),
		filepath.Join(homeDir, ".ssh", "id_ecdsa.pub"),
	}

	var hostKey string
	var keyPath string
	for _, path := range keyPaths {
		if key, err := os.ReadFile(path); err == nil {
			hostKey = string(key)
			keyPath = path
			break
		}
	}

	if hostKey == "" {
		logger.Logger.Warn("No SSH public key found in ~/.ssh/. Please add your public key manually as admin.")
		return nil
	}

	// Add host as admin
	// Normalize key format: "key-type base64-key" (without comment)
	hostKey = strings.TrimSpace(hostKey)
	keyParts := strings.Fields(hostKey)
	if len(keyParts) < 2 {
		logger.Logger.Warn("Invalid SSH public key format")
		return nil
	}
	// Use only key type and base64 key, ignore comment
	normalizedKey := keyParts[0] + " " + keyParts[1]
	
	Data[normalizedKey] = User{
		Name: "host (admin)",
		Perm: "admin",
	}

	// Save to users.json
	file, err := os.Create(filepath.Join(logger.WorkDir, logger.Users))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(Data); err != nil {
		return err
	}

	logger.Logger.Info("Host added as admin", "key_file", keyPath)
	return nil
}

func GetUser(key ssh.PublicKey) string {
	userKey := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())

	user, exist := Data[userKey]
	if !exist {
		return "guest"
	}

	return user.Name
}

func AuthHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	userKey := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())

	user, exist := Data[userKey]
	if !exist {
		username := ctx.User() + "@" + ctx.RemoteAddr().String()

		if !logger.Config.Public {
			logger.Logger.Warn("Unauthorized user tried to connect", "key", username)
			return false
		}

		logger.Logger.Info("New user connecting", "user", username)

		perms := logger.Config.DefaultPerm
		if perms == "" {
			logger.Logger.Error("No default permissions in file", "file", logger.Conf)
			perms = "none"
		}

		Data[userKey] = User{
			Name: username,
			Perm: perms,
		}

		file, err := os.Create(filepath.Join(logger.WorkDir, logger.Users))
		if err != nil {
			logger.Logger.Error("Could not edit users file", "error", err)
			return false
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		err = encoder.Encode(Data)
		if err != nil {
			logger.Logger.Error("Could not edit users file", "error", err)
			return false
		}
	} else {
		logger.Logger.Info("User authenticated", "user", user.Name, "perm", user.Perm)
	}

	return true
}