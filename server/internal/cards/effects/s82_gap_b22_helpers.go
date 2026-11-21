package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82GapB22WhileYouControlOtherSnow is the AppliesTo of "<this creature> has
// <keyword> as long as you control at least <n> other snow permanents": the
// source alone, while its controller controls n or more snow permanents
// besides it, read after the layers (CR 205.4a, so a snow permanent made so by
// an effect counts).
func s82GapB22WhileYouControlOtherSnow(n int) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.InstanceID != source.InstanceID {
			return false
		}
		others := 0
		for _, id := range permanentsControlledByMatching(g, source.Controller, Snow()) {
			if id != source.InstanceID {
				others++
			}
		}
		return others >= n
	}
}

// s82GapB22DefenderRevealsTopLandToHand is Goblin Guide's attack trigger:
// "defending player reveals the top card of their library. If it's a land
// card, that player puts it into their hand."
func s82GapB22DefenderRevealsTopLandToHand(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
	if !defendingPlayerStillIn(g, defender) {
		return nil
	}
	for _, id := range g.RevealTopOfLibraryForEffect(defender, item.SourceCardID, 1, "Goblin Guide — defending player reveals the top card of their library") {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			g.TakeFromLibraryToHandForEffect(defender, []uuid.UUID{id})
		}
	}
	return nil
}
