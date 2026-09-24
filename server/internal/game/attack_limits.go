package game

import "github.com/google/uuid"

// attack_limits.go is CR 508.1c's COUNT restrictions on an attack
// declaration: "no more than one creature can attack each combat"
// (Silent Arbiter, Dueling Grounds), "no more than two creatures can
// attack you each combat" (Crawlspace). #1507, ADR 0045 amendment of
// 2026-09-24, Decisions 43-45.
//
// # Why this is not a Restriction bit, an AttackTax or a BlockRule
//
//   - A Restriction bit is about ONE permanent ("this creature can't
//     attack"). A count is about the whole declaration: no creature is
//     forbidden, some SET of them is.
//   - An AttackTax (ADR 0080) is a price; a limit is a prohibition. A
//     tax can always be paid by attacking with fewer creatures, and so
//     can a limit, but the tax charges and the limit refuses.
//   - A BlockRule is handed a blocker and an attacker. An attack has
//     neither a blocker nor, for the "you" family, a creature to scope
//     by — the scope is the DEFENDING SEAT, the axis AttackTax already
//     keys on. So the attack side is its own small struct, and the
//     "you" in it is structural for AttackTax's reason: a limit
//     protects its source's controller and nobody else, and no card
//     file can typo its way into limiting attacks on the whole table.
//
// # Where it is judged
//
// On the declaration, by both declaration verbs, before anything is
// staged and before the CR 508.1a tax is priced — a declaration the
// limit refuses owes nothing:
//
//   - DeclareAttackerWith, the per-creature verb the client clicks and
//     the legal-move enumerator emits. It is deliberately lax about
//     eligibility for the sandbox's hand-forcing, and a limit is not
//     eligibility, for the reason "can't attack" is not: an effect
//     that says only one creature can attack is the card doing its
//     whole job, and a verb that shrugged it off would make Silent
//     Arbiter a blank.
//   - DeclareAttackersWith, the bulk verb. It skips INELIGIBLE entries
//     silently, but a limit is not an ineligible entry — every
//     creature in an over-full swing is individually fine, and which
//     of them to drop is the attacking player's choice. So the limit
//     is ALL OR NOTHING there, exactly as ADR 0080 made the tax: the
//     eligible set is judged whole, and a refusal stages nothing.
//
// The legal-move enumerator asks the same function of every candidate
// (attacker, target) move and withholds the ones it refuses, so a bot
// is never offered an attack the engine refuses (#544) — and, because
// the enumerator's attacks are one creature at a time, a seat under
// Silent Arbiter is offered attacks until one creature is attacking,
// and then none.

// AttackLimitScope is WHICH attacks a limit counts. Every printed card
// in the family says one of two things, and the difference is the line
// an implementation gets wrong in the direction of playing stronger
// than printed.
type AttackLimitScope uint8

const (
	// AttackLimitAttackingYou is "no more than N creatures can attack
	// YOU each combat" — Crawlspace, Judoon Enforcers. Only a creature
	// attacking the limit's controller — the PLAYER, not a
	// planeswalker they control or a battle they protect, which is
	// AttackTaxOnPlayer's reading of the same word — counts. THE ZERO
	// VALUE, deliberately: it is the narrower scope, so a card file
	// that forgets to choose limits less than printed, never more.
	AttackLimitAttackingYou AttackLimitScope = iota

	// AttackLimitEachCombat is "no more than N creatures can attack
	// each combat" — Silent Arbiter, Dueling Grounds, Caverns of
	// Despair. Every attacking creature counts, whoever it is
	// attacking; the limit binds its own controller's attacks as much
	// as anybody's.
	AttackLimitEachCombat
)

// AttackLimit is one "no more than N creatures can attack …" static
// contributed by a permanent on the battlefield. Declared as a struct
// rather than hooks because every printed card in the family is
// exactly a scope and a number; a conditional one (Mirri, Weatherlight
// Duelist's "as long as Mirri is tapped") is one more field when it is
// written, not a reinterpretation of these two.
type AttackLimit struct {
	// Scope is which attacks count. The zero value is "attacking you".
	Scope AttackLimitScope

	// Max is N. A limit with Max <= 0 is inert: "no creature can
	// attack" is a Restriction bit, not a count.
	Max int
}

// counts reports whether an attack naming `target` counts toward this
// limit, read off `source`, the permanent that prints it.
func (l AttackLimit) counts(target uuid.UUID, source *Card) bool {
	switch l.Scope {
	case AttackLimitEachCombat:
		return true
	case AttackLimitAttackingYou:
		// The seat id itself: an attack on a planeswalker or battle
		// names the permanent, so it never matches.
		return source != nil && target == source.Controller
	}
	return false
}

// CatalogAttackLimits returns the attack limits a catalog entry
// declares. Wired to the card catalog in carddef.go from
// CardDef.AttackLimits; nil in a test that builds cards by hand, the
// contract every other catalog hook has. Tests stub it directly.
//
// Keyed by CatalogAbilityKey, so a Silent Arbiter that has lost all
// its abilities limits nothing (CR 613.1f).
var CatalogAttackLimits func(key string) []AttackLimit

// attackLimitSource is one limit and the permanent it was read from.
type attackLimitSource struct {
	limit  AttackLimit
	source *Card
}

// activeAttackLimitsLocked collects every attack limit on the
// battlefield, in battlefield order. Nil — the answer at nearly every
// table — lets the callers skip building any assignment at all.
//
// Caller must hold g.mu. Reads only.
func (g *Game) activeAttackLimitsLocked() []attackLimitSource {
	if CatalogAttackLimits == nil || g.Battlefield == nil {
		return nil
	}
	var out []attackLimitSource
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		key := CatalogAbilityKey(*src)
		if key == "" {
			continue
		}
		for _, l := range CatalogAttackLimits(key) {
			if l.Max > 0 {
				out = append(out, attackLimitSource{limit: l, source: src})
			}
		}
	}
	return out
}

// AttackLimitError is what the declaration verbs return when a
// declaration would break an attack limit. It wraps ErrAttackLimit, so
// errors.Is callers match, and carries what the ws layer needs for the
// player-facing sentence of the `illegal_attack` frame without
// re-deriving a rule (ADR 0045 §6).
type AttackLimitError struct {
	// Scope and Max are the limit that was broken.
	Scope AttackLimitScope
	Max   int
	// Source is the permanent that prints it, and SourceName its name.
	Source     uuid.UUID
	SourceName string
	// Defender is the seat an "attacking you" limit protects, and
	// DefenderName their display name. uuid.Nil for a combat-wide
	// limit.
	Defender     uuid.UUID
	DefenderName string
	// Attacker is a creature in the refused declaration that the limit
	// counts, so the client has a card to point at.
	Attacker     uuid.UUID
	AttackerName string
}

// Error is the debug form: the sentence as a third party reads it.
func (e *AttackLimitError) Error() string {
	return "game: " + e.Sentence(uuid.Nil)
}

// Unwrap makes errors.Is(err, ErrAttackLimit) hold.
func (e *AttackLimitError) Unwrap() error { return ErrAttackLimit }

// Sentence is the player-facing explanation addressed to `viewer`:
// "No more than two creatures can attack you each combat
// (Crawlspace)." for the seat it protects, the seat's name for anyone
// else. Built server-side so the client never re-derives the limit.
func (e *AttackLimitError) Sentence(viewer uuid.UUID) string {
	s := "No more than " + blockerCountPhrase(e.Max) + " can attack"
	if e.Scope == AttackLimitAttackingYou {
		switch {
		case e.Defender != uuid.Nil && e.Defender == viewer:
			s += " you"
		default:
			s += " " + nameOr(e.DefenderName, "that player")
		}
	}
	s += " each combat"
	if e.SourceName != "" {
		s += " (" + e.SourceName + ")"
	}
	return s + "."
}

// AttackLimitRefusalForEffect reports whether `decls`, applied on top
// of the attacks already declared this combat, would break an attack
// limit, and the refusal when it would. The legal-move enumerator asks
// it of every candidate move, so the moves it offers are exactly the
// ones the declaration verbs accept (#544).
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) AttackLimitRefusalForEffect(decls []AttackDeclaration) *AttackLimitError {
	return g.attackLimitRefusalLocked(decls)
}

// attackLimitRefusalLocked is the one attack-limit check (CR 508.1c,
// #1507): the first limit on the battlefield that `decls` — applied on
// top of every creature attacking now — breaks, or nil.
//
// A limit is broken when it counts MORE attacks after the action than
// its Max AND more than it counts before. The second half mirrors the
// block side's (blockLimitRefusalLocked): a Crawlspace that arrives
// after three creatures were declared does not unmake them, and
// re-pointing one of them elsewhere is never refused for a count it
// lowers. An entry naming a creature that is already attacking
// RE-POINTS it, which is how the sandbox's "re-declare attacker" works,
// so moving an attacker from one Crawlspace player to another counts
// against the second one's limit.
//
// What counts as "attacking now" is every creature with an
// AttackingTarget — including one put onto the battlefield attacking,
// which CR 506.3c says was never declared. That cannot matter to a
// declaration the rules allow: such a creature can only arrive after
// the lock-in, and by then the rules' declaration is over. It can only
// refuse a sandbox LATE declaration, which is the weaker direction.
//
// Caller must hold g.mu with fresh layers. Reads only.
func (g *Game) attackLimitRefusalLocked(decls []AttackDeclaration) *AttackLimitError {
	limits := g.activeAttackLimitsLocked()
	if len(limits) == 0 || len(decls) == 0 {
		return nil
	}
	before := g.currentAttackAssignmentLocked()
	after := make(map[uuid.UUID]uuid.UUID, len(before)+len(decls))
	for a, t := range before {
		after[a] = t
	}
	for _, d := range decls {
		after[d.Attacker] = d.Target
	}
	for _, l := range limits {
		nAfter := l.countIn(after)
		if nAfter <= l.limit.Max || nAfter <= l.countIn(before) {
			continue
		}
		return g.attackLimitErrorLocked(l, decls)
	}
	return nil
}

// currentAttackAssignmentLocked maps every creature attacking now to
// what it attacks — the "before" both attackLimitRefusalLocked and
// AttackLimitRoomForEffect count against.
//
// Caller must hold g.mu. Reads only.
func (g *Game) currentAttackAssignmentLocked() map[uuid.UUID]uuid.UUID {
	out := map[uuid.UUID]uuid.UUID{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget != uuid.Nil {
			out[c.InstanceID] = c.AttackingTarget
		}
	}
	return out
}

// countIn is how many of the attacks in `assign` this limit counts.
func (l attackLimitSource) countIn(assign map[uuid.UUID]uuid.UUID) int {
	n := 0
	for _, target := range assign {
		if l.limit.counts(target, l.source) {
			n++
		}
	}
	return n
}

// AttackLimitRoomForEffect is how many MORE creatures may be declared
// attacking `target` this combat before an attack limit refuses the
// declaration — the smallest allowance among every limit on the
// battlefield that counts an attack on `target` — and whether any
// limit counts it at all (limited == false means no limit applies, and
// room is meaningless). #1533, ADR 0045 Decision 46.
//
// It is attackLimitRefusalLocked's arithmetic read the other way
// round, for creatures that are not attacking yet: k new attackers at
// `target` are accepted exactly when k <= room. A limit already over
// its Max (one that arrived late, Decision 44) leaves room 0, because
// every new attacker raises the count. It says nothing about
// re-pointing a creature that is already attacking, which the client's
// "attack with all" never does.
//
// The client's attack-with-all picker caps its selection at this
// number, so the cap is the engine's and never re-derived (ADR 0045
// §6).
//
// Caller must hold g's lock (read is enough) with fresh layers.
func (g *Game) AttackLimitRoomForEffect(target uuid.UUID) (room int, limited bool) {
	limits := g.activeAttackLimitsLocked()
	if len(limits) == 0 {
		return 0, false
	}
	before := g.currentAttackAssignmentLocked()
	for _, l := range limits {
		if !l.limit.counts(target, l.source) {
			continue
		}
		r := l.limit.Max - l.countIn(before)
		if r < 0 {
			r = 0
		}
		if !limited || r < room {
			room, limited = r, true
		}
	}
	return room, limited
}

// attackLimitErrorLocked builds the refusal for a broken limit, naming
// the first creature in `decls` the limit counts.
//
// Caller must hold g.mu. Reads only.
func (g *Game) attackLimitErrorLocked(l attackLimitSource, decls []AttackDeclaration) *AttackLimitError {
	e := &AttackLimitError{Scope: l.limit.Scope, Max: l.limit.Max}
	if l.source != nil {
		e.Source, e.SourceName = l.source.InstanceID, l.source.Effective().Name
		if l.limit.Scope == AttackLimitAttackingYou {
			e.Defender = l.source.Controller
			if p := g.playerByIDLocked(e.Defender); p != nil {
				e.DefenderName = p.Name
			}
		}
	}
	for _, d := range decls {
		if !l.limit.counts(d.Target, l.source) {
			continue
		}
		e.Attacker = d.Attacker
		if c := findBattlefieldCard(g, d.Attacker); c != nil {
			e.AttackerName = c.Effective().Name
		}
		break
	}
	return e
}
