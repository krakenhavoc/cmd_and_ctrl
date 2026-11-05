package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain Jean-Luc Picard — Legendary Creature — Human Officer {2}{W},
// 2/3 (Star Trek, #4973):
//
//	"Other creatures you control get +1/+1.
//	 Whenever you cast a creature or Spacecraft spell, create a 1/1 red
//	 Officer creature token.
//	 Tap seven untapped Officers you control: Search your library for a
//	 Spacecraft card, put it onto the battlefield with seven charge
//	 counters on it, then shuffle."
//
// The anthem is layer 7c and spares Picard. The cast trigger reads the
// spell's types and subtypes as cast, so a Spacecraft that is not yet a
// creature counts.
//
// The cost is a TapOthers cost of seven, with Picard allowed to pay
// (he is an Officer and the cost does not say "other"). It has no {T}
// symbol, so summoning-sick Officers may be tapped for it (CR 302.6).
// The Spacecraft enters with its seven charge counters on it (CR
// 614.1c, through the entry-counter pipeline), so a station threshold
// at 7+ is met as it arrives.
//
// No simplification.
func init() {
	tapSeven := TapAnUntapped("seven untapped Officers you control", HasSubtype("Officer"))
	tapSeven.TapOthers.Count = 7
	Register(Spec{
		OracleID:     "04624f75-5981-4607-aa8f-706554f826e1",
		Name:         "Captain Jean-Luc Picard",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			Label:    "Other creatures you control get +1/+1.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Creature(), HasSubtype("Spacecraft")),
				"Captain Jean-Luc Picard — create a 1/1 red Officer creature token",
				Do(CreateToken{Template: TokenCard("1/1 red Officer"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap seven untapped Officers you control: Search your library for a Spacecraft card, put it onto the battlefield with seven charge counters on it, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    tapSeven,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:             item.Controller,
					Predicate:          func(c game.Card) bool { return c.HasSubtype("Spacecraft") },
					Dest:               game.ZoneBattlefield,
					Limit:              1,
					Shuffle:            true,
					EntersWithCounters: map[string]int{"charge": 7},
					Reason:             "Captain Jean-Luc Picard — search for a Spacecraft card",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
