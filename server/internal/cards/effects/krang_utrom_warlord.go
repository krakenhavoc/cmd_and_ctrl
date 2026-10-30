package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krang, Utrom Warlord — {9} Legendary Artifact Creature — Utrom Robot 9/9
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying, trample, indestructible, haste
//	 Other artifact creatures you control have flying, trample,
//	 indestructible, and haste."
//
// Four printed keywords and one layer 6 grant (CR 613.1f) of the same four
// to every other artifact creature its controller controls, Cyberdrive
// Awakener's scope with a longer list. The grant reads its words through the
// source's text (CR 612.2), as KeywordGrant does, and a keyword the creature
// already has is not added twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2d56a491-b3eb-4f3e-82ab-981d3036c45d",
		Name:            "Krang, Utrom Warlord",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "indestructible", "haste"},
		Static: []game.StaticAbility{{
			Label: "Other artifact creatures you control have flying, trample, indestructible, and haste.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsArtifact() && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				for _, kw := range []string{"flying", "trample", "indestructible", "haste"} {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, source.TextKeyword(kw))
				}
			},
		}},
	})
}
