package domain

import "testing"

func TestNewAsset(t *testing.T) {
	t.Parallel()

	validMapping := NewDefaultRegisterMapping()

	tests := []struct {
		name        string
		id          AssetID
		assetType   AssetType
		assetName   string
		wantErrText string
	}{
		{
			name:        "empty asset id rejected",
			id:          "",
			assetType:   SolarPanelType,
			assetName:   "panel-a",
			wantErrText: "asset id must not be empty",
		},
		{
			name:        "empty asset type rejected",
			id:          "asset-1",
			assetType:   "",
			assetName:   "panel-a",
			wantErrText: "asset type must not be empty",
		},
		{
			name:        "empty asset name rejected",
			id:          "asset-1",
			assetType:   SolarPanelType,
			assetName:   "",
			wantErrText: "asset name must not be empty",
		},
		{
			name:      "valid asset creation accepted",
			id:        "asset-1",
			assetType: SolarPanelType,
			assetName: "panel-a",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			asset, err := NewAsset(tt.id, tt.assetType, tt.assetName, validMapping)
			if tt.wantErrText == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if asset.ID() != tt.id {
					t.Fatalf("expected asset ID %q, got %q", tt.id, asset.ID())
				}

				if asset.Type() != tt.assetType {
					t.Fatalf("expected asset type %q, got %q", tt.assetType, asset.Type())
				}

				if asset.Name() != tt.assetName {
					t.Fatalf("expected asset name %q, got %q", tt.assetName, asset.Name())
				}

				if asset.RegisterMapping() != validMapping {
					t.Fatalf("expected register mapping %+v, got %+v", validMapping, asset.RegisterMapping())
				}

				return
			}

			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErrText)
			}

			if err.Error() != tt.wantErrText {
				t.Fatalf("expected error %q, got %q", tt.wantErrText, err.Error())
			}
		})
	}
}

func TestNewDefaultAsset(t *testing.T) {
	t.Parallel()

	asset := NewDefaultAsset()
	expectedMapping := NewDefaultRegisterMapping()

	if asset.ID() != DefaultAssetID {
		t.Fatalf("expected asset ID %q, got %q", DefaultAssetID, asset.ID())
	}

	if asset.Type() != SolarPanelType {
		t.Fatalf("expected asset type %q, got %q", SolarPanelType, asset.Type())
	}

	if asset.Name() == "" {
		t.Fatal("expected non-empty asset name")
	}

	if asset.RegisterMapping() != expectedMapping {
		t.Fatalf("expected register mapping %+v, got %+v", expectedMapping, asset.RegisterMapping())
	}
}
