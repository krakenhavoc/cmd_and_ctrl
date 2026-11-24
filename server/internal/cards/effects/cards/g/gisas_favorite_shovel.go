package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gisa's Favorite Shovel — Legendary Artifact — Equipment {1}{B}:
//
//	"Equipped creature gets +2/+0 and has menace.
//	 Whenever equipped creature attacks, defending player sacrifices a creature
//	 of their choice. If they do, you create a Walker token. (It's a 2/2 black
//	 Zombie creature token.)
//	 Equip {4}"
//
// The pump and menace are the usual attached statics. The attack trigger is the
// Equipment's own, matched through the attachment. The edict is Thraximundar's
// (the defending player is the player attacked, or the controller of the
// planeswalker or battle attacked, CR 506.3d, and they choose the creature as it
// resolves, and a creature that can't be sacrificed is not a legal choice) with the "if they do"
// riding it as a keyed run continuation (Doomgape's): the Walker is made, for
// the Shovel's controller, only for a creature that really left the battlefield
// as a sacrifice. A defending player with no creature sacrifices nothing and
// there is no Walker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c298c7b-9c63-4a66-a0fc-977491b623c7",
		Name:         "Gisa's Favorite Shovel",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Gisa's Favorite Shovel — defending player sacrifices a creature of their choice; if they do, you create a Walker token",
				func(g *game.Game, item *game.StackItem) error {
					victim := effects.DefendingPlayerOfAttack(g, item.Trigger.Event)
					if victim == uuid.Nil || g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, victim,
						effects.SacrificeSpec("a creature", effects.Creature()),
						"Gisa's Favorite Shovel — sacrifice a creature", 1,
						gisasFavoriteShovelWalker, []uuid.UUID{item.Controller, victim})
				})),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}

// gisasFavoriteShovelWalker is "if they do, you create a Walker token": the
// carry names the Shovel's controller and the defending player. An on-disk key.
var gisasFavoriteShovelWalker = game.RegisterPromptRunThen("prompt-run/gisas-favorite-shovel-walker", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 2 || len(r.Sacrifices().By(r.Carry[1])) == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie named Walker"), N: 1}.Apply(effects.KeyedContext(g, r.Carry[0], r.Source))
})
