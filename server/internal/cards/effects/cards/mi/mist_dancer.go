package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mist Dancer — Creature — Merfolk Wizard {4}{U}, 3/3 (Lost Caverns of
// Ixalan Commander):
//
//	"Flying
//	 Other Merfolk you control get +1/+0 and have flying.
//	 Encore {5}{U}{U} (...)"
//
// Lyra Dawnbringer's lord shape over Merfolk: a layer 7c +1/+0 and a layer
// 6 flying grant to every other Merfolk you control, read through the
// layers so a changeling counts. Encore is the shared keyword (encore.go);
// the token copies are Merfolk lords as well.
//
// No simplification.
func init() {
	applies := effects.B16OtherCreaturesYouControlOfSubtype("Merfolk")
	effects.Register(effects.Spec{
		OracleID:        "bf76ba01-e06f-4451-a4b4-cf548b43eafc",
		Name:            "Mist Dancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.B16Anthem(applies, 1, 0),
			effects.B16GrantKeywords(applies, "flying"),
		},
		Activated: []effects.ActivatedAbility{effects.Encore("{5}{U}{U}")},
	})
}
