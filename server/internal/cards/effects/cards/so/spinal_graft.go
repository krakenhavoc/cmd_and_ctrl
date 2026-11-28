package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinal Graft — Enchantment — Aura {1}{B} (Tempest, S82 OLD, #5628):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 When enchanted creature becomes the target of a spell or ability, destroy
//	 that creature. It can't be regenerated."
//
// A layer 7c +3/+3 and a trigger on the becomes-target event (one per target slot,
// CR 115.3) for the creature the Aura is attached to, whoever controls the spell
// or ability, the creature's own controller included. The destruction can't be
// regenerated (CR 701.19c), and the creature is destroyed only if it is still on
// the battlefield as the trigger resolves.
//
// No simplification.
func init() {
	const name = "Spinal Graft"
	effects.Register(effects.Spec{
		OracleID:     "70b5eca0-c61b-4066-ae5f-90948912c04e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && source.AttachedTo.Kind == game.TargetCard && ev.CardID == source.AttachedTo.ID
			}, name+" — destroy that creature; it can't be regenerated", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				host := item.Trigger.Event.CardID
				if !effects.OnBattlefield(g, host) {
					return nil
				}
				return effects.DestroyTarget{Target: host, CantBeRegenerated: true}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
