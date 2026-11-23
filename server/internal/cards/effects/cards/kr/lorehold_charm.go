package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lorehold Charm — Instant {R}{W} (Secrets of Strixhaven, #4456):
//
//	"Choose one —
//	 • Each opponent sacrifices a nontoken artifact of their choice.
//	 • Return target artifact or creature card with mana value 2 or less
//	   from your graveyard to the battlefield.
//	 • Creatures you control get +1/+1 and gain trample until end of turn."
//
// The first bullet is an untargeted edict (EachPlayerSacrifices, every
// opponent choosing their own, CR 701.21a) over nontoken artifacts. The
// second is Inscription of Ruin's reanimation with a wider type: the target
// is chosen on the cast and a card that left the graveyard in response
// does nothing (CR 608.2b). The third locks the set of creatures you control
// as it resolves (CR 611.2c): a creature that enters afterwards gets
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5729f11-6cde-4830-b882-ab4c5251c1a9",
		Name:         "Lorehold Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Each opponent sacrifices a nontoken artifact of their choice.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.EachPlayerSacrifices{
						ExceptController: true,
						Match:            effects.And(effects.Artifact(), effects.Not(effects.IsTokenPredicate())),
						Label:            "a nontoken artifact",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Return target artifact or creature card with mana value 2 or less from your graveyard to the battlefield.",
				effects.TargetCardInGraveyard("target artifact or creature card with mana value 2 or less from your graveyard",
					effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature()), effects.ManaValueLE(2)),
				effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
			effects.ModeDoing("Creatures you control get +1/+1 and gain trample until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					mine := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Lorehold Charm — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Lorehold Charm — trample"}.Apply(ctx)
				}),
		),
	})
}
