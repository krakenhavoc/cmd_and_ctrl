package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Druid Class — Enchantment — Class {1}{G} (Adventures in the Forgotten
// Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Landfall — Whenever a land you control enters, you gain 1 life.
//	 {2}{G}: Level 2
//	 You may play an additional land on each of your turns.
//	 {4}{G}: Level 3
//	 When this Class becomes level 3, target land you control becomes a
//	 creature with haste and "This creature's power and toughness are each
//	 equal to the number of lands you control." It's still a land."
//
// One line per level (ADR 0071):
//
//   - Level 1 is the shared Landfall trigger, ungated (CR 716.3).
//   - Level 2 is Exploration's extra land drop, gated at level 2 by
//     Spec.AdditionalLandPlaysAtLevel, read live on every land play.
//   - Level 3 is BecomesLevel's trigger with a land target. Its effect has no
//     duration (CR 611.2a), so it is one record pinned to that land (CR
//     400.7): a layer 4 Creature type added to the land's types (it is
//     still a land), a layer 6 haste, and the granted static of Beorn's
//     Hospitality, a layer 7b set to the number of lands its controller
//     controls, counted live. A land that left in response makes the
//     trigger do nothing (CR 608.2b).
//
// No simplification.
func init() {
	const grantKey = "druid-class/power-toughness-lands"
	Register(Spec{
		OracleID:                   "dcbcbf42-4654-487a-acad-21f2606d229b",
		Name:                       "Druid Class",
		Completeness:               CompletenessFull,
		AdditionalLandPlays:        1,
		AdditionalLandPlaysAtLevel: 2,
		Purpose:                    game.Purpose{ExtraLandDrops: 1},
		Grants: []AbilityGrant{{
			Key: grantKey,
			Static: []game.StaticAbility{{
				Label:    "This creature's power and toughness are each equal to the number of lands you control.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7B_Set,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return source != nil && target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					n := b10LandsControlled(g, source.Controller)
					c.Power, c.Toughness = n, n
				},
			}},
			Text: "This creature's power and toughness are each equal to the number of lands you control.",
		}},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{2}{G}")),
			LevelUp(3, ManaCost("{4}{G}")),
		},
		Triggered: []game.TriggeredAbility{
			Landfall("Druid Class — landfall: you gain 1 life", Do(GainLife{Amount: 1})),
			Targeting(
				BecomesLevel(3, "Druid Class — level 3: target land you control becomes a creature with haste",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantAbilitiesFor{
							Target: id,
							Keys:   []string{grantKey},
							Also: []game.Mod{
								game.AddTypesMod("Creature"),
								game.AddKeywordsMod("haste"),
							},
							Duration: g.PinnedTo(game.IndefiniteDuration(), id),
							Label:    "Druid Class — the land becomes a creature with haste and power and toughness equal to your lands",
						}.Apply(ctx)
					}),
				TargetPermanent("target land you control", Land(), YouControl())),
		},
	})
}
