package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farseek — Sorcery {1}{G}:
//
//	"Search your library for a Plains, Island, Swamp, or Mountain
//	card, put it onto the battlefield tapped, then shuffle."
//
// Note the printed text says card TYPES, not "basic" — in paper
// Farseek fetches a Hallowed Fountain or a Bayou, which is most of
// why it's played over Rampant Growth in three-plus colours.
//
// Sandbox simplification: this fetches a BASIC land that isn't a
// Forest, not any land with those subtypes. Narrower than paper, and
// deliberately so — matching the real clause needs subtype parsing
// that admits duals, and getting that half-right would silently
// fetch illegal cards. The narrower version is always a legal play.
// It also technically admits Wastes, which prints none of the four
// types; harmless in practice since a Wastes deck isn't casting
// Farseek, and noted rather than special-cased.
func init() {
	Register(Spec{
		OracleID:     "495e52e6-4c2b-4574-9474-eadbdcc8b4ac",
		Name:         "Farseek",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Only basic lands are found — it cannot fetch dual lands such as Hallowed Fountain or Bayou, and it can find a Wastes."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     IsBasicLandExcept("forest"),
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Farseek — a Plains, Island, Swamp or Mountain",
			}.Apply(ctx)
		},
	})
}
