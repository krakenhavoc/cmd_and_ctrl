package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wylie Duke, Atiin Hero — Legendary Creature — Human Ranger {1}{G}{W},
// 4/2 (Outlaws of Thunder Junction, #4468):
//
//	"Vigilance
//	 Whenever Wylie Duke becomes tapped, you gain 1 life and draw a
//	 card."
//
// Compassionate Healer's trigger: "becomes tapped" is the tap event and
// the attack that taps it (ThisBecameTapped), so a vigilance attack,
// which taps nothing, is not one. Any other tap counts: a crew or saddle
// payment, a convoke tap, an opponent's tap effect. The life gain and the
// draw are footprinted steps in printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b3ac27d6-7e18-404d-9af2-b03897a5f1d9",
		Name:            "Wylie Duke, Atiin Hero",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Wylie Duke, Atiin Hero — you gain 1 life and draw a card",
				Do(GainLife{Amount: 1}, DrawCards{N: 1})),
		},
	})
}
