package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phoenix Fleet Airship — Artifact — Vehicle {2}{B}{B}, 4/4 (Avatar: The
// Last Airbender, #4459):
//
//	"Flying
//	 At the beginning of your end step, if you sacrificed a permanent
//	 this turn, create a token that's a copy of this Vehicle.
//	 As long as you control eight or more permanents named Phoenix Fleet
//	 Airship, this Vehicle is an artifact creature.
//	 Crew 1"
//
// The end-step trigger is an intervening "if" (CR 603.4) on the turn
// tally of permanents its controller sacrificed (Count Nefaria's read),
// checked as it triggers and again as it resolves; the tally only rises
// within a turn. The token copies the Airship's copiable values (CR
// 707.2), so it has these abilities too and makes copies of itself on
// later turns; an Airship that left in response is copied as it last
// was on the battlefield (CR 608.2h, the third ruling).
//
// The third line is a layer-4 static on the Airship itself: while its
// controller controls eight or more permanents named Phoenix Fleet
// Airship (the Airship among them, a token copy counts), it is an
// artifact creature with its printed 4/4. Crew is the shared ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ffe71c0-acb5-4ba9-8c2f-719330e1ea69",
		Name:            "Phoenix Fleet Airship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{selfIsArtifactCreatureWhile(
			"As long as you control eight or more permanents named Phoenix Fleet Airship, this Vehicle is an artifact creature.",
			func(g *game.Game, source *game.Card) bool { return phoenixFleetAirshipsControlled(g, source) >= 8 })},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.TurnTallyFor(source.Controller).PermanentsSacrificed > 0
			}), "Phoenix Fleet Airship — create a token that's a copy of this Vehicle",
				func(g *game.Game, item *game.StackItem) error {
					if g.TurnTallyFor(item.Controller).PermanentsSacrificed == 0 {
						return nil
					}
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Phoenix Fleet Airship"),
		}},
	})
}

// phoenixFleetAirshipsControlled counts the permanents named Phoenix
// Fleet Airship that the source's controller controls. It is read inside
// the layer pass, so it walks the battlefield as it stands rather than
// asking for fresh characteristics; a name is a copiable value, already
// on the card (CR 707.2).
func phoenixFleetAirshipsControlled(g *game.Game, source *game.Card) int {
	n := 0
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == source.Controller && game.HasName(*c, "Phoenix Fleet Airship") {
			n++
		}
	}
	return n
}
