package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Death Begets Life — Sorcery {5}{B}{G}{U} (Tarkir: Dragonstorm, #4463):
//
//	"Destroy all creatures and enchantments. Draw a card for each
//	 permanent destroyed this way."
//
// Fumigate's shape over two types: one sweep (a creature that is also an
// enchantment is destroyed once and counted once) whose keyed tail draws
// a card for each permanent that actually went to a graveyard, so an
// indestructible permanent, or one a replacement exiled, is not counted
// (CR 701.8a); a permanent destroyed that ends up elsewhere (a finality
// counter, a commander sent to the command zone) still was destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53e86135-3b24-4618-bcc0-af4d81e672dd",
		Name:         "Death Begets Life",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Or(Creature(), Enchantment()), ThenTail: ctx.Tail(deathBegetsLifeDraw)}.Apply(ctx)
		},
	})
}

// deathBegetsLifeDraw is "draw a card for each permanent destroyed this way"
// (Death Begets Life).
var deathBegetsLifeDraw = MassTail("mass/death-begets-life-draw", func(ctx *Context, _ []game.Card, moved int) error {
	if moved <= 0 {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: moved}.Apply(ctx)
})
