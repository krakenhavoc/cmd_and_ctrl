package game

import "github.com/google/uuid"

// untap.go is the engine's ONE untap path, and the CR 502 untap
// step's turn-based action.
//
// # Why it exists (#74)
//
// "Whenever a permanent becomes untapped" (Mesmeric Orb) was
// unwritable, and the reason was structural rather than a missing
// event kind. EventUntapCard has existed since S14 and two of the
// engine's three untap sites emitted it. The third —
// untapAllForLocked, which is the untap step, which is where the
// overwhelming majority of untapping in a game of Magic happens —
// cleared Card.Tapped in a bare loop and announced nothing. So the
// one step named after the action was the one step invisible to
// anything watching for it.
//
// This is #539's shape exactly: a window opened by individual
// callers instead of by the primitive they share, so the next caller
// reintroduces the gap for free. The fix is the same. Every write of
// `Tapped = false` that is an UNTAP now goes through
// untapPermanentLocked, and the callers above are reduced to
// descriptions of a SET of permanents.
//
// Two writes of `Tapped = false` are deliberately NOT untaps and
// keep their own code:
//
//   - MoveCard's battlefield-exit cleanup (zone.go). CR 701.20a is
//     an action performed on a permanent, and a card that has left
//     the battlefield is not a permanent. Announcing it would fire
//     Mesmeric Orb on every creature that dies tapped, which is not
//     what the card says.
//   - The dev spawner's per-instance reset (dev.go), which
//     initialises a template before the card exists anywhere.
//
// # The two shapes, and why neither is a new event kind
//
// Cards ask two different questions about untapping and the brief
// for this work called them out as distinct. They are, and they are
// answered by two mechanisms that both already existed:
//
//	"Whenever a permanent becomes untapped"  (Mesmeric Orb)
//
// is a per-permanent TRIGGER, and it must fire for untaps outside
// the untap step too — Voltaic Key, Fabled Passage, a Seedborn Muse
// untap on someone else's turn. That is EventUntapCard, unchanged,
// now actually emitted from the step. No new kind: a
// step-scoped "the untap step happened" event would be the wrong
// grain for this card, because the card is about the permanent and
// fires once per permanent.
//
//	"Untap all permanents you control during each other player's
//	 untap step"                              (Seedborn Muse)
//
// is NOT a trigger at all. Nothing goes on the stack, nobody gets
// priority, and there is nothing to respond to (CR 502.4). It is a
// modification of the untap step's TURN-BASED ACTION: CR 502.1 says
// the active player determines which permanents they control untap,
// and Seedborn Muse widens that set. Modelling it as a trigger is
// what Quest for Renewal had to settle for before this file, and its
// caveat spelled out the cost — an untap that can be responded to,
// a beat late, in the wrong step.
//
// So the step half is a DERIVED question asked at the moment the
// answer is needed, which is the shape Spec.NoMaxHandSize already
// established for the other player-scoped continuous effect in the
// catalog (see game.EffectiveMaxHandSizeLocked). Nothing is written
// down, nothing is registered on entry and nothing has to be
// unregistered on exit; two Seedborn Muses and one of them dying
// come out right with no bookkeeping, for the same reason two
// Thought Vessels do.
//
// What this deliberately does not add:
//
//   - A third step-entry event kind. EventBeginUpkeep /
//     EventBeginEndStep / EventBeginDrawStep exist because cards
//     name those steps in "at the beginning of" TRIGGERS; the untap
//     step grants no priority and no card triggers on it. Emitting
//     EventBeginUntapStep would create a trigger nothing can legally
//     use and would tempt the next author into writing Seedborn Muse
//     as a trigger anyway. EventStepBegan already announces the step
//     for the game log, and its own doc comment says outright that
//     adding a rules listener to it would be a mistake.
//
//   - A "this permanent doesn't untap" restriction (Winter Orb,
//     Tangle's rider, Icy Manipulator's "doesn't untap during its
//     controller's next untap step"). It is the same set
//     computation from the other direction and untapStepSetLocked is
//     where it goes when the first such card is written; it is not
//     written speculatively here, because nothing in the catalog
//     needs it yet and an unused predicate is a guess about a card
//     nobody has read.

// UntapStepPermission declares one card's contribution to the set of
// permanents that untap during a player's untap step (CR 502.1) —
// Seedborn Muse, Unwinding Clock, Drumbellower, Bender's Waterskin,
// the second half of Quest for Renewal.
//
// The shape mirrors StaticAbility / ReplacementEffect /
// TriggeredAbility: a cheap predicate pair the engine evaluates, and
// no state of its own. Both predicates run under g.mu held in write
// mode and MUST NOT call public locking mutators.
//
// Declared on effects.Spec.UntapStep and bridged into this package
// by the CatalogUntapStepPermissions hook.
type UntapStepPermission struct {
	// AppliesTo reports whether this source contributes anything
	// during the untap step of `activePlayer`. Every card in the
	// family is printed "during each OTHER player's untap step",
	// which is `activePlayer != source.Controller` — the active
	// player's own permanents already untap by CR 502.1 and a
	// permission that fired on your own untap step would be a no-op
	// at best.
	//
	// Quest for Renewal's "as long as there are four or more quest
	// counters" is an intervening condition on the same clause and
	// belongs here too: the permission is continuous, so it is
	// checked at the instant the step asks, with no trigger
	// announce / resolve gap for the counters to change across.
	//
	// Nil means "never contributes" and the permission is skipped —
	// declaring one without both predicates is a programming error.
	AppliesTo func(g *Game, source *Card, activePlayer uuid.UUID) bool

	// Untaps reports whether `target` is one of the permanents this
	// source adds to the set. Seedborn Muse is
	// `target.Controller == source.Controller`; Unwinding Clock adds
	// `target.IsArtifact()` on top; Bender's Waterskin is
	// `target.InstanceID == source.InstanceID`.
	//
	// Consulted only for permanents that are actually TAPPED and not
	// already in the set, so the predicate never has to check either.
	// Type tests should read the effective characteristics
	// (Card.IsArtifact and friends do), because "artifacts you
	// control" means the ones that are artifacts right now.
	Untaps func(g *Game, source *Card, target *Card) bool

	// Label names the clause for logs and debugging. Not player-
	// facing — nothing about this goes on the stack, so there is no
	// stack overlay to title.
	Label string
}

// CatalogUntapStepPermissions returns the untap-step permissions
// declared by the given oracle ID (one entry per
// effects.Spec.UntapStep element), or nil when the card declares
// none — which is every card but a handful. Populated at init time
// by the cards/effects package alongside the other catalog hooks.
//
// Nil hook ⇒ no catalog wired ⇒ the untap step untaps exactly what
// CR 502.1 says and nothing else, which is the pre-existing
// behaviour and what the game package's own tests see.
//
// Consulted once per untap step, not once per event: see
// untapStepSetLocked.
var CatalogUntapStepPermissions func(oracleID string) []UntapStepPermission

// boundUntapPermission is one declared permission paired with the
// battlefield card that declared it. Pointers into
// g.Battlefield.Cards, valid only for the duration of the
// write-locked step that built them.
type boundUntapPermission struct {
	permission UntapStepPermission
	source     *Card
}

// untapPermanentLocked turns one permanent from sideways to upright
// (CR 701.20a) and announces it as EventUntapCard. Returns whether
// anything happened.
//
// A permanent that is ALREADY untapped does not become untapped, and
// gets no event: CR 701.20a describes a change of state, and "untap
// target permanent" pointed at an upright one does nothing at all.
// This matters to exactly the card that motivated the file — a
// Mesmeric Orb that milled for every no-op untap would mill for the
// whole board every untap step regardless of what was tapped.
//
// Actor carries the permanent's controller, captured here rather
// than resolved by the consumer, because "that permanent's
// controller" is the only player the event is ever about and the
// permanent may have moved by the time a downstream effect resolves.
// (The TAP direction's Actor is the activating player, which is a
// different question with a different answer; the two are not
// symmetric and are not merged.)
//
// Caller must hold g.mu in write mode.
func (g *Game) untapPermanentLocked(c *Card) {
	if c == nil || !c.Tapped {
		return
	}
	c.Tapped = false
	g.EmitEvent(Event{
		Kind:   EventUntapCard,
		Actor:  c.Controller,
		CardID: c.InstanceID,
	})
}

// untapPermanentByIDLocked is untapPermanentLocked addressed by
// instance ID. Used by callers that chose their set of permanents up
// front and cannot hold pointers across the emits in between.
// Caller must hold g.mu in write mode.
func (g *Game) untapPermanentByIDLocked(cardID uuid.UUID) {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.untapPermanentLocked(&g.Battlefield.Cards[i])
			return
		}
	}
}

// activeUntapStepPermissionsLocked gathers the untap-step
// permissions that are live right now for `activePlayer`'s untap
// step. One walk of the battlefield per untap step; the per-card
// oracle lookup is the same map hit the trigger harvester does, and
// the AppliesTo evaluation is one predicate per DECLARED permission,
// of which a whole board typically has zero.
//
// Caller must hold g.mu in write mode.
func (g *Game) activeUntapStepPermissionsLocked(activePlayer uuid.UUID) []boundUntapPermission {
	if g.Battlefield == nil || CatalogUntapStepPermissions == nil {
		return nil
	}
	var out []boundUntapPermission
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		// CatalogAbilityKey: "untap all permanents you control
		// during each other player's untap step" is a static
		// ability, and a Seedborn Muse that has lost all its
		// abilities grants nothing.
		oracle := CatalogAbilityKey(*src)
		if oracle == "" {
			continue
		}
		for _, p := range CatalogUntapStepPermissions(oracle) {
			if p.AppliesTo == nil || p.Untaps == nil {
				continue
			}
			if !p.AppliesTo(g, src, activePlayer) {
				continue
			}
			out = append(out, boundUntapPermission{permission: p, source: src})
		}
	}
	return out
}

// untapStepSetLocked answers CR 502.1's question — which permanents
// untap during `activePlayer`'s untap step — as a list of instance
// IDs.
//
// The set is chosen UP FRONT and returned by ID rather than untapped
// in place, for two reasons that happen to agree. CR 502.2 untaps
// them simultaneously, so the set cannot depend on what has already
// untapped within the same step. And each untap emits, which runs
// listeners, which is not a walk you want to be holding pointers
// into g.Battlefield.Cards across — the same reason #529's mill
// chooses its cards before it moves any of them, and the reason the
// catalog's own b16UntapAllYouControlMatching snapshots first.
//
// Already-untapped permanents are filtered here as well as in the
// primitive: it keeps the permission predicates off the vast
// majority of the board on a turn where little is tapped, which is
// most turns.
//
// Caller must hold g.mu in write mode.
func (g *Game) untapStepSetLocked(activePlayer uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	permissions := g.activeUntapStepPermissionsLocked(activePlayer)
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.Tapped {
			continue
		}
		// CR 502.1: the active player's own permanents, always.
		if c.Controller == activePlayer {
			ids = append(ids, c.InstanceID)
			continue
		}
		for _, bp := range permissions {
			if bp.permission.Untaps(g, bp.source, c) {
				ids = append(ids, c.InstanceID)
				break
			}
		}
	}
	return ids
}

// performUntapStepLocked is the untap step's turn-based action
// (CR 502.1-502.3): the active seat's permanents untap, plus
// whatever an UntapStepPermission adds, and the active seat's
// permanents stop being summoning-sick.
//
// The sickness clear and the untap are SEPARATE walks on purpose,
// and this is the one place the difference is visible. CR 302.6 ends
// summoning sickness for permanents their controller has controlled
// continuously since THEIR most recent turn began — it is about
// whose turn it is, not about untapping. A Seedborn Muse controller
// untapping on an opponent's turn unquestionably untaps; their
// creatures are just as unquestionably still sick. The old code
// could write one loop because the two sets were the same set; they
// are not any more.
//
// The clear stays unconditional (rather than creature-gated) because
// SummonedThisTurn is the raw "entered this turn" marker and CR
// 302.6's creature test is applied at read time in
// HasSummoningSickness — see #537, which moved the test there and
// must not be undone by re-adding one here.
//
// Caller must hold g.mu in write mode.
func (g *Game) performUntapStepLocked(seat int) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	activePlayer := g.Seats[seat].ID
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Controller == activePlayer {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	}
	for _, id := range g.untapStepSetLocked(activePlayer) {
		g.untapPermanentByIDLocked(id)
	}
}

// untapAllForLocked untaps every battlefield card controlled by the
// given seat and clears their "entered this turn" markers. This is
// the SANDBOX untap-everything action behind the public UntapAll
// mutation — a player pressing a button, not a step happening — so
// it deliberately does not consult UntapStepPermission: nothing
// about one player's manual untap is "during an untap step".
//
// A negative or out-of-range seat is a no-op. Caller must hold g.mu.
func (g *Game) untapAllForLocked(seat int) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	playerID := g.Seats[seat].ID
	if g.Battlefield == nil {
		return
	}
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller != playerID {
			continue
		}
		g.Battlefield.Cards[i].SummonedThisTurn = false
		if g.Battlefield.Cards[i].Tapped {
			ids = append(ids, g.Battlefield.Cards[i].InstanceID)
		}
	}
	for _, id := range ids {
		g.untapPermanentByIDLocked(id)
	}
}
