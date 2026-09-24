package game

import "github.com/google/uuid"

// life_lock.go — CR 119.7 and CR 119.8: a player whose LIFE TOTAL
// CAN'T CHANGE, and the one predicate everything that could move a
// life total asks.
//
// ADR 0085, #1200. The clause Teferi's Protection has carried a caveat
// for since #1197, and Platinum Emperion's whole printed text:
//
//	"Your life total can't change. (You can't gain or lose life. You
//	 can't pay any amount of life except 0.)"
//
// WHY IT IS NOT IN TurnScopedReplacements, which is the registry
// #1200 named. That slice holds ReplacementEffect values, and a
// ReplacementEffect is two closures — counted in
// ContinuationCensus.TurnScopedReplacements and classified `dropped`
// by the snapshot drift test, because a closure cannot be written to
// disk. "Until your next turn" is a whole turn cycle, so storing the
// clause there would make the one card famous for buying you a
// rotation also the one card that takes undo and persistence away from
// the table for that rotation. ADR 0084's closing section makes the
// argument at length; ADR 0063 Decision 8's deferral of durations on
// that registry therefore still stands, because the card said to need
// one does not.
//
// WHERE IT IS INSTEAD: PlayerStatic.LifeTotalLocked, a third payload
// beside Keyword and Timing on the slice #1197 built
// (player_statics.go). Plain data, a CR 611.2 Duration swept through
// the one durationExpiredLocked, cloned by value, mirrored into the
// snapshot, classified `carried`. "Your life total can't change" is a
// statement about a PLAYER, in the same grammatical position as "you
// have hexproof" and "you may cast spells as though they had flash",
// so it wants the same home.
//
// TWO SOURCES, AND ONLY ONE OF THEM IS STORED — ADR 0072's amendment's
// split, verbatim, because the catalog needs both halves:
//
//   - DERIVED — a permanent on the battlefield whose printed static is
//     "Your life total can't change" (Platinum Emperion). Declared as
//     effects.Spec.PlayerLifeTotalLocked, read through
//     CatalogPlayerLifeTotalLocked on every query and written nowhere,
//     so two Emperions compose and one of them leaving cannot revoke
//     the other's lock.
//   - GRANTED — a resolved spell, "until your next turn, your life
//     total can't change" (Teferi's Protection, Teferi's Reproach).
//     This one HAS to be stored: the spell is in exile a moment after
//     it resolves.
//
// FOUR CONSUMERS, and two of them were already written and
// unreachable:
//
//	every life change   lifeTotalCantChangeReplacement (below)   CR 614, through #482's one window
//	a life PAYMENT      payLifeAsCostLocked (life_tail.go)       CR 119.8, CR 614.17b — the cancel already refuses the cost
//	damage to a player  applyResolvedDamageToPlayerLocked        CR 120.3a — still DEALT, the total does not move
//	a cost VALIDATOR    CanPayLifeLocked (below), eight sites     CR 614.17b — so a refused cost is never offered
//
// WHAT IT IS NOT. Not "you can't lose the game" (CR 104.3, Platinum
// Angel and Phyrexian Unlife) — a locked total cannot reach 0, but
// CR 104.3's other clauses are untouched and are their own seam
// (#883). Not a counter prohibition either: poison and every other
// counter on a player run their own CR 614 window (RepEventCounter,
// ADR 0056), which this built-in does not watch.

// CatalogPlayerLifeTotalLocked reports whether a battlefield permanent
// with this catalog key locks its CONTROLLER's life total — true for
// Platinum Emperion, false for everything else. Populated at init time
// by the cards/effects package from effects.Spec.PlayerLifeTotalLocked.
// A nil hook (no catalog wired) means no permanent locks anything.
//
// Derived rather than written, for the reason CatalogNoMaxHandSize and
// CatalogPlayerKeywords both spell out: a "set on enter, restore on
// leave" design has to answer "restore to what?", and gets two real
// cases wrong — two copies, and a player whose state was changed by
// something else in between. Derivation has no stored value to strand,
// so neither case exists, and undo needs no new state to clone.
var CatalogPlayerLifeTotalLocked func(oracleID string) bool

// GrantLifeTotalLockForEffect locks `player`'s life total for a
// duration — "until your next turn, your life total can't change".
// The *ForEffect surface: caller must hold g.mu (write), which a
// resolution frame already does.
//
// It writes a PlayerStatic carrying neither a Keyword nor a Timing,
// which is what keeps it out of playerAbilityTokensLocked's answer and
// out of castTimingVerdictLocked's: the three kinds of entry on that
// slice are told apart by their payload rather than by a discriminator
// field, the posture #1195 already took for the timing statement.
//
// Build the duration with the constructors in duration.go —
// g.UntilYourNextTurnDuration(p) for both of this seam's spells. An
// unstamped Duration is stamped "until end of turn" on the way in,
// exactly as GrantCastTimingForEffect does, so a card file that forgot
// the window gets the narrowest real one rather than a permanent lock.
func (g *Game) GrantLifeTotalLockForEffect(player uuid.UUID, label string, source uuid.UUID, d Duration) {
	if player == uuid.Nil {
		return
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return
	}
	if d == (Duration{}) {
		d = g.UntilEndOfTurnDuration()
	}
	p.Statics = append(p.Statics, PlayerStatic{
		LifeTotalLocked: true,
		Source:          source,
		Label:           label,
		Duration:        d,
	})
}

// playerLifeTotalCantChangeLocked is THE reader: may this player's
// life total move right now? Derived locks first, stored ones after,
// with an early-out on the first true — the life half's twin of
// playerAbilityTokensLocked, and like it a walk rather than a slice,
// because every consumer asks a yes/no question and the common answer
// is "nothing is locking anything".
//
// THE DURATION IS TESTED HERE as well as in the sweep, for the reason
// the keyword reader gives: the sweep is hygiene run at known moments
// (cleanup, the beginning of a turn, the top of a layer recompute),
// and the reader is the truth between them. A lock that ends as your
// next turn begins must not still be answering during the priority
// round that ends the previous turn.
//
// A nil player is not locked — a caller that has lost the seat has an
// existence problem, which the life and damage tails report separately
// (ErrPlayerNotFound).
//
// Caller must hold g.mu (read or write).
func (g *Game) playerLifeTotalCantChangeLocked(p *Player) bool {
	if p == nil {
		return false
	}
	if CatalogPlayerLifeTotalLocked != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller != p.ID {
				continue
			}
			// CatalogAbilityKey, not CatalogKey: "Your life total
			// can't change" is a static ability of the permanent, so
			// an Emperion that has lost its abilities (layer 6) stops
			// locking. The empty KEY is the skip, so a token is walked
			// like any other permanent (ADR 0083 decision 3).
			key := CatalogAbilityKey(*c)
			if key == "" {
				continue
			}
			if CatalogPlayerLifeTotalLocked(key) {
				return true
			}
		}
	}
	for _, s := range p.Statics {
		if !s.LifeTotalLocked {
			continue
		}
		if g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		return true
	}
	return false
}

// PlayerLifeTotalCantChangeLocked is the exported surface over the
// reader, for the protocol projection's badge, for the bot's move
// enumerator and for card files that want to ask.
//
// Caller must hold g.mu (read or write).
func (g *Game) PlayerLifeTotalCantChangeLocked(p *Player) bool {
	return g.playerLifeTotalCantChangeLocked(p)
}

// CanPayLifeLocked is CR 119.4 and CR 119.8 in one answer: may `p` pay
// `amount` life right now?
//
//   - CR 119.4 — a player may pay life only from a life total at least
//     as large as the payment. Paying down to exactly 0 is legal; the
//     state-based check then ends the game.
//   - CR 119.8 — "if an effect says that a player can't lose life, ...
//     a cost that involves having that player pay life can't be paid",
//     generalised by CR 614.17b: an event that can't happen can't be
//     chosen as part of a cost.
//
// ONE PREDICATE AT EVERY VALIDATOR, which is the point of it. Before
// ADR 0085 the affordability test was open-coded as `p.Life < n` at
// eight sites; payLifeAsCostLocked's CR 119.8 refusal was the backstop
// with no gate in front of it, and its own comment asked for this
// ("a card that adds one should add that check to the cost validators
// it reaches"). The argument for gating rather than only refusing is
// #695's, already made once for CR 119.4: announce refused the cast,
// the view showed the offer anyway, and a button that can only fail is
// worse than no button.
//
// A non-positive amount is payable by anybody — it is not a payment —
// which is what keeps "pay 0 life" and an absent life component out of
// the lock's way (and is Platinum Emperion's own parenthetical: "You
// can't pay any amount of life except 0").
//
// Caller must hold g.mu (read or write).
func (g *Game) CanPayLifeLocked(p *Player, amount int) bool {
	if amount <= 0 {
		return true
	}
	if p == nil || p.Life < amount {
		return false
	}
	return !g.playerLifeTotalCantChangeLocked(p)
}

// lifeChangeIsLockedLocked is the built-in's AppliesTo: is this a life
// change to a player whose total can't move?
//
// A zero delta is left alone. Nothing is being changed, there is
// nothing for the prohibition to stop, and cancelling it would report
// a replacement to a caller whose continuation is going to be told
// zero either way.
//
// Caller must hold g.mu.
func (g *Game) lifeChangeIsLockedLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventLife || ev.LifeDelta == 0 {
		return false
	}
	return g.playerLifeTotalCantChangeLocked(g.playerByIDLocked(ev.LifePlayer))
}

// lifeTotalCantChangeReplacement implements CR 119.7 and CR 119.8 on
// the CR 614 life window: a life change to a player whose total can't
// change simply does not happen.
//
// The fourth engine-owned replacement, and a built-in for the reason
// the other three are: it is printed in the Comprehensive Rules rather
// than on any object. The lock is a static of a permanent (Platinum
// Emperion) or a stored grant (Teferi's Protection), but the
// PROHIBITION is a rule, and there is no card for the pipeline to hang
// it on — a grant outlives its source, which is the whole reason it is
// stored on the player at all.
//
// ONE PREDICATE COVERS EVERY LIFE CHANGE IN THE GAME, because #482
// already made every writer of a life total run this window: a catalog
// GainLife, a drain, "each opponent loses 3 life", the CR 702.15b
// lifelink credit, the life half of an exchange, a Phyrexian symbol's
// two life, a shockland's two life, and the public ChangePlayerLife
// sandbox verb. None of them learns the rule.
//
// WHAT HAPPENS DOWNSTREAM was already written, and already named this
// card's clause. changeLifeThroughReplacementsLocked's cancel branch
// (life_tail.go): "CR 614.10 with a null replacement — 'your life
// total can't change'. No mutation, and no EventChangeLife for a
// 'whenever you gain life' trigger to see, because nothing happened."
// The caller's continuation still runs, with zero, so a drain adding up
// what several players lost is correct across a table where one seat is
// locked. And payLifeAsCostLocked turns the same cancel into a REFUSED
// cost (CR 119.8, CR 614.17b) rather than a payment that bought
// something for nothing.
//
// PREEMPTIVE. It applies before every other applicable replacement and
// asks no CR 616 ordering question, for three reasons in the order they
// matter. (1) There is only one answer: an amount replacement — Rhox
// Faithmender's doubler, Bloodletter of Aclazotz's — rewrites the delta
// and this cancels whatever it rewrote it to, so both orders reach
// zero. That is the argument allPureCancels (#710) and sameModification
// (#792) already make for two other windows. (2) A charged prevention
// shield must not spend itself on a change that was never going to
// happen, which is #420's argument for CR 702.16e one event kind over.
// (3) A life PAYMENT cannot pause at all (mustSettleNow, CR 601.2h), so
// the ordering question could not be put on the path that most needs
// the answer. The declared cost is ADR 0072 §4's: CR 616.1 would hand
// the choice to the affected player, and a replacement that did
// something BESIDES rewriting the amount would get to do its other
// thing if it were ordered first. No printed card in this catalog is
// that shape.
//
// DAMAGE IS NOT THIS WINDOW and must not be. Damage reduces life
// directly once the DAMAGE replacements have settled (CR 120.3), so a
// life replacement getting a second bite would halve a Bolt twice. The
// damage half of this rule is four lines in
// applyResolvedDamageToPlayerLocked: the damage is still DEALT —
// EventDealDamage fires, the CR 903.10a commander tally accrues,
// lifelink credits — and only the life total stays put.
//
// No Controller: CR 616.1 would ask the affected player, and a
// preemptive effect is never ordered against anything, so there is
// nobody to ask.
var lifeTotalCantChangeReplacement = ReplacementEffect{
	Watches:    []EventKind{EventChangeLife},
	Preemptive: true,
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		return g.lifeChangeIsLockedLocked(ev)
	},
	Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
		// CR 614.10 with a null replacement: no mutation and no
		// EventChangeLife, so a "whenever you gain life" trigger sees
		// nothing, because nothing happened.
		ev.Cancel()
		return nil
	},
	Label: "Life total can't change",
}
