package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seedcradle Witch — Creature — Elf Shaman {G/W} 1/1:
//
//	"{2}{G}{W}: Target creature gets +3/+3 until end of turn. Untap that
//	 creature."
//
// The pump and the untap both act on the one target as the ability resolves,
// and neither happens if the target has become illegal (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "278ea996-3094-4a76-b6e1-db1d002e5502",
		Name:         "Seedcradle Witch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{W}: Target creature gets +3/+3 until end of turn. Untap that creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit, Pump: &game.Pump{Power: 3, Toughness: 3}})},
			Cost:    effects.ManaCost("{2}{G}{W}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Seedcradle Witch — +3/+3 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.UntapTarget{Target: id}.Apply(ctx)
			},
		}},
	})
}
