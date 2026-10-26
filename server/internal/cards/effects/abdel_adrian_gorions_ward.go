package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abdel Adrian, Gorion's Ward — Legendary Creature — Human Warrior
// {4}{W}, 4/4:
//
//	"When Abdel Adrian enters, exile any number of other nonland
//	 permanents you control until Abdel Adrian leaves the battlefield.
//	 Create a 1/1 white Soldier creature token for each permanent exiled
//	 this way.
//	 Choose a Background"
//
// The permanents are chosen, not targeted, as the trigger resolves (any
// number, none included). Each is exiled through ExileUntil{ThisLeaves},
// so all of them come back the moment Abdel leaves (CR 610.3), and an
// Abdel that left before the trigger resolved exiles nothing and makes
// no Soldiers (CR 610.3c). The Soldiers are counted from the exiles
// that really happened, so a pick that left in the meantime makes none.
// Choose a Background is deck construction (ADR 0144).
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	Register(Spec{
		OracleID:     "cab092f9-b7ff-43b9-935f-310869a4daf8",
		Name:         "Abdel Adrian, Gorion's Ward",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Abdel Adrian — exile other nonland permanents until he leaves, then a Soldier for each", abdelAdrianEnters),
		},
	})
}

func abdelAdrianEnters(g *game.Game, item *game.StackItem) error {
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && !c.IsLand() && c.InstanceID != item.SourceCardID {
			mine = append(mine, c.InstanceID)
		}
	}
	if len(mine) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Abdel Adrian — exile any number of other nonland permanents you control until he leaves the battlefield",
		Cards:    mine,
		Min:      0,
		Max:      0, // "any number"
		ThenKey:  abdelExilePick,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// abdelExilePick is the picks being exiled one by one, keyed (#4253): the
// asking item rides the prompt as data.
var abdelExilePick = askedPick("choose-cards/abdel-adrian-exile", func(ctx *Context, r game.CardsChosen) error {
	return abdelExileNext(ctx, r.Picked, 0)
})

// abdelExileNext exiles the picks one by one through each exile's
// continuation, counting the ones that landed, and makes the Soldiers
// once every pick has settled.
func abdelExileNext(ctx *Context, picked []uuid.UUID, exiled int) error {
	if len(picked) == 0 {
		if exiled == 0 {
			return nil
		}
		return CreateToken{Template: TokenCard("1/1 white Soldier"), N: exiled}.Apply(ctx)
	}
	id, rest := picked[0], picked[1:]
	return ExileUntil{
		Target:     id,
		ThisLeaves: true,
		Label:      "Abdel Adrian — returns when Abdel Adrian leaves the battlefield",
		Then: func(ctx *Context, ok bool) error {
			if ok {
				exiled++
			}
			return abdelExileNext(ctx, rest, exiled)
		},
	}.Apply(ctx)
}
