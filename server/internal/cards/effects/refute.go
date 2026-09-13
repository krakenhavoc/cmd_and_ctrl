package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Refute — Instant {1}{U}{U} (EDHREC rank 1899):
//
//	"Counter target spell. Draw a card, then discard a card."
//
// Cancel with a loot stapled on. The counter is the ordinary
// CounterTarget; the loot is lootOne, which draws first and queues
// the discard second so the drawn card is a legal discard, as in
// paper. If the target left the stack in response, Refute is
// countered by game rules before it resolves (CR 608.2b) and the
// loot never happens — also as in paper.
//
// Same declared gap as every counterspell in the catalog: only a
// SPELL can be targeted — an activated or triggered ability on the
// stack is not offered (the targeting spec has no ability mode).
func init() {
	Register(Spec{
		OracleID:     "eb1dfa29-7371-4cb6-bfa2-16f7820b69be",
		Name:         "Refute",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only spells can be countered — an activated or triggered ability on the stack can't be picked."},
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return lootOne(ctx.Game, item, 1)
		},
	})
}
