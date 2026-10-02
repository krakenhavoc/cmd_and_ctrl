package game

import (
	"fmt"

	"github.com/google/uuid"
)

// scoped_replacements.go is ADR 0041 phase 3's tier 3b for replacement
// effects (Decision P8, #1497): a replacement effect a resolving spell
// or ability creates (CR 611.2) is a ScopedEffect record, not a
// closure.
//
// WHAT IT REPLACED. `Game.TurnScopedReplacements` was a slice of
// `ReplacementEffect` values — three closures each — emptied wholesale
// at cleanup. A table holding one was not a restore point
// (`ContinuationCensus.TurnScopedReplacements`), an undo could not
// rewind a prevention shield's spent charge (the charge was a variable
// the closure captured), and the Whip of Erebos / unearth redirect
// ended at cleanup whatever had happened to the creature (#1591).
//
// WHAT IT IS NOW. Mod kinds whose reader is the replacement gather
// rather than the layer pass (modKindSpec.reader):
//
//   - preventCombatDamage — Fog, Holy Day, Tangle, Constant Mists, and
//     with Player set, Druid's Deliverance;
//   - preventDamage — Mending Hands' charged shield (CR 615.7);
//   - preventNextFromSource — the Circles of Protection's one-use
//     shield against the next damage from a source (CR 615.8, ADR 0107
//     §6, prevent_next_from_source.go);
//   - gainNoLife — Flames of the Blood Hand (ADR 0107 §5);
//   - exileInsteadOfLeaving — the Whip's and unearth's redirect,
//     INDEFINITE and pinned to the returned object, so it lasts exactly
//     as long as that object is on the battlefield (#1591);
//   - exileInsteadOfYourGraveyard — Yawgmoth's Will (ADR 0108 §4): every
//     card moved or discarded into one player's graveyard this turn;
//   - exileInsteadOfGraveyard — Cosmic Intervention, whose per-card
//     return is a registered delayed-trigger body named by Mod.Then.
//
// The gather adapts each live record into the ReplacementEffect it
// already consumes, with closures the RUNNING binary builds from the
// record — the same "rebuilt, never persisted" status the layer
// adapter has. Every closure reads the record back through the game it
// is handed, by Seq, at call time: an effect held by an open CR 616
// prompt resolves against whatever the registry says NOW, which is
// what makes a shield spent by an earlier event, a record an undo
// removed, and a redirect whose object has gone all answer "does not
// apply" rather than acting on a stale copy.

// scopedReplacementModStride is how many mods of one record the ID
// encoding can tell apart: ID = base + Seq×stride + mod index. Every
// record the engine writes has one replacement mod; the stride leaves
// room without letting two records' IDs overlap.
const scopedReplacementModStride = 8

// scopedReplacementID mints the ReplacementEffectID for mod `mod` of the
// record with Seq `seq`. ok is false for anything the scheme cannot
// represent — a record with no Seq, a mod past the stride, or a Seq
// large enough to run into the test range.
func scopedReplacementID(seq int64, mod int) (ReplacementEffectID, bool) {
	if seq <= 0 || mod < 0 || mod >= scopedReplacementModStride {
		return 0, false
	}
	span := uint64(testReplacementIDBase - scopedReplacementIDBase)
	if uint64(seq) >= span/scopedReplacementModStride {
		return 0, false
	}
	return scopedReplacementIDBase + ReplacementEffectID(uint64(seq)*scopedReplacementModStride+uint64(mod)), true
}

// decodeScopedReplacementID is scopedReplacementID's exact inverse.
func decodeScopedReplacementID(id ReplacementEffectID) (seq int64, mod int, ok bool) {
	if id < scopedReplacementIDBase || id >= testReplacementIDBase {
		return 0, 0, false
	}
	n := uint64(id - scopedReplacementIDBase)
	seq, mod = int64(n/scopedReplacementModStride), int(n%scopedReplacementModStride)
	if seq <= 0 {
		return 0, 0, false
	}
	return seq, mod, true
}

// scopedEffectIndexBySeqLocked finds the live record with Seq `seq`.
// Caller must hold g.mu.
func (g *Game) scopedEffectIndexBySeqLocked(seq int64) (int, bool) {
	if seq <= 0 {
		return -1, false
	}
	for i := range g.ScopedEffects {
		if g.ScopedEffects[i].Seq == seq {
			return i, true
		}
	}
	return -1, false
}

// scopedReplacementModLocked is the live record and mod an adapted
// closure names, or false when the record is gone (swept, spent, undone)
// or no longer holds a replacement mod of that kind at that index.
// Caller must hold g.mu.
func (g *Game) scopedReplacementModLocked(seq int64, mod int, kind ModKind) (ScopedEffect, Mod, bool) {
	i, ok := g.scopedEffectIndexBySeqLocked(seq)
	if !ok {
		return ScopedEffect{}, Mod{}, false
	}
	e := g.ScopedEffects[i]
	if mod < 0 || mod >= len(e.Mods) || e.Mods[mod].Kind != kind {
		return ScopedEffect{}, Mod{}, false
	}
	return e, e.Mods[mod], true
}

// replacementModProblem is registration's check on a replacement mod's
// parameters — "" when there is none. Registration panics on a problem,
// as it does on an unknown kind: every caller is an engine function
// below, and a bad parameter is a programming error in it.
func replacementModProblem(m Mod) string {
	switch m.Kind {
	case ModPreventDamage:
		if m.Amount < 1 {
			return fmt.Sprintf("a preventDamage shield needs a charge of at least 1, got %d", m.Amount)
		}
	case ModExileInsteadOfGraveyard:
		if m.Then == "" || !KnownEffectBody(m.Then) {
			return fmt.Sprintf("exileInsteadOfGraveyard names delayed-trigger body %q, which is not registered", m.Then)
		}
	case ModGainNoLife:
		if m.Player == uuid.Nil {
			return "a gainNoLife replacement names no player"
		}
	case ModExileInsteadOfYourGraveyard:
		if m.Player == uuid.Nil {
			return "an exileInsteadOfYourGraveyard replacement names no player"
		}
	}
	return ""
}

var (
	watchDamage   = []EventKind{EventDealDamage}
	watchZoneMove = []EventKind{EventZoneMove}
	watchLife     = []EventKind{EventChangeLife}
	// watchMoveOrDiscard is "from anywhere": a discard is its own
	// replacement window (RepEventDiscard, #650), not a RepEventMove.
	watchMoveOrDiscard = []EventKind{EventZoneMove, EventDiscardCard}
)

// scopedReplacementWatches is the watch key of a replacement kind — the
// pre-filter the gather applies before it builds anything.
func scopedReplacementWatches(kind ModKind) []EventKind {
	switch kind {
	case ModPreventCombatDamage, ModPreventDamage, ModPreventNextFromSource:
		return watchDamage
	case ModExileInsteadOfLeaving, ModExileInsteadOfGraveyard:
		return watchZoneMove
	case ModGainNoLife:
		return watchLife
	case ModExileInsteadOfYourGraveyard:
		return watchMoveOrDiscard
	}
	return nil
}

// ---------------------------------------------------------------
// Registration — the only writers of these kinds
// ---------------------------------------------------------------

// PreventCombatDamageThisTurnForEffect is "prevent all combat damage
// that would be dealt this turn" (CR 615.1, Fog) — or, with `player`
// set, "… that would be dealt to <player> this turn" (Druid's
// Deliverance). Non-combat damage is untouched. Reports whether a
// record was written.
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) PreventCombatDamageThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModPreventCombatDamage, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// PreventNextDamageThisTurnForEffect is the charged shield (CR 615.7):
// "prevent the next `amount` damage that would be dealt to <target>
// this turn". `target` is a player or a permanent. A shield on a
// permanent is pinned to that object (CR 400.7), so it neither follows
// a flicker nor outlives the permanent. Registers nothing for a charge
// below 1 or a target that is neither a player nor a permanent.
//
// Caller must hold g.mu (write).
func (g *Game) PreventNextDamageThisTurnForEffect(sourceID, target uuid.UUID, amount int, combatOnly bool, label string) bool {
	if target == uuid.Nil || amount < 1 {
		return false
	}
	m := Mod{Kind: ModPreventDamage, Amount: amount, CombatOnly: combatOnly}
	d := g.UntilEndOfTurnDuration()
	if g.playerByIDLocked(target) != nil {
		m.Player = target
		return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil, []Mod{m}, d, label)
	}
	affected := g.PinnedObjectsLocked(target)
	if len(affected) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, uuid.Nil, []Mod{m},
		g.PinnedTo(d, target), label, timeNowUnixNano())
}

// ExileInsteadOfLeavingBattlefieldForEffect is "if it would leave the
// battlefield, exile it instead of putting it anywhere else" on the one
// permanent `cardID` (Whip of Erebos, unearth — CR 702.82a).
// `controller` controls the replacement effect.
//
// The duration is INDEFINITE and pinned to the object (#1591): the
// printed clause has no duration, so it lasts exactly as long as that
// object is on the battlefield — past cleanup too, when the end-step
// exile was countered. Registers nothing for a card that is not on the
// battlefield.
//
// Caller must hold g.mu (write).
func (g *Game) ExileInsteadOfLeavingBattlefieldForEffect(sourceID, cardID, controller uuid.UUID, label string) bool {
	affected := g.PinnedObjectsLocked(cardID)
	if len(affected) == 0 {
		return false
	}
	return g.appendScopedEffectLocked(sourceID, affected, ScopeNone, controller,
		[]Mod{{Kind: ModExileInsteadOfLeaving}}, g.PinnedTo(IndefiniteDuration(), cardID),
		label, timeNowUnixNano())
}

// ExileInsteadOfGraveyardThisTurnForEffect is "if a permanent you
// control would be put into a graveyard from the battlefield this turn,
// exile it instead" (Cosmic Intervention), with `then` — a registered
// delayed-trigger body — scheduled at the next end step for each card
// it redirects, carrying that card. "You" is `controller`, and the set
// is read live, so a permanent that enters after the spell resolved is
// saved too (a replacement effect is not a characteristic, so CR 611.2c
// does not lock it).
//
// Caller must hold g.mu (write).
func (g *Game) ExileInsteadOfGraveyardThisTurnForEffect(sourceID, controller uuid.UUID, then BodyRef, label string) bool {
	if then.key == "" {
		effectKeyFault(fmt.Sprintf("game: %q exiles instead of the graveyard with no delayed-trigger body — dropped", label))
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeYourPermanents, controller,
		[]Mod{{Kind: ModExileInsteadOfGraveyard, Then: then.key}}, g.UntilEndOfTurnDuration(), label)
}

// ExileInsteadOfYourGraveyardThisTurnForEffect is "if a card would be
// put into your graveyard from anywhere this turn, exile that card
// instead" (CR 614.1a, Yawgmoth's Will). `player` is "you" and the
// graveyard owner the clause names. The record is a game-wide one with
// no object pinned: it covers a card from any zone, the resolving
// spell itself included, and a token never (a token is not a card).
//
// Caller must hold g.mu (write) — every caller is a resolving effect.
func (g *Game) ExileInsteadOfYourGraveyardThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModExileInsteadOfYourGraveyard, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// ---------------------------------------------------------------
// The gather adapter
// ---------------------------------------------------------------

// gatherScopedReplacementsLocked appends every scoped replacement that
// applies to ev and has not applied to it yet. Called from
// gatherActiveReplacementsLocked, between the catalog and the test
// registries. Caller must hold g.mu.
func (g *Game) gatherScopedReplacementsLocked(ev *ReplacementEvent, applied map[ReplacementEffectID]bool, out []activeReplacement) []activeReplacement {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Seq == 0 {
			continue
		}
		for j, m := range e.Mods {
			if modKinds[m.Kind].reader != readerReplacement {
				continue
			}
			if !eventKindMatches(scopedReplacementWatches(m.Kind), ev.Kind) {
				continue
			}
			id, ok := scopedReplacementID(e.Seq, j)
			if !ok || applied[id] {
				continue
			}
			if !scopedReplacementAppliesLocked(g, *e, m, ev) {
				continue
			}
			out = append(out, activeReplacement{effect: scopedReplacementEffect(e.Seq, j, m.Kind, e.Label), id: id})
		}
	}
	return out
}

// scopedReplacementLabelLocked is the prompt label of a scoped
// replacement's ID, "" when it no longer resolves. Caller must hold g.mu.
func (g *Game) scopedReplacementLabelLocked(id ReplacementEffectID) string {
	seq, _, ok := decodeScopedReplacementID(id)
	if !ok {
		return ""
	}
	i, ok := g.scopedEffectIndexBySeqLocked(seq)
	if !ok {
		return ""
	}
	return g.ScopedEffects[i].Label
}

// scopedReplacementsWatchLocked reports whether any live scoped
// replacement watches `kind` — producesManaReplacementsExistLocked's
// cheap gate. Caller must hold g.mu.
func (g *Game) scopedReplacementsWatchLocked(kind ReplacementEventKind) bool {
	for i := range g.ScopedEffects {
		for _, m := range g.ScopedEffects[i].Mods {
			if modKinds[m.Kind].reader == readerReplacement && eventKindMatches(scopedReplacementWatches(m.Kind), kind) {
				return true
			}
		}
	}
	return false
}

// scopedReplacementEffect is the ReplacementEffect the gather hands the
// pipeline for mod `mod` of record `seq`. Every closure re-reads the
// record through the game it is handed (see the file comment); nothing
// is captured but the record's name and the kind.
func scopedReplacementEffect(seq int64, mod int, kind ModKind, label string) ReplacementEffect {
	eff := ReplacementEffect{
		Watches: scopedReplacementWatches(kind),
		// CR 615.1a: the three damage shields say "prevent", so CR 615.12
		// reads them (unpreventable_damage.go). The two exile
		// redirects and "gains no life instead" are replacements.
		Prevention: scopedKindPrevents(kind),
		AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
			e, m, ok := g.scopedReplacementModLocked(seq, mod, kind)
			return ok && scopedReplacementAppliesLocked(g, e, m, ev)
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			e, m, ok := g.scopedReplacementModLocked(seq, mod, kind)
			if !ok {
				return nil
			}
			return g.applyScopedReplacementLocked(e, mod, m, ev)
		},
		Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
			switch kind {
			case ModPreventCombatDamage, ModPreventDamage, ModPreventNextFromSource:
				// CR 616.1 gives the ordering choice to the AFFECTED
				// player — whoever is being dealt the damage — so a
				// prevention shield reports no controller (S17's Fog).
				return uuid.Nil
			}
			e, _, ok := g.scopedReplacementModLocked(seq, mod, kind)
			if !ok {
				return uuid.Nil
			}
			if kind == ModExileInsteadOfYourGraveyard {
				// CR 616.1: the affected object's controller orders the
				// replacements — its owner when it has no controller
				// (every zone but the battlefield, where the two agree).
				if c, found := g.LookupCardForEffect(ev.CardID); found {
					return c.Controller
				}
			}
			return e.Controller
		},
		Label: label,
	}
	return eff
}

// scopedKindPrevents reports whether a replacement kind is a CR 615
// prevention effect.
func scopedKindPrevents(kind ModKind) bool {
	return kind == ModPreventCombatDamage || kind == ModPreventDamage || kind == ModPreventNextFromSource
}

// scopedReplacementAppliesLocked is the AppliesTo of each kind. Caller
// must hold g.mu.
func scopedReplacementAppliesLocked(g *Game, e ScopedEffect, m Mod, ev *ReplacementEvent) bool {
	switch m.Kind {
	case ModPreventCombatDamage:
		// Non-combat damage is untouched: ReplacementEvent.IsCombatDamage
		// is set by the combat-damage resolver and by nothing else.
		return ev.Kind == RepEventDamage && ev.IsCombatDamage &&
			(m.Player == uuid.Nil || ev.DamageTarget == m.Player)
	case ModPreventDamage:
		if ev.Kind != RepEventDamage || m.Amount < 1 {
			return false
		}
		if m.CombatOnly && !ev.IsCombatDamage {
			return false
		}
		if e.Scope == ScopeGame {
			return m.Player != uuid.Nil && ev.DamageTarget == m.Player
		}
		return scopedAffectsLiveObjectLocked(g, e, ev.DamageTarget)
	case ModExileInsteadOfLeaving:
		// `NewZone != ZoneExile` is what stops the effect replacing its
		// own result — CR 614.5 forbids it anyway, and it would be an
		// infinite loop if it did not.
		return ev.Kind == RepEventMove && ev.OldZone == ZoneBattlefield && ev.NewZone != ZoneExile &&
			scopedAffectsLiveObjectLocked(g, e, ev.CardID)
	case ModExileInsteadOfGraveyard:
		if ev.Kind != RepEventMove || ev.OldZone != ZoneBattlefield || ev.NewZone != ZoneGraveyard {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.Controller == e.Controller
	case ModExileInsteadOfYourGraveyard:
		if (ev.Kind != RepEventMove && ev.Kind != RepEventDiscard) || ev.NewZone != ZoneGraveyard {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || c.IsToken() {
			return false
		}
		// The graveyard a card goes to is its owner's.
		dest := ev.NewZoneOwner
		if dest == uuid.Nil {
			dest = c.Owner
		}
		return dest == m.Player
	case ModGainNoLife:
		// CR 119.10: "if a player would gain life" — a positive change
		// only. A gain of 0 is no life gain event at all.
		return ev.Kind == RepEventLife && ev.LifeDelta > 0 && ev.LifePlayer == m.Player
	case ModPreventNextFromSource:
		return g.nextFromSourceAppliesLocked(e, m, ev)
	}
	return false
}

// scopedAffectsLiveObjectLocked reports whether the object `id` is a
// member of e's pinned set NOW: the same instance, still the object
// that carried the pinned entry stamp (CR 400.7). The card is read
// wherever it is — a battlefield exit is judged before the move, while
// the permanent still carries its stamp.
func scopedAffectsLiveObjectLocked(g *Game, e ScopedEffect, id uuid.UUID) bool {
	if id == uuid.Nil || len(e.Affected) == 0 {
		return false
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return false
	}
	return affectedPredicate(e.Affected)(&c, g, nil)
}

// applyScopedReplacementLocked is the Replace of each kind. Caller must
// hold g.mu (write).
func (g *Game) applyScopedReplacementLocked(e ScopedEffect, mod int, m Mod, ev *ReplacementEvent) error {
	switch m.Kind {
	case ModPreventCombatDamage:
		ev.Cancel()
	case ModPreventNextFromSource:
		g.applyNextFromSourceLocked(e, mod, m, ev)
	case ModPreventDamage:
		// CR 615.7's arithmetic, not "cancel if the shield covers any
		// of it": a 4-point shield facing 6 damage prevents 4 and lets
		// 2 through; facing 3 it prevents all 3 and keeps 1.
		left := 0
		if ev.DamageAmount <= m.Amount {
			left = m.Amount - ev.DamageAmount
			ev.Cancel()
		} else {
			ev.DamageAmount -= m.Amount
		}
		g.setShieldChargeLocked(e.Seq, mod, left)
	case ModExileInsteadOfLeaving:
		ev.NewZone = ZoneExile
	case ModExileInsteadOfYourGraveyard:
		ev.NewZone = ZoneExile
		ev.NewZoneOwner = uuid.Nil
	case ModGainNoLife:
		// "That player gains no life instead": CR 614.10's null
		// replacement. No EventChangeLife, so "whenever you gain life"
		// sees nothing.
		ev.Cancel()
	case ModExileInsteadOfGraveyard:
		ev.NewZone = ZoneExile
		ev.NewZoneOwner = uuid.Nil
		label := e.Label
		if e.SourceName != "" {
			label = e.SourceName + " — return the exiled permanent"
		}
		// One delayed trigger per redirected card, because the
		// replacement fires once per move event — a wrath is one
		// event per creature.
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:   e.Controller,
			SourceCardID: e.Source.ID,
			Label:        label,
			At:           StepEnd,
			Cards:        []uuid.UUID{ev.CardID},
			Body:         BodyRef{key: m.Then},
		})
	}
	return nil
}

// setShieldChargeLocked writes a prevention shield's new charge — the
// one change a record ever sees (ADR 0041 P8). COPY ON WRITE: the
// registry slice and the record's Mods are both rebuilt, never written
// in place, because a clone taken before the damage shares them and
// must keep the old charge (an undo rewinds it). A shield with no
// charge left is removed. Caller must hold g.mu (write).
func (g *Game) setShieldChargeLocked(seq int64, mod int, left int) {
	i, ok := g.scopedEffectIndexBySeqLocked(seq)
	if !ok {
		return
	}
	rec := g.ScopedEffects[i]
	next := make([]ScopedEffect, 0, len(g.ScopedEffects))
	next = append(next, g.ScopedEffects[:i]...)
	if left > 0 || len(rec.Mods) > 1 {
		rec.Mods = cloneMods(rec.Mods)
		rec.Mods[mod].Amount = left
		next = append(next, rec)
	}
	next = append(next, g.ScopedEffects[i+1:]...)
	if len(next) == 0 {
		next = nil
	}
	g.ScopedEffects = next
}

// maxScopedEffectSeq is the largest Seq among the records — what restore
// sets Game.scopedEffectSeq to.
func maxScopedEffectSeq(records []ScopedEffect) int64 {
	var n int64
	for i := range records {
		if records[i].Seq > n {
			n = records[i].Seq
		}
	}
	return n
}
