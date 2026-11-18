package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riders of Gavony — Creature — Human Knight {2}{W}{W}, 3/3:
//
//	"Vigilance
//	 As this creature enters, choose a creature type.
//	 Human creatures you control have protection from creatures of the
//	 chosen type."
//
// The type is the permanent's NamedTribe (Adaptive Automaton's as-enters
// prompt). The grant is a Layer 6 static over the Humans you control
// (changelings included, through the layered view) that adds
// "protection from creatures of type <T>": a creature source with that
// type, never a non-creature object that merely carries it. Until the
// type is chosen the grant gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60dabe9e-4872-4aec-9e8e-35358b3b7dd5",
		Name:            "Riders of Gavony",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		AsEnters:        ChooseCreatureTypeAsEnters("Riders of Gavony"),
		Static: []game.StaticAbility{{
			Label: "Human creatures you control have protection from creatures of the chosen type.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return source.NamedTribe != "" &&
					TribeFilter{Tribes: []string{"Human"}, YoursOnly: true}.Matches(target, g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				if source.NamedTribe == "" {
					return
				}
				appendKeywordsTo(c, []string{game.ProtectionFromCreaturesOfType(source.NamedTribe)})
			},
		}},
	})
}
