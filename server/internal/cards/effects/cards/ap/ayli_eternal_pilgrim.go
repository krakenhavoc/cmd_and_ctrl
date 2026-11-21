package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ayli, Eternal Pilgrim — Legendary Creature — Kor Cleric {W}{B}, 2/3
// (Foundations, #4465):
//
//	"Deathtouch
//	 {1}, Sacrifice another creature: You gain life equal to the
//	 sacrificed creature's toughness.
//	 {1}{W}{B}, Sacrifice another creature: Exile target nonland
//	 permanent. Activate only if you have at least 10 life more than your
//	 starting life total."
//
// The first row is Bushmeat Poacher's without the tap or the draw: the
// sacrifice is a cost, so the toughness is the creature's last-known one
// (CR 608.2h), counters and anthems included; zero or less gains nothing.
// The second is Speaker of the Heavens' condition (CR 602.1b),
// LifeAtLeastAboveStarting(10), 50 or more at a Commander table, on
// Gideon Blackblade's exile. Neither row taps Ayli, so both work the turn
// she arrives, at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e6a35729-edee-4a8c-9aa1-041114d66c69",
		Name:            "Ayli, Eternal Pilgrim",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, Sacrifice another creature: You gain life equal to the sacrificed creature's toughness.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					fed, ok := effects.B17PermanentSacrificedToPay(ctx.Game, ctx.Item)
					if !ok {
						return nil
					}
					if t := effects.DepartedCreatureToughness(ctx.Game, fed); t > 0 {
						return effects.GainLife{Player: ctx.Controller(), Amount: t}.Apply(ctx)
					}
					return nil
				})),
			},
			{
				Label:     "{1}{W}{B}, Sacrifice another creature: Exile target nonland permanent. Activate only if you have at least 10 life more than your starting life total.",
				Cost:      effects.Plus(effects.ManaCost("{1}{W}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
				Condition: effects.LifeAtLeastAboveStarting(10),
				Targets:   effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				Effect:    effects.ExileFirstTarget,
			},
		},
	})
}
