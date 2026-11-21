package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Artistic Process — Sorcery {3}{R}{R} (Secrets of Strixhaven):
//
//	"Choose one —
//	 • Artistic Process deals 6 damage to target creature.
//	 • Artistic Process deals 2 damage to each creature you don't control.
//	 • Create a 3/3 blue and red Elemental creature token with flying. It
//	 gains haste until end of turn."
//
// The haste is a keyed tail of the token's creation, so a creation that
// waits on a question (a CR 616 order) still hastes the token that lands;
// it is a grant until end of turn, not a printed keyword on the token.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c061a1b-206a-4fcc-92e5-30c8ed9d65c1",
		Name:         "Artistic Process",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Artistic Process deals 6 damage to target creature.",
				effects.TargetCreature("target creature"),
				effects.SocB20DamageToTheModeTarget(6)), effects.ForTargets(effects.DamageToTarget(0, 6))),
			effects.ModeWithPurpose(effects.ModeDoing("Artistic Process deals 2 damage to each creature you don't control.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 2)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, OpponentsOnly: true}}),
			effects.ModeDoing("Create a 3/3 blue and red Elemental creature token with flying. It gains haste until end of turn.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: item.Controller,
						Source:     item.SourceCardID,
						Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("3/3 blue and red Elemental with flying")), Count: 1}},
					}, ctx.ItemTail(artisticProcessHaste))
				}),
		),
	})
}

// artisticProcessHaste is "it gains haste until end of turn" over the
// token that landed. An on-disk key.
var artisticProcessHaste = effects.CardTail("artistic-process/token-haste", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Artistic Process — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
