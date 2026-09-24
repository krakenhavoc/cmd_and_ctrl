package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mass Diminish — Sorcery for {1}{U}:
//
//	"Until your next turn, creatures target player controls have base
//	 power and toughness 1/1.
//	 Flashback {3}{U}"
//
// The CR 611.2b turn-boundary card, and the reason "until your next
// turn" could not be expressed before S38: `Turn.Round` counts
// ROUNDS, so all four seats in a Commander game share one number and
// there was nothing to compare against. The duration now ends on
// `Player.TurnsBegun` (ADR 0063 Decision 3), so this really does sit
// through all three opponents' turns and end as yours begins.
//
// Two things the card gets from the model and does not have to write:
//
//   - **The affected set is locked in at resolution (CR 611.2c).** A
//     creature that player casts afterwards is a normal creature; the
//     ones on the battlefield when this resolved are 1/1s, and one
//     that leaves and comes back is a new object (CR 400.7) and is
//     not affected either. `SnapshotAffected` is that key.
//   - **Layer 7b, not 7c.** "Base power and toughness" SETS the
//     value, so it applies before every +1/+1 anthem and after every
//     CDA, and a lord's +1/+1 still makes these 2/2s. A 7c "-X/-X"
//     would have been a different card.
//
// Flashback is the engine's ordinary alternative cost, so the
// graveyard cast opens the same targeting prompt and installs the same
// duration — and the second copy's "your next turn" is stamped fresh,
// because the duration is read from the game at resolution rather than
// baked into the card.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:         "3a3fd911-b71c-49eb-aadb-cfa80b9e3a3a",
		Name:             "Mass Diminish",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}")},
		Targets:          TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			victim := item.Targets[0].ID
			applies := SnapshotAffected(ctx, And(Creature(), ControlledBy(victim)))
			if applies == nil {
				return nil
			}
			return StaticForDuration{
				Ability: game.StaticAbility{
					Layer:     game.Layer7PT,
					SubLayer:  game.SubLayer7B_Set,
					AppliesTo: applies,
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power = 1
						c.Toughness = 1
					},
				},
				Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
				Label:    "Mass Diminish — base 1/1 until your next turn",
			}.Apply(ctx)
		},
	})
}
