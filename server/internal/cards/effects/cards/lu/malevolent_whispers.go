package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malevolent Whispers — Sorcery {3}{R} (Commander 2019, #5517):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gets +2/+0 and gains haste until end of turn.
//	 Madness {3}{R} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Bloody Betrayal's three clauses with a +2/+0 in the third, and the one
// Madness string. If the target is illegal on resolution the whole spell
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7ff27b6-7591-44b9-86b9-ee12f795ec25",
		Name:         "Malevolent Whispers",
		Completeness: effects.CompletenessFull,
		Madness:      "{3}{R}",
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Malevolent Whispers — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 2, Label: "Malevolent Whispers — +2/+0 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Malevolent Whispers — haste until end of turn",
			}.Apply(ctx)
		},
	})
}
