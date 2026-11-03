package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "a5729f11-6cde-4830-b882-ab4c5251c1a9",
		Name:         "Lorehold Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Each opponent sacrifices a nontoken artifact of their choice.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return EachPlayerSacrifices{
						ExceptController: true,
						Match:            And(Artifact(), Not(IsTokenPredicate())),
						Label:            "a nontoken artifact",
					}.Apply(ctx)
				}),
			ModeDoing("Return target artifact or creature card with mana value 2 or less from your graveyard to the battlefield.",
				TargetCardInGraveyard("target artifact or creature card with mana value 2 or less from your graveyard",
					YouOwn(), Or(Artifact(), Creature()), ManaValueLE(2)),
				rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
					return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
				})),
			ModeDoing("Creatures you control get +1/+1 and gain trample until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					mine := And(Creature(), YouControl())
					if err := (BoostUntilEOT{Match: mine, Power: 1, Toughness: 1, Label: "Lorehold Charm — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: mine, Keywords: []string{"trample"}, Label: "Lorehold Charm — trample"}.Apply(ctx)
				}),
		),
	})
}
