package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b3_helpers.go — the small pieces the third batch of the Jumpstart
// deck requests (#4892) needed and the catalog did not have yet.

// jmpB3YouGainedThreeOrMoreLife is the trigger-time half of Silversmote
// Ghoul's intervening "if you gained 3 or more life this turn" (CR 603.4):
// the per-turn tally counts only gains, so a gain undone by a loss still
// counts.
func jmpB3YouGainedThreeOrMoreLife(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return b15LifeGainedThisTurn(g, source.Controller) >= 3
}

// jmpB3YouHaveDrawnTwoOrMoreCards is Tome Anima's "as long as you've drawn
// two or more cards this turn", read off the controller's per-turn tally
// whenever a block is checked.
func jmpB3YouHaveDrawnTwoOrMoreCards(g *game.Game, you uuid.UUID, _ game.Card) bool {
	return g.TurnTallyFor(you).CardsDrawn >= 2
}

// jmpB3CreatureCardsInOpponentsGraveyards counts the creature cards in
// the graveyards of every player still in the game other than `you`
// (Wight of Precinct Six). A card is a creature card by its printed type
// line, since a card in a graveyard has no layers applied.
func jmpB3CreatureCardsInOpponentsGraveyards(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				n++
			}
		}
	}
	return n
}
