package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Majestic Genesis — Sorcery {6}{G}{G}:
//
//	"Reveal the top X cards of your library, where X is the greatest
//	 mana value of a commander you own on the battlefield or in the
//	 command zone. You may put any number of permanent cards from among
//	 them onto the battlefield. Put the rest on the bottom of your
//	 library in a random order."
//
// Genesis Wave keyed to the commander. X is read at resolution
// (CR 608.2h) over the commanders the caster OWNS — a stolen opposing
// commander on the caster's side of the table does not count, and the
// caster's own commander under an opponent's control does — in two
// places only: the battlefield and the command zone. A commander in
// hand, library or graveyard contributes nothing, as printed.
//
// The pick is #745's PutFromLibraryOntoBattlefield over "any number of
// permanent cards", entering as one simultaneous batch; the rest go to
// the bottom in an order drawn from the game's seeded RNG.
//
// Mana value is the printed cost's. A commander whose cost the engine
// cannot read reports zero, which can only shrink X.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15039c85-31e2-4a2b-82f8-2f8270ff9a00",
		Name:         "Majestic Genesis",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := greatestOwnedCommanderManaValue(ctx.Game, item.Controller)
			revealed := ctx.Game.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), x,
				"Majestic Genesis — revealed from the top of the library")
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    revealed,
				Optional: true,
				Label:    "Majestic Genesis — put any number of permanent cards onto the battlefield",
				Then:     PutRestOnBottomInRandomOrder,
			}.Apply(ctx)
		},
	})
}

// greatestOwnedCommanderManaValue is "the greatest mana value of a
// commander you own on the battlefield or in the command zone"; zero
// when there is none.
func greatestOwnedCommanderManaValue(g *game.Game, player uuid.UUID) int {
	best := 0
	consider := func(c game.Card) {
		if c.IsCommander && c.Owner == player && c.ManaValue() > best {
			best = c.ManaValue()
		}
	}
	for _, c := range g.Battlefield.Cards {
		consider(c)
	}
	if p := g.PlayerByIDForEffect(player); p != nil && p.Command != nil {
		for _, c := range p.Command.Cards {
			consider(c)
		}
	}
	return best
}
