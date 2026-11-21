package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the spellshapers: mana, {T}, discard a card ----------------------------

// mmqB2Shape activates a spellshaper's only ability with `mana` floated
// (nothing more when empty: a refused activation leaves the pool as it
// was), discarding `discard` (none when uuid.Nil).
func MmqB2Shape(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, discard uuid.UUID, targets []game.TargetRef) error {
	t.Helper()
	if mana != "" {
		DcMana(t, g, me.ID, mana)
	}
	params := game.ActivateAbilityParams{Strict: true, Targets: targets}
	if discard != uuid.Nil {
		params.DiscardIDs = []uuid.UUID{discard}
	}
	return g.ActivateCatalogAbility(me.ID, id, 0, params)
}

// mmqB2Discards puts a card in hand, activates with it as the discard,
// and checks it went to the graveyard and the shaper tapped.
func MmqB2Discards(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, targets []game.TargetRef) {
	t.Helper()
	fodder := HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := MmqB2Shape(t, g, me, id, mana, fodder, targets); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if me.Hand.Contains(fodder) || !me.Graveyard.Contains(fodder) {
		t.Fatal("the card was not discarded as the cost")
	}
	if !B16Tapped(t, g, id) {
		t.Error("the spellshaper did not tap")
	}
	PassPriorityAroundTable(t, g)
}
