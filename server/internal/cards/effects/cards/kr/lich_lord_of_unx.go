package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lichLordOfUnxRecord is the target player and X, the Zombies counted as the
// ability resolves.
const lichLordOfUnxRecord = "lich-lord-of-unx/drain"

// Lich Lord of Unx — Creature — Zombie Wizard {1}{U}{B}, 2/2:
//
//	"{U}{B}, {T}: Create a 1/1 blue and black Zombie Wizard creature
//	 token.
//	 {U}{U}{B}{B}: Target player loses X life and mills X cards, where X
//	 is the number of Zombies you control."
//
// X is counted as the second ability resolves (CR 608.2h), over every
// Zombie permanent its controller controls, the Lich Lord included.
// The life is lost first, then the cards are milled, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6643fc8e-548e-48db-9f67-ffbca9e57016",
		Name:         "Lich Lord of Unx",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}{B}, {T}: Create a 1/1 blue and black Zombie Wizard creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{U}{B}"), effects.TapCost()),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue and black Zombie Wizard"), N: 1}),
			},
			{
				Label:   "{U}{U}{B}{B}: Target player loses X life and mills X cards, where X is the number of Zombies you control.",
				Cost:    effects.ManaCost("{U}{U}{B}{B}"),
				Targets: effects.TargetPlayer("target player"),
				// The mill waits for a prompt the life loss asks (a CR 616 order; CR 608.2c,
				// #4155), and mills the X counted before the loss.
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetPlayer {
								continue
							}
							x := effects.ZombiesYouControl(ctx.Game, ctx.Controller())
							ctx.RecordStep(lichLordOfUnxRecord, []uuid.UUID{t.ID}, x)
							if x == 0 {
								return nil
							}
							return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -x)
						}
						return nil
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						r := ctx.StepRecord(lichLordOfUnxRecord)
						if len(r.IDs) == 0 || r.N() <= 0 {
							return nil
						}
						return effects.MillCards{Player: r.IDs[0], N: r.N()}.Apply(ctx)
					}),
				),
			},
		},
	})
}
