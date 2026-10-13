package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Amulet // Primal Wellspring — Artifact {4} // Land (#3213, the
// Kuja deck):
//
//	Primal Amulet
//	  "Instant and sorcery spells you cast cost {1} less to cast.
//	   Whenever you cast an instant or sorcery spell, put a charge
//	   counter on this artifact. Then if there are four or more charge
//	   counters on it, you may remove those counters and transform it."
//	Primal Wellspring
//	  "({T}: Add one mana of any color. When that mana is spent to cast
//	   an instant or sorcery spell, copy that spell and you may choose
//	   new targets for the copy.)" — see primal_wellspring.go.
//
// Brass's Tunnel-Grinder's shape: the counter goes through the counter
// window (a Doubling Season doubles it), so the four-counter check waits
// for it to land as a continuation, and "remove those counters and
// transform it" removes what is there and turns the same permanent over
// in place. The "may" is a real yes/no, asked only when there are four.
// The discount is the Medallion's, over instants and sorceries.
//
// No simplification on this face; the back face carries the caveat.
func init() {
	Register(Spec{
		OracleID:     primalAmuletOracleID,
		Name:         "Primal Amulet",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Primal Amulet — put a charge counter on it; at four, you may remove them and transform it",
				primalAmuletCharge),
		},
	})
}

// primalAmuletOracleID is shared with the back face, which registers
// under it plus "#1" (game.CatalogKey).
const primalAmuletOracleID = "8e4d0da0-c7d8-4a20-9bfd-02c1331a7a49"

// primalAmuletCharge is the cast trigger, in printed order.
func primalAmuletCharge(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	source := item.SourceCardID
	return g.AddCounterThenForEffect(source, game.CounterCharge, 1, func(g *game.Game, _ int) error {
		if b39CountersOn(g, source, game.CounterCharge) < 4 {
			return nil
		}
		return MayChoice{
			Question: "Primal Amulet — remove its charge counters and transform it?",
			YesLabel: "Transform",
			NoLabel:  "Keep it",
			OnYes: func(ctx *Context) error {
				// "Those counters": whatever it holds now.
				n := b39CountersOn(ctx.Game, source, game.CounterCharge)
				if n <= 0 {
					return nil
				}
				return ctx.Game.AddCounterThenForEffect(source, game.CounterCharge, -n,
					func(g *game.Game, _ int) error {
						return TransformThis{}.Apply(NewContext(g, item))
					})
			},
		}.Apply(NewContext(g, item))
	})
}
