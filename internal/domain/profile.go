package domain

import (
	"context"
	"time"
)

// ProfileTier represents account quota tiers.
type ProfileTier string

const (
	TierFree    ProfileTier = "free"
	TierPro     ProfileTier = "pro"
	TierProPlus ProfileTier = "pro_plus"
)

// Profile represents an isolated Google account / Chrome user profile.
type Profile struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	AccountEmail string      `json:"account_email,omitempty"`
	Tier         ProfileTier `json:"tier"`
	UserDataDir  string      `json:"user_data_dir"`
	CDPPort      int         `json:"cdp_port"`
	IsRunning    bool        `json:"is_running"`
	CreatedAt    time.Time   `json:"created_at"`
}

// ProfileRepository defines persistence operations for Profiles.
type ProfileRepository interface {
	CreateProfile(ctx context.Context, p *Profile) error
	GetProfile(ctx context.Context, id string) (*Profile, error)
	ListProfiles(ctx context.Context) ([]*Profile, error)
	UpdateProfile(ctx context.Context, p *Profile) error
	DeleteProfile(ctx context.Context, id string) error
	GetNextAvailableCDPPort(ctx context.Context, startPort int) (int, error)
}
