package game

// infect_wither_toxic.go is the READ side of ADR 0056: the grammar of
// the three damage-result keywords, and the arithmetic of what damage
// from a source that has them DOES to its target (CR 120.3). It lands
// ahead of the engine wiring that will call it.
//
// WHY IT ARRIVES ON ITS OWN, AND WHAT IT DELIBERATELY DOES NOT DO.
// ADR 0056's PR 2 puts `infect`, `wither` and `toxicTotal` on the
// damage tail (damage_tail.go) and branches on them in two functions.
// Two seams change hands there that this change cannot reach:
//
//   - `DamageAssignmentFrame` (pending_choice.go) has to cache the
//     same three facts, because the CR 510.1c assignment prompt is
//     answered after the attacker may have died and
//     `damageTailFromFrame` can only copy what the frame holds.
//   - The counters are a RESULT, not a write: ADR 0056 Decisions 3
//     to 5 put them through the CR 614 counter window, which needs
//     `ReplacementEvent.CounterPlayer` / `CounterPlacer` /
//     `CounterFromCombatDamage` (replacements.go),
//     `EventPlayerCounterPlaced` (events.go) and the placer on
//     `applyCounterLocked` (mutations.go) — ADR 0056's PR 1.
//
// Branching without those would give poison that no replacement can
// see (Vorinclex, Solemnity, Melira, and the Vizier of Remedies
// ruling ADR 0056 quotes), counters credited to nobody, and combat
// damage that silently loses its result the moment a multi-blocker
// assignment prompt pauses. So the two branches are NOT wired here
// and the tail is unchanged; what is here is every part of the
// decision that carries no such dependency, pinned by tests so the
// wiring PR has one thing left to do and nothing left to decide.
//
// THE TOKENS ARE NOT IN canonicalKeywords YET, ON PURPOSE. That table
// is closed, and "a keyword joins this table in the same change that
// teaches the engine to honour it" (keywords.go). The deck importer
// stamps what the table admits and the ADR 0037 coverage signal reads
// the same list through CanonicalKeywords, so admitting `infect`
// today would mark the 47 cards that print it — and the 30
// keyword-only creatures ADR 0056 lists — as fully playable while
// damage from them still just marked damage. That is a badge
// promising a rule the engine does not keep, the half-a-card failure
// ADR 0037 §5 forbids, and it errs STRONGER than printed. The three
// tokens join the table in the change that branches on them.

import (
	"strconv"
	"strings"
)

// KeywordInfect and KeywordWither are the canonical tokens for
// CR 702.90 and CR 702.80. Named constants for the same reason
// KeywordChangeling is: they are read from the damage tail, the
// importer, the coverage scan and the client's icon map, and a typo
// in any one of them would silently turn an infect creature back
// into a vanilla one.
//
// Neither is in canonicalKeywords yet — see the file comment.
const (
	KeywordInfect = "infect"
	KeywordWither = "wither"
)

// keywordToxic is the bare word CR 702.164 spells with a number after
// it. It is NOT a token on its own: a bare "toxic" names no amount,
// and Scryfall's `keywords` array carries exactly that bare word with
// the N living only in the oracle line. ToxicValue is the only thing
// that reads this constant, and it refuses the bare word for the same
// reason CanonicalKeywords refuses a bare "protection".
const keywordToxic = "toxic"

// maxToxicValue bounds what ToxicValue will accept. Printed toxic
// tops out at 4 (Tyrranax Rex), so anything past three digits is a
// malformed oracle line rather than a card, and refusing it leaves
// the card flagged unimplemented — weaker than printed, never
// stronger, which is the direction every parse in this engine errs.
const maxToxicValue = 999

// ToxicValue parses one ability token as CR 702.164's numbered
// keyword, reporting the N it carries.
//
// The grammar is deliberately tight: the word, exactly one space, and
// a positive decimal integer, with nothing after it. "toxic" alone,
// "toxic 0", "toxic  2", "toxic two" and "toxic 2 1" all answer
// false. Leading zeros are accepted and normalised away, because
// Scryfall's oracle text is the source and it has printed odd
// spacing before; a second space is not the same keyword and guessing
// which one was meant is how a badge ends up lying.
//
// Case and surrounding whitespace are normalised, the same lowercase
// plus trim CanonicalKeywords applies, so a raw oracle line can be
// handed straight in.
func ToxicValue(token string) (int, bool) {
	rest, ok := strings.CutPrefix(strings.ToLower(strings.TrimSpace(token)), keywordToxic)
	if !ok {
		return 0, false
	}
	digits, ok := strings.CutPrefix(rest, " ")
	if !ok || digits == "" {
		return 0, false
	}
	n := 0
	for i := 0; i < len(digits); i++ {
		d := digits[i]
		if d < '0' || d > '9' {
			return 0, false
		}
		n = n*10 + int(d-'0')
		if n > maxToxicValue {
			return 0, false
		}
	}
	if n <= 0 {
		return 0, false
	}
	return n, true
}

// CanonicalToxicToken normalises one printed toxic clause to the
// engine's wire form — "Toxic 01" becomes "toxic 1" — reporting
// whether it is one at all.
//
// The sibling of CanonicalKeyword for the one keyword in ADR 0056
// that carries a parameter, and, like ProtectionTokens, the only
// thing that mints the token. CanonicalKeywords will call it in the
// change that admits toxic to the table; until then nothing in the
// importer path reaches it, which is what keeps the token out of
// Characteristic.Abilities.
func CanonicalToxicToken(s string) (string, bool) {
	n, ok := ToxicValue(s)
	if !ok {
		return "", false
	}
	return keywordToxic + " " + strconv.Itoa(n), true
}

// ToxicTotal is the card's total toxic value: the sum of the N of
// EVERY `toxic N` ability it has, duplicates included (CR 702.164b —
// "a creature's total toxic value is the sum of all N values of toxic
// abilities that creature has").
//
// This is the only way toxic is read. HasKeyword(c, "toxic") is not,
// and cannot be: the token carries its amount, and a Rat that prints
// toxic 1 under Karumonix's "other Rats you control have toxic 1" has
// a total of 2, not "toxic: yes".
//
// It walks the same ability list in the same zone order HasKeyword
// documents (forEachAbilityToken), so a granted toxic and a printed
// one answer alike on the battlefield, and a card in hand answers off
// its own printed keywords. nil card returns 0.
func ToxicTotal(c *Card) int {
	total := 0
	forEachAbilityToken(c, func(a string) bool {
		if n, ok := ToxicValue(a); ok {
			total += n
		}
		return true
	})
	return total
}

// AppendKeywordAbility appends a keyword token to an ability list
// under the rule the keyword itself has: a CUMULATIVE keyword is
// appended every time, a redundant one only once.
//
// Every keyword in this engine before toxic is redundant when
// repeated (CR 702.80d, 702.90f, 702.15f — a second instance of
// wither, infect or lifelink does nothing), which is why the catalog's
// grant helpers all dedupe on apply and why nothing noticed. Toxic is
// the first keyword where the second instance is the whole point, so
// the dedupe has to become a decision instead of a habit: ADR 0056
// Decision 1 gives it one home, and the grant sites move onto it in
// the change that admits the token.
//
// An empty token is dropped: an ability list with "" in it renders as
// a blank badge.
func AppendKeywordAbility(abilities []string, kw string) []string {
	if kw == "" {
		return abilities
	}
	if !keywordIsCumulative(kw) {
		for _, a := range abilities {
			if a == kw {
				return abilities
			}
		}
	}
	return append(abilities, kw)
}

// keywordIsCumulative reports whether repeating this token means
// something. Toxic (CR 702.164b) is the only one today, and it is
// recognised by its grammar rather than by a second table, so a
// `toxic 2` that nothing has taught this file about still counts
// twice when it is granted twice.
func keywordIsCumulative(kw string) bool {
	_, ok := ToxicValue(kw)
	return ok
}

// DamageResultSource is the part of a damage source's snapshot that
// decides WHICH RESULT its damage has (CR 120.3), as opposed to how
// much damage there is (the CR 614 window settles that first) or what
// rides along with it (deathtouch, lifelink, the commander tally).
//
// It is a struct rather than three arguments because all three are
// read off the source at the same moment, for the same reason: the
// damage tail is a snapshot taken when the event is CREATED, so that
// a CR 616 prompt answered after the source has died still deals the
// damage the source had. ADR 0056 Decision 2 puts these three fields
// on `damageTail` and fills them from one reader alongside deathtouch
// and lifelink; this type is the shape that reader returns.
type DamageResultSource struct {
	// Infect is CR 702.90: damage to a player is poison counters
	// instead of life loss, damage to a creature is -1/-1 counters
	// instead of marked damage.
	Infect bool

	// Wither is CR 702.80: damage to a CREATURE is -1/-1 counters
	// instead of marked damage. It says nothing about players, which
	// is why wither damage to a player is ordinary life loss.
	Wither bool

	// ToxicTotal is the source's CR 702.164b total toxic value, summed
	// over every instance. Read by the player branch only, and only
	// for combat damage.
	ToxicTotal int
}

// SourceDamageResultTraits reads the three keywords off a damage
// source card.
//
// It takes a card rather than an ID on purpose: ADR 0056 Decision 2
// makes finding the source — battlefield, then the stack, then
// last-known information — ONE reader shared with deathtouch and
// lifelink, and that reader belongs next to them in the tail. This
// function is what it calls once it has the card, so the "which
// keywords" half cannot drift from the "which card" half by living
// in a different file from either.
//
// nil card carries none of the three. That is what "a source with
// infect" means for a spell that has left the stack or a source the
// engine can no longer see, and it is the same answer deathtouch and
// lifelink give today.
func SourceDamageResultTraits(c *Card) DamageResultSource {
	if c == nil {
		return DamageResultSource{}
	}
	return DamageResultSource{
		Infect:     HasKeyword(c, KeywordInfect),
		Wither:     HasKeyword(c, KeywordWither),
		ToxicTotal: ToxicTotal(c),
	}
}

// Any reports whether this source changes the result of its damage at
// all. The fast path for the branches: a board with no infect, wither
// or toxic on it answers one bool and lands exactly as it does today.
func (s DamageResultSource) Any() bool {
	return s.Infect || s.Wither || s.ToxicTotal > 0
}

// DamageToCreature splits settled damage to a creature into the two
// results CR 120.3d and CR 120.3e allow: -1/-1 counters when the
// source has infect or wither, marked damage otherwise. Never both —
// they are alternatives, not additions, unlike the CR 120.3c loyalty
// clause, which is additive and is not this function's business.
//
// `amount` is the POST-replacement number. A non-positive amount is
// fully prevented damage and has no result at all: no counters, no
// mark (ADR 0056, and the Grafted Exoskeleton ruling — prevented
// damage gives no counters).
//
// Infect and wither together are still one set of counters. CR 702.90
// and CR 702.80 name the same result, and a source with both is not
// dealt with twice.
//
// What this function does NOT answer, because the rules put it
// elsewhere: deathtouch (CR 702.2b is about damage DEALT, so the
// lethal flag is set whether or not the counters are replaced away),
// the planeswalker and battle clauses, and how many counters
// actually land once the counter window has had them — which is the
// Vizier of Remedies ruling, and why the caller places them through
// the window rather than writing them.
func (s DamageResultSource) DamageToCreature(amount int) (minusOneCounters, markedDamage int) {
	if amount <= 0 {
		return 0, 0
	}
	if s.Infect || s.Wither {
		return amount, 0
	}
	return 0, amount
}

// DamageToPlayer splits settled damage to a player into poison
// counters and life loss (CR 120.3a, 120.3b, 120.3g).
//
// Infect replaces the life loss with that much poison. Toxic ADDS its
// total on top, and only for COMBAT damage — the toxic rulings are
// explicit that it "doesn't change the amount of combat damage" and
// does nothing on noncombat damage or on damage to a creature or
// planeswalker. Wither is absent here by design: CR 702.80 is about
// creatures, so a wither source hitting a player just loses them life.
//
// The poison is returned as ONE number, not as an infect number and a
// toxic number, because CR 120.4c processes a damage event into its
// results in one step: under a halving replacement a single event of
// 3 gives 1, where two events of 2 and 1 would give 1 + 0. ADR 0056
// Decision 4 rejects the two-placement shape for exactly that reason.
//
// The toxic total is not scaled by the amount: a damage doubler
// changes the life loss and leaves the poison alone, which is what
// the Pestilent Syphoner ruling says. A non-positive amount is
// prevented damage and gives neither result, so a Fog stops the
// poison too.
//
// The caller still owes the damage event everything that is about
// damage rather than about its result: the CR 903.10a commander
// tally accrues on infect combat damage like any other, lifelink
// credits the damage amount, and EventDealDamage fires either way.
// Only the life change is replaced here, which is why infect damage
// fires no EventChangeLife and no "whenever a player loses life".
func (s DamageResultSource) DamageToPlayer(amount int, combat bool) (poison, lifeLoss int) {
	if amount <= 0 {
		return 0, 0
	}
	if s.Infect {
		poison = amount
	} else {
		lifeLoss = amount
	}
	if combat && s.ToxicTotal > 0 {
		poison += s.ToxicTotal
	}
	return poison, lifeLoss
}
