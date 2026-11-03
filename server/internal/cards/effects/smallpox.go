package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Smallpox — Sorcery {B}{B} (Secrets of Strixhaven Mystical Archive,
// #4834):
//
//	"Each player loses 1 life, discards a card, sacrifices a creature of
//	 their choice, then sacrifices a land of their choice."
//
// Four instructions done to every player, one after the other
// (CR 608.2c), each finishing before the next begins — the rulings spell
// it out: life first, then the discards, then the creatures, then the
// lands.
//
//   - Each player (the caster included) loses 1 life.
//   - Each player discards a card, asked in APNAP order
//     (EachPlayerDiscardsThenKeyedForEffect); a player with an empty hand
//     is not asked (CR 701.9a).
//   - Once every discard has landed, each player sacrifices a creature
//     (EachPlayerSacrifices): APNAP choices, each seeing the earlier ones,
//     and every chosen creature leaves together (CR 101.4, 608.2e).
//   - Once those have gone, the same for a land. Lands and creatures that
//     are both are one permanent and may be chosen for either step.
//
// The continuations are registered (smallpoxSacrificeCreatures,
// smallpoxSacrificeLands), so a table waiting on any seat's choice is a
// restore point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7bd22c98-0782-447f-a5fa-4281585a3eab",
		Name:         "Smallpox",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, id := range tablePlayers(ctx) {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -1); err != nil {
						return err
					}
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				caster := ctx.Controller()
				return ctx.Game.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
					game.DiscardPrompt{
						Source:   ctx.Source(),
						N:        1,
						Question: "Smallpox — discard a card",
					},
					smallpoxSacrificeCreatures, []uuid.UUID{caster})
			}),
		),
	})
}

// smallpoxSacrificeCreatures is "sacrifices a creature of their choice",
// run once the last discard has landed.
var smallpoxSacrificeCreatures = RunThen("prompt-run/smallpox-sacrifice-creatures", func(ctx *Context, _ game.RunDone) error {
	return EachPlayerSacrifices{
		Match: Creature(),
		Label: "a creature",
		Keyed: smallpoxSacrificeLands,
	}.Apply(ctx)
})

// smallpoxSacrificeLands is "then sacrifices a land of their choice", run
// once the creatures have left.
var smallpoxSacrificeLands = RegisterEdictThen("prompt-run/smallpox-sacrifice-lands", func(ctx *Context, _ game.PromptedSacrifices) error {
	return EachPlayerSacrifices{
		Match: Land(),
		Label: "a land",
	}.Apply(ctx)
})
