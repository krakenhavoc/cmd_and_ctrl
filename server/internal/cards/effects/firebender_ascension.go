package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firebender Ascension — Enchantment {1}{R}:
//
//	"When this enchantment enters, create a 2/2 red Soldier creature
//	 token with firebending 1.
//	 Whenever a creature you control attacking causes a triggered ability
//	 of that creature to trigger, put a quest counter on this
//	 enchantment. Then if it has four or more quest counters on it, you
//	 may copy that ability. You may choose new targets for the copy."
//
// The Soldier is a catalog token (ADR 0083, `token:soldier-firebending`)
// whose firebending 1 is its own attack trigger, Avatar Roku's Dragon's
// shape.
//
// The second ability watches triggered abilities being put on the stack
// (EventTriggerPlaced, Aboleth Spawn's seam), so it goes on the stack
// above the ability that caused it and its copy resolves first (the
// ruling). It triggers for an ability whose source is a creature you
// control and whose triggering event is that creature being declared
// as an attacker. Each such ability is its own trigger. As it resolves
// it puts a quest counter on the Ascension, then counts the counters on
// it (as it last existed, if it has left); at four or more you may copy
// the ability (CR 707.10), keeping its modes and with new targets
// allowed (the rulings).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "96cc63af-69d1-4493-b241-0cb244a9e48a",
		Name:         "Firebender Ascension",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Firebender Ascension — create a 2/2 red Soldier creature token with firebending 1",
				Do(CreateToken{Template: firebendingSoldierToken(), N: 1})),
			On(game.EventTriggerPlaced, firebenderAscensionAttackTrigger,
				"Firebender Ascension — put a quest counter on this enchantment, then you may copy that ability",
				firebenderAscensionBody),
		},
	})
}

// firebenderAscensionAttackTrigger is "a creature you control attacking
// causes a triggered ability of that creature to trigger": the placed
// item is a triggered ability of a creature, its cause that creature's
// attack, and the creature was yours as it attacked.
func firebenderAscensionAttackTrigger(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.StackItemID == uuid.Nil {
		return false
	}
	item := g.StackItemForEffect(ev.StackItemID)
	if item == nil || item.Kind != game.StackItemTriggered || item.Trigger == nil {
		return false
	}
	cause := item.Trigger.Event
	if cause.Kind != game.EventAttack || cause.CardID == uuid.Nil || cause.CardID != item.SourceCardID {
		return false
	}
	if obj := item.Trigger.Object; obj != nil {
		return obj.HasType("Creature") && obj.Controller == source.Controller
	}
	c, ok := g.LookupCardForEffect(cause.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller
}

func firebenderAscensionBody(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if info, ok := ctx.SourcePermanent(); ok && !info.Left {
		if err := (AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	info, ok := ctx.SourcePermanent()
	if !ok || info.Counters["quest"] < 4 {
		return nil
	}
	return MayChoice{
		Question: "Firebender Ascension — copy that ability? (You may choose new targets for the copy.)",
		Then:     firebenderAscensionCopyAnswer,
	}.Apply(ctx)
}

// firebenderAscensionCopyAnswer is the "yes": copy the ability that
// caused the trigger. Keyed (#4291).
var firebenderAscensionCopyAnswer = MayYes("confirm/firebender-ascension-copy", func(ctx *Context, _ game.ConfirmAnswer) error {
	return copyTheTriggeringAbility(ctx.Game, ctx.Item)
})

// firebendingSoldierToken is Firebender Ascension's 2/2 red Soldier with
// firebending 1.
func firebendingSoldierToken() game.Card { return tokenFromCatalog(printedFirebendingSoldierToken) }
