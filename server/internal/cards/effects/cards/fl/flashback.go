package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flashback — Instant {R} (Secrets of Strixhaven, #4456):
//
//	"Target instant or sorcery card in your graveyard gains flashback until
//	 end of turn. The flashback cost is equal to its mana cost. (You may cast
//	 that card from your graveyard for its flashback cost. Then exile it.)"
//
// Snapcaster Mage's clause (GrantFlashbackToCard) as a spell: the permission
// names that card object (CR 400.7), its cost is the card's own mana cost,
// and because it is flashback (CR 702.34a) the card is exiled instead of
// going anywhere else as it leaves the stack. The permission lapses at the
// end of the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02070488-9203-4304-9392-a111d20218c5",
		Name:         "Flashback",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard",
			effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			return effects.GrantFlashbackToCard{
				Target: ts[0].ID,
				Label:  "Flashback — its mana cost (Flashback)",
			}.Apply(ctx)
		},
	})
}
