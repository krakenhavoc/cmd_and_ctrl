package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grinning Totem — Artifact {4} (Mirage, S82 OLD b808):
//
//	"{2}, {T}, Sacrifice this artifact: Search target opponent's library for a
//	 card and exile it. Then that player shuffles. Until the beginning of your
//	 next upkeep, you may play that card. At the beginning of your next upkeep,
//	 if you haven't played it, put it into its owner's graveyard."
//
// The search is Bribery's over a target opponent's library: the activator
// chooses any card (a non-empty library must give one up, CR 701.23d) and it
// is exiled face up, the opponent's library shuffled either way. The grant is
// a permission for the activator to play it, land or spell, that ends as their
// next upkeep begins (CR 611.2). The delayed trigger waits for that same
// upkeep (CR 603.7); the card is put into its owner's graveyard only if it is
// still in exile, which is exactly "you haven't played it", since playing it
// moves it out. The sacrifice is a cost, so it resolves with the Totem gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa798a3d-7797-4795-b251-a45f0429d169",
		Name:         "Grinning Totem",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search target opponent's library for a card and exile it. Then that player shuffles. Until the beginning of your next upkeep, you may play that card. At the beginning of your next upkeep, if you haven't played it, put it into its owner's graveyard.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return effects.SearchLibrary{
					Player:       ctx.Controller(),
					LibraryOwner: victim,
					Dest:         game.ZoneExile,
					Limit:        1,
					Shuffle:      true,
					Reason:       "Grinning Totem — search target opponent's library for a card and exile it",
					Source:       ctx.Source(),
					ThenTail:     ctx.ItemTail(effects.OldB808GrinningTotemGrant),
				}.Apply(ctx)
			},
		}},
	})
}
