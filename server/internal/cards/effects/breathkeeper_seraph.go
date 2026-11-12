package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Breathkeeper Seraph — {4}{W}{W} Creature — Angel 4/4
// (Innistrad: Crimson Vow Commander, #5675):
//
//	"Flying, soulbond (You may pair this creature with another unpaired
//	 creature when either enters. They remain paired for as long as you
//	 control both of them.)
//	 As long as Breathkeeper Seraph is paired with another creature,
//	 each of those creatures has "When this creature dies, you may
//	 return it to the battlefield under its owner's control at the
//	 beginning of your next upkeep.""
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go). The
// second line is a layer-6 grant to both creatures of the pair; a dies
// trigger looks back at the creature as it last existed (CR 603.10a),
// when the two were still paired.
//
// The dies trigger sets up a delayed trigger (CR 603.7) for the next
// upkeep of the player who controlled the creature. As that resolves
// they may return the card, if it is still the object that died, in a
// graveyard (CR 400.7). It returns under its OWNER's control, whoever
// controlled it when it died. A token that died is gone and returns
// nothing. The returned creature is a new object and is unpaired.
//
// No simplification.
const breathkeeperSeraphGrant = "breathkeeper-seraph/return"

func init() {
	Register(Spec{
		OracleID:        "a41e632f-9d36-44d2-b1dd-e17df0d2486e",
		Name:            "Breathkeeper Seraph",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordSoulbond},
		Triggered:       Soulbond("Breathkeeper Seraph"),
		Grants: []AbilityGrant{{
			Key: breathkeeperSeraphGrant,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Breathkeeper Seraph — you may return this creature at the beginning of your next upkeep", breathkeeperSeraphSchedule),
			},
			Text: "When this creature dies, you may return it to the battlefield under its owner's control at the beginning of your next upkeep.",
		}},
		Static: []game.StaticAbility{
			PairedCreaturesHave("As long as Breathkeeper Seraph is paired with another creature, each of those creatures has \"When this creature dies, you may return it to the battlefield under its owner's control at the beginning of your next upkeep.\"",
				breathkeeperSeraphGrant),
		},
	})
}

// breathkeeperSeraphSchedule resolves the dies trigger: a delayed
// trigger at the controller's next upkeep, naming the card that died.
// Nothing is scheduled for a creature that is not that object in a
// graveyard (a token, or a card that has already moved on).
func breathkeeperSeraphSchedule(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Breathkeeper Seraph — you may return " + c.Name + " to the battlefield under its owner's control",
		Cards:              []uuid.UUID{id},
		Body:               breathkeeperSeraphReturnBody,
		Params:             game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// breathkeeperSeraphReturnBody is the delayed trigger: its controller
// may return the card <Object> names, if it is still that object in a
// graveyard. An on-disk key: never renamed, never reused.
var breathkeeperSeraphReturnBody = game.DelayedBody("breathkeeper-seraph/may-return-the-object",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if !objectStillInAGraveyard(g, p.Object) {
			return nil
		}
		c, _ := g.LookupCardForEffect(p.Object.ID)
		return MayChoice{
			Player:   item.Controller,
			Question: "Breathkeeper Seraph — return " + c.Name + " to the battlefield under its owner's control?",
			YesLabel: "Return it",
			NoLabel:  "Leave it",
			Then:     breathkeeperSeraphAnswer,
			Carry:    []uuid.UUID{p.Object.ID},
		}.Apply(NewContext(g, item))
	})

// breathkeeperSeraphAnswer is the "yes": the card (the carried ID)
// returns to the battlefield under its owner's control. Keyed.
var breathkeeperSeraphAnswer = MayYes("confirm/breathkeeper-seraph-return", func(ctx *Context, r game.ConfirmAnswer) error {
	return ReturnFromGraveyard{Target: carryAt(r, 0), Dest: game.ZoneBattlefield}.Apply(ctx)
})
