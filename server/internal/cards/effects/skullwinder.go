package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skullwinder — Creature — Snake, {2}{G}, 1/1:
//
//	"Deathtouch
//	 When this creature enters, return target card from your graveyard
//	 to your hand, then choose an opponent. That player returns a card
//	 from their graveyard to their hand."
//
// The Regrowth you have to share, and the second half of #929: the
// controller names the opponent, and then that OPPONENT picks the
// card. Two prompts to two different seats out of one trigger — the
// first a target chosen as the ability goes on the stack (CR 603.3d),
// the second a player chosen while it resolves, which nothing can
// respond to.
//
// The opponent's pick is over their OWN graveyard, a public zone, so
// it needs no reveal: the redaction pass hands the chooser cards they
// are already a knower of. That is what makes this the cheap half of
// the "pick from a graveyard" shape — the expensive one is a pick over
// somebody ELSE's graveyard, which is a different seam.
//
// "Return target card from your graveyard" is mandatory, so with an
// empty graveyard the trigger has no legal target and is removed
// without a prompt (CR 603.3d) — and the opponent's half goes with it,
// which is the printed outcome.
//
// Deathtouch is printed card data and arrives from the deck import,
// not from this entry.
func init() {
	Register(Spec{
		OracleID:     "d3d8dd6e-a80b-4063-a542-ff67ef8dad4f",
		Name:         "Skullwinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Skullwinder — return a card, then an opponent returns one", skullwinderReturn),
				TargetCardInGraveyard("target card from your graveyard", YouOwn()),
			),
		},
	})
}

// skullwinderReturn is the controller's half, then the choice.
func skullwinderReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := b34ReturnChosenGraveyardCardToHand(ctx); err != nil {
		return err
	}
	return ChoosePlayer{
		Among:    Opponents,
		Question: "Skullwinder — choose an opponent to return a card from their graveyard",
		Then:     skullwinderOpponentReturn,
	}.Apply(ctx)
}

// skullwinderOpponentReturn queues the chosen opponent's own pick.
//
// Mandatory for them — "that player returns a card" — so the floor is
// one, and an opponent with an empty graveyard is not prompted at all
// (CR 608.2, as much as it can).
func skullwinderOpponentReturn(ctx *Context) error {
	opp := ctx.ChosenPlayer()
	cards := graveyardCardsOfPlayer(ctx.Game, opp)
	if opp == uuid.Nil || len(cards) == 0 {
		return nil
	}
	// The resolving item, lifted out so the continuation can bind a
	// FRESH Context to it against whichever *Game the resolver hands
	// back — the undo rule every chained prompt here follows (see
	// may_choice.go).
	resolving := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opp,
		FromPlayer: opp,
		Source:     ctx.Source(),
		Question:   "Skullwinder — return a card from your graveyard to your hand",
		Cards:      cards,
		Min:        1,
		Max:        1,
		// Re-checked on submit: a card can leave the graveyard
		// between the question and the answer.
		Zone: game.ZoneGraveyard,
		Then: ReturnPickedToHand(resolving),
	})
	return nil
}

// graveyardCardsOfPlayer lists a player's graveyard in zone order —
// the candidate set behind "a card from their graveyard".
//
// Caller must hold g.mu — it is an effect-time read.
func graveyardCardsOfPlayer(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	if playerID == uuid.Nil {
		return nil
	}
	p := g.PlayerByIDForEffect(playerID)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}
