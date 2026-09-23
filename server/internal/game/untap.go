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
//   - MoveCard's battlefield-exit cleanup (zone.go). CR 701.26b is
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
// modification of the untap step's TURN-BASED ACTION: CR 502.3 says
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
// Static restrictions and next-untap markers narrow that same set.
// Stun counters instead replace every individual untap in the shared
// primitive, including untaps outside an untap step.

// UntapStepPermission declares one card's contribution to the set of
// permanents that untap during a player's untap step (CR 502.3) —
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
	// player's own permanents already untap by CR 502.3 and a
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

// UntapStepRestriction is a static effect that keeps a permanent from
// untapping during that permanent's controller's own untap step.
type UntapStepRestriction struct {
	Restricts func(target *Card, g *Game, source *Card) bool
	Label     string
}

// UntapSkip is one one-shot next-untap-step marker. Nil Player follows the
// permanent's controller; a non-nil Player names that player's next step.
type UntapSkip struct{ Player uuid.UUID }

type untapSkipSnapshot struct {
	Player uuid.UUID `json:"player"`
}

func snapshotUntapSkips(in []UntapSkip) []untapSkipSnapshot {
	if len(in) == 0 {
		return nil
	}
	out := make([]untapSkipSnapshot, len(in))
	for i := range in {
		out[i].Player = in[i].Player
	}
	return out
}

func restoreUntapSkips(in []untapSkipSnapshot) []UntapSkip {
	if len(in) == 0 {
		return nil
	}
	out := make([]UntapSkip, len(in))
	for i := range in {
		out[i].Player = in[i].Player
	}
	return out
}

// CatalogUntapStepPermissions returns the untap-step permissions
// declared by the given oracle ID (one entry per
// effects.Spec.UntapStep element), or nil when the card declares
// none — which is every card but a handful. Populated at init time
// by the cards/effects package alongside the other catalog hooks.
//
// Nil hook ⇒ no catalog wired ⇒ the untap step untaps exactly what
// CR 502.3 says and nothing else, which is the pre-existing
// behaviour and what the game package's own tests see.
//
// Consulted once per untap step, not once per event: see
// untapStepSetLocked.
var CatalogUntapStepPermissions func(oracleID string) []UntapStepPermission

var CatalogUntapStepRestrictions func(oracleID string) []UntapStepRestriction

// boundUntapPermission is one declared permission paired with the
// battlefield card that declared it. Pointers into
// g.Battlefield.Cards, valid only for the duration of the
// write-locked step that built them.
type boundUntapPermission struct {
	permission UntapStepPermission
	source     *Card
}

type boundUntapRestriction struct {
	restriction UntapStepRestriction
	source      *Card
}

// untapPermanentLocked turns one permanent from sideways to upright
// (CR 701.26b) and announces it as EventUntapCard, unless a stun
// counter replaces the untap with removing that counter.
//
// A permanent that is ALREADY untapped does not become untapped, and
// gets no event: CR 701.26b describes a change of state, and "untap
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
	if c.Counters[CounterStun] > 0 {
		_ = g.applyCounterLocked(c.InstanceID, CounterStun, -1)
		return
	}
	c.Tapped = false
	g.EmitEvent(Event{
		Kind:   EventUntapCard,
		Actor:  c.Controller,
		CardID: c.InstanceID,
	})
}

// SkipNextUntapForEffect adds a deduplicated next-untap marker to a
// battlefield permanent. Caller already holds g.mu.
func (g *Game) SkipNextUntapForEffect(cardID, player uuid.UUID) error {
	if g.Battlefield == nil {
		return nil
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != cardID {
			continue
		}
		for _, skip := range c.NextUntapSkips {
			if skip.Player == player {
				return nil
			}
		}
		c.NextUntapSkips = append(c.NextUntapSkips, UntapSkip{Player: player})
		return nil
	}
	return nil
}

func (c Card) hasNextUntapSkipFor(player uuid.UUID) bool {
	for _, skip := range c.NextUntapSkips {
		if skip.Player == player || (skip.Player == uuid.Nil && c.Controller == player) {
			return true
		}
	}
	return false
}

// activeUntapStepRestrictionsLocked binds every live restriction source once
// for a step-sized query. Caller holds g.mu.
func (g *Game) activeUntapStepRestrictionsLocked() []boundUntapRestriction {
	if g.Battlefield == nil || CatalogUntapStepRestrictions == nil {
		return nil
	}
	var out []boundUntapRestriction
	for i := range g.Battlefield.Cards {
		source := &g.Battlefield.Cards[i]
		key := CatalogAbilityKey(*source)
		if key == "" {
			continue
		}
		for _, restriction := range CatalogUntapStepRestrictions(key) {
			if restriction.Restricts != nil {
				out = append(out, boundUntapRestriction{restriction, source})
			}
		}
	}
	return out
}

func untapStepRestrictedBy(c *Card, g *Game, restrictions []boundUntapRestriction) bool {
	if c == nil {
		return false
	}
	for _, bound := range restrictions {
		if bound.restriction.Restricts(c, g, bound.source) {
			return true
		}
	}
	return false
}

// UntapStepRestrictedLocked is the pure static half of the untap-step set.
// Caller holds g.mu; it deliberately does not recompute layers.
func (g *Game) UntapStepRestrictedLocked(c *Card) bool {
	return untapStepRestrictedBy(c, g, g.activeUntapStepRestrictionsLocked())
}

func (g *Game) consumeUntapSkipsLocked(activePlayer uuid.UUID) {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		out := c.NextUntapSkips[:0]
		for _, skip := range c.NextUntapSkips {
			if skip.Player == activePlayer || (skip.Player == uuid.Nil && c.Controller == activePlayer) {
				continue
			}
			out = append(out, skip)
		}
		if len(out) == 0 {
			c.NextUntapSkips = nil
		} else {
			c.NextUntapSkips = out
		}
	}
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

// untapStepSetLocked answers CR 502.3's question — which permanents
// untap during `activePlayer`'s untap step — as a list of instance
// IDs.
//
// The set is chosen UP FRONT and returned by ID rather than untapped
// in place, for two reasons that happen to agree. CR 502.3 untaps
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
	restrictions := g.activeUntapStepRestrictionsLocked()
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.Tapped {
			continue
		}
		// CR 502.3: the active player's own permanents, always.
		if c.Controller == activePlayer {
			if untapStepRestrictedBy(c, g, restrictions) || c.hasNextUntapSkipFor(activePlayer) {
				continue
			}
			ids = append(ids, c.InstanceID)
			continue
		}
		if c.hasNextUntapSkipFor(activePlayer) {
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
// The set is read off FRESH layers. The step-entry hook recomputes
// before it gets here, but this is the read that went wrong when it
// did not — cleanup's "until end of turn" sweep bumps the layer
// version and the cursor recurses straight into this step, so a
// Seedborn Muse whose "loses all abilities until end of turn" had
// just ended still read as silenced and untapped nothing, and the
// turn wrap's own bump (a "during your turn" static changes its
// answer there) went unread the same way. The recompute stays here
// too, next to the reads, so a caller that is not the hook gets it.
//
// THE STEP CAN PAUSE (ADR 0070, #826). CR 502.3's first sentence is a
// DETERMINATION — "the active player determines which permanents they
// control will untap" — and two printed families make it a real
// decision: a cap ("players can't untap more than one land during
// their untap steps") and an opt-out ("you may choose not to untap
// this during your untap step"). When one of them is live and actually
// binds, this function queues ONE prompt to the active player and
// returns true, having untapped NOTHING: CR 502.3 untaps them all
// simultaneously, so the determination finishes first and the whole
// set untaps in one loop from the prompt's continuation
// (finishUntapStepLocked). See untap_choice.go.
//
// The summoning-sickness clear and the marker sweep stay eager, and
// deliberately so. The first is a different turn-based action that does
// not depend on the answer (CR 302.6, above). The second is used up by
// the step that happened rather than by the permanents that untapped
// (ADR 0058 Decision 2), and re-running it after a pause would consume
// twice.
//
// Returns whether the step PAUSED. A paused step has not moved the
// cursor either; the caller must not advance it. See
// exitUntapStepLocked for the one place the step ends.
//
// Caller must hold g.mu in write mode.
func (g *Game) performUntapStepLocked(seat int) (paused bool) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	activePlayer := g.Seats[seat].ID
	// CR 502.1, the FIRST of this step's three turn-based actions and
	// the one this function is not named after: phasing (#1199, ADR
	// 0084, phasing.go). Ahead of the CR 302.6 sickness clear and well
	// ahead of untapStepSetLocked, because a permanent phasing in has
	// to be in the untap set and one phasing out has to be out of it.
	//
	// HERE RATHER THAN IN THE STEP-ENTRY ARM, for the reason this
	// file's contract gives about consumeUntapSkipsLocked: the step
	// can pause on CR 502.3's determination and its continuation
	// (finishUntapStepLocked) only untaps and exits, so everything
	// ahead of the pause runs exactly once. CR 502.1 is a turn-based
	// action of this step, not of the step entry, and moving it would
	// have to be re-reasoned the day a second pause is added.
	//
	// The layers are recomputed again afterwards: phasing changes what
	// is on the battlefield, so the untap-step permissions and
	// restrictions below (Seedborn Muse, Winter Orb) must be read off
	// a board that no longer has the phased-out permanents in it.
	g.performPhasingLocked(activePlayer)
	g.RecomputeLayersIfStaleLocked()
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Controller == activePlayer {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	}
	ids := g.untapStepSetLocked(activePlayer)
	// Markers are consumed for the actual step even when their card was
	// already upright and therefore absent from this set. Kept before any
	// emits so listeners cannot observe a half-used step.
	g.consumeUntapSkipsLocked(activePlayer)
	if plan := g.untapChoicePlanLocked(activePlayer, ids); plan != nil {
		g.queueUntapChoiceLocked(activePlayer, plan)
		return true
	}
	for _, id := range ids {
		g.untapPermanentByIDLocked(id)
	}
	return false
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
