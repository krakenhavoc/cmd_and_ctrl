package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bestial Fury — Enchantment — Aura {2}{R} (Alliances, #5634):
//
//	"Enchant creature
//	 When this Aura enters, draw a card at the beginning of the next turn's
//	 upkeep.
//	 Whenever enchanted creature becomes blocked, it gets +4/+0 and gains
//	 trample until end of turn."
//
// The enters trigger schedules the delayed draw (CR 603.7a) for the Aura's
// controller. The block trigger is the host's EventBecomesBlocked, which fires
// once however many creatures block it (CR 509.1h), and pumps and grants trample
// to the creature the event names while it is still on the battlefield.
//
// No simplification.
func init() {
	const name = "Bestial Fury"
	effects.Register(effects.Spec{
		OracleID:     "dfc08963-132f-4ebf-9384-15d17d1098f7",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — draw a card at the beginning of the next turn's upkeep", effects.Do(effects.DrawACardAtTheNextUpkeep(name+" — draw a card"))),
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, name+" — enchanted creature gets +4/+0 and gains trample until end of turn", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				id := item.Trigger.Event.CardID
				if !effects.OnBattlefield(g, id) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: id, Power: 4, Label: name + " — +4/+0"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: name + " — trample until end of turn"}.Apply(ctx)
			}),
		},
	})
}
