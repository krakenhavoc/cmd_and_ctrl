package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mondrak, Glory Dominus — Legendary Creature — Phyrexian Horror
// {2}{W}{W}, 4/4 (EDHREC rank 428):
//
//	"If one or more tokens would be created under your control, twice
//	 that many of those tokens are created instead.
//	 {1}{W/P}{W/P}, Sacrifice two other artifacts and/or creatures:
//	 Put an indestructible counter on Mondrak."
//
// The white Dominus, and the fourth printing of Doubling Season's
// token half — the same declared effect as Parallel Lives and
// Anointed Procession in a body, which is what makes a token deck's
// "×2, ×2, ×2" an ordinary CR 616.1 chain rather than a special case
// (#762).
//
// Its wording is the one the doubling cycle prints without the "under
// your control ... it creates" framing, but the clause is the same:
// only tokens created under Mondrak's controller's control double.
//
// The activated ability is Zopandrel's and Drivnod's: Phyrexian mana
// ({W/P} paid with {W} or 2 life, announced with the activation) plus a
// sacrifice of exactly two other artifacts and/or creatures paid at
// announce, and the indestructible counter is a keyword counter
// (CR 122.1b).
func init() {
	Register(Spec{
		OracleID:     "fe83087d-c6c1-40be-9295-baaa1c6b2db1",
		Name:         "Mondrak, Glory Dominus",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			TokensDoubled("Mondrak, Glory Dominus: double tokens"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W/P}{W/P}, Sacrifice two other artifacts and/or creatures: Put an indestructible counter on Mondrak.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}{W/P}{W/P}"), SacrificeAnotherN(2, "two other artifacts and/or creatures", Or(Artifact(), Creature()))),
			Effect:  putIndestructibleCounterOnSource(),
		}},
	})
}
