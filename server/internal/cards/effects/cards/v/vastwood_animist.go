package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vastwood Animist — Creature — Elf Shaman Ally {2}{G}, 1/1 (Worldwake, S82 OLD
// b487):
//
//	"{T}: Target land you control becomes an X/X Elemental creature until end
//	 of turn, where X is the number of Allies you control. It's still a land."
//
// Avalanche Caller's animation: one layer-4 type change (creature, Elemental,
// still a land) and a base P/T of X (layer 7b), until end of turn on the object
// the target was as the ability resolved (CR 611.2c). X is counted as the ability
// resolves (the Animist itself included while it is still there), and the target
// is re-checked then (CR 608.2b). The land has been under its controller's control
// since the turn began or not: it is a creature, so it can attack only if it has
// (CR 302.6), the ordinary summoning-sickness rule.
//
// No simplification.
func init() {
	const label = "Vastwood Animist — target land becomes an X/X Elemental creature until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "0b331f27-0d78-4d22-b546-36c5dec37606",
		Name:         "Vastwood Animist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target land you control becomes an X/X Elemental creature until end of turn, where X is the number of Allies you control. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				land := effects.FirstLegalBattlefieldTarget(ctx)
				if land == uuid.Nil {
					return nil
				}
				x := effects.ZenB2AlliesYouControl(ctx)
				return effects.BecomeCreatureUntilEOT{Target: land, Types: []string{"Creature"}, Subtypes: []string{"Elemental"}, SetPower: x, SetToughness: x, Label: label}.Apply(ctx)
			},
		}},
	})
}
