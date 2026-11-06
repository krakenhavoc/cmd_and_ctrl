package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alesha's Legacy — Instant {1}{B} (Tarkir: Dragonstorm, #4463):
//
//	"Target creature you control gains deathtouch and indestructible until
//	 end of turn."
//
// One until-end-of-turn keyword grant of both words on the target. The
// target is re-checked as the spell resolves (CR 608.2b): a creature that
// changed controller in response is no longer "you control" and gains
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4229cc8-11e5-4221-b00d-86969701e21b",
		Name:         "Alesha's Legacy",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose: game.Purpose{
			Answers: game.AnswerSurvive | game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch", "indestructible")),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"deathtouch", "indestructible"},
				Label:    "Alesha's Legacy — deathtouch and indestructible",
			}.Apply(ctx)
		},
	})
}
