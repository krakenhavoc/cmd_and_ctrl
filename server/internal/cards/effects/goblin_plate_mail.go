package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Plate Mail — Artifact — Equipment {1}{B/R} (The Hobbit, HOB):
//
//	"When this Equipment enters, amass Goblins 1, then attach this Equipment
//	 to the amassed Army.
//	 Equipped creature gets +1/+0 and has menace.
//	 Equip {4}"
//
// The attach is the amass's continuation (CR 701.47c, "the amassed Army"), so
// it names the Army the amass chose and reads the counters it just placed.
// It is skipped quietly when the Equipment is no longer the permanent that
// asked (CR 701.3b) and when there was no Army to choose (uuid.Nil).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eda99a16-6a7c-4f39-8a6b-a284e6afd3fc",
		Name:         "Goblin Plate Mail",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0), GrantToAttached("menace")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Goblin Plate Mail — amass Goblins 1, then attach this Equipment to the amassed Army",
				func(g *game.Game, item *game.StackItem) error {
					return Amass{Subtype: "Goblin", N: 1, ThenTail: goblinPlateMailAttach}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}

// goblinPlateMailAttach is "then attach this Equipment to the amassed Army"
// as a keyed tail. An on-disk key: never renamed, never reused.
var goblinPlateMailAttach = AmassTail("goblin-plate-mail/attach", func(ctx *Context, army uuid.UUID) error {
	if army == uuid.Nil {
		return nil
	}
	return ctx.Game.AttachSourceForEffect(ctx.Item, game.TargetRef{Kind: game.TargetCard, ID: army})
})
