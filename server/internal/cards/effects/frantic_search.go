package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frantic Search — Instant for {2}{U}:
//
//	"Draw two cards, then discard two cards. Untap up to three lands."
//
// Free if you untap three, which is why it shows up in storm decks
// and why it's here: two discard triggers for no net mana.
//
// "Up to three lands" is a choice the card makes trivial — they're
// your own tapped lands and untapping them is strictly good — so the
// sandbox untaps up to three rather than opening a picker. See
// untapUpToLands for when that stops being true.
func init() {
	Register(Spec{
		OracleID:     "16e015b2-f8a3-4b1a-80be-58a8f5fb5e8c",
		Name:         "Frantic Search",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You can't choose which lands untap — it automatically untaps up to three of your own tapped lands and can never untap another player's land."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := lootOne(ctx.Game, item, 2); err != nil {
				return err
			}
			return untapUpToLands(ctx.Game, item.Controller, 3)
		},
	})
}
