package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valor's Flagship — Legendary Artifact — Vehicle {4}{W}{W}{W}, 7/7:
//
//	"Flying, first strike, lifelink
//	 Crew 3
//	 Cycling {X}{2}{W} ({X}{2}{W}, Discard this card: Draw a card.)
//	 When you cycle this card, create X 1/1 colorless Pilot creature
//	 tokens with "This token saddles Mounts and crews Vehicles as though
//	 its power were 2 greater.""
//
// Shark Typhoon's shape: the cycling ability draws the card and a
// separate trigger, fired from the graveyard (CR 702.29c), makes the
// tokens. EventCycle carries the X the cost was announced with, so
// countering the cycling ability leaves the Pilots standing, as
// printed.
//
// No simplification.
func init() {
	cycling := Cycling("{2}{W}")
	cycling.Label = "Cycling {X}{2}{W} ({X}{2}{W}, Discard this card: Draw a card.)"
	cycling.Cost = Plus(ManaCost("{X}{2}{W}"), DiscardThis())
	Register(Spec{
		OracleID:        "a3023de2-61e7-4f46-aa4a-09d08ee7d939",
		Name:            "Valor's Flagship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "lifelink"},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Valor's Flagship"),
			},
			cycling,
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventCycle, Self, "Valor's Flagship — cycled: create X Pilot tokens, X the cycling cost's X",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return CreateToken{
						Controller: item.Controller,
						Template:   PilotToken(),
						N:          item.Trigger.Event.Amount,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}
