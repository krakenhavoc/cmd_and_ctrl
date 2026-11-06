package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blooming Stinger — Creature — Plant Scorpion {1}{G}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Deathtouch
//	 When this creature enters, another target creature you control
//	 gains deathtouch until end of turn."
//
// Deathtouch is a printed keyword. The entry trigger's clause is
// "another" by object identity (Another, #1738), so the Stinger can
// never pick itself; with no other creature to target the trigger is
// removed. The grant is a layer 6 keyword until end of turn, skipped
// when the pick is no longer a legal target (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ab282bb1-900e-40b8-9821-8e2de88927bc",
		Name:            "Blooming Stinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Blooming Stinger — another target creature you control gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return GrantKeywordUntilEOT{
							Target:   t.ID,
							Keywords: []string{"deathtouch"},
							Label:    "Blooming Stinger — deathtouch until end of turn",
						}.Apply(ctx)
					}
					return nil
				}),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
