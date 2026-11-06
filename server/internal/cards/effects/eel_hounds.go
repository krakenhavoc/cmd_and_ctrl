package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "25432a3a-bcfd-4c43-b662-07676e9b3fdd",
		Name:            "Eel-Hounds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WheneverThisAttacks("Eel-Hounds — another target creature you control gets +2/+2 and gains trample",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Eel-Hounds — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
						return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Eel-Hounds — trample"}.Apply(ctx)
					}),
				Another(TargetCreature("another target creature you control", YouControl()))),
				ForTargets(BenefitUntilEOT(0, 2, 2, "trample"))),
		},
	})
}
