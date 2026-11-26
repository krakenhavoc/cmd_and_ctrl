package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vedalken Anatomist — Creature — Phyrexian Vedalken Wizard {2}{U}, 1/2
// (Mirrodin Besieged, S82 OLD batch 464, #5576):
//
//	"{2}{U}, {T}: Put a -1/-1 counter on target creature. You may tap or
//	 untap that creature."
//
// The counter, then Trickster Mage's "you may tap or untap" prompt (tap it,
// untap it or leave it) for the same target while it is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7719fd5d-c9dc-49b8-87b4-24f7d42af62f",
		Name:         "Vedalken Anatomist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Put a -1/-1 counter on target creature. You may tap or untap that creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				if !effects.OnBattlefield(g, id) {
					return nil
				}
				return effects.SocB23MayTapOrUntap(ctx, "Vedalken Anatomist", id)
			},
		}},
	})
}
