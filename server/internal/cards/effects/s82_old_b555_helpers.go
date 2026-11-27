package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB555ReturnTargetCreatureHasteIfGoblin is Warren Pilferers' "return
// target creature card from your graveyard to your hand. If that card is a
// Goblin card, this creature gains haste until end of turn": the target
// still legal as the trigger resolves (CR 608.2b) is read for its types in the
// graveyard, where the card was, and then returned. The haste is for the
// Pilferers only while it is still the permanent that entered (CR 400.7).
func S82OldB555ReturnTargetCreatureHasteIfGoblin(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	card, found := g.LookupCardForEffect(id)
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
		return err
	}
	if !found || !card.HasSubtypeWord("Goblin") || !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return ThisGainsKeywordUntilEndOfTurn("haste", "Warren Pilferers — haste until end of turn")(g, item)
}
