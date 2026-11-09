package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Lotus Lounge — Plane — MagicCon (punk, #4870):
//
//	"When you planeswalk here and at the beginning of your upkeep, create
//	 a Black Lotus artifact token.
//	 Whenever Chaos ensues, each player draws a card for each artifact
//	 token they control."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// first is one ability with two trigger conditions. The chaos ability (CR
// 311.7) counts each player's artifact tokens as it resolves, and each draws
// that many cards, in turn order from the active player (CR 101.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ddea0dcf-e8d1-44ea-9ce4-009fa8bbfb95",
		Name:         "Black Lotus Lounge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Black Lotus Lounge — create a Black Lotus artifact token", Do(CreateToken{Template: B11BlackLotusToken(), N: 1})),
			WheneverChaosEnsues("Black Lotus Lounge — each player draws a card for each artifact token they control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, p := range seatsFromActive(g) {
						n := len(MatchingBattlefield(ctx, And(Artifact(), IsTokenPredicate(), ControlledBy(p))))
						if err := (DrawCards{Player: p, N: n}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
