package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burn Down the House — Sorcery {3}{R}{R} (Midnight Hunt, #5497):
//
//	"Choose one —
//	 • Burn Down the House deals 5 damage to each creature and each
//	   planeswalker.
//	 • Create three 1/1 red Devil creature tokens with "When this token dies,
//	   it deals 1 damage to any target." They gain haste until end of turn."
//
// Brotherhood's End's sweep at 5 (one instance to every creature and
// planeswalker on the battlefield, the caster's included), and Goldwardens'
// Gambit's token-then-haste shape for the second bullet: the Devils are the
// catalog's (devilToken(), Spiked Corridor's), and each one that is still on
// the battlefield after the creation settles gains haste until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ec54683-a65c-461d-862e-585b6499d6a5",
		Name:         "Burn Down the House",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Burn Down the House deals 5 damage to each creature and each planeswalker.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return damageEachMatching(ctx, Or(Creature(), Planeswalker()), 5)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 5}}),
			ModeWithPurpose(ModeDoing("Create three 1/1 red Devil creature tokens with \"When this token dies, it deals 1 damage to any target.\" They gain haste until end of turn.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups: []game.TokenGroup{{
							Template: ctx.TextToken(devilToken()),
							Count:    3,
						}},
					}, ctx.ItemTail(burnDownTheHouseHaste))
				}), game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 3, CreatureTokens: 3}),
		),
	})
}

// burnDownTheHouseHaste is "They gain haste until end of turn": every Devil
// the creation left on the battlefield. An on-disk key.
var burnDownTheHouseHaste = CardTail("burn-down-the-house/haste", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Burn Down the House — haste until end of turn",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
})
