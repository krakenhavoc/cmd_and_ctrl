package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Orzhov Guildmage — Creature — Human Wizard {W/B}{W/B}, 2/2 (Ravnica:
// Clue Edition, Ravnica Allegiance):
//
//	"{2}{W}: Target player gains 1 life.
//	 {2}{B}: Each player loses 1 life."
//
// Two independent activated abilities. The life gain targets, and a target
// player who has gone is skipped (CR 608.2b). The loss is life loss, not
// damage, for every player including the controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "877bfa3f-6e30-4cb3-831e-d56849b01334",
		Name:         "Orzhov Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: Target player gains 1 life.",
			Purpose: ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 1}),
			Cost:    ManaCost("{2}{W}"),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				p, ok := firstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return GainLife{Player: p, Amount: 1}.Apply(ctx)
			},
		}, {
			Label:   "{2}{B}: Each player loses 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue}, // a drain answers nothing (ADR 0142)
			Cost:    ManaCost("{2}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return eachPlayerLifeChanges(NewContext(g, item), -1)
			},
		}},
	})
}
