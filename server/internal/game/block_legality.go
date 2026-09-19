package game

// block_legality.go is the one answer to "may this creature block that
// attacker?" (CR 509.1b), for every caller that asks it: DeclareBlocker,
// the legal-move enumerator (internal/legal) and the #328 auto-pass
// signal (seatOwesBlockDecisionLocked). ADR 0045 §3 is the reason there
// is exactly one: a bot is only ever offered a block the engine accepts
// (#544) because all three read this function rather than a copy of it.
//
// It is a method on *Game, not a two-card function, because the rules
// it has to answer read the table. Landwalk (CR 702.14c) looks at the
// DEFENDING player's lands, and the block rules of #750 (slot 4
// below, block_rules.go) read a third card's live power. The free function
// CanBlock(attacker, blocker) this replaces could see neither; it was
// deleted rather than kept as a wrapper, because a second entry point
// that can't see the game is precisely the drifting copy §3 forbids.
// See docs/decisions/0045-combat-restrictions.md, addendum Decisions
// 7-9.

import "github.com/google/uuid"

// BlockReason is the stable snake_case token naming why a block was
// refused, in the style of Restriction.Names(). It is the `reason`
// field of the `illegal_block` error frame, so a token never changes
// spelling once shipped. The zero value means the block is legal.
//
// Only the reasons some check actually produces are declared. The
// addendum reserves more (too_few_blockers, too_many_blockers,
// declaration_limit, not_defending, tapped); each joins
// this list in the PR that first returns it, so the wire never
// advertises a refusal nothing can send. The count reasons are still
// reserved on purpose: a bound is judged on a whole declaration, and
// this engine's per-pair DeclareBlocker has no set to judge, so
// nothing can send them yet (see blockers.go).
type BlockReason string

const (
	// BlockReasonCantBlock — a "~ can't block" restriction on the
	// blocker (Pacifism, Carrion Feeder). CR 509.1b.
	BlockReasonCantBlock BlockReason = "cant_block"
	// BlockReasonCantBeBlocked — a "~ can't be blocked" restriction on
	// the attacker (Whispersilk Cloak, Rogue's Passage). CR 509.1b.
	BlockReasonCantBeBlocked BlockReason = "cant_be_blocked"
	// BlockReasonFlying — the attacker flies and the blocker has
	// neither flying nor reach. CR 702.9b.
	BlockReasonFlying BlockReason = "flying"
	// BlockReasonLandwalk — the attacker has a landwalk ability and the
	// defending player controls a land it names. CR 702.14c.
	BlockReasonLandwalk BlockReason = "landwalk"
	// BlockReasonFear — fear allows only artifact or black blockers.
	BlockReasonFear BlockReason = "fear"
	// BlockReasonIntimidate — intimidate allows only artifact blockers or
	// blockers sharing a color with the attacker.
	BlockReasonIntimidate BlockReason = "intimidate"
	// BlockReasonShadow — creatures with shadow can block or be blocked only
	// by creatures with shadow.
	BlockReasonShadow BlockReason = "shadow"
	// BlockReasonHorsemanship — horsemanship can be blocked only by
	// horsemanship.
	BlockReasonHorsemanship BlockReason = "horsemanship"
	// BlockReasonSkulk — skulk cannot be blocked by a creature with greater
	// power.
	BlockReasonSkulk BlockReason = "skulk"
	// BlockReasonProtection — the attacker has protection from a
	// quality the blocker has (CR 702.16f). #662 claimed the token
	// ADR 0045's addendum reserved.
	BlockReasonProtection BlockReason = "protection"
	// BlockReasonCantBeBlockedBy — a block rule on the attacker's side
	// naming creatures that may NOT block it: "can't be blocked by
	// creatures with power 2 or less" (Legolas Greenleaf), "can't be
	// blocked by more than one creature"'s pair-shaped cousins. #750.
	BlockReasonCantBeBlockedBy BlockReason = "cant_be_blocked_by"
	// BlockReasonCantBeBlockedExceptBy — the closed form: everything
	// outside the named set is refused ("can't be blocked except by
	// Walls", Prowler's Helm). #750.
	BlockReasonCantBeBlockedExceptBy BlockReason = "cant_be_blocked_except_by"
	// BlockReasonCantBlockAttacker — a block rule on the BLOCKER's
	// side: "creatures with power less than this creature's power
	// can't block creatures you control" (Champion of Lambholt). #750.
	BlockReasonCantBlockAttacker BlockReason = "cant_block_attacker"
)

// BlockReasons lists every reason the engine can return today, in the
// order BlockPairRefusalLocked checks them. The ws layer's test walks
// it so a new reason cannot ship without a player-facing sentence.
func BlockReasons() []BlockReason {
	return []BlockReason{
		BlockReasonCantBlock,
		BlockReasonCantBeBlocked,
		BlockReasonFlying,
		BlockReasonLandwalk,
		BlockReasonFear,
		BlockReasonIntimidate,
		BlockReasonShadow,
		BlockReasonHorsemanship,
		BlockReasonSkulk,
		BlockReasonProtection,
		BlockReasonCantBeBlockedBy,
		BlockReasonCantBeBlockedExceptBy,
		BlockReasonCantBlockAttacker,
	}
}

// BlockRefusal is the answer BlockPairRefusalLocked gives. The zero
// value, BlockOK, is a legal block.
type BlockRefusal struct {
	// Reason is "" for a legal block.
	Reason BlockReason
	// Source is the permanent carrying the refusing ability or
	// restriction: the blocker for cant_block, the attacker for
	// cant_be_blocked, flying, landwalk, fear, intimidate, horsemanship
	// and skulk; shadow uses the creature that has shadow, whether it is
	// attacking or blocking. The engine does not record
	// which effect granted a keyword or wrote a restriction bit (a
	// Pacifism, a Lord of Atlantis), so for those it is the affected
	// creature rather than the effect's source.
	Source uuid.UUID
	// N is the bound for the count reasons the addendum's PR 2 adds.
	// Zero for every reason declared today.
	N int
	// Label is the block rule's parameter as its card prints it —
	// "Walls", "creatures with power 2 or less" — carried so the
	// player-facing sentence can name it without the client
	// re-deriving the rule (ADR 0045 §6). Empty for every reason a
	// bit or a keyword produced. #750.
	Label string
}

// BlockOK is the legal answer.
var BlockOK = BlockRefusal{}

// Legal reports whether the refusal is no refusal at all.
func (r BlockRefusal) Legal() bool { return r.Reason == "" }

// BlockPairRefusalLocked reports whether `blocker` may block `attacker`
// and, when it may not, why. It checks per-pair legality only: tap
// state, the step, who controls the blocker and block counts (menace)
// are the declaration's business, not this function's.
//
// The checks run in the addendum's Decision 9 order and the first
// refusal is returned. The order decides only which reason is reported,
// never whether the pair is legal:
//
//  1. the restriction bits: CantBlock on the blocker, CantBeBlocked on
//     the attacker (CR 509.1b);
//  2. evasion keywords: flying (CR 702.9b), landwalk (CR 702.14c), fear,
//     intimidate, shadow, horsemanship and skulk;
//  3. protection (CR 702.16f), #662;
//  4. block rules from permanents (#750) — the restrictions that
//     carry a parameter, read from their sources at check time
//     (block_rules.go).
//
// READ-ONLY, and that is a contract with a test
// (TestBlockLegalityDoesNotMutate): the enumerator and the view call
// this from inside ReadSnapshot, where a write is a data race. It never
// recomputes layers. Caller holds g.mu (read or write) and layers are
// fresh — DeclareBlocker recomputes before calling, and ReadSnapshot
// recomputes before running its callback.
//
// nil arguments refuse: a missing attacker "can't be blocked" and a
// missing blocker "can't block". No caller passes nil; the answer only
// has to be a refusal.
func (g *Game) BlockPairRefusalLocked(attacker, blocker *Card) BlockRefusal {
	if blocker == nil {
		return BlockRefusal{Reason: BlockReasonCantBlock}
	}
	if attacker == nil {
		return BlockRefusal{Reason: BlockReasonCantBeBlocked}
	}

	// 1. CR 509.1b restrictions. Read before the evasion keywords
	// because they are absolute: no defensive keyword answers "can't
	// be blocked" the way reach answers flying.
	if Restricted(blocker, CantBlock) {
		return BlockRefusal{Reason: BlockReasonCantBlock, Source: blocker.InstanceID}
	}
	if Restricted(attacker, CantBeBlocked) {
		return BlockRefusal{Reason: BlockReasonCantBeBlocked, Source: attacker.InstanceID}
	}

	// 2. Evasion keywords. Their order is stable because it controls the
	// reason the client receives when an attacker has more than one.
	if HasKeyword(attacker, "flying") && !HasKeyword(blocker, "flying") && !HasKeyword(blocker, "reach") {
		return BlockRefusal{Reason: BlockReasonFlying, Source: attacker.InstanceID}
	}
	if _, land := g.landwalkBlockingLandLocked(attacker); land != nil {
		return BlockRefusal{Reason: BlockReasonLandwalk, Source: attacker.InstanceID}
	}
	if HasKeyword(attacker, "fear") && !blocker.IsArtifact() && !blocker.HasColor("B") {
		return BlockRefusal{Reason: BlockReasonFear, Source: attacker.InstanceID}
	}
	if HasKeyword(attacker, "intimidate") && !blocker.IsArtifact() && !sharesColor(attacker, blocker) {
		return BlockRefusal{Reason: BlockReasonIntimidate, Source: attacker.InstanceID}
	}
	if attackerShadow, blockerShadow := HasKeyword(attacker, "shadow"), HasKeyword(blocker, "shadow"); attackerShadow != blockerShadow {
		source := attacker.InstanceID
		if !attackerShadow {
			source = blocker.InstanceID
		}
		return BlockRefusal{Reason: BlockReasonShadow, Source: source}
	}
	if HasKeyword(attacker, "horsemanship") && !HasKeyword(blocker, "horsemanship") {
		return BlockRefusal{Reason: BlockReasonHorsemanship, Source: attacker.InstanceID}
	}
	if HasKeyword(attacker, "skulk") && blocker.PowerForComparison() > attacker.PowerForComparison() {
		return BlockRefusal{Reason: BlockReasonSkulk, Source: attacker.InstanceID}
	}

	// 3. Protection (CR 702.16f): "attacking creatures with protection
	// can't be blocked by creatures that have the stated quality".
	// #662 filled in the slot ADR 0045's addendum reserved, with the
	// call it predicted: the blocker's effective characteristics,
	// which this function already holds.
	//
	// ONE DIRECTION ONLY, and the rule is asymmetric on purpose.
	// Protection on the ATTACKER refuses a blocker with the quality;
	// protection on the BLOCKER refuses nothing — a pro-red creature
	// may block a red attacker all day, it simply takes no damage
	// (CR 702.16e). Writing the test both ways would be the
	// commonest misreading of the keyword.
	if HasProtection(attacker) && ProtectedFrom(attacker, SourceCharacteristics(blocker)) {
		return BlockRefusal{Reason: BlockReasonProtection, Source: attacker.InstanceID}
	}

	// 4. Block rules ("can't be blocked except by Walls", "can't be
	// blocked by creatures with power 2 or less"), read from their
	// sources at check time — #750, addendum Decision 11. Last
	// because a rule is the only check here that walks the whole
	// battlefield, and because a bit or a keyword naming the same
	// refusal is the more specific answer for the player.
	//
	// Block COUNTS are not here. A bound is a property of a whole
	// declaration, not of a pair, so it is judged at the declaration's
	// lock-in (revertIllegalBlockCountsLocked, blockers.go).
	if r := g.blockRuleRefusalLocked(attacker, blocker); !r.Legal() {
		return r
	}

	return BlockOK
}

// CanBlockLocked reports whether `blocker` may block `attacker`: the
// boolean form of BlockPairRefusalLocked, with the same lock and
// fresh-layers contract.
func (g *Game) CanBlockLocked(attacker, blocker *Card) bool {
	return g.BlockPairRefusalLocked(attacker, blocker).Legal()
}

// BlockRefusedError is what DeclareBlocker returns for a pair
// BlockPairRefusalLocked refuses. It wraps ErrIllegalBlock, so
// errors.Is(err, ErrIllegalBlock) callers keep working, and carries
// what the ws layer needs to build the player-facing sentence of the
// `illegal_block` error frame without re-deriving a rule.
type BlockRefusedError struct {
	BlockRefusal
	Blocker      uuid.UUID
	Attacker     uuid.UUID
	BlockerName  string
	AttackerName string
	// Keyword is the evasion keyword that refused the block, for any
	// evasion reason ("flying", "islandwalk", "fear", …).
	Keyword string
	// Defender is the defending player for the attack, and
	// DefenderName their display name. uuid.Nil when the attacker is
	// not attacking anything.
	Defender     uuid.UUID
	DefenderName string
	// LandName and LandKind name the defending player's land that
	// switched landwalk on: "Tropical Island" and "an Island".
	LandName string
	LandKind string
	// Quality is the protection quality that refused the block, as
	// the attacker's token prints it: "red", "Demons". Empty for
	// every other reason. #662.
	Quality string
}

// Error is the debug form. It is the sentence as a third party would
// read it.
func (e *BlockRefusedError) Error() string {
	return "game: " + e.Sentence(uuid.Nil)
}

// Unwrap makes errors.Is(err, ErrIllegalBlock) hold.
func (e *BlockRefusedError) Unwrap() error { return ErrIllegalBlock }

// Sentence is the player-facing explanation, addressed to `viewer`:
// the defending player reads "you control an Island", anyone else
// reads the defender's name. Built server-side so the client never
// re-derives a block rule (ADR 0045 §6).
func (e *BlockRefusedError) Sentence(viewer uuid.UUID) string {
	blocker := nameOr(e.BlockerName, "That creature")
	attacker := nameOr(e.AttackerName, "That creature")
	switch e.Reason {
	case BlockReasonCantBlock:
		return blocker + " can't block."
	case BlockReasonCantBeBlocked:
		return attacker + " can't be blocked."
	case BlockReasonFlying:
		return attacker + " has flying, and " + blocker + " has neither flying nor reach."
	case BlockReasonLandwalk:
		who := "its defending player controls"
		switch {
		case e.Defender != uuid.Nil && e.Defender == viewer:
			who = "you control"
		case e.DefenderName != "":
			who = e.DefenderName + " controls"
		}
		land := nameOr(e.LandKind, "a land of that kind")
		if e.LandName != "" {
			land += " (" + e.LandName + ")"
		}
		return attacker + " has " + nameOr(e.Keyword, "landwalk") + ", and " + who + " " + land + "."
	case BlockReasonFear:
		return attacker + " has fear, and " + blocker + " is neither an artifact nor black."
	case BlockReasonIntimidate:
		return attacker + " has intimidate, and " + blocker + " is not an artifact and does not share a color with it."
	case BlockReasonShadow:
		return "Only one of " + attacker + " and " + blocker + " has shadow, so they can't block each other."
	case BlockReasonHorsemanship:
		return attacker + " has horsemanship, and " + blocker + " does not."
	case BlockReasonSkulk:
		return attacker + " has skulk, and " + blocker + " has greater power."
	case BlockReasonProtection:
		// CR 702.16f. The quality is named because "has protection"
		// alone does not tell the player which of their creatures
		// could have blocked instead.
		return attacker + " has protection from " + nameOr(e.Quality, "that quality") +
			", so " + blocker + " can't block it."
	case BlockReasonCantBeBlockedBy:
		// #750. The Label is the rule's parameter as the card prints
		// it, so the player reads the clause rather than a token.
		return attacker + " can't be blocked by " + nameOr(e.Label, "that creature") + "."
	case BlockReasonCantBeBlockedExceptBy:
		return attacker + " can't be blocked except by " + nameOr(e.Label, "creatures it names") +
			", and " + blocker + " is not one."
	case BlockReasonCantBlockAttacker:
		return blocker + " can't block " + attacker + ": " + nameOr(e.Label, "an effect says so") + "."
	}
	return blocker + " can't block " + attacker + "."
}

// blockRefusedErrorLocked builds the error for a refused pair. Caller
// holds g.mu with fresh layers; reads only.
func (g *Game) blockRefusedErrorLocked(attacker, blocker *Card, r BlockRefusal) *BlockRefusedError {
	e := &BlockRefusedError{BlockRefusal: r}
	if blocker != nil {
		e.Blocker, e.BlockerName = blocker.InstanceID, blocker.Effective().Name
	}
	if attacker == nil {
		return e
	}
	e.Attacker, e.AttackerName = attacker.InstanceID, attacker.Effective().Name
	if e.Defender = g.defendingPlayerForAttackLocked(attacker.AttackingTarget); e.Defender != uuid.Nil {
		if p := g.playerByIDLocked(e.Defender); p != nil {
			e.DefenderName = p.Name
		}
	}
	switch r.Reason {
	case BlockReasonFlying, BlockReasonFear, BlockReasonIntimidate, BlockReasonShadow, BlockReasonHorsemanship, BlockReasonSkulk:
		e.Keyword = string(r.Reason)
	case BlockReasonLandwalk:
		if kw, land := g.landwalkBlockingLandLocked(attacker); land != nil {
			e.Keyword = kw
			e.LandName = land.Effective().Name
			if spec, ok := landwalkRequirement(kw); ok {
				e.LandKind = spec.describe()
			}
		}
	case BlockReasonProtection:
		e.Keyword = string(r.Reason)
		if q, ok := MatchedProtection(attacker, SourceCharacteristics(blocker)); ok {
			e.Quality = q.Printed
		}
	}
	return e
}

// sharesColor reports whether two permanents have a color in common, using
// their current layer-5 color characteristics. Colorless creatures share no
// color, so intimidate lets them through only when they are artifacts.
func sharesColor(a, b *Card) bool {
	if a == nil || b == nil {
		return false
	}
	for _, color := range a.EffectiveColors() {
		if b.HasColor(color) {
			return true
		}
	}
	return false
}

func nameOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
