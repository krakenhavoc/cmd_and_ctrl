package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of the Woods — Enchantment — Aura {G} (Alliances, #5634):
//
//	"Enchant creature
//	 Whenever enchanted creature blocks or becomes blocked, it gets +0/+3 until
//	 end of turn and you gain 1 life."
//
// Ferocity's one trigger over both events (EventBlock names the blocker,
// EventBecomesBlocked the attacker, once however many block it, CR 509.1h). The
// toughness goes to the creature the event names while it is on the battlefield;
// the life goes to the Aura's controller either way (CR 109.5).
//
// No simplification.
func init() {
	const name = "Gift of the Woods"
	effects.Register(effects.Spec{
		OracleID:     "491c6f69-9b75-4d5d-b41f-a14b2fbfdd07",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return source.IsAttachedTo(ev.CardID)
				},
				name+" — it gets +0/+3 until end of turn and you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if id := item.Trigger.Event.CardID; effects.OnBattlefield(g, id) {
						if err := (effects.BoostUntilEOT{Target: id, Toughness: 3, Label: name + " — +0/+3"}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
				}),
		},
	})
}
