package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwindle — Enchantment — Aura {2}{U} (Core Set 2019, #5527):
//
//	"Enchant creature
//	 Enchanted creature gets -6/-0.
//	 When enchanted creature blocks, destroy it. (The attacking creature
//	 remains blocked.)"
//
// Mystic Subdual's attached modifier (-6/-0) and Psychic Impetus' Aura-owned
// trigger, on EventBlock: the plain "blocks" triggers once however many
// attackers the creature blocks (CR 509.3a, Amount <= 1). "It" is the
// creature that blocked, read off the trigger's event, so it is destroyed
// even if the Aura has left by resolution; the attacker stays blocked
// (CR 506.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48a8fce6-b493-4b1d-8ac8-2a4f4b4781b5",
		Name:         "Dwindle",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(-6, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Target != uuid.Nil && ev.Amount <= 1 && source.IsAttachedTo(ev.CardID)
			}, "Dwindle — destroy the enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTarget{Target: item.Trigger.Event.CardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
