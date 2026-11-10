package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riverchurn Monument — Artifact {1}{U} (Aetherdrift, #4464):
//
//	"{1}, {T}: Any number of target players each mill two cards.
//	 Exhaust — {2}{U}{U}, {T}: Any number of target players each mill cards
//	 equal to the number of cards in their graveyard."
//
// Officious Interrogation's "any number of target players", which may be
// none, you, or anyone, on two activated abilities. Both are applied to
// the players still legal as the ability resolves (CR 608.2b), in the order
// they were chosen. The exhaust count is each player's OWN graveyard,
// counted for every target before any of them mills (the mills are one
// instruction, so a mill that finishes first does not change the next
// player's count); an exhaust ability may be activated only once for each
// object (CR 702.177a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71c714a2-baa9-4082-bc8e-ceadbea231e0",
		Name:         "Riverchurn Monument",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Any number of target players each mill two cards.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Targets: TargetPlayer("any number of target players").WithCount(0, 0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, id := range kozilekTargetPlayers(ctx) {
						if err := (MillCards{Player: id, N: 2}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "Exhaust — {2}{U}{U}, {T}: Any number of target players each mill cards equal to the number of cards in their graveyard.",
				Exhaust: true,
				Cost:    Plus(ManaCost("{2}{U}{U}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Targets: TargetPlayer("any number of target players").WithCount(0, 0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					players := kozilekTargetPlayers(ctx)
					counts := make([]int, len(players))
					for i, id := range players {
						if p := g.PlayerByIDForEffect(id); p != nil && p.Graveyard != nil {
							counts[i] = p.Graveyard.Size()
						}
					}
					for i, id := range players {
						if err := (MillCards{Player: id, N: counts[i]}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
	})
}
