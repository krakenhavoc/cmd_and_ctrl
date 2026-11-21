package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elegant Entourage — Creature — Elf Druid {3}{G}, 4/4 (Streets of New
// Capenna, #5491):
//
//	"Alliance — Whenever another creature you control enters, target
//	 creature other than this creature gets +1/+1 and gains trample until
//	 end of turn."
//
// Alliance is "another creature you control enters". The target is chosen as
// the trigger goes on the stack and excludes the Entourage as an object
// (Another, so a copy of it stays a legal pick); on resolution a target that
// is no longer legal gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Elegant Entourage — alliance: target creature other than this creature gets +1/+1 and gains trample until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "e049eb22-a01a-463e-8de9-705d6eaf84e2",
		Name:         "Elegant Entourage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AnotherCreatureEnteredUnderYourControl, label,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: label}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
					}),
				effects.Another(effects.TargetCreature("target creature other than this creature"))),
		},
	})
}
