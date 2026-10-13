package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Persist — Sorcery {1}{B} (EDHREC rank 1347):
//
//	"Return target nonlegendary creature card from your graveyard to
//	 the battlefield with a -1/-1 counter on it."
//
// Two-mana reanimation with a legend clause and a tax: the
// nonlegendary restriction is the target predicate (Not(Legendary),
// on the printed supertype — a card in a graveyard has no layer
// cache), "your graveyard" is ownership, and the card lands under
// its owner's control, which for a card from your own graveyard is
// you (Zombify's shape, not Reanimate's).
//
// The -1/-1 counter is an "enters with" clause (CR 614.1c): it rides the
// entry event, so a Hardened Scales or Doubling Season sees it, the
// creature's own enters-the-battlefield abilities already see it
// shrunk, and it still lands when the creature first asks a question
// as it enters (a Clone's copy pick, a shockland's payment).
func init() {
	Register(Spec{
		OracleID:     "367d4cf2-270f-4236-af10-7d5e8ea8c9fe",
		Name:         "Persist",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target nonlegendary creature card from your graveyard",
			YouOwn(), Creature(), Not(b05Legendary())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			targets := ctx.Targets()
			if len(targets) == 0 || targets[0].Kind != game.TargetCard {
				return nil
			}
			_, err := ctx.Game.ReturnFromGraveyardWithCountersForEffect(targets[0].ID, uuid.Nil, false,
				map[string]int{"-1/-1": 1})
			if errors.Is(err, game.ErrCardNotFound) {
				return nil // left the graveyard in response (CR 608.2b)
			}
			return err
		},
	})
}
