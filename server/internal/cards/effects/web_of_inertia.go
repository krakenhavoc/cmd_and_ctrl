package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Web of Inertia — Enchantment {2}{U}:
//
//	"At the beginning of combat on each opponent's turn, that player may
//	 exile a card from their graveyard. If the player doesn't, creatures
//	 they control can't attack you this turn."
//
// The opponent whose turn it is chooses up to one card from their own
// graveyard; the prompt stops the table, so declare attackers waits for
// the answer. A player with an empty graveyard can't exile anything and
// is restricted outright. The restriction is the this-turn attack grant
// of ADR 0063's 2026-10-10 amendment (#2719), scoped to the player
// alone: the ruling is that a creature that can't attack you can still
// attack a planeswalker you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a96d24f1-7670-4cb2-9971-5dd1b4c67957",
		Name:         "Web of Inertia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, false), ByAnOpponent),
				"Web of Inertia — that player may exile a card from their graveyard", webOfInertiaCombat),
		},
	})
}

// webOfInertiaCombat asks the active opponent, and restricts them if
// they exile nothing.
func webOfInertiaCombat(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	opponent := item.Trigger.Event.Actor
	cards := echoOfEonsGraveyardCardIDs(g, opponent)
	if len(cards) == 0 {
		webOfInertiaRestrict(g, item, opponent)
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opponent,
		FromPlayer: opponent,
		Source:     item.SourceCardID,
		Question:   "Web of Inertia — exile a card from your graveyard, or your creatures can't attack its controller this turn",
		Cards:      cards,
		Min:        0,
		Max:        1,
		ThenKey:    webOfInertiaPicked,
		LiveZone:   game.ZoneGraveyard,
		Carry:      []uuid.UUID{opponent},
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// webOfInertiaPicked is the opponent's answer, keyed (#4253): the card they
// exile, or the restriction when they exile none. The opponent rides the
// prompt as the first of its carry, and the item as data.
var webOfInertiaPicked = askedPick("choose-cards/web-of-inertia-exile-or-restrict", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		if len(r.Carry) > 0 {
			webOfInertiaRestrict(ctx.Game, ctx.Item, r.Carry[0])
		}
		return nil
	}
	return ExileTarget{Target: r.Picked[0]}.Apply(ctx)
})

// webOfInertiaRestrict is "creatures they control can't attack you this
// turn". Caller holds g.mu.
func webOfInertiaRestrict(g *game.Game, item *game.StackItem, opponent uuid.UUID) {
	g.GrantCantAttackPlayerThisTurnForEffect(opponent, item.Controller,
		game.CantAttackScope{PlayerOnly: true}, "Web of Inertia", item.SourceCardID)
}
