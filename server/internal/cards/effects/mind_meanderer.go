package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Meanderer — Creature — Bird Fish Illusion {3}{G}{U}{U}, 4/4
// (Reality Fracture, tracker #2795):
//
//	"Flying
//	 This creature has vigilance as long as you control a Jace
//	 planeswalker.
//	 When this creature enters, it fights up to one target creature an
//	 opponent controls."
//
// A Jace token (Empower Jace) is a Jace planeswalker, so it turns the
// vigilance on. The fight is the "up to one" shape (Mawloc): the trigger
// always goes on the stack and may take no target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18d1765b-3b60-4dda-93a6-d02a4e53f185",
		Name:            "Mind Meanderer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			KeywordGrant(selfWhileYouControlA(And(Planeswalker(), HasSubtype("Jace"))), "vigilance"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Mind Meanderer — it fights up to one target creature an opponent controls",
			Targets:   TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return b10Fight(ctx, item.SourceCardID, id)
			},
		}},
	})
}
