package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valley Rally — Instant {2}{R} (EDHREC rank 13246):
//
//	"Gift a Food (You may promise an opponent a gift as you cast this
//	 spell. If you do, they create a Food token before its other
//	 effects. It's an artifact with "{2}, {T}, Sacrifice this token:
//	 You gain 3 life.")
//	 Creatures you control get +2/+0 until end of turn. If the gift was
//	 promised, target creature you control gains first strike until
//	 end of turn."
//
// The CR 702.174m card: the unpromised spell has no target at all, and
// the promised one has one. "The spell's controller chooses those
// targets only if the gift was promised" is a clause list that differs
// by the cost, so the promised clause is declared with `.Instead(…)`
// on a card whose printed Targets is empty — the engine asks for the
// target only on a promised cast, and the client opens its picker only
// then.
//
// The Food is the token catalog's (ADR 0083), so the opponent can
// crack it.
//
// The +2/+0 is BoostUntilEOT over "creatures you control", evaluated
// once at resolution (CR 611.2c). The first strike reads the target
// back through ctx.LegalTargets(). A promised Rally whose one target
// left in response does not resolve at all — CR 608.2b counters a
// spell whose every target is illegal, the +2/+0 included — which is
// the printed card's real downside for promising.
func init() {
	Register(Spec{
		OracleID:     "5b919920-b1b6-499b-a7fe-c630778a7831",
		Name:         "Valley Rally",
		Completeness: CompletenessFull,
		Gift:         GiftAFood().Instead(TargetCreature("target creature you control", YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (BoostUntilEOT{
				Match: And(Creature(), YouControl()),
				Power: 2,
				Label: "Valley Rally — creatures you control get +2/+0",
			}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.GiftPromised() {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"first strike"},
					Label:    "Valley Rally — first strike",
				}.Apply(ctx)
			}
			return nil
		},
	})
}
