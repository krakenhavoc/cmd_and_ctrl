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
// opening-hand clause (CR 103.6a) is the one every Leyline in this
// catalog leaves out: there is no pre-game window, so it is cast for
// {2}{G}{G}. That is weaker than printed, which is the only direction a
// simplification may go.
func init() {
	Register(Spec{
		OracleID:     "caab67eb-65e7-4755-b116-6977e97f0844",
		Name:         "Leyline of Mutation",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You can't begin the game with it on the battlefield from your opening hand — it has to be cast.",
		},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{PayWUBRGForSpellsYouCast()},
	})
}
