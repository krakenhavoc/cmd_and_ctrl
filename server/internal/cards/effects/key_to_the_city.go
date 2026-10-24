package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Key to the City — Artifact {2} (deck request #3655):
//
//	"{T}, Discard a card: Up to one target creature can't be blocked
//	 this turn.
//	 Whenever this artifact becomes untapped, you may pay {2}. If you
//	 do, draw a card."
//
// The untap trigger watches the engine's per-permanent untap event for
// this artifact (any untap, not only the untap step), and the draw is
// behind the payment as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a242c9bf-dab6-4996-bb3a-b4356f9ed0c9",
		Name:         "Key to the City",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}, Discard a card: Up to one target creature can't be blocked this turn.",
			// #4302: "can't be blocked" ends with the turn.
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0))},
			Cost:    Plus(TapCost(), DiscardACard()),
			Targets: TargetCreature("up to one target creature").WithCount(0, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Key to the City — can't be blocked",
					}.Apply(ctx)
				}
				return nil
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventUntapCard, Self, "Key to the City — you may pay {2} to draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Key to the City — pay {2} to draw a card?",
						Then:     mayPayDrawACard,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
