package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lembas — Artifact — Food {2}:
//
//	"When this artifact enters, scry 1, then draw a card.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life.
//	 When this artifact is put into a graveyard from the battlefield,
//	 its owner shuffles it into their library."
//
// The scry is a queued prompt, so the draw is the Scry primitive's
// Then: written as the next statement it would take a card the player
// is still deciding about. The shuffle-back is a dies trigger, not a
// replacement — the Lembas really does reach the graveyard first, so a
// graveyard payoff sees it — and it does nothing if the card has left
// the graveyard by the time the trigger resolves (it would be a new
// object, CR 400.7). It is the OWNER'S library and shuffle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f71fcdc3-5e96-416e-a49d-37019806e2e2",
		Name:         "Lembas",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Lembas — scry 1, then draw a card", lembasScryThenDraw),
			WhenThisDies("Lembas — its owner shuffles it into their library", lembasShuffleBack),
		},
		Activated: []ActivatedAbility{samiFoodLifeAbility()},
	})
}

func lembasScryThenDraw(g *game.Game, item *game.StackItem) error {
	c := item.Controller
	return Scry{Player: c, N: 1, Then: func(g *game.Game) error {
		return g.DrawNForEffect(c, 1)
	}}.Apply(NewContext(g, item))
}

func lembasShuffleBack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	if !diedCardStillInGraveyard(ctx, id) {
		return nil
	}
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	owner := card.Owner
	return PutIntoLibrary{Card: id, Then: func(g *game.Game, placed bool) error {
		if !placed {
			return nil
		}
		return g.ShuffleLibraryForEffect(owner)
	}}.Apply(ctx)
}
