package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f17b9356-a329-4f89-b690-afc2eaff4977",
		Name:            "Sunpearl Kirin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Sunpearl Kirin — return up to one other target nonland permanent you control to its owner's hand; if it was a token, draw a card",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						t, ok := firstLegalTarget(ctx)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						draws := 0
						if c, found := g.LookupCardForEffect(t.ID); found && c.IsToken() {
							draws = 1
						}
						return BounceToHand{Target: t.ID, ThenTail: withN(ctx.Tail(thenDrawCards), draws)}.Apply(ctx)
					}),
				Another(TargetPermanent("up to one other target nonland permanent you control", Nonland(), YouControl()).WithCount(0, 1))),
		},
	})
}
