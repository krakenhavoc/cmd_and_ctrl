package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frantic Search — Instant for {2}{U}:
//
//	"Draw two cards, then discard two cards. Untap up to three lands."
//
// Free if you untap three, which is why it shows up in storm decks
// and why it's here: two discard triggers for no net mana.
//
// "Up to three lands" prints no "target" and no "you control", so on
// paper it is a resolution-time choice among every land at the
// table. UntapUpToLands is that choice: a prompt over every tapped
// land at the table, any controller's, queued after the loot.
func init() {
	Register(Spec{
		OracleID:     "16e015b2-f8a3-4b1a-80be-58a8f5fb5e8c",
		Name:         "Frantic Search",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := lootOne(ctx.Game, item, 2); err != nil {
				return err
			}
			return UntapUpToLands{N: 3, Question: "Frantic Search — untap up to three lands"}.Apply(ctx)
		},
	})
}
