package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pestered Wellguard — Creature — Merfolk Soldier {3}{U}, 3/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Whenever this creature becomes tapped, create a 1/1 blue and black
//	 Faerie creature token with flying."
//
// Compassionate Healer's trigger ("becomes tapped" is the tap event and the
// attack that taps it, so a vigilance attack is not one) with the Faerie
// token Lorwyn Eclipsed prints.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0dbfd40-ecc5-493b-976b-bc3f62137fcf",
		Name:         "Pestered Wellguard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Pestered Wellguard — create a 1/1 blue and black Faerie creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 blue and black Faerie with flying"), N: 1})),
		},
	})
}
