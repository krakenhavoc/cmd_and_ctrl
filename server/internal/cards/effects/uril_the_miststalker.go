package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Uril, the Miststalker — Legendary Creature — Beast, {2}{R}{G}{W}, 5/5:
//
//	"Hexproof"
//	"Uril, the Miststalker gets +2/+2 for each Aura attached to it."
//
// One of the two commanders #77 named by hand, and now fully live.
// Both halves are real: hexproof is enforced at the targeting gate
// (S23), and the Aura count is enforced by the layer engine against
// the attachment relation S24 landed (#374) partway through this
// sprint. This file was written a few hours earlier with the count
// declared as a deferral and a note saying it would be "a one-line
// predicate over that field the day it lands" — it was.
//
// # Why the count is a Static and not an ETB or a trigger
//
// "+2/+2 for each Aura attached to it" is a CR 613 continuous effect
// with a variable input, so it is recomputed on every pass rather
// than latched at any moment: put an Aura on, Uril is bigger; Disenchant
// it, Uril shrinks, mid-combat, without anything having to remember
// to undo a counter. Layer 7c (modify P/T), so it composes additively
// with an anthem and stacks on top of a 7b "base P/T becomes N/N"
// rather than being overwritten by it.
//
// The count is one battlefield pass, reading the reverse direction
// of `Card.AttachedTo` the way `game.AttachmentsOf` does. Safe from
// inside an `Apply`: it takes no locks of its own and the recompute
// pass already holds the read side. The `IsAura` filter matters
// because Equipment attaches through the same field — a Sword of
// Feast and Famine on Uril grants its own bonuses and must not also
// be counted here.
//
// # This is why voltron wants a hexproof commander
//
// The two lines are one card. Auras are the archetype's classic
// two-for-one risk: sink three of them into a creature, lose the
// creature to a Doom Blade, lose everything. Hexproof means the
// opponent has to answer Uril with a wrath or an edict, and the
// commander-damage clock (CR 903.10a) means they have about three
// turns to find one — a 5/5 with two Auras is a 9/9 and kills in
// three connections.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "4308a020-48cf-45fe-8074-dd5d0ac6d12d",
		Name:            "Uril, the Miststalker",
		PrintedKeywords: []string{"hexproof"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if g == nil {
					return
				}
				auras := 0
				for _, att := range g.BattlefieldCardsForEffect() {
					if att.IsAttachedTo(source.InstanceID) && att.IsAura() {
						auras++
					}
				}
				c.Power += 2 * auras
				c.Toughness += 2 * auras
			},
		}},
	})
}
