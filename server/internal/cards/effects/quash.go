package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quash — Instant {2}{U}{U}:
//
//	"Counter target instant or sorcery spell. Search its controller's
//	 graveyard, hand, and library for all cards with the same name as
//	 that spell and exile them. Then that player shuffles."
//
// The spell's name and controller are read before it is countered, while
// it is still on the stack: a split card cast as one half has only that
// half's name (CR 709.3b). Then the counter puts the card into its
// owner's graveyard, where the search finds it with the rest: every card
// with that name in the graveyard is exiled, a public zone having no fail
// to find (CR 701.23b), and the caster may leave matches in the hand and
// the library, which are hidden. The library is shuffled whatever was
// taken (extractionSearch, #4033). A spell that can't be countered stays
// on the stack and the search still happens; a target that has left the
// stack makes Quash do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "255131a2-616c-4c8d-a36a-14e5b5ab345e",
		Name:         "Quash",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target instant or sorcery spell", Or(Instant(), Sorcery())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			spell := ctx.Game.StackItemForEffect(target)
			c, found := ctx.Game.LookupCardForEffect(target)
			if spell == nil || !found {
				return nil
			}
			victim, names := spell.Controller, extractionNames(c)
			if err := (CounterTarget{StackID: target}).Apply(ctx); err != nil {
				return err
			}
			return extractionSearch(ctx, "Quash", victim, names, true)
		},
	})
}
