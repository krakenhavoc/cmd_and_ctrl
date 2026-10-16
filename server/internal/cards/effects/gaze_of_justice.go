package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gaze of Justice — Sorcery {W} (#3868):
//
//	"As an additional cost to cast this spell, tap three untapped white
//	 creatures you control.
//	 Exile target creature.
//	 Flashback {5}{W} (You may cast this card from your graveyard for its
//	 flashback cost and any additional costs. Then exile it.)"
//
// The additional cost is New Blood's (TapUntappedCost): three untapped
// white creatures the caster controls, named on cast_spell as
// teamwork_ids. It is an additional cost, so the flashback cast pays it
// too, beside {5}{W} (CR 702.34a, 601.2f). Colour is read off the
// creature as it is now, so a creature painted white counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "07c7e4d4-6452-437b-9f17-b0ec1fd9f702",
		Name:             "Gaze of Justice",
		Completeness:     CompletenessFull,
		AdditionalCost:   TapUntappedCost(3, "three untapped white creatures you control", Creature(), OfColor("W")),
		Targets:          TargetCreature("target creature"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{W}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
