package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Overseer — {3}{W}{W} Creature — Angel 5/3 (Innistrad, S82 OLD b445,
// #5572):
//
//	"Flying
//	 As long as you control a Human, this creature has hexproof and
//	 indestructible."
//
// Two layer 6 grants on the source alone, gated on its controller controlling a
// Human, counted after the layers so a changeling counts.
//
// No simplification.
func init() {
	whileHuman := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return effects.SelfOnly(target, g, source) &&
			len(effects.PermanentsControlledByMatching(g, source.Controller, effects.Subtype(source.TextSubtype("Human")))) > 0
	}
	effects.Register(effects.Spec{
		OracleID:        "ec22386d-42fd-40f9-b800-a043e80faabe",
		Name:            "Angelic Overseer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.KeywordGrant(whileHuman, "hexproof"),
			effects.KeywordGrant(whileHuman, "indestructible"),
		},
	})
}
