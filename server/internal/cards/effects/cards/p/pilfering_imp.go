package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pilfering Imp — Creature — Imp {B}, 1/1:
//
//	"Flying
//	 {1}{B}, {T}, Sacrifice this creature: Target opponent reveals their
//	 hand. You choose a nonland card from it. That player discards that
//	 card. Activate only as a sorcery."
//
// Thoughtseize's pick (ADR 0116) on an activated ability whose cost
// sacrifices its source. The chooser is the ability's controller, the
// player who activated it (CR 113.8), so the Imp being in the graveyard
// by resolution changes nothing. "Activate only as a sorcery" is the
// sorcery timing of CR 602.5d. The {T} needs the Imp free of summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61017a67-653c-4f75-9ec3-22d2149340a8",
		Name:            "Pilfering Imp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}, {T}, Sacrifice this creature: Target opponent reveals their hand. You choose a nonland card from it. That player discards that card. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect:       effects.TargetRevealsYouChooseDiscardAbility(effects.Nonland(), "nonland card"),
		}},
	})
}
