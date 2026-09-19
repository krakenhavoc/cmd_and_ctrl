package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revitalize — Instant {1}{W} (EDHREC rank 4242):
//
//	"You gain 3 life.
//	 Draw a card."
//
// A cantrip that gains life, which is the whole card: it replaces
// itself, so a lifegain deck can run it as a free trigger for
// Archangel of Thune, Well of Lost Dreams, Exemplar of Light or
// Drogskol Reaver — the last of which is in this same batch and turns
// it into two cards.
//
// Order matters and is printed: the life comes first, so the lifegain
// trigger it is really being cast for goes on the stack above nothing
// and the draw has already happened by the time that trigger
// resolves. Writing the draw first would let a "whenever you draw"
// payoff see the card before the lifegain payoff sees the life, which
// is a different card in any deck running both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1385b03-cb4b-4812-857f-7421f1df39af",
		Name:         "Revitalize",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GainLife{Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
