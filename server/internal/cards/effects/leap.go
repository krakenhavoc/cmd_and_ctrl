package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leap — Instant {U} (EDHREC rank 4212):
//
//	"Target creature gains flying until end of turn.
//	 Draw a card."
//
// A one-mana cantrip that pushes a creature past a ground blocker, or
// saves one from a ground blocker, and replaces itself either way. It
// is played almost entirely by Voltron and "whenever you cast your
// second spell" decks, where a free-ish spell that draws is worth a
// slot on its own.
//
// Both halves are one effect of one spell, and that has a consequence
// worth stating: the card has ONE target, so if that creature has
// left or gained protection by the time Leap resolves, the spell does
// not resolve at all and the draw does not happen either (CR
// 608.2b). Writing the draw as an unconditional second statement
// would be a strictly better card — it would cantrip through a
// removal spell in response — and stronger than printed is the one
// direction this catalog does not ship (#259).
//
// The engine's fizzle check upstream is what enforces that; the guard
// below is only for a target slot that never got filled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e89a3ce0-6b38-4326-a7af-8575af371baa",
		Name:         "Leap",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (GrantKeywordUntilEOT{
				Target:   item.Targets[0].ID,
				Keywords: []string{"flying"},
				Label:    "Leap — flying until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
