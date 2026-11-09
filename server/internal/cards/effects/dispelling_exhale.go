package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dispelling Exhale — Instant {1}{U} (Tarkir: Dragonstorm, #4463):
//
//	"As an additional cost to cast this spell, you may behold a Dragon.
//	 (You may choose a Dragon you control or reveal a Dragon card from
//	 your hand.)
//	 Counter target spell unless its controller pays {2}. If a Dragon was
//	 beheld, counter that spell unless its controller pays {4} instead."
//
// The behold is the optional additional cost of OptionalBehold (Hulk's
// Thunderclap): the caster names the Dragon on reveal_ids when they take
// the offer. "If a Dragon was beheld" is a fact about the announcement
// (CR 701.4b), asked through Context.WasBeheld. The payment is the
// spell's controller's question (CR 118.12, CounterUnlessPaid): {2}, or
// {4} instead for a beholding cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "dd3a4b64-0987-4e9b-a18d-c54365438857",
		Name:          "Dispelling Exhale",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalBehold("a", "Dragon")},
		Targets:       TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			cost := "{2}"
			if ctx.WasBeheld() {
				cost = "{4}"
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     cost,
				Question: "Dispelling Exhale — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
