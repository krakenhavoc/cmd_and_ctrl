package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abundant Maw — Creature — Eldrazi Leech {8}, 6/4:
//
//	"Emerge {6}{B} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, target opponent loses 3 life and you gain
//	 3 life."
//
// Emerge is the shared alternative cost (ADR 0135 §4): the creature is
// sacrificed with the spell on the stack, and its mana value comes off the
// emerge cost's generic part. The cast trigger resolves above the spell,
// so it drains even if the Maw is countered, and it triggers however the
// Maw was cast. A target opponent who has left the game by then is an
// illegal target and the trigger does nothing (CR 608.2b).
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Abundant Maw — target opponent loses 3 life and you gain 3 life",
		func(g *game.Game, item *game.StackItem) error { return targetOpponentLosesAndYouGain(g, item, 3) })
	cast.Targets = TargetPlayer("target opponent", Opponent())
	Register(Spec{
		OracleID:         "3676b745-001d-46e6-880f-2fe26476a38d",
		Name:             "Abundant Maw",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Emerge("{6}{B}")},
		Triggered:        []game.TriggeredAbility{cast},
	})
}
