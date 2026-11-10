package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// North Pole Research Base — Plane — Earth (Doctor Who, #4870):
//
//	"At the beginning of your upkeep, target opponent draws a card and
//	 creates a Treasure token.
//	 Whenever chaos ensues, create a 2/2 white Alien creature token. When
//	 you do, tap target nontoken creature an opponent controls. Put a
//	 stun counter on it. (If a permanent with a stun counter would become
//	 untapped, remove one from it instead.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The upkeep ability targets as it goes on the stack and the chosen
// opponent both draws and creates the Treasure. The chaos ability (CR 311.7)
// creates the Alien and then, as a reflexive trigger (CR 603.12) whose
// target is chosen as it goes on the stack, taps the creature and puts the
// stun counter on it even if it was already tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2dfbb3e-3d5f-4b6a-b910-4c7b0c9dc17a",
		Name:         "North Pole Research Base",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("North Pole Research Base — target opponent draws a card and creates a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := firstLegalTargetPlayer(ctx)
					if p == uuid.Nil {
						return nil
					}
					if err := (DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return CreateToken{Controller: p, Template: TreasureToken(), N: 1}.Apply(ctx)
				}), TargetPlayer("target opponent", Opponent())),
			WheneverChaosEnsues("North Pole Research Base — create a 2/2 white Alien creature token; when you do, tap target nontoken creature an opponent controls and put a stun counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (CreateToken{Template: planesB2Alien(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					return WhenYouDo("North Pole Research Base — tap target nontoken creature an opponent controls and put a stun counter on it",
						northPoleResearchBaseStun).Apply(ctx)
				}),
		},
	})
}

// northPoleResearchBaseStun is the reflexive half. Registered once (ADR 0041
// P9); an on-disk key, never renamed, never reused.
var northPoleResearchBaseStun = game.ReflexiveBody("north-pole-research-base/stun",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		return AddCounter{Target: id, Kind: game.CounterStun, N: 1}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target nontoken creature an opponent controls", OpponentControls(), Not(IsTokenPredicate()))
	}))
