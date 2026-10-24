package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tempt with Discovery — Sorcery {3}{G} (EDHREC rank 869):
//
//	"Tempting offer — Search your library for a land card and put it
//	 onto the battlefield. Each opponent may search their library for
//	 a land card and put it onto the battlefield. For each opponent
//	 who searches a library this way, search your library for a land
//	 card and put it onto the battlefield. Then each player who
//	 searched a library this way shuffles."
//
// The lands deck's Gaea's Cradle tutor, and the first tempting offer
// in the catalog. Four searches at most on a full table, and every
// one of them a real chooser prompt over a real library, ANY land
// card (not just basics), entering UNTAPPED — as printed.
//
// Runs as a chain of SearchLibrary continuations rather than a
// loop: the caster's opening search, then each opponent's optional
// search in seat order, and after every opponent who takes a land,
// another search for the caster — b07TemptWithDiscoveryOffer. The
// chain is strictly sequential so no two prompts are ever open over
// the same library (Cultivate's rule), which is also how the offer
// is made in paper: to each opponent in turn.
//
// Each opponent is asked first whether to search (a confirm prompt);
// "searches a library this way" is that yes, so an opponent who
// searches and takes no land still earns the caster a land. Each search
// shuffles as it finishes rather than all at the end; nothing can
// observe the difference.
func init() {
	Register(Spec{
		OracleID:     "4baa6145-216e-476b-b178-aaaa1e633701",
		Name:         "Tempt with Discovery",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b07TemptWithDiscoveryCasterSearch(ctx, ctx.Opponents())
		},
	})
}
