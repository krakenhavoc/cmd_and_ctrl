package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angel of Deliverance — Creature — Angel {6}{W}{W}, 6/6 (Shadows over
// Innistrad, #5543):
//
//	"Flying
//	 Delirium — Whenever this creature deals damage, if there are four or
//	 more card types among cards in your graveyard, exile target creature an
//	 opponent controls."
//
// An intervening if (CR 603.4) on the controller's own graveyard, checked as
// the damage is dealt and again as the trigger resolves. Damage the Angel
// deals to several recipients at once is one occurrence (CR 603.2c), so the
// ability triggers once per batch. The target is chosen as the trigger goes
// on the stack; a target gone by resolution is not exiled (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f888f3ba-8316-4c10-bf97-873731f7d364",
		Name:            "Angel of Deliverance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.InterveningIf(effects.B24CondDelirium{},
				effects.OncePerBatch(effects.On(game.EventDealDamage, effects.AllOf(effects.ThisDealtDamage, effects.YouHaveDeliriumWhen),
					"Angel of Deliverance — delirium: exile target creature an opponent controls", effects.ExileTheFirstLegalTarget))),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
