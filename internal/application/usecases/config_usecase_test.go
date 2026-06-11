package usecases

import (
	"testing"

	"github.com/dh85/outfitpicker/internal/domain/entities"
)

func TestConfigUseCase_LoadOrCreate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ConfigUseCase
		wantErr bool
		wantNil bool
	}{
		{
			name: "loads existing config",
			setup: func() *ConfigUseCase {
				config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)
				return NewConfigUseCase(&mockConfigRepo{loadResult: config})
			},
		},
		{
			name: "returns nil when no config exists",
			setup: func() *ConfigUseCase {
				return NewConfigUseCase(&mockConfigRepo{loadResult: nil})
			},
			wantNil: true,
		},
		{
			name: "returns error on load failure",
			setup: func() *ConfigUseCase {
				return NewConfigUseCase(&mockConfigRepo{loadError: assert.AnError})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.setup().LoadOrCreate()
			assertError(t, tt.wantErr, err)
			if !tt.wantErr {
				assertNil(t, tt.wantNil, result)
			}
		})
	}
}

func TestConfigUseCase_Save(t *testing.T) {
	config, _ := entities.NewConfig("/test/path", nil, nil, nil, nil)

	tests := []struct {
		name    string
		setup   func() *ConfigUseCase
		wantErr bool
	}{
		{
			name:  "saves successfully",
			setup: func() *ConfigUseCase { return NewConfigUseCase(&mockConfigRepo{}) },
		},
		{
			name:    "returns error on save failure",
			setup:   func() *ConfigUseCase { return NewConfigUseCase(&mockConfigRepo{saveError: assert.AnError}) },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup().Save(config)
			assertError(t, tt.wantErr, err)
		})
	}
}

func TestConfigUseCase_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ConfigUseCase
		wantErr bool
	}{
		{
			name:  "deletes successfully",
			setup: func() *ConfigUseCase { return NewConfigUseCase(&mockConfigRepo{}) },
		},
		{
			name:    "returns error on delete failure",
			setup:   func() *ConfigUseCase { return NewConfigUseCase(&mockConfigRepo{deleteError: assert.AnError}) },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup().Delete()
			assertError(t, tt.wantErr, err)
		})
	}
}
