package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tajic, Blade of the Legion — Legendary Creature — Human Soldier
// {2}{R}{W}, 2/2:
//
//	"Indestructible
//	 Battalion — Whenever Tajic and at least two other creatures attack,
//	 Tajic gets +5/+5 until end of turn."
//
// Battalion is Boros Strike-Captain's trigger (battalionAttack); the
// pump is on Tajic himself, and is skipped when he has left the
// battlefield by the time it resolves (a pump of a pinned object
// applies only to it, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8951d797-6356-4f27-a6de-e013216274d6",
		Name:            "Tajic, Blade of the Legion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, battalionAttack,
				"Tajic, Blade of the Legion — Tajic gets +5/+5 until end of turn (battalion)",
				func(g *game.Game, item *game.StackItem) error {
					if sourceIsNewObject(g, item) {
						return nil
					}
					return BoostUntilEOT{Target: item.SourceCardID, Power: 5, Toughness: 5,
						Label: "Tajic, Blade of the Legion — battalion"}.Apply(NewContext(g, item))
				}),
		},
	})
}
