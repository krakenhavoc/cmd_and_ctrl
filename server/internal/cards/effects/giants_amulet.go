package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant's Amulet — Artifact — Equipment {U}:
//
//	"When this Equipment enters, you may pay {3}{U}. If you do, create a
//	 4/4 blue Giant Wizard creature token, then attach this Equipment to
//	 it.
//	 Equipped creature gets +0/+1 and has "This creature has hexproof as
//	 long as it's untapped."
//	 Equip {2}"
//
// The quoted ability is the equipped creature's own static (ADR 0093),
// so "this creature" is the creature, and its hexproof comes and goes
// with the creature's tap state. It is a granted layer-6 static, which
// the layer pass could not apply before #3784 (ADR 0093 amendment
// 2026-10-15). A tap or untap already invalidates the layer cache, so
// the hexproof is re-read the moment the creature taps.
//
// The enter trigger is optional mana on resolution (MayPay), and the
// attach rides the token's creation, as job select's does: if the
// Amulet has left in response, the Giant is made and nothing is
// attached.
//
// No simplification.
func init() {
	const hexproofKey = "giants-amulet/untapped-hexproof"
	Register(Spec{
		OracleID:     "016a6ec3-f4f6-4b47-9931-e0b14a25bd5e",
		Name:         "Giant's Amulet",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: hexproofKey,
			Static: []game.StaticAbility{{
				Label: "This creature has hexproof as long as it's untapped.",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID && !target.Tapped
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, game.KeywordHexproof)
				},
			}},
			Text: "This creature has hexproof as long as it's untapped.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(0, 1),
			GrantAbilitiesToAttached(hexproofKey),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Giant's Amulet — you may pay {3}{U} to create a 4/4 Giant Wizard and attach this to it",
				giantsAmuletMayPayForAGiant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// giantsAmuletMayPayForAGiant is the enter trigger's body: the offer,
// then the Giant and the attach on a payment. Package-level, so it
// captures nothing across an undo.
func giantsAmuletMayPayForAGiant(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{3}{U}",
		Question: "Giant's Amulet — pay {3}{U} to create a 4/4 blue Giant Wizard and attach the Amulet to it?",
		OnPay: func(ctx *Context) error {
			return createTokenAndAttachSource(ctx.Game, ctx.Item, giantsAmuletGiantToken, true)
		},
	}.Apply(NewContext(g, item))
}

// giantsAmuletGiantToken is the token the Amulet makes.
const giantsAmuletGiantToken = "4/4 blue Giant Wizard"
