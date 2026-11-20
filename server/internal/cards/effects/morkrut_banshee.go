package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morkrut Banshee — Creature — Spirit {3}{B}{B}, 4/4 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Morbid — When this creature enters, if a creature died this turn,
//	 target creature gets -4/-4 until end of turn."
//
// Morbid is an intervening "if" (CR 603.4, condACreatureDiedThisTurn:
// any creature, anyone's, tokens included), checked as the Banshee
// enters and again as the trigger resolves, so the trigger is not put on
// the stack on a quiet turn. The target is any creature, chosen as the
// trigger goes on the stack; one that has left by resolution is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2fe1847-4003-41c5-a64b-03afd888b81e",
		Name:         "Morkrut Banshee",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(IfVentureCond(condACreatureDiedThisTurn,
				Targeting(WhenThisEnters("Morkrut Banshee — morbid: target creature gets -4/-4 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return BoostUntilEOT{Target: id, Power: -4, Toughness: -4, Label: "Morkrut Banshee — -4/-4 until end of turn"}.Apply(ctx)
					}),
					TargetCreature("target creature"))),
				ForTargets(HarmUntilEOT(0, -4, -4))),
		},
	})
}
