package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dominaria's Judgment — Instant {2}{W} (Planeshift, S82 OLD b643, #5615):
//
//	"Until end of turn, creatures you control gain protection from white if you
//	 control a Plains, from blue if you control an Island, from black if you control
//	 a Swamp, from red if you control a Mountain, and from green if you control a
//	 Forest."
//
// Each "if" is read once as the spell resolves (CR 608.2c): a land type is
// controlled when one of the caster's lands has it, read from the layered land.
// The creatures are the ones the caster controls then (CR 611.2c) and each gets
// every protection that applies, until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41998d4d-ca01-4a9c-8bd6-399e4c3e25b1",
		Name:         "Dominaria's Judgment",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, p := range []struct{ landType, color string }{
				{"Plains", "white"}, {"Island", "blue"}, {"Swamp", "black"}, {"Mountain", "red"}, {"Forest", "green"},
			} {
				if !effects.OldB642ControlsLandType(ctx.Game, ctx.Controller(), p.landType) {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"protection from " + p.color},
					Label:    "Dominaria's Judgment — protection from " + p.color,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
