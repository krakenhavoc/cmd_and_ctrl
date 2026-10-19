package cards

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestFlavorNamesResolveWithoutChangingIdentity(t *testing.T) {
	original := Card{ID: uuid.New(), OracleID: uuid.New(), Name: "Loxodon Warhammer", SetType: "expansion", Lang: "en"}
	alternate := original
	alternate.ID = uuid.New()
	alternate.SetType = "box"
	alternate.FlavorName = "Ghal Maraz, the Great Shatterer"
	alternate.CardFaces = []CardFace{{Name: original.Name, FlavorName: "Alternate face title"}}
	for _, method := range []string{"put", "load"} {
		t.Run(method, func(t *testing.T) {
			idx := NewIndex()
			if method == "put" {
				idx.Put(original)
				idx.Put(alternate)
			} else {
				raw, err := json.Marshal([]Card{original, alternate})
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), "cards.json")
				if err := os.WriteFile(path, raw, 0o600); err != nil {
					t.Fatal(err)
				}
				if _, err := idx.Load(path); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"  GHAL MARAZ, the Great Shatterer  ", "Alternate face title", "Alternate face title // Other side"} {
				got, ok := idx.FindByName(name)
				if !ok || got.ID != alternate.ID || got.Name != original.Name || got.OracleID != original.OracleID {
					t.Errorf("FindByName(%q) = %+v, %v", name, got, ok)
				}
			}
			if got, _ := idx.FindByName(original.Name); got.ID != original.ID {
				t.Errorf("canonical name no longer uses its default printing: %s", got.ID)
			}
		})
	}
}

func TestFlavorAliasNeverShadowsCanonicalName(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		idx := NewIndex()
		real := Card{ID: uuid.New(), Name: "Real Card", SetType: "box"}
		alias := Card{ID: uuid.New(), Name: "Other Card", FlavorName: real.Name, SetType: "expansion"}
		if reverse {
			idx.Put(alias)
			idx.Put(real)
		} else {
			idx.Put(real)
			idx.Put(alias)
		}
		if got, _ := idx.FindByName(real.Name); got.ID != real.ID {
			t.Fatal("alias shadowed canonical name")
		}
	}
	idx := NewIndex()
	idx.Put(Card{ID: uuid.New(), Name: "Token", FlavorName: "Token alias", Layout: "token"})
	if _, ok := idx.FindByName("Token alias"); ok {
		t.Fatal("token title became an import alias")
	}
}
