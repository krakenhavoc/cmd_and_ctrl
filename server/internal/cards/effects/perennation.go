package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// perennationOracle is Perennation's Scryfall oracle ID.
const perennationOracle = "be5138fd-3e92-42b9-93c5-311e66bd443f"

// Perennation — Sorcery {3}{W}{B}{G}:
//
//	"Return target permanent card from your graveyard to the
//	 battlefield with a hexproof counter and an indestructible counter
//	 on it."
//
// The card #1753 was waiting for (#1117's last seam-blocked card). The
// counters are keyword counters (CR 122.1b), and since ADR 0101 the
// engine reads them itself: nothing on the battlefield has to carry
// the grant, and they work on ANY permanent — an enchantment brought
// back this way is indestructible too.
//
// "With … counters on it" is a CR 614.1c "enters with" clause, so the
// counters ride the entry event (ReturnFromGraveyardWithCountersForEffect,
// Feign Death's door): a counter replacement sees them and an enters
// trigger finds them already there. "From your graveyard" is the owner,
// so the permanent returns under its owner's control — the caster.
func init() {
	Register(Spec{
		OracleID:     perennationOracle,
		Name:         "Perennation",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			_, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(id, ctx.Controller(), false, map[string]int{
				game.CounterHexproof:       1,
				game.CounterIndestructible: 1,
			})
			if errors.Is(err, game.ErrCardNotFound) {
				return nil // CR 400.7: it left the graveyard in response
			}
			return err
		},
	})
}
