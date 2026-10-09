package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Konstrari Charm — Instant {R}{G}:
//
//	"Choose one —
//	 • Konstrari Charm deals 6 damage to target creature with flying.
//	 • Put two +1/+1 counters on target creature. It gains trample
//	   until end of turn.
//	 • Add {C}{C}{C}."
//
// Mode one's "with flying" is a target restriction read through
// HasKeyword, so a granted flying counts. Mode three is a spell
// adding mana (it uses the stack and can be countered), not a mana
// ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40961b23-b153-434a-aeb9-b9c2ff613dcb",
		Name:         "Konstrari Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Konstrari Charm deals 6 damage to target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying"))),
			Mode("Put two +1/+1 counters on target creature. It gains trample until end of turn.",
				TargetCreature("target creature")),
			Mode("Add {C}{C}{C}."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				t, ok := ModeTarget(ctx, 0)
				if !ok {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 6}.Apply(ctx)
			case ctx.HasMode(1):
				t, ok := ModeTarget(ctx, 0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"trample"},
					Label:    "Konstrari Charm — trample",
				}.Apply(ctx)
			case ctx.HasMode(2):
				return AddMana{Produced: "{C}{C}{C}"}.Apply(ctx)
			}
			return nil
		},
	})
}
