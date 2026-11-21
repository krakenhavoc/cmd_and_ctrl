package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wicked Wolf — Creature — Wolf {2}{G}{G}, 3/3 (Jumpstart 2022, #5948):
//
//	"When this creature enters, it fights up to one target creature you
//	 don't control.
//	 Sacrifice a Food: Put a +1/+1 counter on this creature. It gains
//	 indestructible until end of turn. Tap it."
//
// The fight is Agatha's Champion's (CR 701.14): up to one target, and with none
// chosen, or the Wolf gone, nothing happens. "Sacrifice a Food" is Bosco's
// cost: any Food artifact its controller controls, a token or not. The body is
// the counter, then indestructible, then the tap, each only while the Wolf is
// still the permanent that activated it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19cede61-1982-47d1-9e0e-49ac6de00798",
		Name:         "Wicked Wolf",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Wicked Wolf — it fights up to one target creature you don't control", b30SourceFightsFirstLegalTarget),
				TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Food: Put a +1/+1 counter on this creature. It gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))},
			Effect:  gapB153WickedWolfSacrificeFood,
		}},
	})
}
