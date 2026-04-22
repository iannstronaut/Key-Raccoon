package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"keyraccoon/internal/config"
	"keyraccoon/internal/database/repositories"
	"keyraccoon/internal/services"
)

var (
	cliInitConfig   = config.Init
	cliInitDatabase = config.InitDatabase
	cliGetDB        = config.GetDB
)

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "keyraccoon",
		Short: "KeyRaccoon CLI",
	}

	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newSetupCommand())
	rootCmd.AddCommand(newCreateUserCommand())
	return rootCmd
}

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print effective runtime config summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Init()
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "environment=%s\nserver=%s:%s\ndatabase=%s:%s/%s\nredis=%s:%s\n",
				cfg.Environment,
				cfg.ServerHost,
				cfg.ServerPort,
				cfg.DBHost,
				cfg.DBPort,
				cfg.DBName,
				cfg.RedisHost,
				cfg.RedisPort,
			)
			return nil
		},
	}
}

func newSetupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Setup superadmin account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, userService, userRepo, err := bootstrapUserDependencies()
			if err != nil {
				return err
			}

			if _, err := userRepo.GetByEmail(cfg.AdminEmail); err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Superadmin already exists")
				return nil
			} else if !errors.Is(err, repositories.ErrUserNotFound) {
				return err
			}

			user, err := userService.CreateUser(cfg.AdminEmail, cfg.AdminPassword, "Super Admin", "superadmin", -1, -1)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Superadmin created successfully\nEmail: %s\nRole: %s\nID: %d\n", user.Email, user.Role, user.ID)
			return nil
		},
	}
}

func newCreateUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-user",
		Short: "Create new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			email, _ := cmd.Flags().GetString("email")
			password, _ := cmd.Flags().GetString("password")
			name, _ := cmd.Flags().GetString("name")
			role, _ := cmd.Flags().GetString("role")
			tokenLimit, _ := cmd.Flags().GetInt64("token-limit")
			creditLimit, _ := cmd.Flags().GetFloat64("credit-limit")

			_, userService, _, err := bootstrapUserDependencies()
			if err != nil {
				return err
			}

			user, err := userService.CreateUser(email, password, name, role, tokenLimit, creditLimit)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "User created successfully\nEmail: %s\nName: %s\nRole: %s\n", user.Email, user.Name, user.Role)
			return nil
		},
	}

	cmd.Flags().String("email", "", "User email")
	cmd.Flags().String("password", "", "User password")
	cmd.Flags().String("name", "", "User name")
	cmd.Flags().String("role", "user", "User role (superadmin/admin/user)")
	cmd.Flags().Int64("token-limit", 0, "User token limit")
	cmd.Flags().Float64("credit-limit", 0, "User credit limit")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func bootstrapUserDependencies() (*config.Config, *services.UserService, *repositories.UserRepository, error) {
	cfg, err := cliInitConfig()
	if err != nil {
		return nil, nil, nil, err
	}

	if cliGetDB() == nil {
		if err := cliInitDatabase(cfg); err != nil {
			return nil, nil, nil, err
		}
	}

	db := cliGetDB()
	if db == nil {
		return nil, nil, nil, errors.New("database is not initialized")
	}

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo)
	return cfg, userService, userRepo, nil
}

func normalizeLine(input string) string {
	return strings.ReplaceAll(input, "\r\n", "\n")
}
