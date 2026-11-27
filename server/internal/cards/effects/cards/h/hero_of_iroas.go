package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hero of Iroas — {1}{W} Creature — Human Soldier 2/2 (Born of the Gods):
//
//	"Aura spells you cast cost {1} less to cast.
//	 Heroic — Whenever you cast a spell that targets this creature, put a +1/+1 counter on
//	 this creature."
//
// Transcendent Envoy's reducer (CR 601.2f, generic mana only) and Bloodcrazed Hoplite's
// heroic. An Aura spell targeting the Hero triggers both.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d44bd43-77da-449e-b771-8a14da88b677",
		Name:         "Hero of Iroas",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Aura spells you cast cost {1} less to cast.", effects.YourSpell(), effects.SocB14AuraSpell()),
		},
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Hero of Iroas — put a +1/+1 counter on it (heroic)", effects.ThisGetsAPlusOneCounter),
		},
	})
}
