package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S30CreatureCardsIn is the number of creature cards in `player`'s
// graveyard and the greatest power among them (0 when there are none).
func cmdr2S30CreatureCardsIn(g *game.Game, player uuid.UUID) (count, greatestPower int) {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0, 0
	}
	for _, c := range p.Graveyard.Cards {
		if !c.IsCreature() {
			continue
		}
		count++
		if c.Power > greatestPower {
			greatestPower = c.Power
		}
	}
	return count, greatestPower
}

// cmdr2S30NontokenCreatureYouControlDied is "whenever a nontoken creature you
// control dies" for a source that is not itself the creature: a creature card
// that went from the battlefield to a graveyard, last controlled by the
// source's controller as it died (CR 603.10a).
func cmdr2S30NontokenCreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller
}

// cmdr2S30OldStickfingersReveal is Old Stickfingers' cast trigger: reveal
// cards from the top of the controller's library until X creature cards are
// revealed (X on item.Params.Amount), put those into the graveyard, then put
// the rest of the revealed cards on the bottom in a random order.
func cmdr2S30OldStickfingersReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hits := revealUntilCount(ctx, player, func(c game.Card) bool { return c.IsCreature() },
		item.Params.Amount, "Old Stickfingers — revealed until X creature cards")
	if len(run) == 0 {
		return nil
	}
	if err := restIntoGraveyard(g, hits); err != nil {
		return err
	}
	return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, cardsStillInALibrary(g, run))
}
