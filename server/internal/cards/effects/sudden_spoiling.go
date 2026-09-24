package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sudden Spoiling — Instant {1}{B}{B}:
//
//	"Split second (As long as this spell is on the stack, players
//	 can't cast spells or activate abilities that aren't mana
//	 abilities.)
//	 Until end of turn, creatures target player controls lose all
//	 abilities and have base power and toughness 0/2."
//
// A combat trick for the whole of one player's board, behind split
// second (#1519) so they can't pump, sacrifice or flicker in response.
//
// Two continuous effects from one resolution, both pinned to the same
// CR 611.2c set — the creatures that player controls as this
// resolves; one cast afterwards is untouched, and one that blinks is a
// new object (CR 400.7):
//
//   - layer 6, "lose all abilities" (CR 613.1f) — RemovesAbilities,
//     so catalogued triggered, activated, static and mana abilities
//     all go quiet as well as the keyword badges (the LoseAllAbilities
//     shape, pinned to the set rather than to an Aura's host);
//   - layer 7b, "base power and toughness 0/2" — SETS the value, so
//     an anthem's +1/+1 still applies on top (a 1/3), which is the
//     printed interaction and the reason this is not a 7c shrink.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dce202c7-fe8e-462a-858e-7a5a69bd5b6b",
		Name:            "Sudden Spoiling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		Targets:         TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			applies := SnapshotAffected(ctx, And(Creature(), ControlledBy(item.Targets[0].ID)))
			if applies == nil {
				return nil
			}
			duration := DurationUntilEndOfTurn(ctx)
			if err := (StaticForDuration{
				Ability: game.StaticAbility{
					Layer:            game.Layer6Ability,
					RemovesAbilities: true,
					AppliesTo:        applies,
					Apply:            func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
				},
				Duration: duration,
				Label:    "Sudden Spoiling — loses all abilities",
			}).Apply(ctx); err != nil {
				return err
			}
			return StaticForDuration{
				Ability: game.StaticAbility{
					Layer:     game.Layer7PT,
					SubLayer:  game.SubLayer7B_Set,
					AppliesTo: applies,
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power = 0
						c.Toughness = 2
					},
				},
				Duration: duration,
				Label:    "Sudden Spoiling — base 0/2",
			}.Apply(ctx)
		},
	})
}
