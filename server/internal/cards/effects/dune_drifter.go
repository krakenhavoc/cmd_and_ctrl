package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dune Drifter — Artifact — Vehicle {X}{W}{B}, 3/3 (Aetherdrift, #4464):
//
//	"When this Vehicle enters, return target artifact or creature card
//	 with mana value X or less from your graveyard to the battlefield.
//	 Crew 2"
//
// Kinetic Ooze's shape: X is the value announced for the spell (CR 107.3m),
// read off the permanent as the trigger is built (Card.CastX), so the
// target is bounded as it is chosen (CR 603.3d). A Drifter that wasn't
// cast, or was cast without paying its mana cost, has X = 0 and can return
// only a card with mana value 0 (the 2024 ruling). Crew is the engine's
// (CR 702.122a).
//
// No simplification.
func init() {
	const label = "Dune Drifter — return target artifact or creature card with mana value X or less from your graveyard to the battlefield"
	trig := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
			return TargetCardInGraveyard("target artifact or creature card with mana value X or less from your graveyard",
				YouOwn(), Or(Artifact(), Creature()), ManaValueLE(source.CastX()))
		},
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label)
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
		},
	}
	Register(Spec{
		OracleID:     "b32649cb-bf5c-45fa-86f0-8bf4e8555f06",
		Name:         "Dune Drifter",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trig},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Dune Drifter"),
		}},
	})
}
