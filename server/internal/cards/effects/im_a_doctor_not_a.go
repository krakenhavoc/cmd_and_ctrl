package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// I'm a Doctor, Not a . . . — {G} Instant (Star Trek, #4973):
//
//	"Put a +1/+1 counter on target creature you control. Untap it. Until
//	 end of turn, it gets +1/+1 and becomes a Doctor. (It loses all other
//	 creature types.)"
//
// Soldier On's counter-then-untap on the one target, then two until-end-
// of-turn effects: the +1/+1 (layer 7c) and a layer-4 replacement of the
// creature types alone (SetCreatureTypesMod, CR 205.1a), so it keeps its
// card types and abilities and is only a Doctor until the turn ends. A
// target that is gone or no longer yours does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad1d8eca-e3bf-479f-a442-bcd8c31bfe7f",
		Name:         "I'm a Doctor, Not a . . .",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1)),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "I'm a Doctor, Not a . . . — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes("Doctor")...)},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "I'm a Doctor, Not a . . . — a Doctor until end of turn",
			}.Apply(ctx)
		},
	})
}
