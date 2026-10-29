package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrowin of Clan Undurr — Legendary Creature — Dwarf Cleric
// {2}{W}{B}, 3/3:
//
//	"When Barrowin enters, venture into the dungeon.
//	 Whenever Barrowin attacks, return up to one creature card with mana
//	 value 3 or less from your graveyard to the battlefield if you've
//	 completed a dungeon."
//
// The attack trigger's "if" ends the sentence, so it is not an
// intervening "if" (CR 603.4): the ability always triggers, and the
// condition is read as it resolves. The card is chosen then, not
// targeted, so a creature card that entered the graveyard in response
// may be chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c16d5daf-4995-4342-815c-a6b29d594f26",
		Name:         "Barrowin of Clan Undurr",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersVenture("Barrowin of Clan Undurr"),
			WheneverThisAttacks("Barrowin of Clan Undurr — return up to one creature card with mana value 3 or less from your graveyard to the battlefield if you've completed a dungeon",
				barrowinReturn),
		},
	})
}

// barrowinReturn is the attack trigger's effect.
func barrowinReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := item.Controller
	if !YouCompletedADungeon(g, you) {
		return nil
	}
	fits := And(Creature(), ManaValueLE(3))
	cands := graveyardCardIDs(ctx, you, func(c game.Card) bool { return fits(g, you, c) })
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     you,
		Source:      item.SourceCardID,
		Question:    "Barrowin of Clan Undurr — return up to one creature card with mana value 3 or less to the battlefield",
		Cards:       cands,
		Min:         0,
		Max:         1,
		LiveZone:    game.ZoneGraveyard,
		Destination: game.ChooseOntoBattlefield,
		ThenKey:     barrowinReturnPick,
		Asked:       game.NewAskedItem(item),
	})
	return nil
}

// barrowinReturnPick is the chosen creature card returning to the
// battlefield, keyed (#4253): the asking item rides the prompt as data.
var barrowinReturnPick = askedPick("choose-cards/barrowin-of-clan-undurr-return", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
