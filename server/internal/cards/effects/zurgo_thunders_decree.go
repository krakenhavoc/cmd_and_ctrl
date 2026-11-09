package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zurgo, Thunder's Decree — Legendary Creature — Orc Warrior {R}{W}{B},
// 2/4 (Tarkir: Dragonstorm, #4463):
//
//	"Mobilize 2 (Whenever this creature attacks, create two tapped and
//	 attacking 1/1 red Warrior creature tokens. Sacrifice them at the
//	 beginning of the next end step.)
//	 During your end step, Warrior tokens you control have "This token
//	 can't be sacrificed.""
//
// Mobilize is mobilize.go (CR 702.181a). The second line is a layer-6
// grant of "can't be sacrificed" (Garland, Royal Kidnapper's static over
// cant_be_sacrificed.go's engine half) to the Warrior tokens the
// controller controls, while it is the controller's own end step. So the
// Warriors a Zurgo attack made survive the sacrifice that the mobilize
// delayed trigger tries at the beginning of that turn's end step: the
// instruction does nothing to them (CR 101.2) and, the trigger being
// spent, they stay. Tokens of other subtypes, tokens an opponent controls
// and the first end step of an opponent's turn are untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "393d0bf8-5f2e-4eaa-91a1-9132429a7371",
		Name:         "Zurgo, Thunder's Decree",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Mobilize("Zurgo, Thunder's Decree", 2)},
		Static: []game.StaticAbility{{
			// The answer changes as each step begins (the layer cache is
			// bumped on EventStepBegan while this static is in play).
			DependsOnEvents: []game.EventKind{game.EventStepBegan},
			Label:           "During your end step, Warrior tokens you control have \"This token can't be sacrificed.\"",
			Layer:           game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsToken() && target.HasSubtype(source.TextSubtype("Warrior")) &&
					g.Turn.Step == game.StepEnd && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.CantBeSacrificed = true
			},
		}},
	})
}
