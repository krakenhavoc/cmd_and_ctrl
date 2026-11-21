package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emmara, Soul of the Accord — Legendary Creature — Elf Cleric {G}{W}, 2/2
// (Commander Starter Decks):
//
//	"Whenever Emmara becomes tapped, create a 1/1 white Soldier creature token
//	 with lifelink."
//
// "Becomes tapped" is any tap, an attack included (Badgey's ThisBecameTapped,
// CR 701.26a); an attack that does not tap her (vigilance) is silent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c65ba242-3369-48a9-864f-1b1f85238f67",
		Name:         "Emmara, Soul of the Accord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Emmara, Soul of the Accord — create a 1/1 white Soldier with lifelink",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier with lifelink"), N: 1})),
		},
	})
}
