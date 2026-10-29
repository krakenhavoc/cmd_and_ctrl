package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crude Bent Blade — Artifact — Equipment {2}{B}:
//
//	"When this Equipment enters, target opponent sacrifices a creature
//	 of their choice.
//	 Equipped creature gets +2/+1.
//	 Equip {2}"
//
// An edict on the enters trigger: the opponent is the target and the
// creature is their own choice at resolution, so hexproof on their
// creatures does not matter. PlayerSacrificesForEffect queues that
// prompt. A target opponent who left the game leaves nothing to do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eec81460-899d-4b4c-b33c-ce2ca704df17",
		Name:         "Crude Bent Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Crude Bent Blade — target opponent sacrifices a creature of their choice",
					targetPlayerSacrificesACreature("Crude Bent Blade — sacrifice a creature")),
				TargetPlayer("target opponent", Opponent()),
			),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// targetPlayerSacrificesACreature is "target player sacrifices a creature
// of their choice" as an ability's whole body: the first still-legal
// player target sacrifices, and a target that left leaves nothing to do
// (CR 608.2b). The prompt carries `label`. Crude Bent Blade's and Custodi
// Lich's body.
func targetPlayerSacrificesACreature(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		for _, t := range NewContext(g, item).LegalTargets() {
			if t.Kind == game.TargetPlayer {
				g.PlayerSacrificesForEffect(item.SourceCardID, t.ID, sacrificeSpec("a creature", Creature()), label)
				return nil
			}
		}
		return nil
	}
}
