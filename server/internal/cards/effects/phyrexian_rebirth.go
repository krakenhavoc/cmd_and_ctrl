package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Rebirth — Sorcery {4}{W}{W} (EDHREC rank 4281):
//
//	"Destroy all creatures, then create an X/X colorless Phyrexian
//	 Horror artifact creature token, where X is the number of
//	 creatures destroyed this way."
//
// A board wipe that leaves you the biggest creature on the table. Six
// mana is the price of the asymmetry, and the token's size is the
// whole reason to cast it into a full board rather than an empty one.
//
// "Destroyed this way" is the count the `Then` continuation is handed,
// and it counts what LANDED: an indestructible creature is not
// destroyed and does not grow the Horror, and a commander whose owner
// takes the CR 903.9 offer went to the command zone rather than dying,
// so it does not either (CR 400.7). It counts EVERY creature
// destroyed, yours included — the Horror is bigger for the creatures
// you lost, which is exactly how the card is meant to be cast.
//
// The token is built by hand rather than read out of the token table:
// its power and toughness are decided at resolution, and a table keyed
// on "X/X colorless Phyrexian Horror" would need a row per possible
// size. Zero creatures destroyed still makes a token, a 0/0 that the
// state-based-action sweep immediately puts in the graveyard — the
// printed outcome, not a special case.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ef3c75d-6af2-4ea0-b98d-96c5d7d3af58",
		Name:         "Phyrexian Rebirth",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: Creature(),
				Then: func(ctx *Context, _ []game.Card, destroyed int) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   b41PhyrexianHorrorToken(destroyed),
						N:          1,
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
