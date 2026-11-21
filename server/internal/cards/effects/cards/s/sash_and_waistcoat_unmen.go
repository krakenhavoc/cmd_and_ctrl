package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sash and Waistcoat, Unmen — Legendary Creature — Insect Rogue {1}{U}, 2/2
// (Mystery Booster Commander, #5948):
//
//	"Whenever a face-down creature attacks one of your opponents, it can't be
//	 blocked this combat.
//	 {3}: Manifest the top card of your library. Any player may activate this
//	 ability but only as a sorcery."
//
// The first ability watches every attacker, whoever controls it: a face-down
// creature (as it was declared) aimed at a player other than Sash's
// controller (CR 508.3a: an attack on a planeswalker or battle is not an
// attack on its controller). The evasion is a "can't be blocked" restriction
// on the attacker as an object until end of turn, which ends with the combat.
//
// The manifest is an any-player row (ADR 0106): "you" is the activator, who
// pays the {3} from their own pool, manifests the top card of their own
// library and controls the face-down 2/2. "Only as a sorcery" is the
// activator's: their own turn, a main phase, an empty stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ff470ee-4532-4ded-a684-a2c0d1253fb3",
		Name:         "Sash and Waistcoat, Unmen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				defender := effects.AttackedPlayer(g, ev)
				if defender == uuid.Nil || defender == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.FaceDown
			}, "Sash and Waistcoat, Unmen — that creature can't be blocked this combat",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					return effects.RestrictUntilEOT{
						Target:       ctx.Trigger().Object.Ref().ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Sash and Waistcoat, Unmen — can't be blocked this combat",
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}: Manifest the top card of your library. Any player may activate this ability but only as a sorcery.",
			Cost:         effects.ManaCost("{3}"),
			AnyPlayer:    true,
			SorcerySpeed: true,
			Effect:       effects.Do(effects.Manifest{N: 1}),
		}},
	})
}
