package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sentinel Sarah Lyons — Legendary Creature — Human Knight {3}{R}{W}, 4/4
// (Fallout, #5948):
//
//	"Haste
//	 As long as an artifact entered the battlefield under your control this
//	 turn, creatures you control get +2/+2.
//	 Battalion — Whenever Sentinel Sarah Lyons and at least two other creatures
//	 attack, Sentinel Sarah Lyons deals damage equal to the number of artifacts
//	 you control to target player."
//
// The anthem is a layer 7c static over the creatures its controller controls,
// on while the turn's entry tally for artifacts under that controller is not
// empty (EnteredWithCardTypeThisTurn: tokens count, and so does one that has
// since left; the Mechan Shieldmate reading). Battalion is the trigger of her
// own attack with two other attackers in the declaration (battalionAttack);
// the target player is chosen as it goes on the stack and the artifacts are
// counted as it resolves, the damage dealt by Sarah as she last existed
// (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b368b1bd-3273-40a9-8102-27501f4d0e15",
		Name:            "Sentinel Sarah Lyons",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{{
			Label:    "As long as an artifact entered the battlefield under your control this turn, creatures you control get +2/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					g.EnteredWithCardTypeThisTurn(source.Controller, "artifact") > 0
			},
			Apply: pumpBy(2, 2),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventAttack, battalionAttack,
				"Sentinel Sarah Lyons — damage equal to the number of artifacts you control to target player (battalion)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := len(MatchingBattlefield(ctx, And(Artifact(), YouControl())))
					if n <= 0 {
						return nil
					}
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
						}
					}
					return nil
				}), TargetPlayer("target player")),
		},
	})
}
