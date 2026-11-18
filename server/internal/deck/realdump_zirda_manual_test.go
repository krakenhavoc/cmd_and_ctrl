package deck

import (
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// realdump_zirda_manual_test.go — Zirda, the Dawnwaker's companion
// condition (companion_zirda.go, #5909) on real Scryfall records, gated
// on CMDCTRL_SCRYFALL_DUMP like the other realdump tests:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  scripts/go-docker.sh go test ./internal/deck -run TestRealDumpZirda
func TestRealDumpZirdaCondition(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	for name, fails := range map[string]bool{
		// A "[Cost]: [Effect]" line, a mana ability among them.
		"Zirda, the Dawnwaker": false,
		"Sol Ring":             false,
		"Mishra's Bauble":      false,
		"Blinkmoth Nexus":      false,
		"Command Tower":        false,
		"Elspeth Tirel":        false,
		// Activated keywords.
		"Sword of Sinew and Steel": false,
		"Kitesail":                 false,
		"Smuggler's Copter":        false,
		"Forgotten Cave":           false,
		"Generous Ent":             false,
		"Abzan Falconer":           false,
		// A basic land, whose text is reminder text (CR 305.6).
		"Forest": false,
		// Front faces only (CR 712.8a): the flip card's unflipped half has
		// one; Delver and Catapult Fodder have one only on the back, and
		// Drowner of Truth only on its land back.
		"Orochi Eggwatcher // Shidako, Broodmistress": false,
		"Delver of Secrets // Insectile Aberration":   true,
		"Catapult Fodder // Catapult Captain":         true,
		"Drowner of Truth // Drowned Jungle":          true,
		// Adventurer: the creature's own text (CR 715.4).
		"Order of Midnight // Alter Fate": true,
		// Granted, triggered and reminder-text colons.
		"Underworld Connections": true,
		"Goldspan Dragon":        true,
		"Urza's Saga":            true,
		"Treasure Vault":         false,
		"Smothering Tithe":       true,
		"Monastery Swiftspear":   true,
		"Lodestone Golem":        true,
		"Ivory Tower":            true,
		// Not a permanent card.
		"Swords to Plowshares": false,
	} {
		c, ok := idx.FindByName(name)
		if !ok {
			t.Errorf("%s: not in the dump", name)
			continue
		}
		if got := permanentWithoutActivatedAbility(c); got != fails {
			t.Errorf("%s (%s): fails = %v, want %v", name, c.Layout, got, fails)
		}
	}
}
