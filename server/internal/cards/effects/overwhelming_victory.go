package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overwhelming Victory — Instant — Lesson {4}{R} (Avatar: The Last
// Airbender Eternal, #5075):
//
//	"Overwhelming Victory deals 5 damage to target creature. Each creature
//	 you control gains trample and gets +X/+0 until end of turn, where X
//	 is the amount of excess damage dealt this way."
//
// The excess is CR 120.4a's: the damage that landed beyond what the
// creature still needed to die, damage already marked included, and with
// a deathtouch source only the first point is needed
// (socB20DealDamageWithExcess). A prevented blow leaves less excess. The
// creatures you control are read once after the damage (CR 611.2c); they
// gain trample even when there is no excess.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20d94561-7526-4305-8d2d-7daf62729ff4",
		Name:         "Overwhelming Victory",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			excess, err := socB20DealDamageWithExcess(ctx, id, 5)
			if err != nil {
				return err
			}
			yours := And(Creature(), YouControl())
			if err := (GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: "Overwhelming Victory — trample"}).Apply(ctx); err != nil {
				return err
			}
			return BoostUntilEOT{Match: yours, Power: excess, Label: "Overwhelming Victory — +X/+0"}.Apply(ctx)
		},
	})
}
