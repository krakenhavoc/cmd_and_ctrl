package game

import "github.com/google/uuid"

// sacrifice.go — S21 sub-PR 1: sacrifice as a first-class engine
// operation (CR 701.17). A permanent is sacrificed by its
// controller, as a cost (Goblin Bombardment, Treasure's mana
// ability) or as an effect's instruction. It is NOT destroyed:
// indestructible and regeneration don't apply, and a replacement
// keyed on destruction never sees it.
//
// The mechanics are deliberately thin — the permanent takes the
// ordinary route to its owner's graveyard, so dies-triggers, the
// CR 903.9 commander-zone replacement and LKI all keep working
// untouched. The only new thing is EventSacrifice, emitted while
// the card is still on the battlefield so "whenever you sacrifice"
// payoffs can read its characteristics before it moves.

// sacrificePermanentLocked emits EventSacrifice for the permanent's
// controller and routes it to its owner's graveyard. Caller must
// hold g.mu.
func (g *Game) sacrificePermanentLocked(cardID uuid.UUID) error {
	controller := g.controllerOfBattlefieldCardLocked(cardID)
	if controller == uuid.Nil {
		return ErrCardNotFound
	}
	g.EmitEvent(Event{
		Kind:   EventSacrifice,
		Actor:  controller,
		CardID: cardID,
	})
	return g.routeBattlefieldCardToOwnerGraveyardLocked(cardID)
}

// SacrificePermanent is the locking entry point: a player sacrifices
// a permanent they control. Rejects a permanent controlled by
// someone else (CR 701.17b — only a permanent's controller may
// sacrifice it) so the action layer can expose it directly.
//
// Caller must NOT hold g.mu.
func (g *Game) SacrificePermanent(playerID, cardID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	controller := g.controllerOfBattlefieldCardLocked(cardID)
	if controller == uuid.Nil {
		return ErrCardNotFound
	}
	if controller != playerID {
		return ErrCardCallerMismatch
	}
	if err := g.sacrificePermanentLocked(cardID); err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}

// EachPlayerSacrificesForEffect queues one PendingChoiceSacrifice per
// affected player, in APNAP seat order starting from the active
// player — "each player sacrifices a creature" (Fleshbag Marauder),
// "each other player sacrifices a creature" (Grave Pact), "each
// opponent sacrifices a creature" (Butcher of Malakir).
//
// `spec` narrows what may be chosen; nil means any permanent.
// `except` is the player who does NOT sacrifice ("each OTHER
// player" / "each opponent"), or uuid.Nil when everyone does.
//
// Each player chooses their own, which is why this is a fan-out of
// prompts rather than one effect: the controller of Grave Pact does
// not get to pick which of your creatures dies. Players with no legal
// permanent are skipped at queue time — the requirement is "if you
// can", and prompting them with an empty list would wedge the queue.
//
// Returns the number of prompts queued, so a caller can tell "nobody
// had a creature" from "everyone was asked".
//
// The choices are queued in APNAP order and answered in whatever
// order the players click. Strictly, CR 701.17a makes the sacrifices
// simultaneous after all choices are made; sequential resolution is
// observable only through a payoff that counts them (a Blood Artist
// sees the same number of deaths either way, just spread across more
// trigger batches). Simultaneous choice-then-sacrifice would need the
// whole fan-out held in a resume frame, which is a lot of machinery
// for a difference no card in the catalog can see today.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) EachPlayerSacrificesForEffect(source uuid.UUID, except uuid.UUID, spec *TargetSpec, reason string) int {
	queued := 0
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return 0
	}
	start := g.Turn.ActiveSeat
	for i := 0; i < numSeats; i++ {
		p := g.Seats[(start+i)%numSeats]
		if p == nil || p.Eliminated || p.ID == except {
			continue
		}
		queued += g.PlayerSacrificesForEffect(source, p.ID, spec, reason)
	}
	return queued
}

// PlayerSacrificesForEffect queues ONE sacrifice prompt, for one
// player — "sacrifice another permanent" (Korvold), "sacrifice a
// creature" as the rider on a resolving effect rather than as a
// cost. The single-seat half of EachPlayerSacrificesForEffect, which
// is now written in terms of it.
//
// Same contract as the fan-out version: `spec` narrows what may be
// chosen (nil means any permanent they control), the player chooses
// their own, and a player with no legal permanent is skipped rather
// than prompted with an empty list — a mandatory sacrifice with
// nothing to sacrifice does nothing (CR 701.17b).
//
// Returns 1 when a prompt was queued and 0 when it was skipped, so a
// caller with a "if you do" rider can tell the two apart.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) PlayerSacrificesForEffect(source, playerID uuid.UUID, spec *TargetSpec, reason string) int {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Eliminated {
		return 0
	}
	options := g.sacrificeCandidatesLocked(playerID, spec)
	if len(options) == 0 {
		return 0
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceSacrifice,
		Chooser:          playerID,
		FromPlayer:       playerID,
		Count:            1,
		Source:           source,
		Reason:           reason,
		SacrificeOptions: options,
	})
	return 1
}

// sacrificeCandidatesLocked lists the permanents a player controls
// that match spec. Nil spec means every permanent they control.
//
// Note this is NOT LegalTargetsForEffect: the effect doesn't target,
// so a hexproof or protected creature is still a legal choice. Only
// the spec's own card predicate and control matter.
func (g *Game) sacrificeCandidatesLocked(playerID uuid.UUID, spec *TargetSpec) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.Controller != playerID {
			continue
		}
		if spec != nil && spec.CardOK != nil && !spec.CardOK(g, playerID, c, ZoneBattlefield) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}
