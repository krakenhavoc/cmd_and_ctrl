package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eccentric Apprentice — Creature — Tiefling Wizard {2}{U}, 2/2:
//
//	"Flying
//	 When this creature enters, venture into the dungeon.
//	 At the beginning of combat on your turn, if you've completed a
//	 dungeon, up to one target creature becomes a Bird with base power
//	 and toughness 1/1 and flying until end of turn."
//
// "Becomes a Bird" sets its creature types to Bird (CR 205.1b) and keeps
// its card types, colours and other abilities; the base 1/1 is layer
// 7b, so counters and pumps still apply on top. An intervening "if"
// (CR 603.4).
//
// No simplification.
func init() {
	const label = "Eccentric Apprentice — up to one target creature becomes a 1/1 Bird with flying until end of turn"
	Register(Spec{
		OracleID:        "27dbe95e-6274-416a-a45b-a2a0a64d5403",
		Name:            "Eccentric Apprentice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersVenture("Eccentric Apprentice"),
			IfVentureCond(CondYouCompletedADungeon, Targeting(
				AtYourStep(game.StepBeginCombat, label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := untilEndOfTurn(ctx, t.ID, nil, label,
							game.SetCreatureTypesMod(ctx.TextSubtypes("Bird")...),
							game.SetBasePowerMod(1), game.SetBaseToughnessMod(1),
							game.AddKeywordsMod("flying")); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetCreature("up to one target creature").WithCount(0, 1))),
		},
	})
}
