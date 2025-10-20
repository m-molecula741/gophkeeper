package main

import (
	"fmt"
	"os"

	"github.com/m-molecula741/gophkeeper/internal/client"
	"github.com/spf13/cobra"
)

var (
	// Версия и дата сборки (устанавливаются при компиляции через -ldflags)
	version   = "dev"
	buildDate = "unknown"
	
	// Глобальные флаги
	serverURL string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper - secure password manager",
		Long:  "GophKeeper CLI client for managing your passwords and secrets",
	}
	
	// Глобальные флаги
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "Server URL (default: from config or http://localhost:8080)")
	
	// Добавляем команды
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(registerCmd())
	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(secretsCmd())
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// versionCmd команда для показа версии
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper CLI\n")
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Build Date: %s\n", buildDate)
		},
	}
}

// registerCmd команда регистрации
func registerCmd() *cobra.Command {
	var email, password string
	
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := client.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			// Используем serverURL из флага, если указан
			if serverURL != "" {
				cfg.ServerURL = serverURL
			}
			
			apiClient := client.NewAPIClient(cfg.ServerURL, "")
			
			if err := apiClient.Register(email, password); err != nil {
				return fmt.Errorf("registration failed: %w", err)
			}
			
			fmt.Println("✅ Registration successful! You can now login.")
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	
	return cmd
}

// loginCmd команда входа
func loginCmd() *cobra.Command {
	var email, password string
	
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := client.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			// Используем serverURL из флага, если указан
			if serverURL != "" {
				cfg.ServerURL = serverURL
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}
			}
			
			apiClient := client.NewAPIClient(cfg.ServerURL, "")
			
			token, err := apiClient.Login(email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}
			
			// Сохраняем токен
			if err := cfg.SetToken(token); err != nil {
				return fmt.Errorf("failed to save token: %w", err)
			}
			
			fmt.Println("✅ Login successful!")
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&email, "email", "e", "", "Email address (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("password")
	
	return cmd
}

// secretsCmd главная команда для работы с секретами
func secretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage your secrets",
		Long:  "Create, list, view, update and delete your secrets",
	}
	
	cmd.AddCommand(addSecretCmd())
	cmd.AddCommand(listSecretsCmd())
	cmd.AddCommand(getSecretCmd())
	cmd.AddCommand(deleteSecretCmd())
	
	return cmd
}

// addSecretCmd команда добавления секрета
func addSecretCmd() *cobra.Command {
	var secretType, data, metadata string
	
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := client.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			if cfg.Token == "" {
				return fmt.Errorf("not authenticated. Please run 'gophkeeper login' first")
			}
			
			if serverURL != "" {
				cfg.ServerURL = serverURL
			}
			
			apiClient := client.NewAPIClient(cfg.ServerURL, cfg.Token)
			
			secret, err := apiClient.CreateSecret(secretType, data, metadata)
			if err != nil {
				return fmt.Errorf("failed to create secret: %w", err)
			}
			
			fmt.Printf("✅ Secret created successfully!\n")
			fmt.Printf("ID: %s\n", secret.ID)
			fmt.Printf("Type: %s\n", secret.Type)
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&secretType, "type", "t", "text", "Secret type (login, text, binary, card)")
	cmd.Flags().StringVarP(&data, "data", "d", "", "Secret data (required)")
	cmd.Flags().StringVarP(&metadata, "metadata", "m", "{}", "Metadata in JSON format")
	cmd.MarkFlagRequired("data")
	
	return cmd
}

// listSecretsCmd команда вывода списка секретов
func listSecretsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all secrets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := client.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			if cfg.Token == "" {
				return fmt.Errorf("not authenticated. Please run 'gophkeeper login' first")
			}
			
			if serverURL != "" {
				cfg.ServerURL = serverURL
			}
			
			apiClient := client.NewAPIClient(cfg.ServerURL, cfg.Token)
			
			secrets, err := apiClient.ListSecrets()
			if err != nil {
				return fmt.Errorf("failed to list secrets: %w", err)
			}
			
			if len(secrets) == 0 {
				fmt.Println("No secrets found.")
				return nil
			}
			
			fmt.Printf("Found %d secret(s):\n\n", len(secrets))
			for i, secret := range secrets {
				fmt.Printf("%d. ID: %s\n", i+1, secret.ID)
				fmt.Printf("   Type: %s\n", secret.Type)
				fmt.Printf("   Metadata: %s\n", secret.Metadata)
				fmt.Printf("   Created: %s\n", secret.CreatedAt)
				fmt.Println()
			}
			
			return nil
		},
	}
}

// getSecretCmd команда получения секрета
func getSecretCmd() *cobra.Command {
	var secretID string
	
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a secret by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := client.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			if cfg.Token == "" {
				return fmt.Errorf("not authenticated. Please run 'gophkeeper login' first")
			}
			
			if serverURL != "" {
				cfg.ServerURL = serverURL
			}
			
			apiClient := client.NewAPIClient(cfg.ServerURL, cfg.Token)
			
			secret, err := apiClient.GetSecret(secretID)
			if err != nil {
				return fmt.Errorf("failed to get secret: %w", err)
			}
			
			fmt.Printf("Secret Details:\n\n")
			fmt.Printf("ID: %s\n", secret.ID)
			fmt.Printf("Type: %s\n", secret.Type)
			fmt.Printf("Data: %s\n", secret.Data)
			fmt.Printf("Metadata: %s\n", secret.Metadata)
			fmt.Printf("Created: %s\n", secret.CreatedAt)
			fmt.Printf("Updated: %s\n", secret.UpdatedAt)
			
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&secretID, "id", "i", "", "Secret ID (required)")
	cmd.MarkFlagRequired("id")
	
	return cmd
}

// deleteSecretCmd команда удаления секрета
func deleteSecretCmd() *cobra.Command {
	var secretID string
	
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := client.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			
			if cfg.Token == "" {
				return fmt.Errorf("not authenticated. Please run 'gophkeeper login' first")
			}
			
			if serverURL != "" {
				cfg.ServerURL = serverURL
			}
			
			apiClient := client.NewAPIClient(cfg.ServerURL, cfg.Token)
			
			if err := apiClient.DeleteSecret(secretID); err != nil {
				return fmt.Errorf("failed to delete secret: %w", err)
			}
			
			fmt.Println("✅ Secret deleted successfully!")
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&secretID, "id", "i", "", "Secret ID (required)")
	cmd.MarkFlagRequired("id")
	
	return cmd
}

