package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brave the Wilds — Sorcery {G} (Wilds of Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 If this spell was bargained, target land you control becomes a 3/3
//	 Elemental creature with haste that's still a land.
//	 Search your library for a basic land card, reveal it, put it into your
//	 hand, then shuffle."
//
// The land target exists only when the spell is bargained (the rulings):
// the bargain cost is WhenPaid's, which appends the "target land you
// control" clause to a spell that prints none, so an unbargained cast
// names no target and a bargained one can't be cast without a land to
// name. If the land is an illegal target as the spell resolves, the spell
// doesn't resolve and doesn't search (CR 608.2b, the rulings).
//
// The animation has no duration (CR 611.2a), so it lasts as long as the
// land stays that object (CR 400.7): a layer 4 type and subtype change, a
// layer 6 haste grant and a layer 7b 3/3 base, in one record pinned to the
// land, Awaken's animation with a 3/3 base. It is still a land, so it
// keeps its land type and abilities. The search is the Down in the Valley
// chapter's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "49d189dd-9ce7-434f-931b-7110cd5255e7",
		Name:          "Brave the Wilds",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tutors: 1},
		OptionalCosts: []game.AdditionalCost{WhenPaid(g2Bargain(), AwakenLandTargets())},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if !g2WasBargained(ctx) {
					return nil
				}
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ScopedEffectFor{
					Target: id,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
						game.AddKeywordsMod("haste"),
						game.SetBasePowerMod(3),
						game.SetBaseToughnessMod(3),
					},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Brave the Wilds — becomes a 3/3 Elemental creature with haste",
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:    ctx.Controller(),
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Brave the Wilds — a basic land card",
					Source:    ctx.Source(),
				}.Apply(ctx)
			}),
		),
	})
}
