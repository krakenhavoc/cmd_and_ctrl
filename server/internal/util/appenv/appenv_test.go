package appenv

import "testing"

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		raw     string
		want    Env
		wantErr bool
	}{
		{"", EnvProd, false},
		{"prod", EnvProd, false},
		{"production", EnvProd, false},
		{"PROD", EnvProd, false},
		{"  dev  ", EnvDev, false},
		{"develop", EnvDev, false},
		{"development", EnvDev, false},
		{"staging", EnvProd, true},
		{"1", EnvProd, true},
	} {
		got, err := Parse(tc.raw)
		if (err != nil) != tc.wantErr {
			t.Errorf("Parse(%q) err = %v, wantErr %v", tc.raw, err, tc.wantErr)
		}
		if got != tc.want {
			t.Errorf("Parse(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

// The load-bearing invariant: prod gets no dev features, no matter
// what the per-feature variables say.
func TestLoadFeaturesProdIgnoresOverrides(t *testing.T) {
	for _, fv := range featureVars {
		t.Setenv(fv.name, "1")
	}
	if got := LoadFeatures(EnvProd); got != (Features{}) {
		t.Fatalf("LoadFeatures(prod) = %+v, want zero value — dev features must be unreachable in production", got)
	}
}

func TestLoadFeaturesDevDefaultsOn(t *testing.T) {
	got := LoadFeatures(EnvDev)
	want := Features{CardSpawn: true, SeatSwap: true, FrameInspector: true, ReplayScrubber: true}
	if got != want {
		t.Fatalf("LoadFeatures(dev) = %+v, want %+v", got, want)
	}
}

func TestLoadFeaturesDevOverrideSubtracts(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_CARD_SPAWN", "0")
	t.Setenv("CMDCTRL_DEV_SEAT_SWAP", "off")
	got := LoadFeatures(EnvDev)
	if got.CardSpawn {
		t.Error("CardSpawn should be off when CMDCTRL_DEV_CARD_SPAWN=0")
	}
	if got.SeatSwap {
		t.Error("SeatSwap should be off when CMDCTRL_DEV_SEAT_SWAP=off")
	}
	if !got.FrameInspector || !got.ReplayScrubber {
		t.Error("unset feature vars should stay on in dev")
	}
}

func TestStrayProdOverrides(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_CARD_SPAWN", "1")
	t.Setenv("CMDCTRL_DEV_SEAT_SWAP", "0")
	got := StrayProdOverrides(EnvProd)
	if len(got) != 1 || got[0] != "CMDCTRL_DEV_CARD_SPAWN" {
		t.Fatalf("StrayProdOverrides(prod) = %v, want [CMDCTRL_DEV_CARD_SPAWN]", got)
	}
	if got := StrayProdOverrides(EnvDev); got != nil {
		t.Fatalf("StrayProdOverrides(dev) = %v, want nil", got)
	}
}
