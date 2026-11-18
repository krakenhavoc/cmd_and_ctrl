package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelfire Ignition — Sorcery {1}{R}{W} (Midnight Hunt, #5497):
//
//	"Put two +1/+1 counters on target creature. It gains vigilance,
//	 trample, lifelink, indestructible, and haste until end of turn.
//	 Flashback {2}{R}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Two counters, then the five keywords until end of turn; Flashback is the
// shared constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "29e1c6e3-d76e-4388-b4e9-f732fdec4338",
		Name:             "Angelfire Ignition",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   id,
				Keywords: []string{"vigilance", "trample", "lifelink", "indestructible", "haste"},
				Label:    "Angelfire Ignition — vigilance, trample, lifelink, indestructible and haste",
			}.Apply(ctx)
		},
	})
}
