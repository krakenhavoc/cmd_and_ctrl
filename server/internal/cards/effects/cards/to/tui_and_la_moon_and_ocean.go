package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tui and La, Moon and Ocean — Legendary Creature — Fish Spirit {3}{U},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever Tui and La become tapped, draw a card.
//	 Whenever Tui and La become untapped, put a +1/+1 counter on them."
//
// Two triggers, each an existing shape. "Become tapped" is Attentive
// Sunscribe's condition (ThisBecameTapped over the tap and attack
// events, so attacking without vigilance counts). "Become untapped" is
// the per-permanent untap event the engine emits only on a real change of
// state (CR 701.26b): the untap step, an untap effect, either one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "781c9391-0322-40ad-8050-07151e2e6fc7",
		Name:         "Tui and La, Moon and Ocean",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Tui and La — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.On(game.EventUntapCard, effects.Self,
				"Tui and La — put a +1/+1 counter on them", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
