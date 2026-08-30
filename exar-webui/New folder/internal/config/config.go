package config // sets the Go package name to config

import ( // starts the import list for required packages
	"crypto/rand" // executes this step of the current logic
	"encoding/hex" // executes this step of the current logic
	"fmt" // executes this step of the current logic
	"os" // executes this step of the current logic
	"strings" // executes this step of the current logic
) // ends the current grouped declaration

const ( // declares a constant value
	HostEnv       = "HOST" // updates HostEnv with a new value
	PortEnv       = "PORT" // updates PortEnv with a new value
	ConfigFileEnv = "CONFIG_FILE" // updates ConfigFileEnv with a new value
	PageTitle     = "Expense Entry" // updates PageTitle with a new value
) // ends the current grouped declaration

var DBEnvKeys = []string{ // declares a package-level variable
	"DATABASE_URL", // executes this step of the current logic
	"PGHOST", // executes this step of the current logic
	"PGPORT", // executes this step of the current logic
	"PGDATABASE", // executes this step of the current logic
	"PGUSER", // executes this step of the current logic
	"PGPASSWORD", // executes this step of the current logic
} // closes the current block scope

func EnvOrDefault(key string, fallback string) string { // defines EnvOrDefault, which handles this unit of behavior
	value := os.Getenv(key) // initializes value with the result of this expression
	if value == "" { // checks this condition before continuing
		return fallback // returns the computed values to the caller
	} // closes the current block scope
	return value // returns the computed values to the caller
} // closes the current block scope

func HasDatabaseConfig() bool { // defines HasDatabaseConfig, which handles this unit of behavior
	for _, key := range DBEnvKeys { // iterates through items while this loop condition holds
		if EnvOrDefault(key, "") != "" { // checks this condition before continuing
			return true // returns the computed values to the caller
		} // closes the current block scope
	} // closes the current block scope
	return false // returns the computed values to the caller
} // closes the current block scope

var apiToken string // declares a package-level variable

func APIToken() string { // defines APIToken, which handles this unit of behavior
	return apiToken // returns the computed values to the caller
} // closes the current block scope

func RefreshAPIToken(configPath string) (string, error) { // defines RefreshAPIToken, which handles this unit of behavior
	token, err := generateToken(32) // initializes token, err with the result of this expression
	if err != nil { // checks this condition before continuing
		return "", fmt.Errorf("generate api token: %w", err) // returns the computed values to the caller
	} // closes the current block scope
	content, err := upsertToken(configPath, token) // initializes content, err with the result of this expression
	if err != nil { // checks this condition before continuing
		return "", err // returns the computed values to the caller
	} // closes the current block scope
	if writeErr := os.WriteFile(configPath, []byte(content), 0o600); writeErr != nil { // checks this condition before continuing
		return "", fmt.Errorf("write %s: %w", configPath, writeErr) // returns the computed values to the caller
	} // closes the current block scope
	apiToken = token // updates apiToken with a new value
	return token, nil // returns the computed values to the caller
} // closes the current block scope

func generateToken(byteLen int) (string, error) { // defines generateToken, which handles this unit of behavior
	buf := make([]byte, byteLen) // initializes buf with the result of this expression
	if _, err := rand.Read(buf); err != nil { // checks this condition before continuing
		return "", err // returns the computed values to the caller
	} // closes the current block scope
	return hex.EncodeToString(buf), nil // returns the computed values to the caller
} // closes the current block scope

func upsertToken(path string, token string) (string, error) { // defines upsertToken, which handles this unit of behavior
	raw, err := os.ReadFile(path) // initializes raw, err with the result of this expression
	if err != nil && !os.IsNotExist(err) { // checks this condition before continuing
		return "", fmt.Errorf("read %s: %w", path, err) // returns the computed values to the caller
	} // closes the current block scope
	lines := strings.Split(string(raw), "\n") // initializes lines with the result of this expression
	if len(lines) == 1 && lines[0] == "" { // checks this condition before continuing
		lines = []string{} // updates lines with a new value
	} // closes the current block scope

	tokenLine := `token = "` + token + `"` // initializes tokenLine with the result of this expression
	inAuthSection := false // initializes inAuthSection with the result of this expression
	authSectionFound := false // initializes authSectionFound with the result of this expression
	tokenWritten := false // initializes tokenWritten with the result of this expression
	updated := make([]string, 0, len(lines)+3) // initializes updated with the result of this expression

	for _, line := range lines { // iterates through items while this loop condition holds
		trimmed := strings.TrimSpace(line) // initializes trimmed with the result of this expression
		isSection := strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") // initializes isSection with the result of this expression
		if isSection { // checks this condition before continuing
			if inAuthSection && !tokenWritten { // checks this condition before continuing
				updated = append(updated, tokenLine) // updates updated with a new value
				tokenWritten = true // updates tokenWritten with a new value
			} // closes the current block scope
			inAuthSection = trimmed == "[auth]" // executes this step of the current logic
			if inAuthSection { // checks this condition before continuing
				authSectionFound = true // updates authSectionFound with a new value
			} // closes the current block scope
		} // closes the current block scope

		if inAuthSection && strings.HasPrefix(trimmed, "token") && strings.Contains(trimmed, "=") { // checks this condition before continuing
			if !tokenWritten { // checks this condition before continuing
				updated = append(updated, tokenLine) // updates updated with a new value
				tokenWritten = true // updates tokenWritten with a new value
			} // closes the current block scope
			continue // skips to the next loop iteration
		} // closes the current block scope
		updated = append(updated, line) // updates updated with a new value
	} // closes the current block scope

	if inAuthSection && !tokenWritten { // checks this condition before continuing
		updated = append(updated, tokenLine) // updates updated with a new value
		tokenWritten = true // updates tokenWritten with a new value
	} // closes the current block scope
	if !authSectionFound { // checks this condition before continuing
		if len(updated) > 0 && strings.TrimSpace(updated[len(updated)-1]) != "" { // checks this condition before continuing
			updated = append(updated, "") // updates updated with a new value
		} // closes the current block scope
		updated = append(updated, "[auth]", tokenLine) // updates updated with a new value
	} // closes the current block scope

	return strings.Join(updated, "\n"), nil // returns the computed values to the caller
} // closes the current block scope
