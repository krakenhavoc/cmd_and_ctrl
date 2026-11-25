package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska's Conquistador — Creature — Vampire Soldier {1}{B}, 2/1:
//
//	"Whenever this creature attacks or blocks, if you control a Vraska
//	 planeswalker, target opponent loses 2 life and you gain 2 life."
//
// An intervening "if" (CR 603.4): a Vraska planeswalker (any planeswalker with
// that subtype, whoever's it was printed on) must be controlled both as the
// creature attacks or blocks and as the ability resolves; the target
// opponent is chosen as the ability goes on the stack (CR 603.3d). The drain
// is a life loss and a life gain, not damage.
//
// No simplification.
func init() {
	t := effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.Self,
		"Vraska's Conquistador — target opponent loses 2 life and you gain 2 life",
		effects.Do(effects.TargetPlayerLosesLife{N: 2}, effects.GainLife{Amount: 2}))
	t.Targets = effects.TargetPlayer("target opponent", effects.Opponent())
	effects.Register(effects.Spec{
		OracleID:     "d3839a3f-3018-4a8d-b298-3f5b5fc70cd1",
		Name:         "Vraska's Conquistador",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("you control a Vraska planeswalker",
				func(g *game.Game, you uuid.UUID) bool { return effects.ControlsA("Vraska")(g, you) }, t),
		},
	})
}
