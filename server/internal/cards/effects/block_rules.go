package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_rules.go — the card side of #750: constructors for
// Spec.BlockRules, the CR 509.1b block restrictions that carry a
// PARAMETER ("can't be blocked except by Walls", "can't be blocked by
// creatures with power 2 or less", "can't be blocked by more than one
// creature"). The engine half — slot 4 of BlockPairRefusalLocked, the
// count bounds DeclareBlockers judges, the two registries — is
// game/block_rules.go. See ADR 0045's addendum, Decision 11, and its
// 2026-09-24 amendment.
//
// A rule is two things, and the constructors keep them apart:
//
//   - a SCOPE, which says whose creatures the rule binds relative to
//     the permanent that carries it: the permanent itself (OnSelf),
//     the creature it is attached to (OnAttached), creatures its
//     controller controls (ControlledBySourceController), or any
//     creature matching a predicate (OnMatching);
//   - the RULE, which says what is refused for the creatures in
//     scope: every blocker outside a set (CantBeBlockedExceptBy), every
//     blocker inside one (CantBeBlockedBy), every blocker while a
//     condition holds (CantBeBlockedWhile), a bound on how many may
//     block (MaxBlockers, MinBlockers), or — on the other side of the
//     pair — which attackers a creature may not block
//     (CantBlockAttackers).
//
// One rule has no scope, because its printed line has none:
// NoMoreThanNCanBlockEachCombat, the whole-combat limit (#1507).
//
//	BlockRules: []game.BlockRule{
//	    CantBeBlockedExceptBy(OnAttached(), OfCreatureType("Wall"), "Walls"),   // Prowler's Helm
//	    CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"), // Legolas Greenleaf
//	    MaxBlockers(OnAttached(), 1),                                            // Vorrac Battlehorns
//	},
//
// Why the scope is not optional. The rule is read off every permanent
// on the battlefield for every (attacker, blocker) pair the engine
// considers, so a Pair function that forgets to ask "is this MY
// attacker?" binds every attacker at the table — a Prowler's Helm
// that makes the whole board unblockable except by Walls. Taking the
// scope as an argument makes that mistake impossible to write.
//
// Predicates are the targeting vocabulary (CardPredicate, targets.go),
// evaluated with the rule source's CONTROLLER as the caster, so
// YouControl() inside a rule means "the controller of the permanent
// that prints it". Everything reads effective characteristics live,
// at the moment the block is checked (CR 509.1b), and nothing is
// re-checked after the declaration.

// BlockScope picks the creatures a block rule binds, relative to the
// permanent carrying the rule. `source` is nil for an until-end-of-turn
// rule (game.TurnScopedBlockRules), whose effect has no permanent; the
// scopes that read a source answer false for it, and the UntilEOT
// constructors build their own scope from a snapshotted set instead.
//
// Reads only — it runs inside the engine's read-only block check.
type BlockScope func(g *game.Game, c, source *game.Card) bool

// OnSelf — "this creature": the permanent that carries the rule.
func OnSelf() BlockScope {
	return func(_ *game.Game, c, source *game.Card) bool {
		return source != nil && c.InstanceID == source.InstanceID
	}
}

// OnAttached — "equipped creature" / "enchanted creature": the
// creature the carrying Equipment or Aura is attached to. Re-read on
// every check, so moving the Equipment moves the rule with it.
func OnAttached() BlockScope {
	return func(_ *game.Game, c, source *game.Card) bool {
		return source != nil && source.IsAttachedTo(c.InstanceID)
	}
}

// ControlledBySourceController — "creatures you control", where "you"
// is the controller of the permanent carrying the rule.
func ControlledBySourceController() BlockScope {
	return func(_ *game.Game, c, source *game.Card) bool {
		return source != nil && c.Controller == source.Controller
	}
}

// OnMatching binds every creature the predicate accepts, with the
// rule source's controller as the predicate's caster ("Slivers",
// "non-Spirit creatures"). Usable on either side of the pair.
func OnMatching(pred CardPredicate) BlockScope {
	return func(g *game.Game, c, source *game.Card) bool {
		return pred(g, blockRuleCaster(source), *c)
	}
}

// PowerLessThanSource — "creatures with power less than this
// creature's power" (Champion of Lambholt). Both sides read the power
// the comparison rules use (CR 208.1, negative power included), live.
func PowerLessThanSource() BlockScope {
	return func(_ *game.Game, c, source *game.Card) bool {
		return source != nil && c.PowerForComparison() < source.PowerForComparison()
	}
}

// blockRuleCaster is the "you" a rule's predicates are evaluated
// against: the controller of the permanent carrying it, or nobody for
// a rule with no source.
func blockRuleCaster(source *game.Card) uuid.UUID {
	if source == nil {
		return uuid.Nil
	}
	return source.Controller
}

// --- attacker-side pair rules -----------------------------------

// CantBeBlockedExceptBy — "<scope> can't be blocked except by
// <allowed>" (Prowler's Helm's Walls, Departed Deckhand's Spirits,
// Canopy Cover's "creatures with flying or reach"). Every blocker the
// predicate does NOT accept is refused. `label` is the allowed set as
// the card prints it; the player reads it in the refusal sentence.
func CantBeBlockedExceptBy(attackers BlockScope, allowed CardPredicate, label string) game.BlockRule {
	return game.BlockRule{
		Reason: game.BlockReasonCantBeBlockedExceptBy,
		Label:  label,
		Pair: func(g *game.Game, attacker, blocker, source *game.Card) bool {
			return attackers(g, attacker, source) && !allowed(g, blockRuleCaster(source), *blocker)
		},
	}
}

// CantBeBlockedBy — "<scope> can't be blocked by <forbidden>" (Legolas
// Greenleaf's "creatures with power 2 or less"). The open form: only
// blockers the predicate accepts are refused.
func CantBeBlockedBy(attackers BlockScope, forbidden CardPredicate, label string) game.BlockRule {
	return game.BlockRule{
		Reason: game.BlockReasonCantBeBlockedBy,
		Label:  label,
		Pair: func(g *game.Game, attacker, blocker, source *game.Card) bool {
			return attackers(g, attacker, source) && forbidden(g, blockRuleCaster(source), *blocker)
		},
	}
}

// CantBeBlockedWhile — "<scope> can't be blocked as long as <cond>"
// (Thieves' Tools: "as long as its power is 3 or less"). The
// condition is on the ATTACKER and is read when the block is checked,
// so it refuses every blocker or none. It reports the plain
// cant_be_blocked reason — the player reads "can't be blocked", which
// is what the card says while the condition holds.
func CantBeBlockedWhile(attackers BlockScope, cond CardPredicate) game.BlockRule {
	return game.BlockRule{
		Reason: game.BlockReasonCantBeBlocked,
		Pair: func(g *game.Game, attacker, _, source *game.Card) bool {
			return attackers(g, attacker, source) && cond(g, blockRuleCaster(source), *attacker)
		},
	}
}

// --- blocker-side pair rules ------------------------------------

// CantBlockAttackers — "<blockers> can't block <attackers>". The rule
// on the BLOCKER's side of the pair: Champion of Lambholt's "creatures
// with power less than this creature's power can't block creatures
// you control" is
//
//	CantBlockAttackers(PowerLessThanSource(), ControlledBySourceController(), …)
//
// `label` is the whole clause as a sentence fragment; the refusal
// reads "<blocker> can't block <attacker>: <label>."
func CantBlockAttackers(blockers, attackers BlockScope, label string) game.BlockRule {
	return game.BlockRule{
		Reason: game.BlockReasonCantBlockAttacker,
		Label:  label,
		Pair: func(g *game.Game, attacker, blocker, source *game.Card) bool {
			return blockers(g, blocker, source) && attackers(g, attacker, source)
		},
	}
}

// CantBlockOrBeBlockedBy — "<scope> can't block or be blocked by
// <other>" (Avatar Kuruk's Spirit token: "This token can't block or be
// blocked by non-Spirit creatures"). Two rules, one per side of the
// pair, because they refuse different pairs and report different
// reasons: the scoped creature attacking and an `other` creature
// blocking it, and the scoped creature blocking an `other` attacker.
//
// `label` names the other creatures as printed ("non-Spirit
// creatures").
func CantBlockOrBeBlockedBy(scope BlockScope, other CardPredicate, label string) []game.BlockRule {
	return []game.BlockRule{
		CantBeBlockedBy(scope, other, label),
		CantBlockAttackers(scope, OnMatching(other), "it can't block "+label),
	}
}

// --- count rules ------------------------------------------------

// MaxBlockers — "<scope> can't be blocked by more than N creatures"
// (Hungering Hydra, Vorrac Battlehorns, Alpha Authority: one). Judged
// on the whole declaration by DeclareBlockers, which refuses a set
// that would exceed it with too_many_blockers; the tightest maximum on
// the battlefield wins.
func MaxBlockers(attackers BlockScope, n int) game.BlockRule {
	return game.BlockRule{
		Count: func(g *game.Game, attacker, source *game.Card) (int, int) {
			if !attackers(g, attacker, source) {
				return 0, 0
			}
			return 0, n
		},
	}
}

// MinBlockers — "<scope> can't be blocked except by N or more
// creatures" (Rampaging Ceratops: three). Menace's two is built into
// the engine and needs no rule; the largest minimum wins, so a menace
// creature under this rule needs three. A minimum above a maximum
// makes the attacker unblockable (ADR 0045 addendum, Decision 12).
func MinBlockers(attackers BlockScope, n int) game.BlockRule {
	return game.BlockRule{
		Count: func(g *game.Game, attacker, source *game.Card) (int, int) {
			if !attackers(g, attacker, source) {
				return 0, 0
			}
			return n, 0
		},
	}
}

// --- whole-combat limit -----------------------------------------

// NoMoreThanNCanBlockEachCombat — "No more than N creatures can block
// each combat" (Silent Arbiter and Dueling Grounds: one; Caverns of
// Despair: two). BlockRule.Limit, the whole-combat bound ADR 0045's
// Decision 12 reserved and #1507 built: every blocker in the combat
// counts, whoever controls it and whichever attacker it blocks, and
// DeclareBlockers refuses a declaration that would go past it with
// declaration_limit (Decision 43).
//
// No scope argument, unlike every rule above, because the printed
// line has none: it binds every creature at the table, the Arbiter's
// own controller's included. The attack half of the same card is an
// AttackLimit (attack_limits.go), not a BlockRule.
func NoMoreThanNCanBlockEachCombat(n int) game.BlockRule {
	return game.BlockRule{
		Limit: func(_ *game.Game, _, _ *game.Card) int { return n },
	}
}

// EachOpponentCantBlockWithMoreThanN — "each opponent can't block with
// more than N creatures this combat" (Mirri, Weatherlight Duelist's
// attack trigger: one). A Limit rule with LimitPerDefender, so each
// opponent's blockers are counted on their own: in a four-player game
// every opponent may block with one creature, and one opponent's
// block never uses up another's (#1534, ADR 0045 Decision 47).
//
// Registered on resolution into the turn-scoped registry, as
// BlockRuleUntilEOT's rules are, but WITHOUT a snapshot: the line is
// a rule about players, not a change to any creature's
// characteristics, so CR 611.2c does not lock the affected set, and a
// creature an opponent flashes in afterwards is bound too. "Each
// opponent" is read at resolution against the ability's controller,
// and the rule outlives Mirri (CR 611.2b).
//
// "This combat" rides the until-end-of-turn registry because the
// engine has no extra combats yet (#753): the turn's one combat is the
// only one the rule can meet. When extra combats land, this must end
// at the end of combat instead, or it binds the next one too.
//
// `label` is the whole clause the refusal sentence reads, card name
// included, because a turn-scoped rule has no source permanent to
// name.
type EachOpponentCantBlockWithMoreThanN struct {
	N     int
	Label string
}

// Apply registers the rule. Caller is inside the resolution frame
// (holds g.mu write).
func (r EachOpponentCantBlockWithMoreThanN) Apply(ctx *Context) error {
	if r.N <= 0 {
		return nil
	}
	you, n := ctx.Controller(), r.N
	ctx.Game.RegisterTurnScopedBlockRuleLocked(game.BlockRule{
		Limit: func(_ *game.Game, blocker, _ *game.Card) int {
			if blocker == nil || blocker.Controller == you {
				return 0
			}
			return n
		},
		LimitPerDefender: true,
		Label:            eotLabel(r.Label, "each opponent's blocks are limited this combat"),
	})
	return nil
}

// --- until end of turn ------------------------------------------

// BlockRuleUntilEOT registers a block rule for the rest of the turn —
// "<this creature / target creature> can't be blocked this turn except
// by creatures with haste" (Gingerbrute, Departed Deckhand's {3}{U}).
// It goes into game.TurnScopedBlockRules, which the turn-end sweep
// empties (ADR 0045 addendum, Decision 11).
//
// The affected set is snapshotted at resolution (CR 611.2c), exactly
// as RestrictUntilEOT's is: a creature that enters later is not
// covered, and one that leaves and returns is a new object (CR 400.7)
// and is not covered either. `Rule` receives the scope for that set
// and builds the rule from the ordinary constructors, so the
// turn-scoped twin of every rule above is one line:
//
//	BlockRuleUntilEOT{Target: id, Rule: func(s BlockScope) game.BlockRule {
//	    return CantBeBlockedExceptBy(s, HasKeyword("haste"), "creatures with haste")
//	}}
//
// A turn-scoped rule has no source permanent (the effect outlives the
// ability that made it, CR 611.2b), so predicates inside it see no
// caster: YouControl() matches nothing there. No printed turn-scoped
// rule needs one.
type BlockRuleUntilEOT struct {
	// Target pins the rule to one permanent. Ignored when Match is set.
	Target uuid.UUID
	// Match selects the affected permanents, evaluated ONCE at
	// resolution with the ability's controller as the caster.
	Match CardPredicate
	// Rule builds the rule for the snapshotted set.
	Rule func(scope BlockScope) game.BlockRule
	// Label names the effect in logs and the continuation census.
	Label string
}

// Apply registers the rule. Caller is inside the resolution frame
// (holds g.mu write).
func (r BlockRuleUntilEOT) Apply(ctx *Context) error {
	if r.Rule == nil {
		return nil
	}
	set := eotSnapshot(ctx, r.Target, r.Match)
	if set == nil {
		return nil
	}
	inSet := set.appliesTo()
	rule := r.Rule(func(g *game.Game, c, _ *game.Card) bool { return inSet(c, g, nil) })
	if rule.Label == "" {
		// The continuation census notes a turn-scoped rule by its
		// Label. A rule with a printed parameter keeps it, because
		// the refusal sentence reads it too; one without (a count, a
		// "can't be blocked while") is named after the effect.
		rule.Label = eotLabel(r.Label, "block rule until end of turn")
	}
	ctx.Game.RegisterTurnScopedBlockRuleLocked(rule)
	return nil
}
