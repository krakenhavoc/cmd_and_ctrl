package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Council of Reeds — Legendary Creature — Human Scientist Hero {2}{U}, 2/2
// (Marvel Super Heroes Commander, #4707):
//
//	"The "legend rule" doesn't apply to creatures you control.
//	 At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, create a token that's a copy of Council of Reeds."
//
// The first line is a legend-rule exemption scoped to the controller's
// CREATURES (LegendRuleExemptYourCreatures, CR 704.5j; ADR 0032's 2026-10-30
// amendment): two legendary creatures with the same name stay, while two
// legendary noncreature permanents they control, and an opponent's duplicate
// legends, still answer to the rule. The exemption is read live off the
// battlefield, so the token copies (legendary, and creatures) coexist with
// their original and with each other.
//
// The second is an intervening "if" (CR 603.4), Franklin Richards' shape:
// checked as combat begins and again as the ability resolves. The token is a
// copy of the source (CR 707.2), so it carries both abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:             "33fa949c-f29e-4a98-9381-ecfa9e5c3f6d",
		Name:                 "Council of Reeds",
		Completeness:         CompletenessFull,
		LegendRuleExemptions: LegendRuleDoesntApplyToYourCreatures(),
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), franklinCastNoncreature),
				"Council of Reeds — create a token that's a copy of Council of Reeds",
				func(g *game.Game, item *game.StackItem) error {
					if g.CastTallyFor(item.Controller).Noncreature == 0 {
						return nil
					}
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
