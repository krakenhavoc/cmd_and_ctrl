package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b331_helpers.go — card-side helpers of set OLD, batch 331 (Dragons
// of Tarkir: Dash, Megamorph, S82, #5551).

// ForItsDashCost passes on a spell cast for a dash cost (game.CostQuery.AltDash,
// CR 702.109a): the "dash costs" of "Dash costs you pay cost {2} less". The same
// creature cast for its mana cost, or for any other alternative cost, does not
// pass.
func ForItsDashCost() CostPredicate {
	return func(q game.CostQuery) bool { return q.AltDash }
}

// DashCostsYouPayCostLess is Warbringer's "Dash costs you pay cost {N} less
// (as long as this creature is on the battlefield)": a reduction of generic
// mana (CR 118.7a) on a dash cost its controller pays, and on nothing else. A
// Spec.CostModifiers entry, so it lasts exactly as long as its source is on the
// battlefield.
func DashCostsYouPayCostLess(n int) game.CostModifier {
	return CostsLess(n, "Dash costs you pay cost {2} less.", YourSpell(), ForItsDashCost())
}

// EnteredCreatureGetsPlusTwoPlusTwo is "that creature gets +2/+2 until end of
// turn" for the creature whose entry triggered the ability (Ambuscade Shaman),
// applied only while it is still on the battlefield as the ability resolves
// (Ogre Battledriver's read).
func EnteredCreatureGetsPlusTwoPlusTwo(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		info, ok := ctx.TriggeringPermanent()
		if !ok || info.Left {
			return nil
		}
		return BoostUntilEOT{Target: ctx.Trigger().Object.Ref().ID, Power: 2, Toughness: 2, Label: name + " — +2/+2 until end of turn"}.Apply(ctx)
	}
}

// ExileThisThenReturnItFaceDown is "Exile this creature, then return it to the
// battlefield face down under its owner's control." (Shorecrasher Elemental):
// the Flicker shape with a face-down entry (CR 708.2a: the card is a new
// object, CR 400.7, and enters face down as a 2/2 with no abilities, so no
// enters trigger runs). The exile is the continuation's gate, as Flicker's is:
// a card that did not reach exile is not returned. A Shorecrasher that left and
// came back before the ability resolves is a different object and is left alone.
func ExileThisThenReturnItFaceDown(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	return ExileTarget{
		Target:   item.SourceCardID,
		ThenTail: WithCarry(ctx.ItemTail(s82OldB331ReturnFaceDown), item.SourceCardID),
	}.Apply(ctx)
}

var s82OldB331ReturnFaceDown = CardTail("old-b331/return-face-down", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 || len(o.Tail.Carry) == 0 {
		return nil
	}
	// Nil controller: under its owner's control.
	if _, err := ctx.Game.ReturnFromExileFaceDownForEffect(o.Tail.Carry[0], [16]byte{}, false, nil); err != nil && !errors.Is(err, game.ErrCardNotFound) {
		return err
	}
	return nil
})
