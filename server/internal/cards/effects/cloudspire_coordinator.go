package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudspire Coordinator — Creature — Human Pilot {R}{W}, 3/1:
//
//	"When this creature enters, scry 2.
//	 {T}: Create X 1/1 colorless Pilot creature tokens, where X is the
//	 number of Mounts and/or Vehicles that entered the battlefield under
//	 your control this turn. The tokens have "This token saddles Mounts
//	 and crews Vehicles as though its power were 2 greater.""
//
// X is read as the ability resolves, from the per-turn entry tally by
// subtype (Game.EnteredWithSubtypeThisTurn), so a Mount or Vehicle that
// entered and has since left still counts, and one that entered on an
// earlier turn does not. Mount is a creature type and Vehicle an
// artifact type, and no printed permanent is both, so the two counts
// add without counting anything twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "65d8b481-ebe9-4029-98a8-dd4693340bff",
		Name:         "Cloudspire Coordinator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cloudspire Coordinator — scry 2", Do(Scry{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Create X 1/1 colorless Pilot creature tokens, where X is the number of Mounts and/or Vehicles that entered the battlefield under your control this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				// "Mount" is a word of the source's text (CR 612.2, #3984);
				// "Vehicle" is an artifact type, which no text change names.
				x := g.EnteredWithSubtypeThisTurn(item.Controller, g.TextReaderForEffect(item).TextSubtype("Mount")) +
					g.EnteredWithSubtypeThisTurn(item.Controller, "Vehicle")
				if x <= 0 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: PilotToken(), N: x}.Apply(NewContext(g, item))
			},
		}},
	})
}
