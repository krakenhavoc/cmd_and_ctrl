package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Savage Beating — Instant {3}{R}{R}:
//
//	"Cast this spell only during combat on your turn.
//	 Choose one —
//	 • Creatures you control gain double strike until end of turn.
//	 • Untap all creatures you control. After this phase, there is an
//	   additional combat phase.
//	 Entwine {1}{R} (Choose both if you pay the entwine cost.)"
//
// Entwine is CR 702.42 (#3866). The cast window is Take the Bait's
// combat check with the turn the other way round. The double strike is
// a layer-6 grant over the creatures you control as it resolves (CR
// 611.2c); the second bullet is Aurelia's untap and the turn plan's
// added combat after this one (ADR 0059). Entwined, the double strike
// is granted first (CR 702.42b), so it lasts through the extra combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "89d1cd45-8148-4396-a9b8-e0f495e94b5b",
		Name:         "Savage Beating",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat && activePlayerIDOf(g) == controller
		},
		CastConditionLabel: "Cast this spell only during combat on your turn.",
		OptionalCosts:      []game.AdditionalCost{Entwine("{1}{R}")},
		Modes: Entwining(ChooseOne(
			ModeDoing("Creatures you control gain double strike until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"double strike"},
						Label:    "Savage Beating — double strike",
					}.Apply(ctx)
				}),
			ModeDoing("Untap all creatures you control. After this phase, there is an additional combat phase.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return untapAllYouControlThenExtraCombat(ctx.Game, item)
				}),
		)),
	})
}
