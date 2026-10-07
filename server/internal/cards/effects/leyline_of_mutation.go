package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leyline of Mutation — Enchantment {2}{G}{G}:
//
//	"If this card is in your opening hand, you may begin the game with
//	 it on the battlefield.
//	 You may pay {W}{U}{B}{R}{G} rather than pay the mana cost for
//	 spells you cast."
//
// Fist of Suns' static on an enchantment (ADR 0118 §3, #2163). The
// opening-hand clause (CR 103.6a) is Spec.OpeningHand (ADR 0133): the
// seat holding it is asked as the mulligan window closes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "caab67eb-65e7-4755-b116-6977e97f0844",
		Name:                    "Leyline of Mutation",
		Completeness:            CompletenessFull,
		OpeningHand:             BeginTheGameOnTheBattlefield(),
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayWUBRGForSpellsYouCast()},
	})
}
