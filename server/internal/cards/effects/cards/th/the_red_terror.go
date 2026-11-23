package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Red Terror — Legendary Creature — Tyranid {3}{R}, 4/3 (Warhammer 40,000
// Commander, S82 GAP batch 194, #5948):
//
//	"Advanced Species — Whenever a red source you control deals damage to one or
//	 more permanents and/or players, put a +1/+1 counter on The Red Terror."
//
// A "one or more" clause (CR 603.2c): each damage event the engine emits per
// recipient is collapsed per source within one simultaneous batch, so a red
// creature striking three creatures at once triggers it once, while two red
// sources are two triggers. The source is read as it dealt the damage, wherever
// it is (a red instant on the stack counts) and is any source you control,
// The Red Terror itself included; damage prevented to 0 does not trigger it.
// The counter goes on The Red Terror only while it is still that permanent.
//
// No simplification.
func init() {
	trigger := effects.OncePerBatch(effects.On(game.EventDealDamage, effects.GapB194ARedSourceYouControlDealtDamage,
		"The Red Terror — a +1/+1 counter on it", effects.PlusOneCountersOnThis(1)))
	trigger.BatchKey = effects.GapB194PerSource
	effects.Register(effects.Spec{
		OracleID:     "557bd3c9-1c60-480d-9650-30c317506097",
		Name:         "The Red Terror",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
