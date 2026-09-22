package game

import "github.com/google/uuid"

// player_statics.go — CR 702.16i and CR 702.11d: the abilities a
// PLAYER has, and the one place they are read.
//
// ADR 0072 §10 listed this as out of scope with one reason: "Player
// carries no ability slice, so 'you have protection from everything'
// has nowhere to live and no consumer". The ADR's 2026-09-22
// amendment (#1197) gives it both, and this file is the both.
//
// NOTHING IN THE GRAMMAR CHANGES. A player's protection is the same
// token a permanent's is, parsed by the same closed-grammar reader in
// protection.go, matched against the same source snapshot. What is new
// is a second kind of thing that can HAVE a quality. That is why
// PlayerStatic.Keyword is a string in the engine's existing vocabulary
// rather than a player-side enum: "protection from everything" means
// one thing in this engine and exactly one file parses it.
//
// TWO SOURCES, AND ONLY ONE OF THEM IS STORED:
//
//   - DERIVED — a permanent on the battlefield whose printed static is
//     "You have hexproof" (Leyline of Sanctity, Aegis of the Gods).
//     Declared as effects.Spec.PlayerKeywords, read through
//     CatalogPlayerKeywords on every query and written nowhere. Two
//     Leylines therefore compose, and one of them leaving cannot
//     revoke the other's grant — the argument CatalogNoMaxHandSize
//     (#338) and land_drops.go both make at length.
//
//   - GRANTED — a resolved spell or triggered ability, "you gain
//     protection from everything until your next turn" (Teferi's
//     Protection, The One Ring). This one HAS to be stored: the source
//     is in a graveyard a moment after it resolves, which is the same
//     reason ScopedStatic exists. It carries a CR 611.2 Duration and
//     is swept through durationExpiredLocked, the one function ADR
//     0063 says decides when any continuous effect in this game ends.
//
// WHY NOT A ScopedStatic. The registry next door looks like the
// obvious home and is the wrong one: a ScopedStatic is adapted into a
// ContinuousEffect and applied by the CR 613 layer pass, whose Apply
// signature is (*Characteristic, *Card) — a characteristic of an
// OBJECT. A player has no Characteristic and no layer, exactly as
// Spec.NoMaxHandSize and Spec.CostModifiers already argue for
// themselves. So this is a third thing beside the two registries, and
// a small one: a token, an attribution and a duration.
//
// THREE CONSUMERS, one predicate each, all of them a branch that
// already existed and returned early:
//
//	targeting   canPlayerBeTargetedByLocked   targets.go               CR 702.11d / 702.16i
//	damage      protectionPreventsDamageLocked builtin_replacements.go CR 702.16e
//	attachment  attachmentLegalLocked          attach.go               CR 702.16c
//
// The bot's move enumerator and the client's legal_targets follow by
// construction: both read legalTargetsLocked, which is the targeting
// consumer. Neither learns the rule, so neither can disagree with it.

// KeywordHexproof is CR 702.11's token. Named here because it is now
// read from two packages and spelled in card files, and a typo in a
// hand-written token grants nothing at all — the same argument
// ProtectionFromChosenPlayer carries in protection.go.
const KeywordHexproof = "hexproof"

// PlayerStatic is ONE ability a player has, with the duration it has
// it for. The player-level twin of ScopedStatic, and deliberately
// much less: no closure, no layer, no timestamp.
//
// IMMUTABILITY CONTRACT, the same one ScopedStatic and CastPermission
// carry: every field is written once at registration and never
// mutated. Clone copies the slice into a fresh backing array and the
// sweep replaces the slice rather than compacting it, so an undo
// snapshot taken mid-turn still holds the grants that were live when
// it was taken.
//
// PLAIN DATA by construction, which is what lets it be mirrored into
// the snapshot rather than rebuilt, and what keeps it out of the
// restorability census: there is no closure here for a restore to
// fail to bring back.
type PlayerStatic struct {
	// Keyword is an engine ability token, in the same vocabulary a
	// Card's abilities use: "hexproof", or a protection token that
	// ParseProtectionQuality accepts ("protection from everything").
	//
	// A token this engine cannot parse grants nothing at all, which
	// is the direction protection.go's closed grammar already errs
	// in. Build protection tokens with the constants and constructors
	// in protection.go, never by hand.
	//
	// EMPTY on a timing statement (see Timing below), which is what
	// keeps one out of playerAbilityTokensLocked's answer.
	Keyword string `json:"keyword"`

	// Timing is a granted cast-TIMING statement — "you may cast
	// spells this turn as though they had flash" (Emergence Zone),
	// "until your next turn, you may cast sorcery spells as though
	// they had flash" (Teferi, Time Raveler's +1). CR 307.1 and
	// CR 702.8; #1195, cast_timing.go.
	//
	// The zero value (`Timing.Timing == TimingNormal`) says nothing
	// and is what every ability grant carries, so the two kinds of
	// entry are told apart by the payload rather than by a
	// discriminator field.
	//
	// It is HERE rather than in a registry of its own because a
	// granted timing statement and a granted "you have hexproof" are
	// the same kind of thing — an ability a PLAYER has for a CR 611.2
	// duration — and a second player-level slice would have meant two
	// sweeps, two clones, two snapshot fields and two readings of one
	// Duration. #1195 shipped its own for a day and folded it onto
	// this one. Its READER is castTimingVerdictLocked, not
	// playerAbilityTokensLocked: a timing statement is not a token
	// and has no protection quality to parse.
	//
	// Plain data, like everything else here — CastTimingRule is
	// flags, two strings and a zone.
	Timing CastTimingRule `json:"timing,omitzero"`

	// Source is the card that granted it, for the log and for the
	// view's attribution. Never read by any rule: a granted ability
	// outlives its source, which is the whole reason it is stored
	// here rather than derived from the battlefield.
	Source uuid.UUID `json:"source,omitempty"`

	// Label is human-readable attribution ("Teferi's Protection —
	// protection from everything"). Not a rules input.
	Label string `json:"label,omitempty"`

	// Duration is how long the player has it (CR 611.2), read by the
	// same durationExpiredLocked every other continuous effect in the
	// game is ended by. The zero value is "until end of turn".
	Duration Duration `json:"duration,omitempty"`
}

// CatalogPlayerKeywords returns the player-level abilities a
// battlefield permanent with this catalog key grants its CONTROLLER —
// []string{"hexproof"} for Leyline of Sanctity and Aegis of the Gods,
// nil for everything else. Populated at init time by the cards/effects
// package from effects.Spec.PlayerKeywords. A nil hook (no catalog
// wired) means no player has a derived ability.
//
// Derived rather than written, for the reason CatalogNoMaxHandSize
// spells out: a "set on enter, restore on leave" design has to answer
// "restore to what?", and gets two real cases wrong — two Leylines,
// where the first to leave would strip a hexproof the second is still
// granting, and a player whose abilities were changed by something
// else in between. Derivation has no stored value to strand, so
// neither case exists, and undo needs no new state to clone.
var CatalogPlayerKeywords func(oracleID string) []string

// GrantPlayerStaticForEffect gives `player` an ability for a duration
// — "you gain protection from everything until your next turn". The
// *ForEffect surface: caller must hold g.mu (write), which a
// resolution frame already does.
//
// `keyword` is an engine token; an empty one is a no-op rather than an
// error, so a card file that computed its token and got nothing grants
// nothing rather than granting "".
//
// Build the duration with the constructors in duration.go —
// g.UntilYourNextTurnDuration(p) for both of this seam's cards. The
// zero value is "until end of turn", so a caller that forgets gets the
// shortest answer rather than a permanent one.
func (g *Game) GrantPlayerStaticForEffect(player uuid.UUID, keyword, label string, source uuid.UUID, d Duration) {
	if keyword == "" {
		return
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return
	}
	p.Statics = append(p.Statics, PlayerStatic{
		Keyword:  keyword,
		Source:   source,
		Label:    label,
		Duration: d,
	})
}

// playerAbilityTokensLocked is THE reader: every ability `p` has right
// now, derived grants first and stored ones after, in the order a
// reader would expect to see them explained.
//
// The player-side twin of forEachAbilityToken (keywords.go), and like
// it a walk with an early-out rather than a slice, because every
// consumer is asking a yes/no question and the common answer is "this
// player has nothing at all".
//
// THE DURATION IS TESTED HERE as well as in the sweep. The sweep is
// hygiene, run at known moments (cleanup, the beginning of a turn);
// the reader is the truth and has to be right between them — a
// protection that ends as your next turn begins must not still be
// answering during the priority round that ends the previous turn.
// That is the posture CastPermissionActiveForEffect takes for exactly
// the same reason.
//
// Caller must hold g.mu (read or write).
func (g *Game) playerAbilityTokensLocked(p *Player, fn func(token string) bool) {
	if p == nil {
		return
	}
	if CatalogPlayerKeywords != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller != p.ID || c.OracleID == "" {
				continue
			}
			// CatalogAbilityKey, not CatalogKey: "you have hexproof"
			// is a static ability of the permanent, so a Leyline that
			// has lost its abilities (layer 6) stops granting it.
			for _, tok := range CatalogPlayerKeywords(CatalogAbilityKey(*c)) {
				if !fn(tok) {
					return
				}
			}
		}
	}
	for _, s := range p.Statics {
		// #1195: an entry carrying a cast-timing statement is not an
		// ability token and has no keyword to offer. Skipped before
		// the duration is read, because the cheapest way to not be
		// this walk's business is to say so first.
		if s.Keyword == "" {
			continue
		}
		if g.durationExpiredLocked(s.Duration, false) {
			continue
		}
		if !fn(s.Keyword) {
			return
		}
	}
}

// PlayerAbilitiesForEffect is the *ForEffect surface over the reader:
// every ability this player has right now, as tokens, in the order
// above. For the protocol projection's badge and for card files that
// want to ask.
//
// Caller must hold g.mu (read or write).
func (g *Game) PlayerAbilitiesForEffect(p *Player) []string {
	var out []string
	g.playerAbilityTokensLocked(p, func(tok string) bool {
		out = append(out, tok)
		return true
	})
	return out
}

// PlayerHasKeywordLocked reports whether the player has a bare ability
// token — "hexproof". Not for protection: a protection token carries a
// parameter and is asked with PlayerProtectedFromLocked.
//
// Caller must hold g.mu (read or write).
func (g *Game) PlayerHasKeywordLocked(p *Player, keyword string) bool {
	found := false
	g.playerAbilityTokensLocked(p, func(tok string) bool {
		if tok == keyword {
			found = true
			return false
		}
		return true
	})
	return found
}

// PlayerProtectedFromLocked is the player-side MatchedProtection: does
// this player have protection from a source with these
// characteristics, and if so WHICH quality refused (CR 702.16i)?
//
// The quality is returned for the same reason the object side returns
// it: the player-facing sentence names it ("you have protection from
// everything"), and naming it is the whole reason the token keeps the
// card's own spelling.
//
// A nil src is an unknown source and matches nothing — except
// "protection from everything", which CR 702.16j says names no
// characteristic to look up. Identical to the object side, on purpose.
//
// Caller must hold g.mu (read or write).
func (g *Game) PlayerProtectedFromLocked(p *Player, src *Characteristic) (ProtectionQuality, bool) {
	var (
		out   ProtectionQuality
		found bool
	)
	g.playerAbilityTokensLocked(p, func(tok string) bool {
		q, ok := ParseProtectionQuality(tok)
		if !ok {
			return true
		}
		// No bindProtectionQuality: CR 702.16k's "the chosen player"
		// quality lives on a PERMANENT's as-enters choice (#980) and a
		// player has no such field. A player-quality token granted to
		// a player therefore protects from nobody, which is the weaker
		// half of that gap and the direction this engine errs in. No
		// catalogued card prints it — see ADR 0072's amendment §A4.
		if q.Matches(src) {
			out, found = q, true
			return false
		}
		return true
	})
	return out, found
}

// canPlayerBeTargetedByLocked is the targeting choke point's player
// half: may the spell or ability `src` describes choose this player as
// a target?
//
// The object-side twin is CanBeTargetedBy (keywords.go), whose doc
// comment used to end "players are not covered… TargetPlayer refs pass
// this gate by not reaching it". They reach it now.
//
//   - hexproof (CR 702.11d) — can't be the target of spells or
//     abilities your OPPONENTS control. So the test is the
//     CONTROLLER's, and a hexproof player may still target THEMSELVES.
//     That half is load-bearing rather than a nicety: Leyline of
//     Sanctity must not stop you casting your own Sylvan Library, and
//     a player who could not target themselves could not pay a cost or
//     aim their own drain.
//   - protection (CR 702.16i) — can't be the target of spells with the
//     stated quality, or of abilities from a SOURCE with it. Tested
//     against the source OBJECT, never its controller, exactly as
//     CR 702.16b is for a permanent.
//
// Shroud is absent on purpose: no card prints shroud on a player, and
// a token with no card behind it is what CanonicalKeywords refuses a
// bare "protection" for.
//
// A nil player returns true — a caller that has lost the seat has an
// existence problem, which the walk that wraps this reports
// separately.
//
// ONE WALK, and the source snapshot built at most once inside it.
// This is the hottest reader of the three: the targeting enumeration
// asks it once per seat, and the view asks the enumeration once per
// card a viewer could cast. Nearly every seat in nearly every game
// has no abilities at all, so the walk ends immediately and
// src.Characteristics() — which copies a Characteristic — is never
// reached.
//
// Caller must hold g.mu (read or write).
func (g *Game) canPlayerBeTargetedByLocked(p *Player, src TargetSource) bool {
	if p == nil {
		return true
	}
	var (
		chars     *Characteristic
		charsRead bool
		allowed   = true
	)
	g.playerAbilityTokensLocked(p, func(tok string) bool {
		if tok == KeywordHexproof {
			// CR 702.11d asks WHO, and only who.
			if p.ID != src.Controller {
				allowed = false
				return false
			}
			return true
		}
		q, ok := ParseProtectionQuality(tok)
		if !ok {
			return true
		}
		if !charsRead {
			chars, charsRead = src.Characteristics(), true
		}
		// No bindProtectionQuality here for the same reason
		// PlayerProtectedFromLocked has none — see its comment.
		if q.Matches(chars) {
			allowed = false
			return false
		}
		return true
	})
	return allowed
}

// CanPlayerBeTargetedByForEffect is the *ForEffect surface over the
// targeting predicate, for catalog HasLegalTarget clauses and the
// protocol projection. Caller must hold g.mu.
func (g *Game) CanPlayerBeTargetedByForEffect(playerID uuid.UUID, src TargetSource) bool {
	return g.canPlayerBeTargetedByLocked(g.playerByIDLocked(playerID), src)
}

// sweepPlayerStaticsLocked drops every stored player ability whose
// duration has run out. `endOfTurn` is handed straight to
// durationExpiredLocked, which is the only code that decides what a
// duration means.
//
// Runs beside sweepScopedStaticsLocked at all three of its moments —
// the cleanup step (CR 514.2), the beginning of a turn (the "until
// your next turn" boundary) and the top of a layer recompute — because
// it is the same sweep and a second schedule would be a second thing
// to keep in step.
//
// Allocates a fresh slice rather than compacting in place with s[:0]:
// the backing array is shared with every undo snapshot Clone has
// taken, so an in-place compaction would rewrite history. Same trap
// sweepScopedStaticsLocked exists to avoid.
//
// Idempotent, which matters because CR 514.3a can give a turn a second
// cleanup step and because the layer recompute runs it every pass.
//
// Caller must hold g.mu (write).
func (g *Game) sweepPlayerStaticsLocked(endOfTurn bool) {
	for _, p := range g.Seats {
		if p == nil || len(p.Statics) == 0 {
			continue
		}
		kept := make([]PlayerStatic, 0, len(p.Statics))
		for _, s := range p.Statics {
			if !g.durationExpiredLocked(s.Duration, endOfTurn) {
				kept = append(kept, s)
			}
		}
		if len(kept) == len(p.Statics) {
			continue
		}
		if len(kept) == 0 {
			kept = nil
		}
		p.Statics = kept
	}
}

// clonePlayerStatics copies a player's ability slice into a fresh
// backing array. The entries are plain data written once at
// registration (see the immutability contract on PlayerStatic), so a
// value copy per entry is all the isolation an undo snapshot needs;
// what must not be shared is the ARRAY, because the sweep replaces the
// slice.
//
// Mirrors cloneCastPermissions, the field this one is modelled on.
func clonePlayerStatics(in []PlayerStatic) []PlayerStatic {
	if len(in) == 0 {
		return nil
	}
	out := make([]PlayerStatic, len(in))
	copy(out, in)
	return out
}
