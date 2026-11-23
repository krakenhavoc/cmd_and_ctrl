package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Hussar — Creature — Human Knight {3}{W}{U}, 4/3 (Ravnica Remastered,
// #5948):
//
//	"Flying
//	 When this creature enters, untap all creatures you control.
//	 Forecast — Tap two untapped white and/or blue creatures you control, Reveal
//	 this card from your hand: Draw a card. (Activate only during your upkeep and
//	 only once each turn.)"
//
// The enters trigger is Village Bell-Ringer's. Forecast (CR 702.57a) is an
// activated ability that works from the hand, whose cost is two tapped
// creatures (Persistent Petitioners' shape: not the {T} symbol, CR 302.6, so
// creatures that arrived this turn pay it, and they do not target) and showing
// this card (AbilityCost.RevealThis): the card is revealed as the cost is paid
// (CR 601.2h), so every opponent sees it before they may respond, and it stays
// in the hand (CR 701.20b). The timing is the upkeep of the card's owner and
// once each turn, counted per object (CR 602.5d, CR 400.7).
//
// No simplification.
func init() {
	const label = "Forecast — Tap two untapped white and/or blue creatures you control, Reveal this card from your hand: Draw a card."
	effects.Register(effects.Spec{
		OracleID:        "21b41dd6-bff7-4e3d-aa28-551a30194225",
		Name:            "Sky Hussar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sky Hussar — untap all creatures you control", effects.B34UntapAllCreaturesYouControl),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Cost: effects.Plus(game.AbilityCost{RevealThis: true}, game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count: 2,
				Filter: effects.TargetPermanent("two untapped white and/or blue creatures you control",
					effects.Creature(), effects.Or(effects.OfColor("W"), effects.OfColor("U"))),
				Label: "two untapped white and/or blue creatures you control",
			}}),
			Condition: effects.AllConditions(effects.DuringYourUpkeep(), effects.OncePerTurnActivation(label)),
			Effect:    effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
