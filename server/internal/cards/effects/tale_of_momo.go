package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tale of Momo — Sorcery {2}{W} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"This spell costs {2} less to cast if a creature left the
//	 battlefield under your control this turn.
//	 Search your library and/or graveyard for an Ally creature card,
//	 reveal it, and put it into your hand. If you search your library
//	 this way, shuffle."
//
// The discount reads the per-player CreaturesLeft tally as the cost is
// determined (CR 601.2f): any creature you controlled that left the
// battlefield this turn, by any route, token or not. The search is
// Delivery Moogle's library-and/or-graveyard search (#3217): the caster
// picks the zones first, and the library is shuffled only if they
// searched it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "728157a3-fcd4-4295-85ff-1162dbb64f8a",
		Name:         "Tale of Momo",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if a creature left the battlefield under your control this turn.",
				func(q game.CostQuery) bool { return avB28CreatureLeftUnderYourControl(q.Game, q.Controller) }),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     func(c game.Card) bool { return c.IsCreature() && c.HasSubtype(ctx.TextSubtype("Ally")) },
				Dest:          game.ZoneHand,
				Limit:         1,
				Reveal:        true,
				Shuffle:       true,
				AlsoGraveyard: true,
				Reason:        "Tale of Momo — an Ally creature card",
			}.Apply(ctx)
		},
	})
}
