package effects

// Doubling Cube — Artifact {2} (EDHREC rank 3544):
//
//	"{3}, {T}: Double the amount of each type of unspent mana you
//	 have."
//
// The big-mana deck's second engine. A MANA ability (the printed
// ruling: it could add mana, has no target and is not a loyalty
// ability — CR 605.1a), so it never uses the stack and can be
// activated while paying for a spell. The Cabal Coffers shape: a
// mana component in the cost and a ProducedFunc that reads the
// board after the cost is paid — here the controller's own pool,
// with the {3} already gone from it, which is the printed order
// (pay three, then double what is left). One plain slot of the same
// colour per token in the pool; a restricted token (Delighted
// Halfling's) is doubled by an unrestricted one, as the printed card
// does. An empty pool after paying adds nothing.
//
// Not auto-tappable, like every ability with a mana component — the
// player floats the {3} and clicks. See signets.go for the reasoning.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9afd8f12-0796-4500-aaa3-10b4a46ef6ec",
		Name:         "Doubling Cube",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true, Mana: "{3}"},
			ProducedFunc: b33DoubleUnspentMana,
			Label:        "{3}, {T}: Double the amount of each type of unspent mana you have",
		}},
	})
}
