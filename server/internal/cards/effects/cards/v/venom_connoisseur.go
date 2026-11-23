package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venom Connoisseur — Creature — Human Druid {1}{G}, 2/2 (Foundations,
// #4465):
//
//	"Alliance — Whenever another creature you control enters, this
//	 creature gains deathtouch until end of turn. If this is the second
//	 time this ability has resolved this turn, all creatures you control
//	 gain deathtouch until end of turn."
//
// The count is Belladonna Took's: this Connoisseur's own resolutions of
// this ability this turn, including the one resolving
// (Game.ResolvedThisTurn), per object (CR 400.7). Every resolution gives the
// Connoisseur deathtouch; the second also gives it to each creature you
// control as it resolves (CR 611.2c: only those on the battlefield then).
// A third resolution is only the Connoisseur again.
//
// No simplification.
func init() {
	const label = "Venom Connoisseur — this creature gains deathtouch until end of turn; the second time this turn, all creatures you control do"
	effects.Register(effects.Spec{
		OracleID:     "03ebc100-68ac-4d53-adc2-d537212f7d31",
		Name:         "Venom Connoisseur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl(label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"deathtouch"},
						Label:    "Venom Connoisseur — deathtouch until end of turn",
					}).Apply(ctx); err != nil {
						return err
					}
					if g.ResolvedThisTurn(item.SourceCardID, label) != 2 {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"deathtouch"},
						Label:    "Venom Connoisseur — creatures you control gain deathtouch until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
