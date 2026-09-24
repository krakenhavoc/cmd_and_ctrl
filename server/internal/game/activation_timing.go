package game

import (
	"github.com/google/uuid"
)

// activation_timing.go — #1208, ADR 0066's and ADR 0073's amendments
// of 2026-09-23: the per-PLAYER half of CR 602.5d and CR 606.3.
//
// #1195 built the CAST side of this sentence (cast_timing.go): "you
// may cast spells as though they had flash" is a statement about a
// PLAYER, not about a card, and it wanted a type, two homes and one
// read. This is its ACTIVATION twin, and the three cards behind it
// print the same grammar one verb over:
//
//	The Wandering Emperor  "As long as The Wandering Emperor entered
//	                        this turn, you may activate her loyalty
//	                        abilities any time you could cast an
//	                        instant."
//	Teferi, Master of Time "You may activate loyalty abilities of
//	                        Teferi on any player's turn any time you
//	                        could cast an instant."
//	Leonin Shikari         "You may activate equip abilities any time
//	                        you could cast an instant."
//
// Before this file, `ActivatedAbilityShape.SorcerySpeed` and
// `AbilityCost.Loyalty != nil` were the whole of the engine's opinion
// about when an ability may be activated, read in three places that
// each wrote the rule out again. Nothing could open that window and
// nothing could narrow it.
//
// TWO TYPES, ONE VOCABULARY. This is NOT `CastTimingRule` with a
// "casts / activations" scope bolted on, and the reason is the homes:
//
//   - A cast timing statement has a STORED home (Emergence Zone's
//     "this turn", Teferi, Time Raveler's "+1"), so it must be pure
//     data — which is why its narrowing is `PermissionFilter`,
//     `FromZone` and a `CastTimingAffects` enum rather than a
//     predicate.
//   - An activation timing statement has only a DERIVED home. Every
//     card that prints one is a permanent whose static says it, and
//     the statement lasts exactly as long as that permanent is
//     there. Nothing is stored, so `Covers` can be a PREDICATE — and
//     it has to be, because what the printed cards narrow on is the
//     ability's SOURCE ("loyalty abilities of Teferi") and the
//     ABILITY ("equip abilities", "loyalty abilities"), two
//     different things that `PermissionFilter` — a filter over the
//     object being CAST — says neither of.
//
// Widening one type to carry both would have given every field two
// meanings and made every read start by asking which half it was
// looking at, which is exactly the argument ADR 0073's #1210
// amendment makes for `ActivationQuery` not being a widened
// `CastQuery`. So the SHAPE is copied and the TYPE is not: the fold
// order is CR 101.2's, `GrantTiming` is ADR 0066's own enum
// unchanged, `ActivationQuery` and `ActivationAbility` are #1210's
// unchanged, and the collection walk is
// `ActivationRestrictionsForCard`'s.
//
// MANA ABILITIES ARE NOT HERE AT ALL, and that is a rule rather than
// a carve-out — see ActivationTimingOpenLocked.

// ActivationTiming is one per-player activation-timing statement a
// permanent makes while it is on the battlefield: what it says, and
// which activations it says it about.
//
// Declared as a struct of hooks rather than as flags for the reason
// `ActivationRestriction` next door is: a card file writes a literal,
// nothing is ever stored, and the two narrowings the printed cards
// use (the ability's source, and which KIND of ability it is) are
// questions about the `ActivationQuery` the gate already builds.
//
// It carries no Player and no Duration, because it needs neither: the
// player is whoever the predicate admits, measured against
// `q.Source.Controller` exactly as a cast timing statement's
// `Affects` clause is; and the duration is the source's presence on
// the battlefield, which is what derivation MEANS (the argument
// `standingCastPermissionsLocked` and `CatalogPlayerKeywords` both
// make at length — two sources compose, and one leaving cannot revoke
// the other's grant).
type ActivationTiming struct {
	// Label is the clause as printed ("You may activate equip
	// abilities any time you could cast an instant"). Required:
	// effects.Register refuses a statement without one, because it
	// is what the next reader of the card file matches against the
	// oracle text.
	Label string

	// Timing is what the statement SAYS. `TimingFlash` opens the
	// instant-speed window on an ability that would otherwise answer
	// to CR 602.5d or CR 606.3; `TimingSorcery` and
	// `TimingYourTurnOnly` narrow it. `TimingNormal` — the zero
	// value — says nothing and is refused at Register, because a
	// statement that says nothing is a card file that meant to say
	// something.
	//
	// ADR 0066's enum, unchanged and not extended: one timing
	// vocabulary for the engine, which is the posture
	// `TimingYourTurnOnly` was added under.
	//
	// NO CATALOGUED CARD DECLARES THE RESTRICTING VALUES. A printed
	// restriction on somebody's activations says "can't be
	// activated" (Cursed Totem, Linvala, Grand Abolisher) and is a
	// BAN, which goes through `ActivationGateLocked` — see that
	// file. They are folded here anyway because the fold order IS
	// CR 101.2 and a grant-only read would have to be rewritten to
	// state it; three lines is a cheap way to have the rule written
	// down once.
	Timing GrantTiming

	// Covers decides whether this statement is ABOUT this
	// activation: which player's, of which ability, of which source.
	// Nil covers nothing, which makes an under-declared card file a
	// no-op rather than a board-wide timing change; Register refuses
	// it outright.
	//
	// Evaluated once per activation per statement, under g.mu.
	// Read-only — a *ForEffect accessor or a plain field read, never
	// a locking mutator. The same contract `ActivationRestriction.
	// Forbids` carries, because it is handed the same query.
	Covers func(q ActivationQuery) bool

	// ActiveWhen is the CR 716 / 719 / 721 designation gate, the same
	// field `CostModifier`, `CastRestriction` and
	// `ActivationRestriction` carry and evaluated in the same
	// accessor, so a gated-off statement never reaches the read and
	// the engine, the enumerator and the view cannot disagree about
	// whether it exists (ADR 0071).
	ActiveWhen Designation
}

// Grants reports whether this statement OPENS the instant-speed
// window. The mirror of CastTimingRule.Grants.
func (t ActivationTiming) Grants() bool { return t.Timing == TimingFlash }

// Restricts reports whether this statement NARROWS the window
// (CR 101.2's "can't" half). The mirror of CastTimingRule.Restricts.
func (t ActivationTiming) Restricts() bool {
	return t.Timing == TimingSorcery || t.Timing == TimingYourTurnOnly
}

// CatalogActivationTimings is the catalog hook the effects package
// wires at init, mirroring CatalogActivationRestrictions and
// CatalogCastTimings. Nil, or a nil return, means the card says
// nothing about anybody's activation timing — which is nearly every
// card.
var CatalogActivationTimings func(oracleID string) []ActivationTiming

// ActivationTimingsForCard returns the timing statements a permanent
// makes right now: none for a permanent under a CR 613.1f
// ability-removing effect (CatalogAbilityKey returns the empty key),
// and none for one whose designation gate is unsatisfied.
//
// Field-for-field ActivationRestrictionsForCard, on purpose.
func ActivationTimingsForCard(c Card) []ActivationTiming {
	if CatalogActivationTimings == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogActivationTimings(key), func(t ActivationTiming) Designation {
		return t.ActiveWhen
	})
}

// activationTimingVerdict is what every statement about one
// activation folds down to: three independent bits, and no record of
// which permanent said what. Built and consumed inside one call, so
// it allocates nothing. The mirror of castTimingVerdict.
type activationTimingVerdict struct {
	// Flash is set by any live TimingFlash statement covering this
	// activation.
	Flash bool

	// SorceryOnly is set by any live TimingSorcery statement.
	SorceryOnly bool

	// YourTurnOnly is set by any live TimingYourTurnOnly statement.
	YourTurnOnly bool
}

// activationTimingVerdictLocked folds every statement on the
// battlefield that speaks about this activation.
//
// THE ORDER OF THE WALK DOES NOT MATTER, deliberately: the verdict is
// three independent bits and CR 101.2's precedence is applied by the
// reader below, not by the walk. A precedence that depended on which
// permanent the battlefield scan reached first would be a rule nobody
// could state — the same sentence castTimingVerdictLocked carries.
//
// `q` arrives with everything but Source filled in; the walk stamps
// Source per candidate, exactly as ActivationGateLocked does.
//
// Caller must hold g.mu (read or write).
func (g *Game) activationTimingVerdictLocked(q ActivationQuery) activationTimingVerdict {
	var v activationTimingVerdict
	if g == nil || g.Battlefield == nil || CatalogActivationTimings == nil {
		return v
	}
	for i := range g.Battlefield.Cards {
		src := g.Battlefield.Cards[i]
		for _, t := range ActivationTimingsForCard(src) {
			// A statement that says nothing, or covers nothing, is
			// skipped before the predicate runs — the cheapest way
			// to not be this walk's business is to say so first.
			if t.Covers == nil || (!t.Grants() && !t.Restricts()) {
				continue
			}
			q.Source = src
			if !t.Covers(q) {
				continue
			}
			switch t.Timing {
			case TimingFlash:
				v.Flash = true
			case TimingSorcery:
				v.SorceryOnly = true
			case TimingYourTurnOnly:
				v.YourTurnOnly = true
			}
		}
	}
	return v
}

// ActivationTimingOpenLocked is CR 602.5's timing question asked
// once: may `activator` BEGIN to activate this ability of `card`,
// sitting in `zone`, right now?
//
// THREE CALLERS AND A FOURTH, ONE FUNCTION, which is the whole point
// — the same arrangement ADR 0073 §7 makes for the cast gate and
// #1195 makes for the cast timing read:
//
//   - ActivateCatalogAbility, where it replaced the inline
//     `(ab.SorcerySpeed || ab.Cost.Loyalty != nil) &&
//     !g.SorcerySpeedOpenLocked(playerID)`;
//   - legal.abilityMovesForSource, which dropped the `speed`
//     parameter it threaded through two functions to keep a copy of
//     the same three lines, so a bot is never offered an activation
//     the engine will refuse and never denied one a Leonin Shikari
//     opens (#544);
//   - protocol.viewOfActivatedAbilities, where it becomes
//     `timing_closed`, so the client greys an ability row from the
//     engine's answer rather than from a rule it reimplemented —
//     which is the last rules derivation `client/src/lib/timing.ts`
//     carried for catalogued abilities;
//   - ActivateLoyalty (mutations.go), the SANDBOX manual loyalty
//     verb. It is not a catalogued ability and `internal/legal`
//     skips it by design, but it is still CR 606.3's window, and a
//     statement about "loyalty abilities of planeswalkers you
//     control" reaches a planeswalker the catalog has never heard
//     of. The same argument `gatherTapSources` is the activation
//     gate's fourth caller under.
//
// FOUR STEPS, AND THE ORDER IS THE RULES:
//
//  1. MANA ABILITIES ARE NOT ASKED ABOUT (CR 605.3a). A mana ability
//     may be activated whenever its controller has priority AND
//     whenever a payment is being made — during the announcement of
//     a spell, inside another ability's cost, in the middle of a
//     resolution. That window is not the CR 602.5d window this
//     function narrows or opens, so a timing statement has nothing
//     to say about one and this returns true before it walks
//     anything.
//     It is NOT the carve-out a "restriction that swallowed mana
//     abilities would lock a player out of paying for anything"
//     needs — that card exists (Grand Abolisher stops mana abilities
//     too, with no "unless they're mana abilities" clause) and it is
//     a BAN, so it goes through ActivationGateLocked, where #1210
//     put the mana decision on the CARD because half the cards print
//     the exemption and half do not. Here there is no decision to
//     put anywhere: the rules give a mana ability its own window.
//
//  2. THE ABILITY'S OWN TIMING. "Activate only as a sorcery"
//     (CR 602.5d, `ActivationAbility.SorcerySpeed`) and a loyalty
//     ability (CR 606.3, `ActivationAbility.Loyalty`) answer to the
//     sorcery window; every other activated ability is instant-speed
//     on its own (CR 602.5a).
//
//  3. THE PER-PLAYER GRANTS. The Wandering Emperor, Teferi, Master
//     of Time, Leonin Shikari. CR 101.1: a card beats a rule, which
//     is how a printed clause reaches CR 606.3's sorcery half at
//     all.
//
//  4. THE PER-PLAYER RESTRICTIONS, LAST, because CR 101.2 says
//     "can't" beats "can". Nothing declares them today; the
//     placement is the rule, stated once, rather than a rule that
//     would have to be discovered the day a card does.
//
// Then a shut window means `SorcerySpeedOpenLocked` must be open,
// which is CR 307.1's main-phase / empty-stack / active-player test
// unchanged.
//
// It sits BESIDE ActivationGateLocked rather than inside it, for the
// reason ADR 0073's #1195 note gives about the cast pair: "banned"
// and "not yet" are different answers to the player, and
// `cant_activate` means the first.
//
// Caller must hold g.mu (read or write).
func (g *Game) ActivationTimingOpenLocked(activator uuid.UUID, card Card, zone ZoneKind, ability ActivationAbility) bool {
	// 1. CR 605.3a — not this window's business.
	if ability.Mana {
		return true
	}
	// 2. The ability's own timing.
	instantSpeed := !ability.SorcerySpeed && !ability.Loyalty
	v := g.activationTimingVerdictLocked(ActivationQuery{
		Game:       g,
		Card:       card,
		Controller: activator,
		FromZone:   zone,
		Ability:    ability,
	})
	// 3. The grants.
	if v.Flash {
		instantSpeed = true
	}
	// 4. The restrictions. "Only during your own turn" refuses
	// outright rather than narrowing to the sorcery window, for the
	// reason CastTimingOpenLocked gives: it is a gate on WHOSE turn
	// it is and nothing else.
	if v.YourTurnOnly && g.activeSeatIDLocked() != activator {
		return false
	}
	if v.SorceryOnly {
		instantSpeed = false
	}
	if instantSpeed {
		return true
	}
	return g.SorcerySpeedOpenLocked(activator)
}
