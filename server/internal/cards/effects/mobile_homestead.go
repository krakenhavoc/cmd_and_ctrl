package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mobile Homestead — Artifact — Vehicle {2}, 3/3 (Outlaws of Thunder
// Junction, #4468):
//
//	"This Vehicle has haste as long as you control a Mount.
//	 Whenever this Vehicle attacks, look at the top card of your library.
//	 If it's a land card, you may put it onto the battlefield tapped.
//	 Crew 2 (Tap any number of creatures you control with total power 2
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// The haste is a KeywordGrant on itself, read live: it has haste only
// while you control a Mount, which matters the turn it is crewed. The
// attack trigger is Explorer's Scope's look: a land card on top may go
// onto the battlefield tapped, and any other card stays where it is.
// Crew is Adventurer's Airship's row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b71239fb-d7dc-4bac-b8aa-26b454409344",
		Name:         "Mobile Homestead",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID &&
				b40ControlsA(g, source.Controller, HasSubtype(source.TextSubtype("Mount")))
		}, "haste")},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Mobile Homestead — look at the top card of your library; you may put a land card onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 1),
						Match:    Land(),
						Max:      1,
						Optional: true,
						Tapped:   true,
						Label:    "Mobile Homestead — you may put the land onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Mobile Homestead"),
		}},
	})
}
