package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flumph — Creature — Jellyfish {1}{W}, 0/4 (EDHREC rank 2919):
//
//	"Defender, flying
//	 Whenever this creature is dealt damage, you and target opponent
//	 each draw a card."
//
// The group-hug wall. Defender and flying ride PrintedKeywords; the
// draw is a targeted trigger on EventDealDamage aimed at the Flumph
// (b17SourceDealtDamageToSelf — combat or not, any source), with
// "target opponent" as its target clause, so the controller picks
// the opponent when the trigger fires. Both draw at resolution, the
// controller first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bc328acc-8521-4581-88c3-99a7cb2c1cdb",
		Name:            "Flumph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b17SourceDealtDamageToSelf(ev, source)
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Flumph — you and target opponent each draw a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return DrawCards{Player: t.ID, N: 1}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
