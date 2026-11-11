package cards

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// The booster fields (#5437) are additive: a card with none of them set
// marshals exactly as it did before they existed, so nothing that reads
// a Card on the wire sees a change for a card that carries no printing
// data, and a dump record without them still loads.
func TestBoosterFieldsAreAdditiveOnTheWire(t *testing.T) {
	plain := Card{ID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Name: "Sol Ring"}
	b, err := json.Marshal(plain)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"rarity"`, `"booster"`, `"promo_types"`} {
		if strings.Contains(string(b), key) {
			t.Errorf("an unset %s is on the wire: %s", key, b)
		}
	}

	var back Card
	if err := json.Unmarshal([]byte(`{"id":"11111111-1111-4111-8111-111111111111","name":"Sol Ring"}`), &back); err != nil {
		t.Fatal(err)
	}
	if back.Rarity != "" || back.Booster || back.PromoTypes != nil {
		t.Errorf("a record without the fields decoded as %q %v %v", back.Rarity, back.Booster, back.PromoTypes)
	}

	if err := json.Unmarshal([]byte(`{"id":"11111111-1111-4111-8111-111111111111","rarity":"mythic","booster":true,"promo_types":["boosterfun"]}`), &back); err != nil {
		t.Fatal(err)
	}
	if back.Rarity != "mythic" || !back.Booster || len(back.PromoTypes) != 1 || back.PromoTypes[0] != "boosterfun" {
		t.Errorf("the fields did not decode from Scryfall's names: %+v", back)
	}
}
