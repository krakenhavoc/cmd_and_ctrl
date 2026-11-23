package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eel-Hounds — Creature — Fish Dog {3}{G}, 4/2 (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Trample (This creature can deal excess combat damage to the player
//	 it's attacking.)
//	 Whenever this creature attacks, another target creature you control
//	 gets +2/+2 and gains trample until end of turn."
//
// The target is chosen as the trigger is put on the stack (CR 603.3d);
// "another" excludes Eel-Hounds itself. A target gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "25432a3a-bcfd-4c43-b662-07676e9b3fdd",
		Name:            "Eel-Hounds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WheneverThisAttacks("Eel-Hounds — another target creature you control gets +2/+2 and gains trample",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Eel-Hounds — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Eel-Hounds — trample"}.Apply(ctx)
					}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample"))),
		},
	})
}
