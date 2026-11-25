package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain's Hook — Artifact — Equipment {3}:
//
//	"Equipped creature gets +2/+0, has menace, and is a Pirate in addition to
//	 its other creature types.
//	 Whenever this Equipment becomes unattached from a permanent, destroy that
//	 permanent.
//	 Equip {1}"
//
// Three static effects on the equipped creature (a layer-4 subtype addition, a
// layer-6 keyword and a layer-7c pump), all following whatever it is attached
// to. The trigger is Grafted Exoskeleton's unattach shape, destroying instead
// of sacrificing: it fires when the Equipment moves to another creature, is
// unattached by a state-based action, or leaves the battlefield while attached
// (the permanent it was last attached to is the one destroyed, if it is still on
// the battlefield).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "25143290-2dd5-4def-a857-a1db6d3ed2eb",
		Name:         "Captain's Hook",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("menace"),
			{
				Label:     "Equipped creature is a Pirate in addition to its other creature types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.Subtypes = append(c.Subtypes, source.TextSubtypes("Pirate")...)
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventUnattach, game.EventLTB}, effects.Self,
				"Captain's Hook — destroy the permanent it became unattached from",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					host := effects.S82OldB220UnattachedHost(ctx)
					if host == [16]byte{} {
						return nil
					}
					if z := g.FindCardZoneForEffect(host); z == nil || z.Kind != game.ZoneBattlefield {
						return nil
					}
					return effects.DestroyTarget{Target: host}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
