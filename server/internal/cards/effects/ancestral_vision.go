package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestral Vision — Sorcery with NO mana cost: "Suspend 4—{U} …
// Target player draws three cards."
//
// The card that makes CR 118.6 visible. An empty mana cost is not
// free, it is UNPAYABLE (CR 118.6), so this card can never be cast
// from hand — `HasNoManaCost` plus `castPaysPrintedCost` refuse it,
// and the legal-move enumerator never offers it. Suspend is the whole
// of its playability: pay {U}, wait four of your upkeeps, and cast it
// without paying its mana cost, which CR 118.6a expressly allows.
//
// So this card is also the proof that the suspend free cast prices
// itself as "{0}" rather than as an empty override. An empty cost
// would mean "pay the printed one", and the printed one is the
// unpayable one — the grant would hand the player a card they still
// could not cast (#659, and the reason cascade's grant spells its
// price the same way).
func init() {
	Register(Spec{
		OracleID:     "9728dec9-d482-4c7a-8cdc-44d010dc878d",
		Name:         "Ancestral Vision",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		SpecialActions: []game.SpecialAction{
			Suspend(4, "{U}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return DrawCards{Player: item.Targets[0].ID, N: 3}.Apply(ctx)
		},
	})
}
