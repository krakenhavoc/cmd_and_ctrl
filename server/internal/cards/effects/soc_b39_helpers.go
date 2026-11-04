package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b39_helpers.go — shared pieces for the thirty-ninth Secrets of
// Strixhaven batch (main set sos, #4456).

// socB39LookAtFiveAndSplit is Zimone's Experiment's look: the top five cards
// of the controller's library, privately, then a choose_cards prompt for up
// to two creature and/or land cards among them. With none of either among
// the five there is nothing to reveal, so all five go to the bottom in a
// random order at once.
func socB39LookAtFiveAndSplit(ctx *Context) error {
	g := ctx.Game
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 5)
	candidates := libraryCardsMatching(g, player, looked, Or(Creature(), Land()))
	if len(candidates) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, cardsStillInALibrary(g, looked))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Zimone's Experiment — reveal up to two creature and/or land cards from among them (lands enter tapped, creatures go to your hand)",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		LiveZone: game.ZoneLibrary,
		ThenKey:  socB39ZimonesExperimentRevealed,
		Carry:    looked,
	})
	return nil
}

// socB39ZimonesExperimentRevealed is the answer: the chosen cards are
// revealed while they are still in the library, then "put all land cards
// revealed this way onto the battlefield tapped" and, behind it as its keyed
// tail, "put all creature cards revealed this way into your hand" and the rest
// on the bottom in a random order. A card that is both a land and a creature is
// a land card: the land sentence is first and it leaves the library, so the
// creature sentence no longer finds it. A card that left the library in the
// meantime (the prompt's live-zone check) is simply not revealed. An on-disk
// key: never renamed, never reused.
var socB39ZimonesExperimentRevealed = game.RegisterChooseCardsThen("choose-cards/zimones-experiment-reveal", func(g *game.Game, r game.CardsChosen) error {
	player := r.Chooser
	var lands, creatures []uuid.UUID
	for _, id := range cardsStillInALibrary(g, r.Picked) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if c.IsLand() {
			lands = append(lands, id)
		} else {
			creatures = append(creatures, id)
		}
	}
	revealTaken(g, player, r.Source, "Zimone's Experiment — revealed from among the top five cards", append(append([]uuid.UUID(nil), lands...), creatures...))
	return PutFromLibraryOntoBattlefield{
		Player:    player,
		Cards:     r.Carry,
		Match:     b10OneOf(lands),
		All:       true,
		Tapped:    true,
		ThenCarry: creatures,
		ThenN:     len(creatures),
		ThenTail:  socB39ZimonesExperimentCreatures,
	}.Apply(keyedContext(g, player, r.Source))
})

// socB39ZimonesExperimentCreatures is the creature sentence and the rest: the
// tail's first N carried cards are the revealed creature cards, the others the
// five looked at. The creatures go to the hand, and whatever of the five is
// still in the library (the unrevealed cards) goes to the bottom in a random
// order (takeRestBottomRandom). An on-disk key.
var socB39ZimonesExperimentCreatures = CardTail("zimones-experiment/creatures-then-rest", func(ctx *Context, o game.TailOutcome) error {
	carry := o.Tail.Carry
	n := o.Tail.N
	if n > len(carry) {
		n = len(carry)
	}
	return TakeFromLibraryToHand{
		Player: o.Tail.Player,
		Cards:  carry[n:],
		Match:  b10OneOf(carry[:n]),
		All:    true,
		Key:    takeRestBottomRandom,
	}.Apply(ctx)
})
