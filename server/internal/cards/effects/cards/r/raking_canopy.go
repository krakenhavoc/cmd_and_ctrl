package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raking Canopy — Enchantment {1}{G}{G} (Shadowmoor, S82 OLD b530, #5591):
//
//	"Whenever a creature with flying attacks you, this enchantment deals 4
//	 damage to it."
//
// Barbed Foliage's "attacks you" is the player: an EventAttack names the
// attacked player or planeswalker in Target, so a flier attacking one of your
// planeswalkers does not trigger it. Flying is read as the creature attacks
// (CR 603.2). The enchantment deals the damage as an object, so one that has
// left the battlefield still deals it; a creature that has left is dealt none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75f381a6-49f9-4caf-a96a-91c75640b9b0",
		Name:         "Raking Canopy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil || ev.Target != source.Controller {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.HasKeyword(&attacker, "flying")
			}, "Raking Canopy — 4 damage to the attacking creature with flying",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					return effects.DealDamage{SourceObject: &ref, Target: ctx.Trigger().Event.CardID, Amount: 4}.Apply(ctx)
				}),
		},
	})
}
