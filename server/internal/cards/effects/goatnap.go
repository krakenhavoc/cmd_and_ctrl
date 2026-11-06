package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goatnap — Sorcery {2}{R} (Lorwyn Eclipsed, #4458):
//
//	"Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn. If that creature is a
//	 Goat, it also gets +3/+0 until end of turn."
//
// Act of Treason's three clauses in printed order, then Little Bear's
// subtype check: "Goat" is read once the creature is yours and untapped,
// from its current types, so a changeling is a Goat. "Goat" is a word of
// this card's text (CR 612.2), so a text change moves it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "060e7187-0c7d-4022-9a1f-36a6f75314ca",
		Name:         "Goatnap",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Goatnap — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Goatnap — haste until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			c, found := ctx.Game.LookupCardForEffect(target)
			if !found || !onBattlefield(ctx.Game, target) || !c.HasSubtype(ctx.TextSubtype("Goat")) {
				return nil
			}
			return BoostUntilEOT{Target: target, Power: 3, Label: "Goatnap — +3/+0 for a Goat"}.Apply(ctx)
		},
	})
}
