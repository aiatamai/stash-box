package api

import (
	"context"

	"github.com/spf13/viper"
	"github.com/stashapp/stash-box/internal/config"
	"github.com/stashapp/stash-box/internal/models"
)

func (r *mutationResolver) UpdateConfig(ctx context.Context, input models.ConfigUpdateInput) (*models.StashBoxConfig, error) {
	// Update viper configuration
	if input.Title != nil {
		viper.Set("title", *input.Title)
	}
	if input.HostURL != nil {
		viper.Set("host_url", *input.HostURL)
	}
	if input.RequireInvite != nil {
		viper.Set("require_invite", *input.RequireInvite)
	}
	if input.RequireActivation != nil {
		viper.Set("require_activation", *input.RequireActivation)
	}
	if input.ActivationExpiry != nil {
		viper.Set("activation_expiry", *input.ActivationExpiry)
	}
	if input.EmailCooldown != nil {
		viper.Set("email_cooldown", *input.EmailCooldown)
	}
	if input.DefaultUserRoles != nil {
		viper.Set("default_user_roles", input.DefaultUserRoles)
	}
	if input.VotePromotionThreshold != nil {
		viper.Set("vote_promotion_threshold", *input.VotePromotionThreshold)
	}
	if input.VoteApplicationThreshold != nil {
		viper.Set("vote_application_threshold", *input.VoteApplicationThreshold)
	}
	if input.VotingPeriod != nil {
		viper.Set("voting_period", *input.VotingPeriod)
	}
	if input.MinDestructiveVotingPeriod != nil {
		viper.Set("min_destructive_voting_period", *input.MinDestructiveVotingPeriod)
	}
	if input.VoteCronInterval != nil {
		viper.Set("vote_cron_interval", *input.VoteCronInterval)
	}
	if input.GuidelinesURL != nil {
		viper.Set("guidelines_url", *input.GuidelinesURL)
	}
	if input.EditUpdateLimit != nil {
		viper.Set("edit_update_limit", *input.EditUpdateLimit)
	}
	if input.RequireSceneDraft != nil {
		viper.Set("require_scene_draft", *input.RequireSceneDraft)
	}
	if input.RequireTagRole != nil {
		viper.Set("require_tag_role", *input.RequireTagRole)
	}

	// Email settings
	if input.EmailHost != nil {
		viper.Set("email_host", *input.EmailHost)
	}
	if input.EmailPort != nil {
		viper.Set("email_port", *input.EmailPort)
	}
	if input.EmailUser != nil {
		viper.Set("email_user", *input.EmailUser)
	}
	if input.EmailPassword != nil {
		viper.Set("email_password", *input.EmailPassword)
	}
	if input.EmailFrom != nil {
		viper.Set("email_from", *input.EmailFrom)
	}

	// Image settings
	if input.ImageLocation != nil {
		viper.Set("image_location", *input.ImageLocation)
	}
	if input.ImageBackend != nil {
		viper.Set("image_backend", *input.ImageBackend)
	}
	if input.ImageJpegQuality != nil {
		viper.Set("image_jpeg_quality", *input.ImageJpegQuality)
	}
	if input.ImageMaxSize != nil {
		viper.Set("image_max_size", *input.ImageMaxSize)
	}

	// Image resizing settings
	if input.ImageResizingEnabled != nil {
		viper.Set("image_resizing.enabled", *input.ImageResizingEnabled)
	}
	if input.ImageResizingCachePath != nil {
		viper.Set("image_resizing.cache_path", *input.ImageResizingCachePath)
	}
	if input.ImageResizingMinSize != nil {
		viper.Set("image_resizing.min_size", *input.ImageResizingMinSize)
	}

	// S3 settings
	if input.S3Endpoint != nil {
		viper.Set("s3.endpoint", *input.S3Endpoint)
	}
	if input.S3Bucket != nil {
		viper.Set("s3.bucket", *input.S3Bucket)
	}
	if input.S3AccessKey != nil {
		viper.Set("s3.access_key", *input.S3AccessKey)
	}
	if input.S3Secret != nil {
		viper.Set("s3.secret", *input.S3Secret)
	}
	if input.S3MaxDimension != nil {
		viper.Set("s3.max_dimension", *input.S3MaxDimension)
	}

	// Database settings
	if input.PostgresMaxOpenConns != nil {
		viper.Set("postgres.max_open_conns", *input.PostgresMaxOpenConns)
	}
	if input.PostgresMaxIdleConns != nil {
		viper.Set("postgres.max_idle_conns", *input.PostgresMaxIdleConns)
	}
	if input.PostgresConnMaxLifetime != nil {
		viper.Set("postgres.conn_max_lifetime", *input.PostgresConnMaxLifetime)
	}

	// Other settings
	if input.PhashDistance != nil {
		viper.Set("phash_distance", *input.PhashDistance)
	}
	if input.FaviconPath != nil {
		viper.Set("favicon_path", *input.FaviconPath)
	}
	if input.DraftTimeLimit != nil {
		viper.Set("draft_time_limit", *input.DraftTimeLimit)
	}
	if input.ProfilerPort != nil {
		viper.Set("profiler_port", *input.ProfilerPort)
	}
	if input.UserLogFile != nil {
		viper.Set("userLogFile", *input.UserLogFile)
	}
	if input.Csp != nil {
		viper.Set("csp", *input.Csp)
	}

	// Write the configuration to the YAML file
	if err := viper.WriteConfig(); err != nil {
		return nil, err
	}

	// Reload the configuration
	if err := config.Initialize(); err != nil {
		return nil, err
	}

	// Return the updated configuration
	return r.GetConfig(ctx)
}
