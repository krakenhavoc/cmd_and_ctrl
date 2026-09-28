package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nowhere to Run — Enchantment {1}{B}:
//
//	"Flash
//	 When this enchantment enters, target creature an opponent controls
//	 gets -3/-3 until end of turn.
//	 Creatures your opponents control can be the targets of spells and
//	 abilities as though they didn't have hexproof. Ward abilities of
//	 those creatures don't trigger."
//
// The proof card for #1560 (ADR 0038's amendment of 2026-09-27). The
// last line is two statics, and neither is a layer effect:
//
//   - The hexproof half is a waiver read at the targeting choke point
//     (Spec.HexproofBypasses). The creature keeps its hexproof. The
//     printed line says "spells and abilities" with no "you control",
//     so it is BySpellsAndAbilities: the hexproof creature's OTHER
//     opponents may target it too, which is what separates this card
//     from Glaring Spotlight.
//   - The ward half is a WardSuppression, read by the ward trigger
//     itself, so every ward in the catalog — printed, granted by an
//     Equipment, an emblem's, a face-down permanent's — stays quiet on
//     an opponent's creature while this is out. Your own creatures'
//     ward still triggers.
//
// Both are read live, which is what the card's 2024-09-20 rulings
// need: a spell aimed at a hexproof creature becomes an illegal
// target if this leaves before it resolves (the same CR 608.2b
// re-check that announced it), and a ward trigger that was suppressed
// does not arrive late when this leaves.
//
// The enters trigger needs nothing new, and its target is chosen with
// the static already in force, so it can reach a hexproof creature and
// ward does not tax it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "da82ca96-a613-4d00-9e6b-1fece4fc23d0",
		Name:            "Nowhere to Run",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Nowhere to Run — target creature an opponent controls gets -3/-3 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return BoostUntilEOT{
							Target:    t.ID,
							Power:     -3,
							Toughness: -3,
							Label:     "Nowhere to Run — -3/-3",
						}.Apply(ctx)
					}
					return nil
				}),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		)},
		HexproofBypasses: []game.HexproofBypass{
			AsThoughNoHexproof(BySpellsAndAbilities, Creature(), OpponentControls()),
		},
		WardSuppressions: []game.WardSuppression{
			WardDoesNotTrigger(Creature(), OpponentControls()),
		},
	})
}
