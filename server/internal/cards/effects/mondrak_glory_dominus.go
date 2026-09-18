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
// Sandbox simplification, declared, and the same one its cycle
// sibling Solphim, Mayhem Dominus carries: the activated ability is
// not offered. Its {W/P} Phyrexian mana would need a pay-2-life
// alternative on an ABILITY cost, which the engine has only for
// spells, and the indestructible counter is not a counter the state
// checks read.
func init() {
	Register(Spec{
		OracleID:     "fe83087d-c6c1-40be-9295-baaa1c6b2db1",
		Name:         "Mondrak, Glory Dominus",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The \"{1}{W/P}{W/P}, Sacrifice two other artifacts and/or creatures\" ability that puts an indestructible counter on Mondrak can't be activated."},
		Replacements: []game.ReplacementEffect{
			TokensDoubled("Mondrak, Glory Dominus: double tokens"),
		},
	})
}
