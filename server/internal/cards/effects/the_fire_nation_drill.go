package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Fire Nation Drill — Legendary Artifact — Vehicle {2}{B}{B}, 6/3:
//
//	"Trample
//	 When The Fire Nation Drill enters, you may tap it. When you do,
//	 destroy target creature with power 4 or less.
//	 {1}: Permanents your opponents control lose hexproof and
//	 indestructible until end of turn.
//	 Crew 2"
//
// The entry is Spider-Man, To the Rescue's shape: the "you may" is the
// parent trigger's optional prompt, tapping the Drill is the doing, and
// the reflexive trigger (CR 603.12) is created only if it was really
// tapped, so a Drill that is already tapped or has left gives nothing.
// Its target is chosen as it goes on the stack and rechecked as it
// resolves (CR 608.2b), so a creature pumped past 4 power survives.
//
// The {1} is Shadowspear's strip: a layer 6 removal over the permanents
// the ability's controller's opponents control as it resolves (CR
// 611.2c), so a permanent that enters or gains either keyword later
// keeps it (the rulings), and it lasts the turn even if the Drill
// leaves. Crew 2 is the ordinary CR 702.122 crew ability.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "3d4b9d71-4275-4c18-a4dc-9aad234eb92c",
		Name:            "The Fire Nation Drill",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("The Fire Nation Drill — you may tap it", fireNationDrillTapThenReflexive),
				"Tap The Fire Nation Drill? (When you do, destroy target creature with power 4 or less.)"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}: Permanents your opponents control lose hexproof and indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerRestrict},
				Cost:    ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return untilEndOfTurn(NewContext(g, item), uuid.Nil, OpponentControls(),
						"The Fire Nation Drill — opponents' permanents lose hexproof and indestructible",
						game.RemoveKeywordsMod("hexproof", "indestructible"))
				},
			},
			{
				Label:   "Crew 2",
				Cost:    CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("The Fire Nation Drill"),
			},
		},
	})
}

// fireNationDrillTapThenReflexive taps the Drill and, only if it did,
// creates the reflexive destroy trigger.
func fireNationDrillTapThenReflexive(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourcePermanent()
	if !ok || src.Left || src.Tapped || sourceIsNewObject(g, item) {
		return nil
	}
	if err := (TapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return WhenYouDo("The Fire Nation Drill — destroy target creature with power 4 or less", fireNationDrillDestroyBody).Apply(ctx)
}

var fireNationDrillDestroyBody = game.ReflexiveBody("the-fire-nation-drill/destroy", simpleBody(destroyFirstLegalCardTarget),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target creature with power 4 or less", PowerLE(4))
	}))
