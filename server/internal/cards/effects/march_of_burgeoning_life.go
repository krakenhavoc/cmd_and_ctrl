package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// March of Burgeoning Life — Instant {X}{G} (Kamigawa: Neon Dynasty):
//
//	"As an additional cost to cast this spell, you may exile any number
//	 of green cards from your hand. This spell costs {2} less to cast for
//	 each card exiled this way.
//	 Choose target creature with mana value less than X. Search your
//	 library for a creature card with the same name as that creature, put
//	 it onto the battlefield tapped, then shuffle."
//
// The cost is March of Swirling Mist's. "Mana value less than X" is the
// announced-X bound shifted down by one (TargetSpec.WithManaValueLessThanX),
// judged at announce and again at the CR 608.2b re-check. The name is the
// target's, read as it is now (the spell only resolves while the target is
// still legal); the search is for a creature card with exactly that name and
// may find nothing (CR 701.23b). The library is shuffled either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "a81fd70c-c2bc-4863-97a0-facde86f024c",
		Name:           "March of Burgeoning Life",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: ExileAnyNumberFromHandCost("green cards", "G"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerExiledFromHand("{2}", "This spell costs {2} less to cast for each card exiled this way"),
		},
		Targets: TargetCreature("target creature with mana value less than X").WithManaValueLessThanX(),
		Purpose: game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID)
			if !ok {
				return nil
			}
			name := target.Name
			return SearchLibrary{
				Player: item.Controller,
				Predicate: func(c game.Card) bool {
					return c.IsCreature() && c.Name == name
				},
				Dest:          game.ZoneBattlefield,
				Limit:         1,
				TappedOnEntry: true,
				Shuffle:       true,
				Reason:        "March of Burgeoning Life — a creature card named " + name,
			}.Apply(ctx)
		},
	})
}
