package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ulamog, the Ceaseless Hunger — Legendary Creature — Eldrazi {10}, 10/10:
//
//	"When you cast this spell, exile two target permanents.
//	 Indestructible
//	 Whenever Ulamog attacks, defending player exiles the top twenty
//	 cards of their library."
//
// The cast trigger is Ugin, Eye of the Storms's FromStack shape: it
// goes on the stack above the spell (CR 603.2, 601.2i) and resolves
// first, so the two permanents are gone before Ulamog is. Both targets
// are chosen as it goes on the stack ("two target permanents", exactly
// two) and each is re-judged at resolution (CR 608.2b): one that left
// in response costs only its own exile. The trigger is controlled by
// the caster, and exists even if the spell is countered, because the
// cast is what triggered it.
//
// The attack trigger exiles from the library of the DEFENDING player,
// the player (or planeswalker controller) this attack is aimed at, read
// off the attack declaration, not an opponent chosen by the attacker.
// Exiling the top of a library is not milling (CR 701.17a): the engine's
// exile-the-top-N opens no mill window, so Bruvac does not double it,
// and a library with fewer than twenty cards exiles what it has without
// anyone losing for it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0bfa4512-e35a-4c93-b324-80ec659f5a97",
		Name:            "Ulamog, the Ceaseless Hunger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			{
				FromStack: true,
				Watches:   []game.EventKind{game.EventCast},
				AppliesTo: Self,
				Key:       "Ulamog, the Ceaseless Hunger — exile two target permanents",
				Targets:   TargetPermanent("two target permanents").WithCount(2, 2),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Ulamog, the Ceaseless Hunger — exile two target permanents")
					item.Controller, item.Owner = ev.Actor, ev.Actor
					return item
				},
				Effect: ExileTargetCards,
			},
			WheneverThisAttacks("Ulamog, the Ceaseless Hunger — defending player exiles the top twenty cards of their library",
				ulamogDefenderExilesTwenty),
		},
	})
}

// ulamogDefenderExilesTwenty exiles the top twenty cards of the
// defending player's library.
func ulamogDefenderExilesTwenty(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	defender := b17DefendingPlayer(g, ctx.Trigger().Event)
	if defender == uuid.Nil {
		return nil
	}
	return MillToZone{Player: defender, N: 20, To: game.ZoneExile}.Apply(ctx)
}
