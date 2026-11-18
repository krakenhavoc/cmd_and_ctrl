package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise of the Ants — Sorcery {4}{G}{G} (Midnight Hunt, #5497):
//
//	"Create two 3/3 green Insect creature tokens. You gain 2 life.
//	 Flashback {6}{G}{G} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Two Insects, then the life, as one resolution; Flashback is the shared
// constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d9d33dcd-3352-478e-afc4-40f2ff369c39",
		Name:             "Rise of the Ants",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 2, CreatureTokens: 2},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{6}{G}{G}")},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("3/3 green Insect"), N: 2},
			GainLife{Amount: 2},
		),
	})
}
