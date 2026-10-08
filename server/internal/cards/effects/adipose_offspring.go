package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Adipose Offspring — Creature — Alien {3}{W}, 2/2:
//
//	"Emerge {5}{W} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When this creature enters, create a 2/2 white Alien creature token.
//	 If this creature's emerge cost was paid, instead create X of those
//	 tokens, where X is the sacrificed creature's toughness."
//
// Emerge is the shared alternative cost (ADR 0135 §4). The enters trigger
// reads the spell's cast record (CR 400.7d): which alternative cost it was
// cast for and which creature that cost sacrificed. Both are taken from
// the permanent as the trigger is harvested (ObjectSnapshot.AltCost and
// AltCostObjects), so the answer is the same if the Offspring has left
// the battlefield by the time the trigger resolves. X is the sacrificed
// creature's toughness as it last existed on the battlefield (CR 608.2h),
// its counters and anthems included, and a toughness below 0 makes no
// tokens. Put onto the battlefield without being cast, or cast for its
// mana cost, it makes one token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1e21e57e-3bc9-41a8-9746-57b583a5ad63",
		Name:             "Adipose Offspring",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Emerge("{5}{W}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Adipose Offspring — create 2/2 white Alien creature tokens", adiposeOffspringAliens),
		},
	})
}

// adiposeOffspringAliens makes one Alien, or, when the emerge cost was
// paid, as many as the sacrificed creature's toughness.
func adiposeOffspringAliens(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 1
	if obj := ctx.Trigger().Object; obj != nil && PaidEmerge(obj.AltCost) {
		n = 0
		if sacrificed := g.AltCostPermanentsForEffect(obj.AltCostObjects); len(sacrificed) > 0 {
			n = max(sacrificed[0].Toughness, 0)
		}
	}
	if n == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("2/2 white Alien"), N: n}.Apply(ctx)
}
