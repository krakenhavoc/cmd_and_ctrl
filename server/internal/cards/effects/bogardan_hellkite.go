package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bogardan Hellkite — Creature — Dragon {6}{R}{R}, 5/5 (EDHREC rank
// 8494):
//
//	"Flash
//	 Flying
//	 When this creature enters, it deals 5 damage divided as you
//	 choose among any number of targets."
//
// Eight mana at instant speed for a 5/5 flier that clears the way as
// it lands. "Any number of targets" is WithCount(0, 0) — unbounded in
// the clause, bounded in practice by the five points, since every
// chosen target must be assigned at least 1 (CR 601.2d) and the gate
// refuses a sixth.
//
// The division is the controller's (#1563): the pick_target prompt
// asks for the targets and each one's share. A target that leaves in
// response takes nothing and its share is lost (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:        "699bc130-2f1f-4bc8-a25f-9329e40efbb1",
		Name:            "Bogardan Hellkite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets:   TargetAny().WithCount(0, 0).Dividing(Divide(5)),
			Key:       "Bogardan Hellkite — 5 damage divided among any number of targets",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DealDividedDamage(NewContext(g, item))
			},
		}},
	})
}
