package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// V'Ger, the Intruder — Legendary Artifact Creature — Construct {6}{U},
// 5/7 (Star Trek):
//
//	Flying
//	Whenever V'Ger enters or attacks, choose one —
//	• Look at target opponent's hand, then draw a card.
//	• Creatures your opponents control get -1/-0 until your next turn.
//
// One modal trigger on WhenThisEntersOrAttacks (Avengers Quinjet's shape).
// The look marks the opponent's hand known to V'Ger's controller only
// ("look at" is not "reveal"); the draw happens even if that opponent's hand
// is empty. The -1/-0 is a layer-7c effect whose set is fixed as the
// ability resolves (CR 611.2c), so a creature an opponent gets later is not
// affected, and it lasts until V'Ger's controller's next turn begins.
//
// No simplification.
func init() {
	trigger := WhenThisEntersOrAttacks("V'Ger, the Intruder — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("Look at target opponent's hand, then draw a card.",
			TargetPlayer("target opponent", Opponent()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				if t, ok := ModeTarget(ctx, occ); ok && t.Kind == game.TargetPlayer {
					if p := ctx.PlayerByID(t.ID); p != nil && p.Hand != nil {
						for i := range p.Hand.Cards {
							p.Hand.Cards[i].AddKnower(ctx.Controller())
						}
					}
				}
				return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}), game.Purpose{Draws: 1}),
		ModeDoing("Creatures your opponents control get -1/-0 until your next turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return ScopedEffectFor{
					Match:    And(Creature(), OpponentControls()),
					Mods:     []game.Mod{game.ModifyPTMod(-1, 0)},
					Duration: DurationUntilYourNextTurn(ctx, item.Controller),
					Label:    "V'Ger, the Intruder — creatures your opponents control get -1/-0 until your next turn",
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "f50ec16b-13f5-4e20-ae09-9c7dee9d223b",
		Name:            "V'Ger, the Intruder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
