package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curse of the Restless Dead — Enchantment — Aura Curse {2}{B}:
//
//	"Enchant player
//	 Whenever a land enchanted player controls enters, you create a 2/2
//	 black Zombie creature token with decayed. (It can't block. When it
//	 attacks, sacrifice it at end of combat.)"
//
// Curse of Opulence's shape: the enchanted player is the Aura's
// AttachedTo, and the trigger is controlled by the Curse's controller,
// who makes the token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b5f7a1b-fbb9-4659-8fcb-7822dc0263bc",
		Name:         "Curse of the Restless Dead",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source.AttachedTo.Kind != game.TargetPlayer {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand() && c.Controller == source.AttachedTo.ID
			}, "Curse of the Restless Dead — create a 2/2 Zombie with decayed",
				Do(CreateToken{Template: TokenCard("2/2 black Zombie with decayed"), N: 1})),
		},
	})
}
