package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stone-Seeder Hierophant — Creature — Human Druid {2}{G}{G}, 1/1 (Ravnica:
// City of Guilds, S82 OLD batch 576, #5600):
//
//	"Landfall — Whenever a land you control enters, untap this creature.
//	 {T}: Untap target land."
//
// The landfall trigger untaps the Hierophant as it resolves if it is still
// the permanent that triggered; the ability taps it as a cost and untaps the
// target land as it resolves (Blossom Dryad's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccaea194-4b21-4485-ba08-d8d030c9418d",
		Name:         "Stone-Seeder Hierophant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Stone-Seeder Hierophant — untap this creature", effects.UntapThisPermanentEffect),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap target land.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.B22UntapEachLegalTarget,
		}},
	})
}
