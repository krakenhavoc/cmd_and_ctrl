package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luminous Guardian — Creature — Human Nomad {3}{W}, 1/4:
//
//	"{W}: This creature gets +0/+1 until end of turn.
//	 {2}: This creature can block an additional creature this turn."
//
// Two activations on itself: a layer 7c pump (BoostUntilEOT) and
// Coastline Chimera's extra block (#1715).
func init() {
	Register(Spec{
		OracleID:     "0d9a31f5-20d0-4266-95ec-38b4fa96b8aa",
		Name:         "Luminous Guardian",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}: This creature gets +0/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Target:    item.SourceCardID,
						Toughness: 1,
						Label:     "Luminous Guardian — +0/+1",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{2}: This creature can block an additional creature this turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{2}"),
				Effect:  selfBlocksAdditionalThisTurn("Luminous Guardian — can block an additional creature this turn"),
			},
		},
	})
}
