package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	applies := b16OtherCreaturesYouControlOfSubtype("Merfolk")
	Register(Spec{
		OracleID:        "bf76ba01-e06f-4451-a4b4-cf548b43eafc",
		Name:            "Mist Dancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			b16Anthem(applies, 1, 0),
			b16GrantKeywords(applies, "flying"),
		},
		Activated: []ActivatedAbility{Encore("{5}{U}{U}")},
	})
}
