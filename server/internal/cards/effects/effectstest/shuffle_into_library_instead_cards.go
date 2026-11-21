package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shuffle_into_library_instead_cards_test.go — #2038's replacement
// siblings: Darksteel Colossus, Progenitus and Legacy Weapon print
// Blightsteel Colossus's "If ~ would be put into a graveyard from
// anywhere, reveal ~ and shuffle it into its owner's library instead"
// (ShuffleIntoOwnersLibraryInstead). The tests prove what the cards add
// beside it: a competing Rest in Peace is the affected card's
// controller's choice (CR 616.1), and a commander sent to the library
// is still offered the command zone (CR 903.9b, 616.2).

const (
	DarksteelColossusOracle = "1b09d0cf-403c-4a15-aeee-602a1bdaf0c1"
	ProgenitusOracle        = "2d7e00b6-12f0-4b03-82a6-50e3d1b5395e"
	LegacyWeaponOracle      = "aa1a248d-2f76-4f15-b065-08f299af07b9"
)

func DarksteelColossusCard(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Darksteel Colossus", TypeLine: "Artifact Creature — Golem",
		OracleID: DarksteelColossusOracle, ManaCost: "{11}", Power: 11, Toughness: 11, Owner: owner, Controller: owner}
}

// sacrificeNow sacrifices a permanent outside any resolution.
func SacrificeNow(t *testing.T, g *game.Game, controller, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (effects.SacrificePermanent{Target: id}).Apply(effects.NewContext(g, &game.StackItem{Controller: controller})); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
}
