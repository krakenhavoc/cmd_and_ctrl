package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saheeli, Radiant Creator — Legendary Creature — Human Artificer
// {1}{G}{U}{R}, 4/4:
//
//	"Whenever you cast an Artificer or artifact spell, you get {E} (an energy
//	 counter).
//	 At the beginning of combat on your turn, you may pay {E}{E}{E}. When you
//	 do, create a token that's a copy of target permanent you control, except
//	 it's a 5/5 artifact creature in addition to its other types and has
//	 haste. Sacrifice it at the beginning of the next end step."
//
// The first ability reads the spell's types as it was cast. The second is a
// "you may pay" with a reflexive trigger (CR 603.12): the payment is made as
// the combat trigger resolves, and the copy's target is chosen as the
// reflexive trigger goes on the stack. The copy takes the permanent's
// copiable values (CR 707.2) except for a base 5/5, the added artifact and
// creature types and haste (CR 707.9), and is sacrificed at the next end step
// (Reflection of Kiki-Jiki's delayed sacrifice).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20f602b3-814f-406c-b7eb-2d1fc98f8cf1",
		Name:         "Saheeli, Radiant Creator",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(Or(Artifact(), HasSubtype("Artificer"))),
				"Saheeli, Radiant Creator — you get {E}", Do(GetEnergy{N: 1})),
			AtBeginningOfYourCombat("Saheeli, Radiant Creator — you may pay {E}{E}{E}",
				mayPayEnergyThen("Saheeli, Radiant Creator", 3, "create a token copy of a permanent you control", saheeliRadiantCreatorPaid)),
		},
	})
}

// saheeliRadiantCreatorPaid is "you may pay {E}{E}{E}. When you do": the
// energy was paid, so the reflexive trigger is created. An on-disk key.
var saheeliRadiantCreatorPaid = energyPaid("pay-unless/saheeli-radiant-creator-copy",
	func(g *game.Game, item *game.StackItem) error {
		return WhenYouDo("Saheeli, Radiant Creator — create a token that's a copy of target permanent you control, except it's a 5/5 artifact creature with haste; sacrifice it at the beginning of the next end step",
			saheeliRadiantCreatorCopyBody).Apply(NewContext(g, item))
	})

// saheeliRadiantCreatorCopyBody is the reflexive trigger's body. An on-disk
// key.
var saheeliRadiantCreatorCopyBody = game.ReflexiveBody("saheeli-radiant-creator/copy",
	simpleBody(saheeliRadiantCreatorCopy),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target permanent you control", YouControl())
	}))

// saheeliRadiantCreatorCopy makes the hasty 5/5 artifact creature copy of the
// chosen permanent and schedules its sacrifice.
func saheeliRadiantCreatorCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var copyOf uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			copyOf = t.ID
			break
		}
	}
	if copyOf == uuid.Nil {
		return nil
	}
	cursor := b25LastEventSeq(g)
	if err := (CreateTokenCopy{
		Controller: item.Controller,
		Copy:       copyOf,
		N:          1,
		Except: func(tok *game.Card) {
			tok.TypeLine = cmdrB23WithCardTypesInAddition(tok.TypeLine, "Artifact", "Creature")
			tok.Power, tok.Toughness = 5, 5
			tok.VariableToughness = false
			tok.PrintedPTKnown = true
			TokenCopyGainsHaste(tok)
		},
	}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Saheeli, Radiant Creator — sacrifice the token",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
