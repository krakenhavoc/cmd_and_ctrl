package game

import "github.com/google/uuid"

// indestructible.go teaches the engine CR 702.12 — the keyword that
// until S25 was a string the catalog could write and nothing could
// read.
//
// # Why it needed its own file rather than a line in keywords.go
//
// Indestructible is the first keyword in `canonicalKeywords` whose
// consumer is NOT the combat or targeting path. Flying, menace and
// reach are read by `CanBlockLocked` / `BlockerCountValid`; hexproof and
// shroud by `CanBeTargetedBy`. Indestructible is read by the
// DESTRUCTION path, which runs through the SBA loop and the effect
// API — two call sites in two other files that have nothing else in
// common. Putting the rule and its rationale here gives it one home,
// and keeps the diff off `mutations.go`, where every concurrent
// sprint is also editing.
//
// # What indestructible actually stops, and what it does not
//
// CR 702.12b: "A permanent with indestructible can't be destroyed.
// Such permanents aren't destroyed by lethal damage, and they ignore
// the state-based action that checks for lethal damage."
//
// That sentence is narrower than players remember, so the engine
// enumerates it explicitly. Indestructible SAVES a permanent from:
//
//	CR 701.8   a "destroy" effect (Wrath of God, Doom Blade, …)
//	CR 704.5g  the lethal-marked-damage state-based action
//	CR 704.5h  the deathtouch state-based action
//
// and does NOT save it from:
//
//	CR 704.5f  toughness 0 or less — that SBA PUTS the creature into
//	           the graveyard, it does not destroy it. A 2/2 with
//	           indestructible under two -1/-1 counters still dies.
//	CR 704.5i  a planeswalker at 0 loyalty — also "put into its
//	           owner's graveyard", which is why an indestructible
//	           planeswalker still dies to its own minus ability.
//	CR 704.5v  a battle at 0 defense counters — sacrificed.
//	CR 701.21a sacrifice, which is never destruction. This is the
//	           reason the check CANNOT live in
//	           routeBattlefieldCardToOwnerGraveyardLocked: that
//	           function is the shared exit ramp for destruction,
//	           sacrifice AND every zero-counter SBA above
//	           (sacrifice.go:32 routes through it), and a check
//	           there would wrongly protect all five.
//	           Exile, bounce, and "put into a graveyard" likewise.
//
// So the gate goes at the places that mean "destroy" and nowhere
// else: `DestroyPermanentForEffect` (effect_api.go — the entry point
// every catalog `DestroyTarget` reaches), `DestroyPermanentsForEffect`
// (simultaneous.go — the mass entry point every board wipe reaches,
// via `DestructibleForEffect` below) and the two damage branches of
// the SBA `doomed` pre-pass (mutations.go).
//
// S30 (#470 / #446): the mass entry point was the one that got
// missed. S25 shipped the single-target gate hours after S23 shipped
// the batched sweep, and the sweep's own comment still described
// indestructible as "not modelled anywhere in the engine" — a note
// that was true when written and became a lie the same week. Every
// Wrath of God in the catalog destroyed an Avacyn until this was
// fixed, which is the direction this engine least wants to be wrong
// in: a player losing permanents they were entitled to keep.
//
// # Damage is still marked
//
// CR 702.12b again: an indestructible creature is dealt damage
// normally and the damage stays marked until cleanup. Nothing in the
// damage-marking path consults this file. That matters for two
// downstream rules that keep working unchanged: lifelink still gains
// life off damage dealt to an indestructible blocker, and a creature
// that LOSES indestructible later in the turn (a turn-scoped grant
// expiring at cleanup is the common case, but a control-change or
// text-change effect could do it mid-turn) is destroyed by the very
// next SBA pass on the damage it accumulated while protected. Storing
// "was protected when the damage landed" instead would get that
// backwards.
//
// The same reasoning is why the deathtouch branch is filtered HERE
// rather than by skipping `MarkedLethalByDeathtouch` at marking time:
// the flag records a fact about the damage event, and indestructible
// is a fact about the permanent right now.
//
// Added in S25 (#77). Before this, `"indestructible"` reached
// `Characteristic.Abilities` from `Spec.PrintedKeywords`
// (Darksteel Citadel) and from `GrantKeywordUntilEOT` (Boros Charm
// mode 1) and rendered as a badge that promised a rule the engine
// never made.

// IsIndestructible reports whether the permanent currently has
// indestructible (CR 702.12a).
//
// Reads through `HasKeyword`, so it sees the full layer-6 picture:
// the card's own printed keyword (Avacyn, Darksteel Citadel), a
// battlefield static granting it to others (Avacyn's second line,
// Bastion Protector's commander clause) and a turn-scoped grant
// (Heroic Intervention, Boros Charm) all land in the same
// `Effective().Abilities` slice.
//
// CALLER MUST HAVE FRESH LAYERS. `HasKeyword` reads the cached
// `effective` characteristic; a grant registered earlier in the same
// resolution frame is only visible after a recompute. The SBA loop
// already calls `RecomputeLayersIfStaleLocked` at its top;
// `DestroyPermanentForEffect` calls it explicitly for exactly this
// reason (Heroic Intervention resolving in response to a Wrath is
// the case that breaks otherwise).
//
// nil card returns false, matching `HasKeyword`.
func IsIndestructible(c *Card) bool {
	return HasKeyword(c, "indestructible")
}

// destroyBattlefieldPermanentLocked is the engine's single
// destruction verb (CR 701.8): it routes the permanent to its
// owner's graveyard UNLESS the permanent is indestructible, in which
// case nothing happens at all — "the permanent remains on the
// battlefield", not "the permanent is moved and then returned".
//
// Returns nil for an indestructible permanent and for a card that
// isn't on the battlefield, because a destroy that finds nothing to
// destroy is a no-op in the rules, not an error. Wrath-style callers
// iterate a snapshot of the battlefield and would otherwise have to
// swallow an error per protected creature.
//
// Caller must hold g.mu.
func (g *Game) destroyBattlefieldPermanentLocked(cardID uuid.UUID, opts DestroyOptions) error {
	// A grant registered moments ago in this same resolution frame
	// bumps layerVersion but does not refresh `effective` until
	// something asks. Ask.
	g.RecomputeLayersIfStaleLocked()
	if c := findBattlefieldCard(g, cardID); c != nil && IsIndestructible(c) {
		// CR 702.12b: nothing happens at all. Note that a regeneration
		// shield on an indestructible permanent is NOT spent here —
		// there was no destruction for it to replace (#667).
		return nil
	}
	// #667: the destroy route, not the plain battlefield exit
	// routeBattlefieldCardToOwnerGraveyardLocked takes. That is what
	// marks the CR 614 event as a DESTRUCTION, which is the only thing
	// a regeneration shield is allowed to replace, and what carries
	// "it can't be regenerated" (CR 701.19c) down to it.
	return g.routeBattlefieldExitInBatchThenLocked(cardID, destroyRouteWith(opts), nil, nil)
}

// DestructibleForEffect narrows a mass-destruction set down to the
// permanents a "destroy" can actually destroy, preserving order and
// dropping nothing else (CR 702.12b).
//
// # Why the filter lives here rather than inside the destroy loop
//
// The obvious fix for #470 / #446 was to make destroyPermanentsLocked
// skip indestructible permanents as it walks `ids`. That is wrong
// twice over.
//
//   - destroyPermanentsLocked is SHARED with the state-based-action
//     sweep in mutations.go, and that sweep's `doomed` set contains
//     three kinds of permanent indestructible does NOT save: a
//     creature at 0 toughness (CR 704.5f), a planeswalker at 0
//     loyalty (CR 704.5i), a battle at 0 defense (CR 704.5v). All
//     three are "put into a graveyard", not "destroy". A filter one
//     level down would wrongly protect every one of them. The SBA
//     already filters the two branches that ARE destruction while it
//     collects `doomed`, which is why that path was correct before
//     this function existed and stays untouched by it.
//   - The batch is published BEFORE the first move
//     (beginSimultaneousExitLocked) and is what every dies-trigger in
//     the wipe observes (CR 700.4 / 603.10). Skipping inside the loop
//     would announce a death that never happens: a Blood Artist would
//     drain for an Avacyn that is still standing.
//
// So the effect-facing entry point filters first and hands
// destroyPermanentsLocked a set that is already true.
//
// Layers are recomputed before the check for the same reason
// destroyBattlefieldPermanentLocked recomputes: Heroic Intervention
// resolving in response to a Wrath registers its grant in a frame
// that has not refreshed `effective` yet, and that is precisely the
// case a player expects to work.
//
// Exported because the catalog's sweep primitive needs the same
// answer for a second purpose — see effects/mass.go, where "for each
// creature destroyed this way" must not count the survivors. One
// rule, one implementation.
//
// Caller must hold g.mu in write mode.
func (g *Game) DestructibleForEffect(ids []uuid.UUID) []uuid.UUID {
	if len(ids) == 0 {
		return ids
	}
	g.RecomputeLayersIfStaleLocked()
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if c := findBattlefieldCard(g, id); c != nil && IsIndestructible(c) {
			continue
		}
		out = append(out, id)
	}
	return out
}
