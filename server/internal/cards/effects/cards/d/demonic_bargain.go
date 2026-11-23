package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Bargain — Sorcery {2}{B} (Innistrad: Crimson Vow):
//
//	"Exile the top thirteen cards of your library, then search your
//	 library for a card. Put that card into your hand, then shuffle."
//
// Two steps in order. The top thirteen are exiled through MillToZone, so a
// commander among them offers its owner the command zone (CR 903.9), and the
// search waits for the exile to settle. A library of fewer than thirteen
// exiles what is there and searches what is left (possibly nothing). The
// search is Demonic Tutor's: any card, chosen by the controller, then a
// shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae18580a-1753-4dc8-b437-8a779b1a6e4d",
		Name:         "Demonic Bargain",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{
				Player:   ctx.Controller(),
				N:        13,
				To:       game.ZoneExile,
				ThenTail: ctx.ItemTail(demonicBargainSearch),
			}.Apply(ctx)
		},
	})
}

// demonicBargainSearch is the clause after the exile: Demonic Tutor's
// search. An on-disk key.
var demonicBargainSearch = effects.CardTail("then/demonic-bargain-search", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.SearchLibrary{
		Player:  ctx.Controller(),
		Dest:    game.ZoneHand,
		Limit:   1,
		Reveal:  false,
		Shuffle: true,
		Reason:  "Demonic Bargain — search your library for a card",
	}.Apply(ctx)
})
