package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lattice-Blade Mantis — Creature — Phyrexian Insect {3}{G}, 4/3 (Phyrexia:
// All Will Be One):
//
//	"This creature enters with two oil counters on it.
//	 Whenever this creature attacks, you may remove an oil counter from it.
//	 If you do, untap it and it gets +1/+1 until end of turn."
//
// The "you may" is asked as the trigger resolves (CR 603.5); the untap and
// the pump happen only if a counter really came off this same object.
// Untapping does not remove it from combat (CR 506.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6755a7af-5b55-4245-a777-27e69979479b",
		Name:         "Lattice-Blade Mantis",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b10EntersWithCounters(oilCounter, 2, "Lattice-Blade Mantis: enters with two oil counters on it")},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisAttacks("Lattice-Blade Mantis — you may remove an oil counter from it; if you do, untap it and it gets +1/+1",
				Do(Instruction(latticeBladeMantisRemoveUntapPump))),
				"Lattice-Blade Mantis — remove an oil counter to untap it and give it +1/+1?"),
		},
	})
}

func latticeBladeMantisRemoveUntapPump(ctx *Context) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	src := ctx.Item.SourceCardID
	c, ok := ctx.Game.LookupCardForEffect(src)
	if !ok || c.Counters[oilCounter] <= 0 {
		return nil
	}
	if err := ctx.Game.AddCounterForEffect(src, oilCounter, -1); err != nil {
		return err
	}
	if err := (UntapTarget{Target: src}).Apply(ctx); err != nil {
		return err
	}
	return BoostUntilEOT{Target: src, Power: 1, Toughness: 1, Label: "Lattice-Blade Mantis — +1/+1 until end of turn"}.Apply(ctx)
}
