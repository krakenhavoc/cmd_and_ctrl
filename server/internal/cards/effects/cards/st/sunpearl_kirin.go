package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunpearl Kirin — Creature — Kirin {1}{W}, 2/1 (Tarkir: Dragonstorm,
// #4463):
//
//	"Flash
//	 Flying
//	 When this creature enters, return up to one other target nonland
//	 permanent you control to its owner's hand. If it was a token, draw a
//	 card."
//
// Flash and flying are printed keywords. The enters trigger is Boomerang
// Basics' bounce with the draw as its continuation: "if it was a token" is
// a fact about the target read before it moves (a token that leaves the
// battlefield ceases to exist, CR 704.5d), and the draw waits for the
// return rather than racing a prompt the return might ask. With no target
// chosen, or the target gone by resolution, nothing is returned and
// nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f17b9356-a329-4f89-b690-afc2eaff4977",
		Name:            "Sunpearl Kirin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Sunpearl Kirin — return up to one other target nonland permanent you control to its owner's hand; if it was a token, draw a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						draws := 0
						if c, found := g.LookupCardForEffect(t.ID); found && c.IsToken() {
							draws = 1
						}
						return effects.BounceToHand{Target: t.ID, ThenTail: effects.WithN(ctx.Tail(effects.ThenDrawCards), draws)}.Apply(ctx)
					}),
				effects.Another(effects.TargetPermanent("up to one other target nonland permanent you control", effects.Nonland(), effects.YouControl()).WithCount(0, 1))),
		},
	})
}
