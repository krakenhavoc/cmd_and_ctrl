package game

import "github.com/google/uuid"

// sacrifice.go — S21 sub-PR 1: sacrifice as a first-class engine
// operation (CR 701.21). A permanent is sacrificed by its
// controller, as a cost (Goblin Bombardment, Treasure's mana
// ability) or as an effect's instruction. It is NOT destroyed:
// indestructible (CR 702.12b) and regeneration (CR 701.19a) don't
// apply, and a replacement keyed on destruction never sees it. That
// last one is load-bearing since #667 rather than merely true: this
// file routes through the SAME battlefield exit a destruction does,
// and the only thing that tells them apart is that the destroy route
// declares zoneRoute.Destruction and this one does not. #910 gave the
// sacrifice its own template for the announcement below, and built it
// on battlefieldExitRoute precisely so that stays true by
// construction rather than by anyone remembering it.
//
// The mechanics are deliberately thin — the permanent takes the
// ordinary route to its owner's graveyard, so dies-triggers, the
// CR 903.9 commander-zone replacement and LKI all keep working
// untouched. The only new thing is EventSacrifice, emitted while
// the card is still on the battlefield so "whenever you sacrifice"
// payoffs can read its characteristics before it moves.

// sacrificePermanentLocked sacrifices ONE permanent: the announcement,
// then the battlefield exit.
//
// #910 made it one leg of the shared batch body (routeLegLocked with
// the sacrifice template) rather than its own pair of calls, so the
// single sacrifice and a batch cannot drift — the same move, the same
// announcement, in the same order. It stays the FIRE-AND-FORGET form:
// nil means "no error", never "it left the battlefield", because a
// sacrificed commander's CR 903.9 prompt can still be open when this
// returns. A caller that reads the outcome uses SacrificeThenForEffect.
//
// A permanent that is not on the battlefield is ErrCardNotFound and
// nothing is announced, which is the contract every caller already
// relies on to tell "there was nothing to sacrifice" from "it was
// sacrificed".
//
// Caller must hold g.mu.
func (g *Game) sacrificePermanentLocked(cardID uuid.UUID) error {
	return g.sacrificeWithAnswerLocked(cardID, commanderZoneUnasked, false)
}

// sacrificeAnsweredLocked is sacrificePermanentLocked for a sacrifice
// paid as a COST, whose commander's owner answered CR 903.9 before the
// payment began (#1397, cost_commander_choice.go). The answer rides the
// route onto the battlefield exit's event, so a sacrificed commander no
// longer pauses the payment half way through it — the pause that let
// the same commander be spent twice while its prompt was open. The route
// also settles CR 616 ordering inline: CR 602.2b makes the whole payment
// one indivisible step, so it cannot stop on any replacement prompt.
//
// Caller must hold g.mu.
func (g *Game) sacrificeAnsweredLocked(cardID uuid.UUID, answer commanderZoneAnswer) error {
	return g.sacrificeWithAnswerLocked(cardID, answer, true)
}

// sacrificeWithAnswerLocked is the shared sacrifice body. Cost callers
// pass mustSettleNow; effect and manual callers do not, because their CR
// 614/616 window is allowed to ask the affected player a question.
//
// Caller must hold g.mu.
func (g *Game) sacrificeWithAnswerLocked(cardID uuid.UUID, answer commanderZoneAnswer, mustSettleNow bool) error {
	if g.controllerOfBattlefieldCardLocked(cardID) == uuid.Nil {
		return ErrCardNotFound
	}
	r := sacrificeRoute(uuid.Nil)
	r.commanderAnswer = answer
	r.MustSettleNow = mustSettleNow
	return g.routeLegLocked(r, cardID, nil, nil)
}

// announceSacrificeLocked emits EventSacrifice for the permanent's
// controller, `source` being the card that asked for the sacrifice.
//
// It fires while the card is STILL ON THE BATTLEFIELD and before the
// CR 614 window opens over its move, which is the whole reason it is a
// step of its own: a "whenever you sacrifice a permanent" payoff reads
// characteristics (Ziatora's power, Witch's Oven's toughness) that the
// CR 400.7 forget wipes a moment later, and it must fire whatever the
// window then does with the destination — a sacrifice whose card an
// "exile it instead" replacement takes is still a sacrifice.
//
// Caller must hold g.mu.
func (g *Game) announceSacrificeLocked(cardID, source uuid.UUID) error {
	controller := g.controllerOfBattlefieldCardLocked(cardID)
	if controller == uuid.Nil {
		return ErrCardNotFound
	}
	g.EmitEvent(Event{
		Kind:   EventSacrifice,
		Actor:  controller,
		Source: source,
		CardID: cardID,
	})
	return nil
}

// SacrificePermanent is the locking entry point: a player sacrifices
// a permanent they control. Rejects a permanent controlled by
// someone else (CR 701.21a — only a permanent's controller may
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
// THE FIRE-AND-FORGET FORM, and the count is not an outcome: nothing
// has left the battlefield when it returns, and a clause written on
// the next line pays out for a sacrifice nobody has chosen yet. A
// card with anything hanging off the answer uses
// EachPlayerSacrificesThenForEffect (#1019, sacrifice_run.go), which
// is this queue loop with the rest of the card attached.
//
// The choices are queued in APNAP order and answered in whatever
// order the players click. Strictly, CR 701.21a makes the sacrifices
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
// nothing to sacrifice does nothing (CR 701.21a).
//
// Returns 1 when a prompt was queued and 0 when it was skipped.
//
// THE FIRE-AND-FORGET FORM. The 1 says a QUESTION went up, not that
// anything was sacrificed — a caller with an "if you do" rider is
// asking the wrong thing and wants
// PlayerSacrificesThenForEffect (#1019, sacrifice_run.go).
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) PlayerSacrificesForEffect(source, playerID uuid.UUID, spec *TargetSpec, reason string) int {
	if !g.queueSacrificePromptLocked(source, playerID, spec, reason, uuid.Nil) {
		return 0
	}
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
