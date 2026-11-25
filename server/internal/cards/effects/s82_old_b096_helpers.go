package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b096_helpers.go — shared shapes of the Zendikar Rising batch 096
// cards (#5508).

// OldB096PartyOf is the number of creatures in `you`'s party (CR 700.8)
// as the effect resolves.
func OldB096PartyOf(g *game.Game, you uuid.UUID) int {
	return FinB9PartySize(g, you)
}

// OldB096BasicLandMatch is "a basic land card" / "basic land" as a card
// predicate (Murasa Rootgrazer).
func OldB096BasicLandMatch(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return IsBasicLand(c)
}

// OldB096CounterUnlessPaysParty is "Counter target <spell> unless its
// controller pays {1} plus an additional {1} for each creature in your
// party" as a spell's whole effect (Concerted Defense).
func OldB096CounterUnlessPaysParty(name string) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		t, ok := ctx.ClauseTarget(0)
		if !ok || t.Kind != game.TargetCard {
			return nil
		}
		cost := "{" + strconv.Itoa(1+OldB096PartyOf(ctx.Game, ctx.Controller())) + "}"
		return CounterUnlessPaid{
			StackID:  t.ID,
			Cost:     cost,
			Question: name + " — pay " + cost + " or your spell is countered",
		}.Apply(ctx)
	}
}
