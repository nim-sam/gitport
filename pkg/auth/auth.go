package auth

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

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

		if logger.Config["public"] != "true" {
			logger.Logger.Warn("Unauthorized user tried to connect", "key", username)
			return false
		}

		logger.Logger.Info("New user connecting", "user", username)

		perms, exists := logger.Config["default_perm"]
		if !exists {
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