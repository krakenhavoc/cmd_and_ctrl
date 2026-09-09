package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kodama's Reach — Sorcery — Arcane {2}{G}:
//
//	"Search your library for up to two basic land cards, reveal those
//	cards, put one onto the battlefield tapped and the other into
//	your hand, then shuffle."
//
// Functionally identical to Cultivate (the Arcane subtype matters
// only to splice, which the engine has no notion of), so the
// composition is the same two sequential searches: the first with
// shuffle deferred so the second still sees library order, the
// second shuffling once at the end as the rules require.
func init() {
	Register(Spec{
		OracleID: "1593ea18-2f2f-4ab4-83fb-6ccc0bec8a90",
		Name:     "Kodama's Reach",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			if err := (SearchLibrary{
				Player:        controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       false,
				TappedOnEntry: true,
			}).Apply(ctx); err != nil {
				return err
			}
			return SearchLibrary{
				Player:    controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
			}.Apply(ctx)
		},
	})
}
