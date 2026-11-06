package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dubious Delicacy — Artifact — Food {2}{B} (Edge of Eternities, #4461):
//
//	"Flash
//	 When this artifact enters, up to one target creature gets -3/-3 until
//	 end of turn.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life.
//	 {2}, {T}, Sacrifice this artifact: Target opponent loses 3 life."
//
// Flash is a printed keyword. The enters trigger is Ant-Man's Air Force's
// "up to one target" shape: it may go on the stack with no target, and a
// target that has left by resolution is skipped (CR 608.2b). The first
// activated ability is the Food ability every Food has (CR 111.10b); the
// second is its own, a life loss that is not damage. The Food subtype comes
// off the type line, so Food payoffs see the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "697b8045-0edb-4f7b-a933-4d41ec2b53d5",
		Name:            "Dubious Delicacy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Answers: game.AnswerRemove}, // the enters trigger shrinks a creature
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Dubious Delicacy — up to one target creature gets -3/-3",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: -3, Toughness: -3,
						Label: "Dubious Delicacy — -3/-3 until end of turn"}.Apply(ctx)
				}), TargetCreature("up to one target creature").WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Purpose: game.Purpose{Answers: game.AnswerValue}, // ADR 0142
				Effect:  b36GainLife(3),
			},
			{
				Label:   "{2}, {T}, Sacrifice this artifact: Target opponent loses 3 life.",
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect:  Do(TargetPlayerLosesLife{N: 3}),
			},
		},
	})
}
