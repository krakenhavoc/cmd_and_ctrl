package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dancer's Chakrams — Artifact — Equipment {3}{W}:
//
//	"Job select (When this Equipment enters, create a 1/1 colorless Hero
//	 creature token, then attach this to it.)
//	 Equipped creature gets +2/+2, has lifelink and "Other commanders you
//	 control get +2/+2 and have lifelink," and is a Performer in addition
//	 to its other types.
//	 Krishna — Equip {3}"
//
// The quoted ability is a static the equipped creature HAS, not one the
// Equipment has (ADR 0093): it is the creature's, so "other" and "you"
// are the creature's, and a creature that loses its abilities stops
// buffing the commanders while the Equipment stays attached. Writing it
// on the Equipment would have kept it working through that, which is
// stronger than printed; #3784 is the engine half that let the bundle
// carry the lifelink, a layer-6 static, beside the layer-7c +2/+2 (ADR
// 0093 amendment 2026-10-15).
//
// "Commanders" reads Card.IsCommander, as Anara does: a commander you
// have stolen is one you control and is covered, and your own commander
// somebody else controls is not. The +2/+2 asks for a creature, since a
// noncreature commander has no power and toughness to raise; lifelink
// goes to every other commander you control.
//
// No simplification.
func init() {
	const commandersKey = "dancers-chakrams/other-commanders"
	otherCommandersYouControl := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCommander &&
			target.InstanceID != source.InstanceID &&
			target.Controller == source.Controller
	}
	Register(Spec{
		OracleID:     "7465d02f-3eee-483c-8de3-404381aa21eb",
		Name:         "Dancer's Chakrams",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: commandersKey,
			Static: []game.StaticAbility{
				{
					Label:    "Other commanders you control get +2/+2.",
					Layer:    game.Layer7PT,
					SubLayer: game.SubLayer7C_Modify,
					AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
						return target.IsCreature() && otherCommandersYouControl(target, g, source)
					},
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power += 2
						c.Toughness += 2
					},
				},
				{
					Label:     "Other commanders you control have lifelink.",
					Layer:     game.Layer6Ability,
					AppliesTo: otherCommandersYouControl,
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, "lifelink")
					},
				},
			},
			Text: "Other commanders you control get +2/+2 and have lifelink.",
		}},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("lifelink"),
			GrantAbilitiesToAttached(commandersKey),
			AttachedIsAlso("Performer"),
		},
		Triggered: []game.TriggeredAbility{JobSelect("Dancer's Chakrams")},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
