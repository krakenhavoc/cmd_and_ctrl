package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staff of Titania — Artifact — Equipment {2} (EDHREC rank 4331):
//
//	"Equipped creature gets +X/+X, where X is the number of Forests
//	 you control.
//	 Whenever equipped creature attacks, create a 1/1 green Forest
//	 Dryad land creature token. (It's affected by summoning sickness.)
//	 Equip {3}"
//
// A two-mana Equipment that scales with the deck it is in and then
// grows the thing it scales with: every attack makes another Forest,
// which makes the bonus bigger. In a mono-green deck it is a +5/+5 by
// turn six that keeps climbing.
//
// The bonus is a LIVE READ, not a number fixed when the Staff was
// equipped: TribalScalingAnthem's `per` runs on every layer recompute,
// so the Dryad token the attack trigger just made is already counted
// when combat damage is assigned. That is what makes the card an
// engine rather than an anthem.
//
// "The number of FORESTS you control" is the land TYPE, not the basic
// land: a Stomping Ground, a Snow-Covered Forest, a Dryad Arbor and
// the Staff's own Dryad tokens all count, and so does any land under a
// type-changing effect. b08LandsWithSubtypeControlled reads effective
// subtypes for exactly that reason.
//
// The Dryad token is a LAND CREATURE, which is why it is affected by
// summoning sickness (the reminder text spells it out) and why making
// one is not a land play — it does not use your land drop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f99f4b52-21ae-47a2-8ae6-b0aa7386723a",
		Name:         "Staff of Titania",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := b08LandsWithSubtypeControlled(g, source.Controller, "Forest")
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Staff of Titania — create a 1/1 green Forest Dryad",
				Do(CreateToken{Template: TokenCard("1/1 green Dryad"), N: 1})),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}
