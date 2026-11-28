package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b624_helpers.go — shared pieces of Judgment batch 624 (#5611),
// activated abilities.

// OldB624ShroudAndCantBeBlocked is Cephalid Inkshrouder's ability: "This
// creature gains shroud until end of turn and can't be blocked this turn." Both
// halves are applied to the permanent the ability came from; one that left, or
// came back as a new object, in response gets neither (CR 400.7, CR 608.2b).
func OldB624ShroudAndCantBeBlocked(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		if err := ThisGainsKeywordUntilEndOfTurn("shroud", name+" — shroud until end of turn")(g, item); err != nil {
			return err
		}
		return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
			Label: name + " — can't be blocked this turn"}.Apply(NewContext(g, item))
	}
}
