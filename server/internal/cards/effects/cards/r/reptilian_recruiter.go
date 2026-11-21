package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reptilian Recruiter — Creature — Lizard Warrior {3}{R}{R}, 4/2
// (Bloomburrow, #4467):
//
//	"Trample
//	 When this creature enters, choose target creature. If that
//	 creature's power is 2 or less or if you control another Lizard,
//	 gain control of that creature until end of turn, untap it, and it
//	 gains haste until end of turn."
//
// Any creature is a legal target; both conditions are read as the trigger
// resolves (the ruling). "Another Lizard" is a permanent you control with
// the subtype other than this object, the word read through the
// Recruiter's text (CR 612.2). When either holds, the theft is the
// Eldrazi threaten (threatenAndGrant): control until end of turn, the
// untap, and haste until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a12a66b5-ea4a-4a58-b0f0-13152598b7e6",
		Name:            "Reptilian Recruiter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Reptilian Recruiter — if its power is 2 or less or you control another Lizard, gain control of target creature until end of turn",
					reptilianRecruiterThreaten),
				effects.TargetCreature("target creature"),
			),
		},
	})
}

// reptilianRecruiterThreaten takes the target when its power is 2 or less
// or another Lizard you control is on the battlefield.
func reptilianRecruiterThreaten(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	c, found := g.LookupCardForEffect(target)
	if !found {
		return nil
	}
	anotherLizard := len(effects.MatchingBattlefield(ctx, effects.And(effects.Subtype("Lizard"), effects.YouControl(), effects.OtherThan(item.SourceCardID)))) > 0
	if c.CurrentPower() > 2 && !anotherLizard {
		return nil
	}
	return effects.ThreatenAndGrant(ctx, target, "Reptilian Recruiter", "haste")
}
