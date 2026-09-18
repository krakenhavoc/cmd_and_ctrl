package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archetype of Courage — Enchantment Creature — Human Soldier
// {1}{W}{W}, 2/2 (EDHREC rank 4481):
//
//	"Creatures you control have first strike.
//	 Creatures your opponents control lose first strike and can't have
//	 or gain first strike."
//
// The white Archetype. Half of it is an anthem-shaped keyword grant
// and half of it is a keyword LOCKOUT, and the lockout is the half
// that wins games: in a combat-heavy deck, giving your whole team
// first strike while no opponent can ever have it means your blocks
// never trade and your attacks always do.
//
// The grant is implemented and is real: a layer 6 ability grant over
// every creature its controller has, the Archetype included (the
// clause has no "other"). It re-evaluates on every recompute, so a
// creature that enters afterwards gets first strike too, as printed.
//
// DECLARED SIMPLIFICATION (weaker than printed): the second line is
// NOT implemented. "Lose first strike AND CAN'T HAVE OR GAIN IT" is
// a dependency-ordered lockout: the losing half is an ordinary layer
// 6 removal, but the "can't gain" half has to beat every later
// timestamp in that same layer, and the engine applies layer 6
// strictly in timestamp order (CR 613.7). A plain removal static
// would therefore work only against first strike that was already
// there when the Archetype arrived, and would silently stop working
// against a first striker played afterwards — which is worse than
// not implementing it, because it looks like it works.
//
// Nothing here is stronger than the printed card: opponents keep a
// first strike they would have lost, which is a loss for the
// Archetype's controller in every case.
func init() {
	Register(Spec{
		OracleID:     "79b48704-480d-4905-b87a-40b127894670",
		Name:         "Archetype of Courage",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only the first half works: your creatures get first strike. Opponents' creatures keep any first strike they have and can still gain it.",
		},
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "first strike"),
		},
	})
}
