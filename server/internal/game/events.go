package game

import (
	"log/slog"

	"github.com/google/uuid"
)

// events.go holds the per-game event log that the S14+ rules engine
// reads from. Events are emitted by every rules-visible mutation
// (damage dealt, card drawn, zone move, counter placement, spell
// cast / resolved, player eliminated) and appended to Game.Events
// in-order under the write lock. Pull-based — listeners registered
// via RegisterListener see each event synchronously after the
// underlying mutation, and S19+ card-effect triggers will harvest
// matching events from the log rather than being push-fired by
// each mutation site.
//
// The log is append-only per-Game, unbounded for the game's lifetime.
// A typical 30-minute Commander game produces a few thousand events;
// cloned across the 32-frame undo stack that's ~100-150k events in
// memory worst case, well under the per-room budget. Post-game
// cleanup runs on Game.End(), so unbounded growth is a non-issue.
//
// Introduced in S14 sub-PR 1 as infrastructure with zero production
// consumers; sub-PR 3 wires the effect-resolution path to emit, and
// S19 registers the first real listener (trigger harvester).

// EventKind discriminates what the event represents. Keep the set
// tight — it's an enum on the wire (when GameView.events ships in
// sub-PR 3) so every addition is a schema change. All rules-visible
// mutation sites map onto one of these.
type EventKind string

const (
	// EventCast — a spell was cast onto the stack. CardID is the
	// spell; Actor is the caster.
	EventCast EventKind = "cast"

	// EventResolve — a stack item successfully resolved. CardID is
	// the card whose spell or ability resolved.
	EventResolve EventKind = "resolve"

	// EventFizzle — a spell or ability resolved but did nothing
	// (all targets illegal, or the effect short-circuited). CR
	// 608.2b's "countered by game rules."
	EventFizzle EventKind = "fizzle"

	// EventDealDamage — Amount damage was dealt from Source to Target.
	// Target may be a player or a card (distinguished by whether
	// Target resolves in PlayerByID vs the zone scan).
	EventDealDamage EventKind = "deal_damage"

	// EventChangeLife — a player's life total changed by Amount
	// (signed). Actor is who caused the change (uuid.Nil for admin
	// / SBA paths that have no single actor).
	EventChangeLife EventKind = "change_life"

	// EventDrawCard — Actor drew CardID. Fires per card, so "draw
	// 3" produces three events.
	EventDrawCard EventKind = "draw_card"

	// EventDiscardCard — Actor discarded CardID.
	EventDiscardCard EventKind = "discard_card"

	// EventMill — Actor milled CardID from the top of their library.
	// Fires per card.
	EventMill EventKind = "mill"

	// EventZoneMove — CardID moved from OldZone to NewZone. The
	// catch-all event for card motion that doesn't fit a more
	// specific kind. Draw / mill / discard / resolve all also
	// produce ZoneMove events — consumers can pick the granularity
	// they care about.
	EventZoneMove EventKind = "zone_move"

	// EventTapCard — CardID was tapped.
	EventTapCard EventKind = "tap_card"

	// EventUntapCard — CardID was untapped.
	EventUntapCard EventKind = "untap_card"

	// EventCounterPlaced — a counter of Label (see CounterKind /
	// KnownCardCounters) was placed on CardID. Amount is the new
	// count of that counter kind on the card. Fires on AddCounter
	// and on the SBA +1/+1 / -1/-1 cancel.
	EventCounterPlaced EventKind = "counter_placed"

	// EventTokenCreated — a token was created under Actor's control.
	// CardID is the new instance.
	EventTokenCreated EventKind = "token_created"

	// EventSearchLibrary — Actor searched their library. Reserved
	// for S14 catalog effects that fire SearchLibrary; the log
	// entry is the "you searched your library" trigger source
	// that S19 listens for (Panoptic Mirror, etc.).
	EventSearchLibrary EventKind = "search_library"

	// EventCounterSpell — a stack item was countered (spell or
	// ability). Source is the counter itself; Target is the
	// countered item's ID. Routed to owner's graveyard by default,
	// or another zone per the counter's destination.
	EventCounterSpell EventKind = "counter_spell"

	// EventConcede — Actor conceded the game. Precedes the
	// eliminate-via-SBA path.
	EventConcede EventKind = "concede"

	// EventPlayerEliminated — Actor was eliminated from the game
	// (0 life, empty library draw, 21+ commander damage, poison >= 10,
	// concede, or manual eliminate). The single terminating event
	// for a player's participation in the game.
	EventPlayerEliminated EventKind = "player_eliminated"

	// EventETB — a permanent entered the battlefield. CardID is the
	// new permanent. Distinct from ZoneMove so listeners can key
	// off of ETB specifically without pattern-matching the zone
	// fields.
	EventETB EventKind = "etb"

	// EventLTB — a permanent left the battlefield (any reason).
	EventLTB EventKind = "ltb"

	// EventTrigger — a triggered ability was announced onto
	// PendingTriggers. Legacy (manual) announce goes through
	// AnnounceTrigger in S13.1; S19's auto-announce will flow
	// through here too. The event lets listeners observe trigger
	// creation without having to poll PendingTriggers.
	EventTrigger EventKind = "trigger"

	// EventEffectError — an effect primitive (S14 catalog) failed
	// to apply. Caller logs + keeps moving; the event is the
	// debugging breadcrumb. ErrorMsg carries the reason.
	EventEffectError EventKind = "effect_error"

	// EventStepTransition is an engine-internal sentinel used only
	// by the S17 replacement-effect pipeline. Fired from the top of
	// runStepEntryHooksLocked so skip-step replacements (Stasis
	// cancels StepUntap) can intercept. Does NOT emit to the public
	// event log — cards read it only via ReplacementEffect.Watches.
	// Added in S17 sub-PR 2.
	EventStepTransition EventKind = "step_transition"

	// EventManaAbilityActivated — a mana-producing ability fired.
	// Actor = controller, Source = the permanent that produced the
	// mana. S15 sub-PR 2.
	EventManaAbilityActivated EventKind = "mana_ability_activated"

	// EventManaAdded — one mana token landed in a player's pool.
	// Actor = pool owner, Source = the producing permanent (uuid.Nil
	// for non-card sources). The token's color rides Amount as 0
	// (W/U/B/R/G/C carries no numeric weight) — the wire-side mana
	// pool is the canonical projection. S15 sub-PR 2.
	EventManaAdded EventKind = "mana_added"

	// EventManaPoolEmptied — a player's mana pool was cleared at a
	// step / phase boundary (CR 106.4) or by an admin reset. Actor =
	// pool owner. Amount = number of tokens that were dropped.
	// S15 sub-PR 2.
	EventManaPoolEmptied EventKind = "mana_pool_emptied"

	// EventManaSpent — the strict-mode S15 cost gate deducted mana
	// from a caster's pool to pay for a spell. Actor = caster,
	// Source = card being cast. Permissive / forced casts emit
	// EventCostWarning instead. S15 sub-PR 3.
	EventManaSpent EventKind = "mana_spent"

	// EventCostWarning — a cast_spell action proceeded under the S15
	// permissive-default cost path OR under a strict-mode override
	// (ForceCast=true) despite the pool not covering the cost.
	// Breadcrumb for the client's auto-toast UI and a debug signal
	// for operators that want to audit "did the caster actually pay
	// for this?" Actor = caster, Source = card being cast.
	// S15 sub-PR 3.
	EventCostWarning EventKind = "cost_warning"
)

// Event is a single entry in the per-game event log. Tagged union
// shape: Kind selects which of the payload fields are meaningful,
// the rest are zero. Chosen over per-kind interfaces so Clone is
// a simple slice value-copy and wire encoding is flat JSON.
//
// Seq is a monotonic per-Game counter stamped at emit time. First
// event has Seq == 1; zero is the un-stamped sentinel.
type Event struct {
	Seq  uint64    `json:"seq"`
	Kind EventKind `json:"kind"`

	// Actor is the player responsible for the event (caster,
	// controller, drawing player, etc.). uuid.Nil for admin / SBA
	// paths that have no single actor.
	Actor uuid.UUID `json:"actor,omitempty"`

	// Source is the card whose effect produced the event
	// (Lightning Bolt that dealt the damage, Counterspell that
	// countered, the ETB creature itself). uuid.Nil when the event
	// isn't tied to a specific source card (e.g. manual ChangeLife
	// via the player header).
	Source uuid.UUID `json:"source,omitempty"`

	// Target is the thing the event acts on: a player for life /
	// damage events, a card for zone moves / counters / tap / ETB.
	// Consumers key off of Kind to decide how to interpret.
	Target uuid.UUID `json:"target,omitempty"`

	// CardID names the card the event references. For ZoneMove /
	// DrawCard / Mill / Discard it's the moved card. For Cast /
	// Resolve it's the spell. For TokenCreated it's the new token.
	CardID uuid.UUID `json:"card_id,omitempty"`

	// Amount is the signed / count payload: damage dealt, life
	// delta, number of cards, counter count after the change.
	Amount int `json:"amount,omitempty"`

	// Label is a free-text qualifier: counter kind
	// ("+1/+1", "loyalty", "poison"), trigger label, error
	// classification. Kept as a string so adding new counter kinds
	// doesn't require a schema change.
	Label string `json:"label,omitempty"`

	// OldZone / NewZone are the zone kinds for ZoneMove-shaped
	// events. Empty string means "not applicable."
	OldZone ZoneKind `json:"old_zone,omitempty"`
	NewZone ZoneKind `json:"new_zone,omitempty"`

	// ErrorMsg carries the failure reason on EventEffectError.
	ErrorMsg string `json:"error_msg,omitempty"`
}

// EmitEvent appends ev to the game's event log under the existing
// write lock. Caller MUST hold g.mu — the emit is a mutation and
// participates in the same atomic write as the action that produced
// it. Stamps Seq monotonically; dispatches to registered listeners
// synchronously before returning, so a listener that wants to
// trigger a follow-on event sees state consistent with the event
// it's reacting to.
func (g *Game) EmitEvent(ev Event) {
	g.eventSeq++
	ev.Seq = g.eventSeq
	g.Events = append(g.Events, ev)
	// S17 sub-PR 6 diagnostic: effect-error events are otherwise
	// silent (no client toast yet). Surfacing them in the server log
	// so manual-test regressions have a visible breadcrumb. Keep
	// until the client learns to render effect_error as a toast.
	if ev.Kind == EventEffectError {
		slog.Warn("effect error",
			"seq", ev.Seq,
			"game", g.ID.String(),
			"source", ev.Source.String(),
			"error", ev.ErrorMsg,
		)
	}
	g.notifyListenersLocked(ev)
}
