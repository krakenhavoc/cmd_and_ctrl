package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b730_helpers.go — Urza's Saga, batch 730 (#5624, S82).

// OldB730ReturnTheCardToOwnerBody is the delayed half of "return the first
// card to the battlefield under its owner's control at the beginning of the
// next end step" (Lifeline, CR 603.7): the card the trigger was about, if it
// is still in a graveyard as the object that died (CR 400.7), returns under its
// owner's control. A token never reached a graveyard, so nothing returns for
// one. An on-disk key: never renamed, never reused.
var OldB730ReturnTheCardToOwnerBody = game.DelayedBody("lifeline/return-the-card-to-its-owner", oldB730ReturnTheCardToOwner)

func oldB730ReturnTheCardToOwner(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || id == uuid.Nil || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	// The zero Controller is "under its owner's control".
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
}

// OldB730AnotherCreatureIsOnTheBattlefield is Lifeline's intervening "if
// another creature is on the battlefield" (CR 603.4): any creature, whoever
// controls it, read through the layers. The creature that died is already in
// the graveyard, so it is never the one found. The ability checks it as the
// creature dies and again as the trigger resolves.
func OldB730AnotherCreatureIsOnTheBattlefield(g *game.Game) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			return true
		}
	}
	return false
}
