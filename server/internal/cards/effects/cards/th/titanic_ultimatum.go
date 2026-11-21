package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Titanic Ultimatum — Sorcery {R}{R}{G}{G}{G}{W}{W} (Shards of Alara,
// #5588):
//
//	"Until end of turn, creatures you control get +5/+5 and gain first
//	 strike, trample, and lifelink."
//
// Triumph of the Hordes' shape: a layer-7c pump (CR 613.4c) and a
// keyword grant over the creatures its controller controls as it
// resolves. The set is fixed then (CR 611.2c), so a creature that enters
// later in the turn gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0cd28580-d718-4d0b-96e3-a836c2911c27",
		Name:         "Titanic Ultimatum",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			yours := effects.And(effects.Creature(), effects.YouControl())
			if err := (effects.BoostUntilEOT{
				Match:     yours,
				Power:     5,
				Toughness: 5,
				Label:     "Titanic Ultimatum — +5/+5",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"first strike", "trample", "lifelink"},
				Label:    "Titanic Ultimatum — first strike, trample and lifelink",
			}.Apply(ctx)
		},
	})
}
