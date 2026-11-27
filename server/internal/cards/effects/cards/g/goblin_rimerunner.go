package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Rimerunner — Snow Creature — Goblin Warrior {2}{R}, 2/2 (Coldsnap):
//
//	"{T}: Target creature can't block this turn.
//	 {S}: This creature gains haste until end of turn."
//
// The haste ability costs no {T}, so it may be activated the turn the
// Rimerunner arrives and then {T} paid at once (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8505f9a6-a510-43f9-a96d-b840b353e144",
		Name:         "Goblin Rimerunner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label: "{T}: Target creature can't block this turn.",
				Purpose: effects.ForTargets(game.TargetPurpose{
					Slot: 0, UntilEndOfTurn: game.TargetHarm, Pump: &game.Pump{CantBlock: true},
				}),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Goblin Rimerunner — can't block this turn"}.Apply(ctx)
				},
			},
			{
				Label:   "{S}: This creature gains haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{S}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("haste", "Goblin Rimerunner — haste until end of turn"),
			},
		},
	})
}
