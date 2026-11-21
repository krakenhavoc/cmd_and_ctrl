package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adorned Crocodile — Creature — Crocodile {4}{B}, 5/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature dies, create a 2/2 black Zombie Druid creature
//	 token.
//	 Renew — {B}, Exile this card from your graveyard: Put a +1/+1
//	 counter on target creature. Activate only as a sorcery."
//
// The dies trigger is the table's Zombie Druid row; renew is renew.go.
// The two compose in order: the Crocodile dies, makes its Zombie, and can
// later be exiled from the graveyard to grow it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb49c7d2-d750-4676-b90e-76e7fae70a25",
		Name:         "Adorned Crocodile",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Adorned Crocodile — create a 2/2 black Zombie Druid",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie Druid"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			effects.Renew("{B}", "Put a +1/+1 counter on target creature.",
				effects.TargetCreature("target creature"), effects.RenewCountersOnTarget(game.CounterPlusOne, 1)),
		},
	})
}
