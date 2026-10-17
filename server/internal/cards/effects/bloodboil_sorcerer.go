package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "acdcffa0-87e5-48a2-93d2-572febfc9d52",
		Name:         "Bloodboil Sorcerer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Bloodboil Sorcerer"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crown of Madness — {1}{R}, Sacrifice an artifact or creature: Goad target creature.",
			Cost:    Plus(ManaCost("{1}{R}"), SacrificeN(1, "an artifact or creature", Or(Artifact(), Creature()))),
			Targets: TargetCreature("target creature"),
			Effect:  GoadTarget,
		}},
	})
}
