package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titan's Strength — Instant {R} (EDHREC rank 4411):
//
//	"Target creature gets +3/+1 until end of turn. Scry 1."
//
// A combat trick that replaces the worst card in your next draw
// instead of replacing itself — the reason it beat out its many
// one-mana rivals in Theros limited and the reason it still shows up
// in low-curve red lists. Roadmap batch 42 (#449) files it under
// until-end-of-turn, which #279 shipped.
//
// Order matters and is printed: the pump happens, THEN the scry. The
// scry queues a prompt, so anything after it in a card's text has to
// ride Scry.Then; here there is nothing after it, which is why this
// card can put the pump first and the scry last as two plain
// statements.
//
// The scry happens whether or not the pump did. That is the printed
// card: a Titan's Strength whose target became illegal in response is
// countered for having no legal targets (CR 608.2b) and the scry does
// not happen either, but a Titan's Strength that resolves with a legal
// target always scries, and a resolved spell cannot lose the second
// sentence on account of the first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0206b34-68b4-4b1f-ae79-6e2d0432ad4b",
		Name:         "Titan's Strength",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				if err := (BoostUntilEOT{
					Target: item.Targets[0].ID,
					Power:  3, Toughness: 1,
					Label: "Titan's Strength — +3/+1 until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
