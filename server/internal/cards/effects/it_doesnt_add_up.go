package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// It Doesn't Add Up — Instant {3}{B}{B}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield. Suspect it. (It has menace and can't block.)"
//
// Zombify's return, then the suspect on the permanent that came back.
// "It" is the NEW object (CR 400.7), so the suspect is aimed at the id
// the return reports, not at the card's old one.
//
// One declared simplification, weaker than printed: if the creature
// asks a question as it enters (a copy effect, a shockland's payment)
// the entry completes from that answer and the return reports nothing
// here, so it is not suspected.
func init() {
	Register(Spec{
		OracleID:     "dcfc8c59-d032-414d-9da6-125ebcb1366c",
		Name:         "It Doesn't Add Up",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If the creature asks a question as it enters, such as a copy effect, it comes back but isn't suspected."},
		Targets:      targetCreatureInYourGraveyard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			entered, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(target, ctx.Controller(), false, nil)
			if err != nil {
				if errors.Is(err, game.ErrCardNotFound) {
					return nil // CR 608.2b: it left the graveyard in response
				}
				return err
			}
			if entered == uuid.Nil {
				return nil
			}
			return Suspect{Target: entered}.Apply(ctx)
		},
	})
}
