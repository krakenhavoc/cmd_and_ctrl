package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kaya's Wrath — Sorcery {W}{W}{B}{B} (EDHREC rank 4282):
//
//	"Destroy all creatures. You gain life equal to the number of
//	 creatures you controlled that were destroyed this way."
//
// Wrath of God with a rider that pays you for having the worse board,
// which is the Orzhov joke: the double-double cost means you are
// casting it off a real mana base, and the life back matters most in
// the games where you wiped your own aristocrats away.
//
// The rider is why this is DestroyAllMatching with a `Then` rather
// than the fire-and-forget form. `swept` is narrowed to the cards that
// actually LANDED in a graveyard, so an indestructible creature is
// neither destroyed nor counted, and a commander whose owner sends it
// to the command zone instead is not counted either (CR 400.7) — the
// clause says "destroyed this way", and a permanent that survived was
// not. The count is not knowable on the line after the sweep for
// exactly that reason, which is what the continuation buys.
//
// "Creatures you CONTROLLED" reads the pre-move copies: by the time
// the life is gained the cards are in graveyards, where Controller
// still holds who had them on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc8de6c7-c69d-4add-8f25-825d945874f9",
		Name:         "Kaya's Wrath",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{
				Match: Creature(),
				Then: func(ctx *Context, swept []game.Card, _ int) error {
					return GainLife{
						Player: item.Controller,
						Amount: controllersOf(swept)[item.Controller],
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
