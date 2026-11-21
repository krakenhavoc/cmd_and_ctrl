package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etherium Spinner — Artifact Creature — Human Wizard {2}{U}, 2/1
// (Modern Horizons 2, #5501):
//
//	"Whenever you cast a spell with mana value 4 or greater, create a 1/1
//	 colorless Thopter artifact creature token with flying."
//
// Angry Rabble's cast trigger (WheneverYouCast over ManaValueGE(4)) making
// a Thopter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b501128b-2797-401b-8aec-923050081cae",
		Name:         "Etherium Spinner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ManaValueGE(4), "Etherium Spinner — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
