package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horrific Assault — Sorcery {G} (Modern Horizons 3):
//
//	"Target creature you control deals damage equal to its power to target
//	 creature or planeswalker you don't control. If you control an
//	 Eldrazi, you gain 3 life."
//
// Bite Down's one-sided fight (two target clauses, the source's power read as
// the spell resolves) with a life rider. "If you control an Eldrazi" is read
// as the spell resolves, after the damage (CR 608.2c), through effective
// subtypes, so an Eldrazi the damage just killed no longer counts and a
// creature that is an Eldrazi through an effect does. The life is gained
// whether or not the damage was dealt: it is its own sentence.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c7e091e-e3c1-446d-a442-4aacb3cdd575",
		Name:         "Horrific Assault",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetPermanent("target creature or planeswalker you don't control",
				Or(Creature(), Planeswalker()), OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := oneSidedBite(item, ctx); err != nil {
				return err
			}
			if !ControlsA(ctx.TextSubtype("Eldrazi"))(ctx.Game, ctx.Controller()) {
				return nil
			}
			return GainLife{Amount: 3}.Apply(ctx)
		},
	})
}
