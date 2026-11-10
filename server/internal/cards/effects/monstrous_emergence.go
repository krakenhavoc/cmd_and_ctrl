package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monstrous Emergence — Sorcery {1}{G} (Duskmourn: House of Horror,
// #4466):
//
//	"As an additional cost to cast this spell, choose a creature you
//	 control or reveal a creature card from your hand.
//	 Monstrous Emergence deals damage equal to the power of the creature
//	 you chose or the card you revealed to target creature."
//
// The cost is ChooseOrRevealCost (std_b25_choose_or_reveal.go): a behold of
// a card type that records its pick on the payment. The damage is read as
// the spell resolves: a creature still on the battlefield counts with its
// current power, so a pump or a counter put on it in response counts; one
// that has left the battlefield counts as it last existed there (CR 608.2h,
// the printed ruling); a revealed card counts with its printed power. The
// spell itself is the damage source, so it carries no deathtouch or
// lifelink of the chosen creature, as printed. The target is chosen as the
// spell is cast; if it has gone by resolution nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "5542ea4d-34ab-4f89-8894-793a7038f501",
		Name:           "Monstrous Emergence",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove}, // damages a creature
		AdditionalCost: ChooseOrRevealCost("creature", "Choose a creature you control or reveal a creature card from your hand"),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n, ok := b25ChosenPower(ctx)
			if !ok {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: target, Amount: n}.Apply(ctx)
		},
	})
}
