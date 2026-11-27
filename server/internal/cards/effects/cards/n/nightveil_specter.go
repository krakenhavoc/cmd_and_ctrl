package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nightveil Specter — Creature — Specter {U/B}{U/B}{U/B}, 2/3 (Gatecrash, S82
// OLD #5565):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, that player
//	 exiles the top card of their library.
//	 You may play lands and cast spells from among cards exiled with this
//	 creature."
//
// The exile is Ian Malcolm's: the damaged player's own top card, linked to
// this object (CR 607.2a, ExileLinkedForEffect), so a Specter that left and
// came back has no cards. An empty library exiles nothing. The second line is
// Kheru Mind-Eater's standing permission over exile, narrowed to the cards
// exiled with this source, for its controller, at the cards' own cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4ce83f8c-b27e-4e5b-b19b-76b66aaa904b",
		Name:            "Nightveil Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Nightveil Specter — that player exiles the top card of their library",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					p := g.PlayerByIDForEffect(item.Trigger.Event.Target)
					if p == nil || p.Eliminated || p.Library == nil || len(p.Library.Cards) == 0 {
						return nil
					}
					src, _ := g.SourceObjectForEffect(item)
					return g.ExileLinkedForEffect(src, p.Library.Cards[len(p.Library.Cards)-1].InstanceID)
				}),
		},
		CastPermissions: []game.CastPermission{{
			Zone:   game.ZoneExile,
			Filter: game.PermissionFilter{ExiledWithSource: true},
			Label:  "You may play lands and cast spells from among cards exiled with Nightveil Specter",
		}},
	})
}
