package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orthodoxy Enforcer — Creature — Phyrexian Cleric {3}{W}, 2/4 (Phyrexia:
// All Will Be One):
//
//	"Vigilance
//	 This creature gets +2/+0 as long as you control two or more
//	 artifacts."
//
// A layer 7c modification on the Enforcer alone, gated on its controller's
// artifact count read through the layered view on every recompute. It is
// not an artifact itself unless something makes it one, in which case it
// counts itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e8875dc-72a4-47a4-bb99-ca817c177b82",
		Name:            "Orthodoxy Enforcer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:    "Gets +2/+0 as long as you control two or more artifacts.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfCreature(target, g, source) && stickerArtifactsControlledBy(g, source.Controller) >= 2
			},
			Apply: pumpBy(2, 0),
		}},
	})
}
