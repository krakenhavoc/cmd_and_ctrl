package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flood of Tears — Sorcery {4}{U}{U} (EDHREC rank 4492):
//
//	"Return all nonland permanents to their owners' hands. If you
//	 return four or more nontoken permanents you control this way, you
//	 may put a permanent card from your hand onto the battlefield."
//
// A six-mana Evacuation that refunds you a threat. The rider is what
// makes it a build-around rather than a reset button: a deck full of
// cheap nontoken permanents plays this, bounces the table, and comes
// down with the most expensive thing in its hand for free — which in
// Commander is routinely a nine-drop.
//
// Three clauses that each do something specific:
//
//   - "ALL nonland permanents" is symmetrical. Your board goes too;
//     that is the cost.
//   - "FOUR OR MORE NONTOKEN PERMANENTS YOU CONTROL" is counted from
//     what was actually RETURNED, not from what was on the
//     battlefield when the spell resolved. A permanent with a
//     replacement effect that sent it somewhere else did not "return
//     this way" and does not count; a TOKEN never counts, which is
//     why a token deck cannot cheat the rider.
//   - "YOU MAY PUT A PERMANENT CARD" is a real prompt with a real
//     decline, and it happens after the bounce, so the six cards the
//     bounce just handed you are legal picks. The permanent enters
//     from hand, untapped, under your control — it is not cast, so
//     it cannot be countered and nothing that watches for a cast
//     sees it.
//
// The count reads the RETURNED set through BounceAllMatching's Then
// continuation, which is the only place that information survives:
// once the permanents are in hands, "you controlled it" is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d92f2e6-27df-4329-b552-cf905f7616ba",
		Name:         "Flood of Tears",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.Nonland(), ThenTail: ctx.ItemTail(floodOfTearsPut)}.Apply(ctx)
		},
	})
}

// floodOfTearsPut is "if four or more nontoken permanents you controlled
// were returned to hand this way, you may put a permanent card from your
// hand onto the battlefield", keyed (#4253 slice 10a).
var floodOfTearsPut = effects.MassTail("mass/flood-of-tears-put", func(ctx *effects.Context, swept []game.Card, _ int) error {
	controller := ctx.Controller()
	yours := 0
	for _, c := range swept {
		if c.Controller == controller && !effects.IsToken(c) {
			yours++
		}
	}
	if yours < 4 {
		return nil
	}
	return effects.PutFromHandOntoBattlefield{
		Player:   controller,
		Optional: true,
		Label:    "Flood of Tears — you may put a permanent card from your hand onto the battlefield",
	}.Apply(ctx)
})
