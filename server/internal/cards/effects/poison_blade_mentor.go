package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Poison-Blade Mentor — Creature — Human Assassin {1}{B}, 2/1 (Assassin's
// Creed, #5948):
//
//	"Deathtouch
//	 Whenever this creature attacks, another target Assassin you control
//	 gains deathtouch until end of turn."
//
// Captain America's "another target creature you control" attack trigger
// with an Assassin filter and a keyword grant for its body; with no other
// Assassin there is no trigger, and a target that left in response gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60a3345c-3234-4f4b-baab-df12dad6b44b",
		Name:            "Poison-Blade Mentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventAttack, ThisAttacked,
				"Poison-Blade Mentor — another target Assassin you control gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Poison-Blade Mentor"}.Apply(ctx)
				}),
				Another(TargetCreature("another target Assassin you control", HasSubtype("Assassin"), YouControl()))),
		},
	})
}
