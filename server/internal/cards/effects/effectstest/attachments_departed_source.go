package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_departed_source_test.go — #812's card-level proof.
//
// The soak found it on Loxodon Warhammer: its equip was activated and
// the Warhammer was sacrificed to Krark-Clan Ironworks in response.
// CR 608.2 resolves the ability anyway — its target is still a
// creature its controller controls — and CR 301.5c means the attach
// does nothing, because only a permanent on the battlefield can be
// attached. The engine logged an EventEffectError instead, which the
// catalog soak treats as a bug and fails the nightly run on.
//
// The rule lives at one choke point (game.AttachSourceForEffect); no
// card file checks anything. These two tests are the printed card
// going through it.

// countCatalogEvents is the event-log read these tests share.
func CountCatalogEvents(g *game.Game, kind game.EventKind, from int) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == kind {
				n++
			}
		}
	})
	return n
}

// activateEquipWithoutResolving activates the equipment's equip
// ability (always its last declared ability) and leaves it on the
// stack, so the test can act in the response window equipTo closes.
func ActivateEquipWithoutResolving(t *testing.T, g *game.Game, controller, equipment, creature uuid.UUID) {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: OracleOfBattlefieldCard(t, g, equipment)})
	idx := len(abilities) - 1
	if idx < 0 {
		t.Fatal("equipment has no activated ability")
	}
	if err := g.ActivateCatalogAbility(controller, equipment, idx, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
}
