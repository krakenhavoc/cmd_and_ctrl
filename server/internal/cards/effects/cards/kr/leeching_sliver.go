package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leeching Sliver — Creature — Sliver {1}{B}, 1/1 (Secret Lair Drop):
//
//	"Whenever a Sliver you control attacks, defending player loses 1 life."
//
// One trigger per attacking Sliver (CR 603.2c; each creature attacks on its
// own), whichever of Leeching Sliver's controller's Slivers it is, itself
// included. "Sliver" is a word of the source's text (CR 612.2). The defending
// player is the one the attack is aimed at, or the controller of the planeswalker
// or battle it attacks (CR 508.5); they lose the life as the ability resolves,
// from the attacker's event, even if the Sliver has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f104b361-8000-4eda-b530-1502528e09f9",
		Name:         "Leeching Sliver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Sliver"))
			}, "Leeching Sliver — defending player loses 1 life", func(g *game.Game, item *game.StackItem) error {
				defender := effects.B17DefendingPlayer(g, item.Trigger.Event)
				if p := g.PlayerByIDForEffect(defender); p == nil || p.Eliminated {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -1)
			}),
		},
	})
}
