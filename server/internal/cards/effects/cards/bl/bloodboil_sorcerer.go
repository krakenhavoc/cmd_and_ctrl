package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodboil Sorcerer — Creature — Human Shaman Sorcerer {3}{R}, 3/3:
//
//	"When this creature enters, you take the initiative.
//	 Crown of Madness — {1}{R}, Sacrifice an artifact or creature: Goad
//	 target creature. (Until your next turn, that creature attacks each
//	 combat if able and attacks a player other than you if able.)"
//
// The ETB is TakeTheInitiative (ADR 0203, #4096). Crown of Madness is
// an ordinary activated ability: the sacrifice is a cost, paid as the
// ability is activated (CR 602.2b), so the Sorcerer itself is a legal
// sacrifice; GoadTarget goads the target if it is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acdcffa0-87e5-48a2-93d2-572febfc9d52",
		Name:         "Bloodboil Sorcerer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouTakeTheInitiative("Bloodboil Sorcerer"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crown of Madness — {1}{R}, Sacrifice an artifact or creature: Goad target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeN(1, "an artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.GoadTarget,
		}},
	})
}
