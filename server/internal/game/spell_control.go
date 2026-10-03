package game

import (
	"sort"

	"github.com/google/uuid"
)

// spell_control.go — gaining control of a SPELL on the stack (ADR 0104,
// #1745): Invert Polarity, Aethersnatch, Commandeer, Perplexing
// Chimera, Sudden Substitution.
//
// The rules make this layer 2. A control change is a continuous effect
// (CR 611.1) applied in layer 2 (CR 613.1b), a spell is an object with
// a controller (CR 109.4), and CR 400.7a carries an effect that changed
// a permanent spell's controller onto the permanent it becomes. So a
// steal is not a new mechanism here any more than Act of Treason was in
// ADR 0063: it is a `setController` ScopedEffect, the same data record,
// swept by the same duration function and sorted by the same CR 613.7
// timestamp. Three things are new, and they are the whole of this
// file:
//
//  1. A STACK PIN. AffectedObject.OnStack / Epoch and
//     Duration.PinnedOnStack / PinnedEpoch name a spell by instance ID
//     and the ObjectEpoch it has on the stack, so the record follows
//     that object and nothing else (CR 400.7), and the duration sweep
//     collects it once the object has left the stack.
//
//  2. A STACK STEP of the layer pass. stackControlPassLocked runs after
//     the battlefield pass: for each spell it reseeds control from
//     StackItem.BaseController, applies every record pinned to it in
//     timestamp order, and materialises the answer onto
//     StackItem.Controller and the spell card's Card.Controller — the
//     same materialisation materialiseControlLocked does for a
//     permanent, and for the same reason: every reader of "you" is
//     right without being touched. Since ADR 0107 §3 (#1854) the step
//     is layers 2 and 6: stackKeywordPassLocked (spell_keywords.go)
//     follows it with the keywords a spell has been given. Registration
//     refuses any other mod on a stack pin, because no card changes
//     another characteristic of a spell this way yet.
//
//  3. THE HAND-OFF at resolution (inheritSpellControlLocked). A stolen
//     permanent spell enters under the thief, its default controller
//     is the caster (CR 110.2b), and each record is re-pinned to the
//     permanent (CR 400.7a) with its own timestamp.
//
// And one rule the whole engine needed and did not have until this
// seam leaned on it: when a player leaves the game, every effect that
// gives them control ENDS (CR 800.4a) — endControlEffectsForLocked.

// stackCardLocked returns the card on the stack with this instance ID.
//
// Caller must hold g.mu.
func (g *Game) stackCardLocked(id uuid.UUID) (*Card, bool) {
	if g.Stack == nil {
		return nil, false
	}
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == id {
			return &g.Stack.Cards[i], true
		}
	}
	return nil, false
}

// sameEpochPresentLocked reports whether the permanent `id` is on the
// battlefield — or phased out, which is not a zone change
// (CR 702.26d) — as the object with ObjectEpoch `epoch`.
//
// Caller must hold g.mu.
func (g *Game) sameEpochPresentLocked(id uuid.UUID, epoch int) bool {
	if c, ok := g.battlefieldCardLocked(id); ok {
		return c.ObjectEpoch == epoch
	}
	if g.PhasedOut != nil {
		for _, c := range g.PhasedOut.Cards {
			if c.InstanceID == id {
				return c.ObjectEpoch == epoch
			}
		}
	}
	return false
}

// stolenSpellLocked is the spell a control writer may act on: a SPELL
// item whose card is on the stack. Abilities are refused — no printed
// card gains control of one (ADR 0104, "Out of scope").
//
// Caller must hold g.mu.
func (g *Game) stolenSpellLocked(spellID uuid.UUID) (*Card, *StackItem, bool) {
	item := g.StackMeta[spellID]
	if item == nil || item.Kind != StackItemSpell {
		return nil, nil, false
	}
	card, ok := g.stackCardLocked(spellID)
	if !ok {
		return nil, nil, false
	}
	return card, item, true
}

// playerInGameLocked reports whether `id` is a seated player who has
// not left the game. CR 800.4b: nothing changes to the control of a
// player who has left.
//
// Caller must hold g.mu.
func (g *Game) playerInGameLocked(id uuid.UUID) bool {
	p := g.playerByIDLocked(id)
	return p != nil && !p.Eliminated
}

// GainControlOfSpellForEffect gives `player` control of the spell
// `spellID` on the stack, indefinitely (ADR 0104; CR 611.1, 611.2a,
// 613.1b). Aethersnatch, Commandeer and Invert Polarity's won flip.
//
// It refuses — returning false, having registered nothing — anything
// but a spell on the stack, a player who is not in the game
// (CR 800.4b), and a player who already controls it.
//
// The change is materialised before it returns (it recomputes), so the
// caller can offer the new controller new targets at once: target
// legality is judged against StackItem.Controller
// (stackItemSourceLocked), and it has to be the thief's by then.
//
// Caller must hold g.mu (write). Effects call it from inside the
// resolution frame, which already holds it.
func (g *Game) GainControlOfSpellForEffect(sourceID, spellID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || !g.playerInGameLocked(player) {
		return false
	}
	card, item, ok := g.stolenSpellLocked(spellID)
	if !ok || item.Controller == player {
		return false
	}
	if item.BaseController == uuid.Nil {
		item.BaseController = item.Controller
	}
	if !g.registerScopedEffectLocked(sourceID,
		[]AffectedObject{PinStackObject(spellID, card.ObjectEpoch)},
		[]Mod{SetControllerMod(player)},
		g.PinnedToStack(IndefiniteDuration(), spellID), label, timeNowUnixNano()) {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	return true
}

// ExchangeControlOfSpellAndPermanentForEffect exchanges control of the
// spell `spellID` and the permanent `permanentID` (Perplexing Chimera,
// Sudden Substitution; CR 701.12).
//
// ALL OR NOTHING (CR 701.12a): both objects are checked before either
// record is registered, and if either is gone nothing is exchanged.
// If one player controls both, the exchange does nothing (CR 701.12b).
// A player who has left the game is refused (CR 800.4b).
//
// ONE EFFECT, ONE TIMESTAMP (CR 613.7): the two records share a
// timestamp, as ExchangeControlForEffect's two halves do.
//
// Like GainControlOfSpellForEffect it recomputes before it returns, so
// "the spell's controller may choose new targets" asks the right
// player. Caller must hold g.mu (write).
func (g *Game) ExchangeControlOfSpellAndPermanentForEffect(sourceID, spellID, permanentID uuid.UUID, label string) bool {
	card, item, ok := g.stolenSpellLocked(spellID)
	if !ok {
		return false
	}
	perm, ok := g.battlefieldCardLocked(permanentID)
	if !ok {
		return false
	}
	spellCtl, permCtl := item.Controller, perm.Controller
	if spellCtl == permCtl || !g.playerInGameLocked(spellCtl) || !g.playerInGameLocked(permCtl) {
		return false
	}
	if item.BaseController == uuid.Nil {
		item.BaseController = item.Controller
	}
	ts := timeNowUnixNano()
	g.registerScopedEffectLocked(sourceID, g.PinnedObjectsLocked(permanentID),
		[]Mod{SetControllerMod(spellCtl)},
		g.PinnedTo(IndefiniteDuration(), permanentID), label, ts)
	g.registerScopedEffectLocked(sourceID,
		[]AffectedObject{PinStackObject(spellID, card.ObjectEpoch)},
		[]Mod{SetControllerMod(permCtl)},
		g.PinnedToStack(IndefiniteDuration(), spellID), label, ts)
	g.RecomputeLayersIfStaleLocked()
	return true
}

// spellControlGrant is one record's layer-2 contribution to a spell.
type spellControlGrant struct {
	player uuid.UUID
	source uuid.UUID
	ts     int64
}

// spellControlGrantsLocked lists every live `setController` record
// pinned to the stack object `c`, oldest timestamp first (CR 613.7) —
// stable, so two records with one timestamp keep their registration
// order, as the battlefield bucket does.
//
// Caller must hold g.mu.
func (g *Game) spellControlGrantsLocked(c *Card) []spellControlGrant {
	var out []spellControlGrant
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if !pinsStackObject(e.Affected, c.InstanceID, c.ObjectEpoch) {
			continue
		}
		for _, m := range e.Mods {
			if m.Kind == ModSetController {
				out = append(out, spellControlGrant{player: m.Player, source: e.Source.ID, ts: e.Timestamp})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ts < out[j].ts })
	return out
}

// pinsStackObject reports whether an affected set names the spell
// (id, epoch) on the stack.
func pinsStackObject(set []AffectedObject, id uuid.UUID, epoch int) bool {
	for _, a := range set {
		if a.OnStack && a.ID == id && a.Epoch == epoch {
			return true
		}
	}
	return false
}

// stackControlPassLocked is the stack step of the layer pass (ADR 0104):
// layer 2 for every spell on the stack.
//
// For each spell with a stack item it reseeds control from
// StackItem.BaseController (zero: the current controller IS the base),
// applies the records pinned to it in timestamp order, and materialises
// the answer onto StackItem.Controller and onto the spell card's
// Card.Controller. The second write is what keeps "target spell you
// don't control" (a predicate over the stack CARD) in step with the
// item — and it is written for every spell, stolen or not, which is
// what fixes a spell cast off another player's card (#1745 finding 1)
// wherever a recompute has run since.
//
// A spell mid-resolution has no stack item (resolveTopOfStackLocked
// deletes it first) and is skipped: its entry window may have
// re-stamped the card's controller (ADR 0102), and this step must not
// undo that.
//
// CR 708.5: a face-down spell may be looked at by its controller and
// by nobody else, so a face-down spell whose controller changes has
// its knowledge rebuilt from the same viewers rule its landing used —
// the new controller may look, the old one no longer may.
//
// Returns the deltas by value, for recomputeLayersLocked to emit after
// the store (#930's reason).
//
// Caller must hold g.mu (write).
func (g *Game) stackControlPassLocked() []controlChange {
	if g.Stack == nil || len(g.StackMeta) == 0 {
		return nil
	}
	var changed []controlChange
	for i := range g.Stack.Cards {
		c := &g.Stack.Cards[i]
		item := g.StackMeta[c.InstanceID]
		if item == nil || item.Kind != StackItemSpell {
			continue
		}
		ctrl := item.BaseController
		if ctrl == uuid.Nil {
			ctrl = item.Controller
		}
		var source uuid.UUID
		for _, gr := range g.spellControlGrantsLocked(c) {
			ctrl, source = gr.player, gr.source
		}
		c.Controller = ctrl
		if ctrl == item.Controller {
			continue
		}
		changed = append(changed, controlChange{card: c.InstanceID, from: item.Controller, to: ctrl, source: source})
		item.Controller = ctrl
		if c.FaceDownIsPermanent() {
			viewers := g.faceDownViewersLocked(*c)
			c.ClearKnown()
			c.AddKnowersAll(viewers)
		}
	}
	return changed
}

// emitSpellControlChangesLocked emits one EventSpellControlChanged per
// delta the stack step found. Same place and same reason as
// emitControlChangesLocked: after the store, so a listener never reads
// a half-applied pass.
//
// Caller must hold g.mu (write).
func (g *Game) emitSpellControlChangesLocked(changed []controlChange) {
	for _, ch := range changed {
		g.EmitEvent(Event{
			Kind:   EventSpellControlChanged,
			Actor:  ch.to,
			Target: ch.from,
			CardID: ch.card,
			Source: ch.source,
		})
	}
}

// inheritSpellControlLocked is CR 110.2b and CR 400.7a at the moment a
// permanent spell becomes a permanent (ADR 0104 §3). Called from the
// entry landing, after the move and before any event about the entry
// goes out.
//
//   - The permanent's default controller is the player who put the
//     spell on the stack: Card.BaseController is stamped with the
//     item's base rather than captured lazily from Card.Controller
//     (which is the thief). When an ADR 0102 entry-controller effect
//     changed whom the permanent enters under, the default is that
//     player instead — it is who the permanent entered under
//     (CR 110.2).
//   - Every record pinned to the spell (`stackEpoch`, the epoch it had
//     on the stack) is REPLACED by one pinned to the permanent: the
//     same mods, timestamp and sequence number, so it sorts against a
//     later control effect exactly as it did on the stack.
//
// The next pass reseeds from the base, applies the record, and lands
// on the controller the permanent already has, so no control-change
// event fires for the entry itself.
//
// For a spell nothing ever took (its item has no base), which is every
// other permanent spell, only the second half can apply: a keyword an
// effect gave the spell (ADR 0107 §3) follows it onto the permanent,
// which is CR 400.7a's "change the characteristics … of a permanent
// spell". With no such record it is a no-op.
//
// Caller must hold g.mu (write).
func (g *Game) inheritSpellControlLocked(entered uuid.UUID, spellID uuid.UUID, stackEpoch int, item *StackItem, enteredUnder uuid.UUID) {
	if item == nil {
		return
	}
	if item.BaseController == uuid.Nil {
		g.repinSpellControlLocked(spellID, stackEpoch, entered)
		return
	}
	perm, ok := g.battlefieldCardLocked(entered)
	if !ok {
		return
	}
	base := item.BaseController
	if enteredUnder != uuid.Nil && enteredUnder != item.Controller {
		base = enteredUnder
	}
	perm.BaseController = base
	g.repinSpellControlLocked(spellID, stackEpoch, entered)
}

// repinSpellControlLocked moves every record pinned to the spell
// (spellID, stackEpoch) onto the permanent `permanentID` (CR 400.7a).
// The registry slice is rebuilt, never edited in place, because its
// backing array is shared with every undo snapshot.
//
// Caller must hold g.mu (write).
//
// The permanent is pinned by its ObjectEpoch, not its entry stamp: this
// runs as the permanent LANDS, before the zone-move event stamps it,
// and a record pinned to a stamp that is about to change would be
// swept the moment it was. The epoch is final once MoveCard has landed
// the card (PinObjectByEpoch).
func (g *Game) repinSpellControlLocked(spellID uuid.UUID, stackEpoch int, permanentID uuid.UUID) {
	perm, ok := g.battlefieldCardLocked(permanentID)
	if !ok || perm.ObjectEpoch <= 0 {
		return
	}
	pins := []AffectedObject{PinObjectByEpoch(permanentID, perm.ObjectEpoch)}
	moved := false
	out := make([]ScopedEffect, len(g.ScopedEffects))
	copy(out, g.ScopedEffects)
	for i := range out {
		if !pinsStackObject(out[i].Affected, spellID, stackEpoch) {
			continue
		}
		out[i].Affected = append([]AffectedObject(nil), pins...)
		out[i].Mods = cloneMods(out[i].Mods)
		out[i].Duration = g.PinnedToEpoch(unpinnedDuration(out[i].Duration), permanentID)
		moved = true
	}
	if !moved {
		return
	}
	g.ScopedEffects = out
	g.layerVersion.Add(1)
}

// unpinnedDuration is `d` with its pin to the spell taken off, so the
// record can be pinned to the permanent instead (CR 400.7a) without
// losing the duration it was given (ADR 0109 §11 decision 3): a
// keyword a spell gained "until end of turn" is the permanent's until
// end of turn. Every other field — the kind, the player, the turn it
// expires after — is kept.
func unpinnedDuration(d Duration) Duration {
	d.Pinned = uuid.Nil
	d.PinnedEnteredAt = 0
	d.PinnedUnstamped = false
	d.PinnedOnStack = false
	d.PinnedEpoch = 0
	return d
}

// spellControlRecordsLocked returns value copies of every record
// pinned to the spell (spellID, epoch) — for the one path that cannot
// re-pin in place because the spell's object is gone before its
// permanent exists: a COPY of a permanent spell, which ceases to exist
// and then becomes a token (CR 608.3f). The caller re-pins the copies
// onto the token once it has entered.
//
// Caller must hold g.mu.
func (g *Game) spellControlRecordsLocked(spellID uuid.UUID, epoch int) []ScopedEffect {
	var out []ScopedEffect
	for _, e := range g.ScopedEffects {
		if pinsStackObject(e.Affected, spellID, epoch) {
			e.Affected = append([]AffectedObject(nil), e.Affected...)
			e.Mods = cloneMods(e.Mods)
			out = append(out, e)
		}
	}
	return out
}

// adoptSpellControlLocked registers `records` — copies taken with
// spellControlRecordsLocked — pinned to the permanent `permanentID`,
// and stamps its default controller (CR 110.2b, CR 400.7a). Each keeps
// its timestamp and sequence number.
//
// Caller must hold g.mu (write).
func (g *Game) adoptSpellControlLocked(records []ScopedEffect, permanentID, base uuid.UUID) {
	if len(records) == 0 {
		return
	}
	perm, ok := g.battlefieldCardLocked(permanentID)
	if !ok {
		return
	}
	pins := g.PinnedObjectsLocked(permanentID)
	if base != uuid.Nil {
		perm.BaseController = base
	}
	out := make([]ScopedEffect, 0, len(g.ScopedEffects)+len(records))
	out = append(out, g.ScopedEffects...)
	for _, e := range records {
		e.Affected = append([]AffectedObject(nil), pins...)
		e.Duration = g.PinnedTo(unpinnedDuration(e.Duration), permanentID)
		out = append(out, e)
	}
	g.ScopedEffects = out
	g.layerVersion.Add(1)
}

// endControlEffectsForLocked is the middle step of CR 800.4a: when a
// player leaves the game, "any effects which give that player control
// of any objects or players end". Every ScopedEffect record that gives
// `playerID` control — of a spell or of a permanent — is dropped, and
// the layer version is bumped so the next pass hands the objects back
// (a spell to the previous controller or its base, a permanent to its
// Card.BaseController).
//
// Until ADR 0104 nothing did this: every control effect was once an
// Aura's static ability, which leaves with its owner, and ADR 0060 §4
// could say "there is no separate registry of control grants to walk".
// Since #756 a resolving spell's theft (Act of Treason) is a record in
// that registry, and a thief leaving the game left the creature under
// the departed player, to be exiled — where CR 800.4a's own example
// sends it home.
//
// A record is dropped whole: it IS the effect, and the effect ends.
// Reports whether anything was dropped. Caller must hold g.mu (write).
func (g *Game) endControlEffectsForLocked(playerID uuid.UUID) bool {
	if len(g.ScopedEffects) == 0 {
		return false
	}
	kept := make([]ScopedEffect, 0, len(g.ScopedEffects))
	for _, e := range g.ScopedEffects {
		if givesControlTo(e, playerID) {
			continue
		}
		kept = append(kept, e)
	}
	if len(kept) == len(g.ScopedEffects) {
		return false
	}
	if len(kept) == 0 {
		kept = nil
	}
	g.ScopedEffects = kept
	g.layerVersion.Add(1)
	return true
}

// givesControlTo reports whether the record gives `playerID` control
// of something.
func givesControlTo(e ScopedEffect, playerID uuid.UUID) bool {
	for _, m := range e.Mods {
		if m.Kind == ModSetController && m.Player == playerID {
			return true
		}
	}
	return false
}

// stackPinProblem is the check for a stack pin (ADR 0104): a record
// that names a spell may only change its controller (layer 2) or add
// keywords to it (layer 6, ADR 0107 §3), because those are the two
// things the stack step of the layer pass applies. Empty when the
// record is fine.
//
// Run at registration, where a failure is a programming error, and at
// restore (checkEffectKeys), where it is a file from a newer binary
// whose stack step applies something this one would silently drop.
func stackPinProblem(affected []AffectedObject, mods []Mod) string {
	onStack := false
	for _, a := range affected {
		if a.OnStack {
			onStack = true
			break
		}
	}
	if !onStack {
		return ""
	}
	for _, m := range mods {
		if m.Kind != ModSetController && m.Kind != ModAddKeywords {
			return "pins a spell on the stack with a " + string(m.Kind) +
				" mod; the stack step of the layer pass applies layer 2 (setController) and layer-6 keyword grants (addKeywords) only"
		}
	}
	return ""
}

// SpellControllerForEffect is the current controller of the spell
// `spellID` on the stack; false when there is no such spell.
//
// Caller must hold g.mu (read or write).
func (g *Game) SpellControllerForEffect(spellID uuid.UUID) (uuid.UUID, bool) {
	item := g.StackMeta[spellID]
	if item == nil {
		return uuid.Nil, false
	}
	return item.Controller, true
}

// SpellBaseControllerForEffect is the player a spell on the stack was
// put there by (CR 110.2b), or its current controller when nothing has
// ever changed it; false when there is no such spell.
//
// Caller must hold g.mu (read or write).
func (g *Game) SpellBaseControllerForEffect(spellID uuid.UUID) (uuid.UUID, bool) {
	item := g.StackMeta[spellID]
	if item == nil {
		return uuid.Nil, false
	}
	if item.BaseController != uuid.Nil {
		return item.BaseController, true
	}
	return item.Controller, true
}
