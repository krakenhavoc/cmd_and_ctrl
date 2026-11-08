package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e6a35729-edee-4a8c-9aa1-041114d66c69",
		Name:            "Ayli, Eternal Pilgrim",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, Sacrifice another creature: You gain life equal to the sacrificed creature's toughness.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another creature", Creature())),
				Effect: Do(Instruction(func(ctx *Context) error {
					fed, ok := b17PermanentSacrificedToPay(ctx.Game, ctx.Item)
					if !ok {
						return nil
					}
					if t := departedCreatureToughness(ctx.Game, fed); t > 0 {
						return GainLife{Player: ctx.Controller(), Amount: t}.Apply(ctx)
					}
					return nil
				})),
			},
			{
				Label:     "{1}{W}{B}, Sacrifice another creature: Exile target nonland permanent. Activate only if you have at least 10 life more than your starting life total.",
				Cost:      Plus(ManaCost("{1}{W}{B}"), SacrificeAnotherN(1, "another creature", Creature())),
				Condition: LifeAtLeastAboveStarting(10),
				Targets:   TargetPermanent("target nonland permanent", Nonland()),
				Effect:    ExileFirstTarget,
			},
		},
	})
}
