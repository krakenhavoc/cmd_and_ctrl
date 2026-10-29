package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thorin, King of Durin's Folk — Legendary Creature — Dwarf Noble
// {3}{R}{W}, 4/4:
//
//	"Whenever Thorin or another Dwarf you control enters, create a
//	 Treasure token.
//	 Other Dwarves you control get +1/+0 for each artifact token you
//	 control."
//
// The trigger is Balin, Loremaster's "Balin or another Dwarf you control
// enters" (a word of the source's text). The static is a Layer 7c pump
// whose amount is counted on every recompute, so a Treasure that comes
// or goes changes the Dwarves in the same beat; the Treasure Thorin's own
// trigger makes counts as soon as it exists.
//
// No simplification. The Hobbit (#4454).
func init() {
	Register(Spec{
		OracleID:     "6aabead3-3c81-40b8-8a8d-8f817094d84d",
		Name:         "Thorin, King of Durin's Folk",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Other Dwarves you control get +1/+0 for each artifact token you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype(source.TextSubtype("Dwarf")) // a word of the source's text (CR 612.2, #3984)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += artifactTokensControlledBy(g, source)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, balinOrAnotherDwarfEntered, "Thorin, King of Durin's Folk — create a Treasure",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}

// artifactTokensControlledBy counts the artifact tokens source's
// controller controls. It runs inside a layer recompute, so it walks the
// live battlefield slice (see sharedCreatureTypeCount).
func artifactTokensControlledBy(g *game.Game, source *game.Card) int {
	n := 0
	for i := range g.Battlefield.Cards {
		other := &g.Battlefield.Cards[i]
		if other.Controller == source.Controller && other.IsToken() && other.IsArtifact() {
			n++
		}
	}
	return n
}
