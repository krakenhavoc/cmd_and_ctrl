package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Systems Override — Sorcery {2}{R} (Edge of Eternities, #4461):
//
//	"Gain control of target artifact or creature until end of turn.
//	 Untap that permanent. It gains haste until end of turn. If it's a
//	 Spacecraft, put ten charge counters on it. If you do, remove ten
//	 charge counters from it at the beginning of the next end step."
//
// Act of Treason's three clauses over an artifact or a creature. The ten
// charge counters push any Spacecraft past every station threshold, so
// it is an artifact creature (and, with haste, can attack) for the turn.
// "If you do" is the counter placement's continuation: only counters that
// really landed schedule the removal, which is a delayed trigger
// (CR 603.7) at the next end step that takes ten off (or what is left).
// Whoever controls the Spacecraft then, the counters come off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c9d9d47-1eb6-4981-ba83-d5ec04906e71",
		Name:         "Systems Override",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (GainControl{
					Target:   t.ID,
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Systems Override — gain control until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{
					Target:   t.ID,
					Keywords: []string{"haste"},
					Label:    "Systems Override — haste until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || !c.HasSubtypeWord("Spacecraft") {
					return nil
				}
				tail := ctx.ItemTail(stdB3ScheduleChargeRemoval)
				tail.Carry = stdB3Taken(t.ID)
				return ctx.Game.AddCounterByKeyedForEffect(item.Controller, t.ID, game.CounterCharge, 10, tail)
			}
			return nil
		},
	})
}
