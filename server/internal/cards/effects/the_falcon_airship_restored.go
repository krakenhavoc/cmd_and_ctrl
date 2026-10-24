package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Falcon, Airship Restored — Legendary Artifact — Vehicle {2}{B},
// 4/3 (deck request #3655):
//
//	"Flying
//	 Whenever The Falcon deals combat damage to a player, you may
//	 sacrifice it. When you do, return target creature card from your
//	 graveyard to the battlefield.
//	 {4}{B}: Return this card from your graveyard to the battlefield
//	 tapped.
//	 Crew 2"
//
// The damage trigger is the "may do X, when you do Y" pair (CR 603.12):
// the sacrifice is the condition, and the return is a second trigger
// whose target is chosen once it is on the stack, so the Falcon itself
// is never a legal pick (it is a Vehicle, not a creature card, and is
// in the graveyard by then only as an artifact card). The graveyard
// ability is the Reassembling Skeleton shape, activated from the
// graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9f4093a6-83f1-4c1d-8b03-8fe15647fb11",
		Name:            "The Falcon, Airship Restored",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("The Falcon, Airship Restored — you may sacrifice it to return a creature card from your graveyard",
				falconSacrificeOffer),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 2",
				Cost:    CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("The Falcon, Airship Restored"),
			},
			{
				Label:   "{4}{B}: Return this card from your graveyard to the battlefield tapped.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{4}{B}"),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  returnThisFromGraveyardTapped,
			},
		},
	})
}

// falconSacrificeOffer asks whether to sacrifice The Falcon; the
// reflexive return is raised only after it has gone.
func falconSacrificeOffer(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Player:   item.Controller,
		Question: "The Falcon — sacrifice it to return a creature card from your graveyard to the battlefield?",
		Then:     falconAnswer,
	}.Apply(NewContext(g, item))
}

// falconAnswer is the "yes": sacrifice The Falcon, then raise the reflexive
// return once it has gone. Keyed (#4291).
var falconAnswer = MayYes("confirm/the-falcon-sacrifice", func(ctx *Context, _ game.ConfirmAnswer) error {
	item := ctx.Item
	// "Sacrifice it": a Falcon that left and came back is a new object, not
	// "it" (CR 400.7).
	if sourceIsNewObject(ctx.Game, item) {
		return nil
	}
	return ctx.Game.SacrificeThenForEffect(item.SourceCardID, item.SourceCardID, func(g *game.Game, sacrificed bool) error {
		if !sacrificed {
			return nil
		}
		return ReflexiveTrigger{
			Label: "The Falcon, Airship Restored — return target creature card from your graveyard to the battlefield",
			Body:  falconReturnBody,
		}.Apply(NewContext(g, item))
	})
})
