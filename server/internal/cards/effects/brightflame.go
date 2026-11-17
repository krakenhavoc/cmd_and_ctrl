package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brightflame — Sorcery {X}{R}{R}{W}{W}:
//
//	"Radiance — Brightflame deals X damage to target creature and each
//	 other creature that shares a color with it. You gain life equal to
//	 the damage dealt this way."
//
// The creatures are read as the spell resolves (radianceSet, #5736) and
// take their damage as one instruction (radianceDamage). "The damage
// dealt this way" is what the damage walk reports as actually dealt,
// after every prevention and replacement (#807): a creature behind a
// shield or with protection from red or white adds nothing, a doubled
// one adds double, and a creature dealt more than its toughness still
// counts all of it, as damage dealt is not capped by toughness (the
// card's ruling: "You don't gain life for any damage that was
// prevented"). The gain comes from the walk's continuation, so it waits
// with the damage when a CR 616 ordering prompt holds one creature's leg.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6473eac1-6786-47dd-8124-9f1c4b790ddf",
		Name:         "Brightflame",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return radianceDamage(ctx, ctx.X(), ctx.Tail(gainLifeEqualToDamageDealt))
		},
	})
}
