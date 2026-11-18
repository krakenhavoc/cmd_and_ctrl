package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avacyn's Memorial — Legendary Artifact {5}{W}{W}{W} (Innistrad: Midnight
// Hunt Commander):
//
//	"Indestructible
//	 Other legendary permanents you control have indestructible."
//
// Indestructible is printed. The second line is Avacyn, Angel of Hope's
// layer-6 grant narrowed to legendary permanents other than the Memorial
// itself, read live through the layered view so a permanent that gains or
// loses the legendary supertype gains or loses the grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a8abc5db-4e0b-488d-898a-8109690d5099",
		Name:            "Avacyn's Memorial",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static: []game.StaticAbility{{
			Label: "Other legendary permanents you control have indestructible.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.IsLegendary()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !keywordSliceContains(c.Abilities, "indestructible") {
					c.Abilities = append(c.Abilities, "indestructible")
				}
			},
		}},
	})
}
