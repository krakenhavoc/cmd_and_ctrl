package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// r38_helpers.go — helpers roll-up 38 moved into the core so cards in two
// shards can share them (#5948, S82).

// R38ShuffleTargetGraveyardCardsIntoLibrary is "Shuffle any number of
// target cards from your graveyard into your library": every target
// still legal goes into its owner's library, then the controller's
// library is shuffled — once, and even when nothing was named, because
// the instruction is the shuffle. Perpetual Timepiece and The Bath Song
// use it.
func R38ShuffleTargetGraveyardCardsIntoLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
			return err
		}
	}
	return ShuffleLibrary{Player: item.Controller}.Apply(ctx)
}
