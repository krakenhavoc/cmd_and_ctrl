package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silvar, Devourer of the Free — Legendary Creature — Cat Nightmare
// {3}{B}{R}, 4/2 (Commander 2020, #5513):
//
//	"Partner with Trynn, Champion of Freedom (When this creature enters,
//	 target player may put Trynn into their hand from their library, then
//	 shuffle.)
//	 Menace
//	 Sacrifice a Human: Put a +1/+1 counter on Silvar. It gains
//	 indestructible until end of turn."
//
// Partner with is the CR 702.124j entry search (partner_with.go). The
// sacrifice is Fiend of the Shadows' "Sacrifice a Human" cost, paid as the
// ability is activated (CR 602.2b). The counter and the indestructible both
// land on Silvar as it resolves; a Silvar that has left gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "09199ebd-2c09-455f-ac10-3dbff8e6d428",
		Name:            "Silvar, Devourer of the Free",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Silvar, Devourer of the Free", "Trynn, Champion of Freedom"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Human: Put a +1/+1 counter on Silvar. It gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("a Human", HasSubtype("Human"))},
			Effect: Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, Instruction(func(ctx *Context) error {
				if !onBattlefield(ctx.Game, ctx.Source()) {
					return nil
				}
				return GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"indestructible"}, Label: "Silvar — indestructible until end of turn"}.Apply(ctx)
			})),
		}},
	})
}
