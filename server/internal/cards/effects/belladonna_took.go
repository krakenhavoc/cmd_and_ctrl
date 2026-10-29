package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Belladonna Took — Legendary Creature — Halfling Citizen {1}{W}, 2/2
// (The Hobbit, #4454):
//
//	"Whenever a token you control enters, you gain 1 life if this is
//	 the first time this ability has resolved this turn. If it's the
//	 second time, draw a card. If it's the third time, put a +1/+1
//	 counter on each creature you control."
//
// The count is Elrond, Lord of Rivendell's: the per-turn tally of this
// Belladonna's resolutions of this ability, which includes the one
// resolving (Game.ResolvedThisTurn), per object (CR 400.7). The fourth
// and later resolutions do nothing. Any token counts, not just a
// creature token, and a token that enters under another player's control
// does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "011da9c5-aa8a-4fa0-b1f2-62b9f3760476",
		Name:         "Belladonna Took",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, belladonnaTokenYouControlEntered, belladonnaLabel, belladonnaTookResolves),
		},
	})
}

const belladonnaLabel = "Belladonna Took — gain 1 life, then draw a card, then a +1/+1 counter on each creature you control"

// belladonnaTokenYouControlEntered is "a token you control enters".
func belladonnaTokenYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && IsToken(c)
}

// belladonnaTookResolves is the ability's body: which of its three
// sentences applies is how many times it has resolved this turn.
func belladonnaTookResolves(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	switch g.ResolvedThisTurn(item.SourceCardID, belladonnaLabel) {
	case 1:
		return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
	case 2:
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	case 3:
		return b11PutCounterOnEachCreatureYouControl(g, item)
	}
	return nil
}
