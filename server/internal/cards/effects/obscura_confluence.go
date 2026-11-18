package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obscura Confluence — Instant {1}{W}{U}{B} (Streets of New Capenna
// Commander):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Until end of turn, target creature loses all abilities and has base
//	   power and toughness 1/1.
//	 • Target creature connives.
//	 • Target player returns a creature card from their graveyard to their
//	   hand."
//
// Mystic Confluence's shape (CR 700.2d). The first bullet is Warkite
// Marauder's pair of scoped layer effects (layer 6 and 7b) until end of
// turn. The second makes the target creature connive for you (CR 701.50):
// draw, discard, and a +1/+1 counter if the discard was a nonland card; the
// connive still happens if the creature has left. The third has the target
// player choose the creature card as it resolves (Mikey and Mona's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f2fb1dd-82e4-45fa-b2a9-7df7b8a4d533",
		Name:         "Obscura Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeDoing("Until end of turn, target creature loses all abilities and has base power and toughness 1/1.",
				Harmful(TargetCreature("target creature")),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ScopedEffectFor{
						Target:   t.ID,
						Mods:     append([]game.Mod{game.LoseAllAbilitiesMod()}, game.SetBasePTMods(1, 1)...),
						Duration: ctx.Game.UntilEndOfTurnDuration(),
						Label:    "Obscura Confluence — loses all abilities and has base power and toughness 1/1 until end of turn",
					}.Apply(ctx)
				}),
			ModeDoing("Target creature connives.",
				TargetCreature("target creature"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return ConniveEach{Controller: item.Controller, Creatures: []uuid.UUID{t.ID}}.Apply(ctx)
				}),
			ModeDoing("Target player returns a creature card from their graveyard to their hand.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return tmcB8TargetPlayerReturnsFromGraveyard(ctx, t.ID,
						func(c game.Card) bool { return c.IsCreature() },
						"Obscura Confluence — return a creature card from your graveyard to your hand")
				}),
		),
	})
}
