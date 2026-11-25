package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warteye Witch — Creature — Goblin Shaman {2}{B}, 3/2 (Modern Horizons):
//
//	"Whenever this creature or another creature you control dies, scry 1."
//
// Phyrexian Vivisector's trigger: a creature of yours dying, the Witch's own
// death included (CR 603.10a). The scry is the footprinted primitive and its
// fixed amount is declared on the card's Purpose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "594a57f6-7827-41c7-aa37-00b498720b16",
		Name:         "Warteye Witch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Warteye Witch — scry 1", effects.Do(effects.Scry{N: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
