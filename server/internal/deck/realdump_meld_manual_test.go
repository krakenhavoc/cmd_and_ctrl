package deck

import (
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// realdump_meld_manual_test.go — ADR 0145 against the real Scryfall
// dump, gated on CMDCTRL_SCRYFALL_DUMP like realdump_manual_test.go:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/deck/ -run RealDumpMeld -v
//
// Every one of the fourteen meld cards (CR 712.5) imports with its
// pair's combined back face, every pair melds, and every back face is
// refused as a deck card.
func TestRealDumpMeldPairs(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	pairs := [][3]string{
		{"Midnight Scavengers", "Graf Rats", "Chittering Host"},
		{"Hanweir Garrison", "Hanweir Battlements", "Hanweir, the Writhing Township"},
		{"Bruna, the Fading Light", "Gisela, the Broken Blade", "Brisela, Voice of Nightmares"},
		{"Phyrexian Dragon Engine", "Mishra, Claimed by Gix", "Mishra, Lost to Phyrexia"},
		{"The Mightstone and Weakstone", "Urza, Lord Protector", "Urza, Planeswalker"},
		{"Argoth, Sanctum of Nature", "Titania, Voice of Gaea", "Titania, Gaea Incarnate"},
		{"Fang, Fearless l'Cie", "Vanille, Cheerful l'Cie", "Ragnarok, Divine Deliverance"},
	}
	for _, p := range pairs {
		var halves []game.Card
		for _, name := range p[:2] {
			c, ok := idx.FindByName(name)
			if !ok {
				t.Fatalf("%s not in the dump", name)
			}
			gc := ToGameCard(c, false)
			if !gc.IsMeldCard() || gc.Meld.Result.Name != p[2] {
				t.Errorf("%s: meld print %+v, want back face %s", name, gc.Meld, p[2])
				continue
			}
			halves = append(halves, gc)
		}
		if len(halves) == 2 && !game.CanMeld(halves[0], halves[1]) {
			t.Errorf("%s and %s do not form a pair", p[0], p[1])
		}
		back, ok := idx.FindByName(p[2])
		if !ok || !back.IsMeldBackFace() {
			t.Errorf("%s is not read as a back face", p[2])
		}
		if vs := meldBackFaceViolations(&List{Mainboard: []cards.Card{back}}); len(vs) != 1 {
			t.Errorf("%s as a deck card: %v", p[2], vs)
		}
	}
}
