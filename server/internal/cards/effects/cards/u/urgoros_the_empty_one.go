package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urgoros, the Empty One — Legendary Creature — Specter {4}{B}{B}, 4/3
// (Dominaria, S82 OLD, #5530):
//
//	"Flying
//	 Whenever Urgoros deals combat damage to a player, that player discards a
//	 card at random. If the player can't, you draw a card."
//
// The damaged player is read off the damage event. The random discard runs
// keyed (DiscardRandomKeyedForEffect), and its tail draws for Urgoros's
// controller when no card was discarded: an empty hand, or a discard
// something stopped (CR 701.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "43e7b834-d3a9-4297-bf79-2dab5507bebe",
		Name:            "Urgoros, the Empty One",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Urgoros — that player discards a card at random; if they can't, you draw a card", urgorosDiscardAtRandom),
		},
	})
}

func urgorosDiscardAtRandom(g *game.Game, item *game.StackItem) error {
	victim := item.Trigger.Event.Target
	if g.PlayerByIDForEffect(victim) == nil {
		return nil
	}
	return g.DiscardRandomKeyedForEffect(victim, 1, effects.NewContext(g, item).ItemTail(urgorosDrawIfNoDiscard))
}

// urgorosDrawIfNoDiscard draws a card for Urgoros's controller when the
// random discard landed nothing. An on-disk key.
var urgorosDrawIfNoDiscard = effects.CardTail("urgoros/draw-if-no-discard", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) > 0 {
		return nil
	}
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
