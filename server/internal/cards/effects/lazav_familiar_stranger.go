package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lazav, Familiar Stranger — Legendary Creature — Shapeshifter
// {1}{U}{B}, 1/4:
//
//	"Whenever you commit a crime, put a +1/+1 counter on Lazav. Then
//	 you may exile a card from a graveyard. If a creature card was
//	 exiled this way, you may have Lazav become a copy of that card
//	 until end of turn. This ability triggers only once each turn.
//	 (Targeting opponents, anything they control, and/or cards in
//	 their graveyards is a crime.)"
//
// "Commit a crime" reuses Magda, the Hoardmaster's reading of it
// (batch18_helpers.go's b18CommittedCrime, EventBecomesTarget) and the
// once-per-turn tally (b11TriggeredThisTurn) her ability already
// exercises.
//
// # The exile does not target
//
// The printed clause is "exile A card from A graveyard" — no "target"
// — so it is a resolution-time CHOICE (CR 601.2c does not apply to
// it), not a target clause: hexproof and shroud are irrelevant, and
// the pick spans every graveyard at the table, the controller's own
// included. That candidate set crosses more than one player's zone,
// so it cannot use the single-owner Zone re-check
// (game.ChooseCardsPrompt.Zone / FromPlayer, the shape Skullwinder and
// Ripples of Undeath use for a pick over ONE player's graveyard).
// Instead the prompt leaves Zone unset — "no zone re-check; the
// frame's `then` owns the validation" (chained_choice.go) — and
// re-validates through ExileTarget's own existence handling, which
// already does nothing rather than erroring when the chosen card has
// moved by the time the pick is submitted.
//
// # The copy is a SECOND "may", not the trigger's OptionalPrompt
//
// The trigger's own OptionalPrompt (game.TriggerOptionalPrompt)
// decides whether to BUILD the ability at all, before anything has
// happened. This ability has already committed to the counter and the
// exile by the time the copy is offered, so "you may have Lazav become
// a copy" is a plain mid-resolution yes/no — game.ConfirmPrompt, the
// same shape Ponder's "you may shuffle" uses — asked only when a
// creature card was actually exiled (CR 608.2c: a "may" that isn't
// reached is simply not asked).
//
// The copy itself is a duration copy (#1593, become_copy.go), UNTIL
// END OF TURN: the printed sentence has no "except" clause and grants
// nothing back, unlike Lazav, Dimir Mastermind's and Dimir
// Doppelganger's indefinite copies, so there is no ability to keep
// alive across the copy and no grant to declare.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca60623a-1a98-4153-a0ee-2f34981de5cb",
		Name:         "Lazav, Familiar Stranger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18CommittedCrime(ev, source, g) &&
					!b11TriggeredThisTurn(g, source.InstanceID, lazavFamiliarStrangerLabel)
			}, lazavFamiliarStrangerLabel, lazavFamiliarStrangerEffect),
		},
	})
}

const lazavFamiliarStrangerLabel = "Lazav, Familiar Stranger — a +1/+1 counter, then exile a card from a graveyard"

// lazavFamiliarStrangerEffect is the trigger's whole body: the
// counter, then the exile choice, then — only if a creature card came
// of it — the copy offer.
func lazavFamiliarStrangerEffect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := ctx.Source()
	if err := (AddCounter{Target: self, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}

	var candidates []uuid.UUID
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		candidates = append(candidates, graveyardCardsOfPlayer(g, p.ID)...)
	}
	if len(candidates) == 0 {
		return nil
	}

	resolving := item
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   self,
		Question: "Lazav, Familiar Stranger — you may exile a card from a graveyard",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		// No Zone: the candidates span every seat's graveyard, not one
		// player's, so there is no single owner to re-check against.
		// ExileTarget below does its own existence handling for a
		// card that has moved by the time this is answered.
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return ExileTarget{Target: picked[0], Then: func(ctx *Context, exiled bool) error {
				if !exiled {
					return nil
				}
				return lazavFamiliarStrangerOfferCopy(ctx, resolving, self, picked[0])
			}}.Apply(NewContext(g, resolving))
		},
	})
	return nil
}

// lazavFamiliarStrangerOfferCopy is "if a creature card was exiled
// this way, you may have Lazav become a copy of that card until end
// of turn" — asked only when the exile actually produced a creature
// card, per CR 608.2c.
func lazavFamiliarStrangerOfferCopy(ctx *Context, resolving *game.StackItem, self, card uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(card)
	if !ok || !c.IsCreature() {
		return nil
	}
	own := game.OwnPrintedValues(c)
	ctx.Game.QueueConfirmForEffect(game.ConfirmPrompt{
		Chooser:      ctx.Controller(),
		Source:       self,
		Question:     "Lazav, Familiar Stranger — have Lazav become a copy of " + c.Name + " until end of turn?",
		AcceptLabel:  "Become a copy",
		DeclineLabel: "Decline",
		OnAccept: func(g *game.Game) error {
			return BecomeCopy{
				Targets: []uuid.UUID{self},
				Of:      card,
				Values:  &own,
				Label:   "Lazav, Familiar Stranger — becomes a copy until end of turn",
			}.Apply(NewContext(g, resolving))
		},
	})
	return nil
}
