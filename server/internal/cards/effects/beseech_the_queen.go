package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beseech the Queen — Sorcery {2/B}{2/B}{2/B}:
//
//	"Search your library for a card with mana value less than or
//	 equal to the number of lands you control, reveal it, put it into
//	 your hand, then shuffle."
//
// A tutor whose ceiling is your own land count, which is the whole
// design: it is a Demonic Tutor that is dead on turn one and
// unrestricted by turn six.
//
// The land count is read WHEN THE SPELL RESOLVES, not when it was
// cast, and the predicate closes over it rather than recomputing per
// candidate — a search that re-counted lands mid-scan would be
// reading a board nothing can change during a resolution anyway, but
// the closure makes the timing explicit.
//
// # Declared sandbox simplification: THE MONOCOLOUR HYBRID COST
//
// {2/B} means "two generic OR one black", three times over, and the
// card's mana value is 6 regardless. Whether the cost parser handles
// monocolour hybrid is a casting-cost question, not an effect one —
// this file describes what happens on resolution, and that part is
// complete. If the printed cost does not parse, the engine's existing
// cost-warning path handles it the same way it handles every other
// unparsed cost; no behaviour is special-cased here.
func init() {
	Register(Spec{
		OracleID: "cf94cafc-527e-4b27-8a28-7807435aaccf",
		Name:     "Beseech the Queen",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			lands := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == controller && c.IsLand() {
					lands++
				}
			}
			return SearchLibrary{
				Player:    controller,
				Predicate: func(c game.Card) bool { return manaValueOfCard(c) <= lands },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Beseech the Queen — a card with mana value ≤ your land count",
			}.Apply(ctx)
		},
	})
}
