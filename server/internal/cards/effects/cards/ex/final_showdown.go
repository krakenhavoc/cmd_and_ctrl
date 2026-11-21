package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Final Showdown — Instant {W}:
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — All creatures lose all abilities until end of turn.
//	 + {1} — Choose a creature you control. It gains indestructible until
//	   end of turn.
//	 + {3}{W}{W} — Destroy all creatures."
//
// Spree (CR 702.172a): each chosen bullet adds its cost, and the
// bullets resolve in printed order (CR 608.2c) whatever order they were
// announced in, so the indestructible creature is chosen before the
// wrath and survives it. The engine parks the third bullet while the
// second's choice is open (ModeDoing's walk).
//
// The first bullet locks its set as it resolves (CR 611.2c): creatures
// on the battlefield then lose their abilities until end of turn, and
// one that enters later keeps its own. The second is a choice, not a
// target, so a creature with hexproof can be chosen; with no creature
// it asks nothing. The third destroys every creature; regeneration is
// allowed, since the card doesn't forbid it.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e7ec3d6-a84f-4cc3-93f4-4d181d41e126",
		Name:         "Final Showdown",
		Completeness: effects.CompletenessFull,
		Modes: effects.Spree(
			effects.SpreeModeDoing("All creatures lose all abilities until end of turn.", "{1}", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.ScopedEffectFor{
						Match:    effects.Creature(),
						Mods:     []game.Mod{game.LoseAllAbilitiesMod()},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Final Showdown — loses all abilities until end of turn",
					}.Apply(ctx)
				}),
			effects.SpreeModeDoing("Choose a creature you control. It gains indestructible until end of turn.", "{1}", nil,
				finalShowdownIndestructible),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Destroy all creatures.", "{3}{W}{W}", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.WrathDestroyAllCreatures(item, ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}),
		),
	})
}

// finalShowdownIndestructible is the second bullet: a creature its
// controller chooses as it resolves gains indestructible until end of
// turn.
func finalShowdownIndestructible(item *game.StackItem, ctx *effects.Context, _ int) error {
	var ids []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.YouControl())) {
		ids = append(ids, c.InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	return effects.ChoosePermanents{
		Question: "Final Showdown — choose a creature you control to gain indestructible until end of turn",
		Candidates: func(_ *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
			return ids, 1, 1
		},
		Keyed: finalShowdownChosen,
	}.Apply(ctx)
}

// finalShowdownChosen is the chosen creature gaining indestructible until
// end of turn, keyed (#4290).
var finalShowdownChosen = effects.RegisterPermanentsPick("prompt-run/final-showdown-indestructible", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !ctx.Game.Battlefield.Contains(id) {
				continue
			}
			if err := (effects.GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"indestructible"},
				Label:    "Final Showdown — indestructible until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
