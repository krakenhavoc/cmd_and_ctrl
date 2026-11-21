package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vraska Joins Up — Legendary Enchantment {B}{G} (EDHREC rank 2594):
//
//	"When Vraska Joins Up enters, put a deathtouch counter on each
//	 creature you control.
//	 Whenever a legendary creature you control deals combat damage to
//	 a player, draw a card."
//
// The Golgari "Joins Up" card. The ETB puts a deathtouch counter on
// every creature the controller controls (snapshotted before the
// first counter lands, so a token made in response gets none); the
// draw trigger is combatDamageToPlayerBy narrowed to a legendary
// dealer, read post-layer (b24LegendaryCreatureYouControlDealtCombatDamageToPlayer),
// one card per legendary creature that connects.
//
// The deathtouch counters need nothing more: the engine reads keyword
// counters itself (CR 122.1b, ADR 0101), so each creature keeps
// deathtouch for as long as it keeps the counter, whether or not
// Vraska is still on the battlefield. Until ADR 0101 Vraska carried the
// rule in a static of its own, which stopped when Vraska left.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c91b0dd5-4c63-49a5-95bf-c342b6ff2076",
		Name:         "Vraska Joins Up",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vraska Joins Up — put a deathtouch counter on each creature you control", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.B04CreatureIDsControlledBy(g, item.Controller) {
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
						continue
					}
					if err := (effects.AddCounter{Target: id, Kind: game.CounterDeathtouch, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B24LegendaryCreatureYouControlDealtCombatDamageToPlayer(ev, source, g)
			}, "Vraska Joins Up — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
