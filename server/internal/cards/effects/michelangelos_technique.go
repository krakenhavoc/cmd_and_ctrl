package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Michelangelo's Technique — Sorcery {4}{G} (TMT, #4457):
//
//	"Sneak {3}{G} (You may cast this spell for {3}{G} if you also return
//	 an unblocked attacker you control to hand during the declare
//	 blockers step.)
//	 Look at the top eight cards of your library. Put up to two creature
//	 cards with total mana value 6 or less from among them onto the
//	 battlefield and the rest on the bottom of your library in a random
//	 order."
//
// Sneak is Sneak(cost) (CR 702.190). The look and put are Ao, the Dawn
// Sky's first bullet: "total mana value 6 or less" judges the set picked,
// a registered put rule (b23TotalManaValueAtMost), and "up to two" is the
// pick's bounds.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "42f12942-9ac1-497e-a8b7-9f66a5afd0ae",
		Name:             "Michelangelo's Technique",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{3}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := item.Controller
			return PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 8),
				Match:    Creature(),
				Max:      2,
				Optional: true,
				Rule:     michelangelosTechniquePick,
				Label:    "Michelangelo's Technique — put up to two creature cards with total mana value 6 or less onto the battlefield",
				ThenTail: PutRestOnBottomInRandomOrderTail,
			}.Apply(ctx)
		},
	})
}

// michelangelosTechniquePick is "with total mana value 6 or less",
// registered (#4253). An on-disk key.
var michelangelosTechniquePick = RegisterPutRule("choose-cards/put-from-library-michelangelos-technique-total-mana-value-6", b23TotalManaValueAtMost(6))
