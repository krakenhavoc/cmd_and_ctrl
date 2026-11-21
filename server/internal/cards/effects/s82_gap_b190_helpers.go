package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b190_helpers.go — shared bodies of GAP batch 190 (#5948), the
// Warhammer 40,000 Commander creatures.

// printedGapB190RobotToken is the 4/4 colorless Robot artifact creature with
// "This creature can't block." that Cybernetica Datasmith's second target
// player gets. A catalog template because the clause is a static the token
// always prints. Its slug is an on-disk identity: never renamed.
func printedGapB190RobotToken() TokenTemplate {
	return TokenTemplate{
		Slug: "robot-4-4-cant-block",
		Card: game.Card{
			Name:      "Robot",
			TypeLine:  "Token Artifact Creature — Robot",
			Power:     4,
			Toughness: 4,
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This creature can't block.",
	}
}

// GapB190RobotToken is that Robot's card value for CreateToken.
func GapB190RobotToken() game.Card { return TokenFromCatalog(printedGapB190RobotToken) }

// gapB190LuciusReturnBody is the delayed trigger's body for Lucius the
// Eternal: "When that creature leaves the battlefield, return this card from
// exile to the battlefield under its owner's control." The card returns only
// if it is still the object that was exiled, in exile (CR 400.7); a card that
// has since moved on stays where it is. A registered key (ADR 0041 phase 3).
var gapB190LuciusReturnBody = game.DelayedBody("lucius-the-eternal/return-from-exile", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(p.Object.ID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return ReturnFromExile{Target: p.Object.ID, Controller: c.Owner}.Apply(NewContext(g, item))
})

// gapB190LuciusExiled is the rest of Lucius's dies trigger, once the card has
// been exiled: the delayed trigger that waits for the chosen creature to leave
// the battlefield. The chosen creature rides the tail's Carry; the exiled card
// is the tail's settled ID, named with the epoch it has in exile.
var gapB190LuciusExiled = CardTail("lucius-the-eternal/exiled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        "Lucius the Eternal — return it when that creature leaves the battlefield",
		On:           []game.EventKind{game.EventLTB},
		Condition:    theListedObjectLeftCondition,
		Cards:        []uuid.UUID{o.Tail.Carry[0]},
		Body:         gapB190LuciusReturnBody,
		Params:       game.EffectParams{Object: game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}},
	})
	return nil
})

// GapB190LuciusDies is the body of Lucius's dies trigger: exile the card and
// choose the targeted creature. A target that is no longer legal removes the
// trigger (CR 608.2b), so the card stays in the graveyard.
func GapB190LuciusDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	t := ctx.ItemTail(gapB190LuciusExiled)
	t.Carry = []uuid.UUID{target}
	return ExileTarget{Target: item.SourceCardID, ThenTail: t}.Apply(ctx)
}
