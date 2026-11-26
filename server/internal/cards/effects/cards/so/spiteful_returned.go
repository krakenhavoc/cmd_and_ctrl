package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiteful Returned — Enchantment Creature — Zombie {1}{B}, 1/1 (Born of the
// Gods, S82 OLD b377, #5559):
//
//	"Bestow {3}{B}
//	 Whenever this creature or enchanted creature attacks, defending player
//	 loses 2 life.
//	 Enchanted creature gets +1/+1."
//
// One printed trigger, one registered trigger: the attack of the source or of
// its host (S82OldB377ThisOrEnchantedAttacked), the defending player read off
// the attack event (S82OldB377DefenderLoses). Life loss, not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e7e60afe-5ddf-406b-86e1-e82e2a660432",
		Name:             "Spiteful Returned",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Bestow("{3}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.S82OldB377ThisOrEnchantedAttacked,
				"Spiteful Returned — defending player loses 2 life", effects.S82OldB377DefenderLoses(2)),
		},
		Static: []game.StaticAbility{effects.PumpAttached(1, 1)},
	})
}
