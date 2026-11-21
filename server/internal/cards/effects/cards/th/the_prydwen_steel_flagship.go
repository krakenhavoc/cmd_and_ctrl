package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Prydwen, Steel Flagship — Legendary Artifact — Vehicle {4}{W}{W},
// 6/6 (Fallout, #5948):
//
//	"Flying
//	 Whenever another nontoken artifact you control enters, create a 2/2
//	 white Human Knight creature token with "This token gets +2/+2 as
//	 long as an artifact entered the battlefield under your control this
//	 turn."
//	 Crew 2"
//
// A Vehicle with the usual crew ability (CrewCost, CrewEffect). The
// trigger is the other artifact's entry (EventETB, read as it entered,
// CR 603.6a): not the Prydwen itself, not a token, any artifact type
// (an artifact creature counts), under the Prydwen's controller. The
// Knight is the printed token (GapB61HumanKnightToken): its static is on
// the turn the artifact entered, so the new Knight is a 4/4 right away
// and shrinks to 2/2 at the end of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6ea8085-5ba9-43b6-994d-04be35c5793b",
		Name:            "The Prydwen, Steel Flagship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsArtifact() && !c.IsToken() && c.Controller == source.Controller
			}, "The Prydwen — create a 2/2 white Human Knight creature token", effects.Do(effects.CreateToken{Template: effects.GapB61HumanKnightToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("The Prydwen"),
		}},
	})
}
