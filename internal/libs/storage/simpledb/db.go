package simpledb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anvh2/futures-trading/internal/libs/logger"
	"go.uber.org/zap"
)

type DB interface {
	Save(state any) error
	Load(target any) error
	Backup() error
}

// Storage implements file-based state persistence
type Storage struct {
	logger    *logger.Logger
	stateFile string
	backupDir string
}

// NewStorage creates a new file-based persistence manager
func NewStorage(logger *logger.Logger, stateFile, backupDir string) (*Storage, error) {
	// Create directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &Storage{
		logger:    logger,
		stateFile: stateFile,
		backupDir: backupDir,
	}, nil
}

// Save saves the state to file
func (fp *Storage) Save(state any) error {
	// Create a temporary file first
	tempFile := fp.stateFile + ".tmp"

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Atomic move from temp to actual file
	if err := os.Rename(tempFile, fp.stateFile); err != nil {
		// Clean up temp file on error
		os.Remove(tempFile)
		return fmt.Errorf("failed to move temporary state file: %w", err)
	}

	fp.logger.Debug("State saved successfully", zap.String("file", fp.stateFile))
	return nil
}

// Load loads the state from file
func (fp *Storage) Load(target any) error {
	// Check if file exists
	if _, err := os.Stat(fp.stateFile); os.IsNotExist(err) {
		return fmt.Errorf("state file does not exist: %s", fp.stateFile)
	}

	// Read file
	data, err := os.ReadFile(fp.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal JSON
	var state any
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	fp.logger.Info("State loaded successfully", zap.String("file", fp.stateFile))

	return nil
}

// Backup creates a backup of the current state file
func (fp *Storage) Backup() error {
	// Check if state file exists
	if _, err := os.Stat(fp.stateFile); os.IsNotExist(err) {
		fp.logger.Warn("No state file to backup", zap.String("file", fp.stateFile))
		return nil
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(fp.backupDir, fmt.Sprintf("state_backup_%s.json", timestamp))

	// Read current state file
	data, err := os.ReadFile(fp.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file for backup: %w", err)
	}

	// Write backup file
	if err := os.WriteFile(backupFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	fp.logger.Info("State backup created", zap.String("backup_file", backupFile))

	// Clean up old backups (keep only last 10)
	fp.cleanupOldBackups()

	return nil
}

// cleanupOldBackups removes old backup files, keeping only the most recent ones
func (fp *Storage) cleanupOldBackups() {
	const maxBackups = 10

	// Read backup directory
	entries, err := os.ReadDir(fp.backupDir)
	if err != nil {
		fp.logger.Warn("Failed to read backup directory", zap.Error(err))
		return
	}

	// Filter backup files and sort by modification time
	var backupFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			backupFiles = append(backupFiles, filepath.Join(fp.backupDir, entry.Name()))
		}
	}

	// If we have more than maxBackups, remove the oldest ones
	if len(backupFiles) > maxBackups {
		// Sort by modification time (oldest first)
		type fileInfo struct {
			path    string
			modTime time.Time
		}

		var files []fileInfo
		for _, file := range backupFiles {
			if stat, err := os.Stat(file); err == nil {
				files = append(files, fileInfo{
					path:    file,
					modTime: stat.ModTime(),
				})
			}
		}

		// Sort by modification time
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[i].modTime.After(files[j].modTime) {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// Remove excess files
		toRemove := len(files) - maxBackups
		for i := 0; i < toRemove; i++ {
			if err := os.Remove(files[i].path); err != nil {
				fp.logger.Warn("Failed to remove old backup", zap.String("file", files[i].path), zap.Error(err))
			} else {
				fp.logger.Debug("Removed old backup", zap.String("file", files[i].path))
			}
		}
	}
}
