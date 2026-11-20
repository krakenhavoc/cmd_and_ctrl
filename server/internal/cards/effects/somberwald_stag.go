package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Somberwald Stag — Creature — Elk {3}{G}{G}, 4/3 (Eldritch Moon, #5541):
//
//	"When this creature enters, you may have it fight target creature you
//	 don't control."
//
// Affectionate Indrik's trigger, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1b67851-23d3-4799-ab29-15bce7bb2c39",
		Name:         "Somberwald Stag",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WhenThisEnters("Somberwald Stag — it fights target creature you don't control",
					b30SourceFightsFirstLegalTarget),
				TargetCreature("target creature you don't control", OpponentControls())),
				"Somberwald Stag — have it fight target creature you don't control?"),
		},
	})
}
