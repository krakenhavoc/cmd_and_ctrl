package game

import (
	"fmt"
)

// scoped_block_rules.go is ADR 0041 phase 3's tier 3b for block-rule
// effects (Decision P8, #1497): a CR 509.1b block restriction a
// resolving spell or ability creates is a ScopedEffect record, not a
// closure in Game.TurnScopedBlockRules.
//
// WHAT IT REPLACED. Game.TurnScopedBlockRules was a slice of BlockRule
// values — three closures each (Pair, Count, Limit) — emptied wholesale
// at cleanup. A table holding one was not a restore point
// (ContinuationCensus.TurnScopedBlockRules).
//
// WHAT IT IS NOW. Two mod kinds whose reader is the block-rule walk
// rather than the layer pass (modKindSpec.reader):
//
//   - cantBeBlockedExceptBy — Gingerbrute's activated ability, Departed
//     Deckhand's granted evasion: the pinned attacker(s), reading
//     Keywords and Subtypes (each an any-of) and Text, the printed
//     allowed set;
//   - limitBlockersPerDefender — Mirri, Weatherlight Duelist's attack
//     trigger: ScopeOpponentsCreatures, reading Amount.
//
// forEachBlockRuleLocked's third walk (forEachScopedBlockRuleLocked)
// adapts every live record into the BlockRule the pair, count and
// limit checks already consume — built by the RUNNING binary from the
// record's data, the same "rebuilt, never persisted" status the layer
// adapter and the replacement gather have.
//
// UNLIKE THE REPLACEMENT KINDS, a block-rule closure is built fresh on
// every walk and captures the record's fields BY VALUE rather than
// reading them back by Seq at call time. That is safe here for a
// reason the replacement kinds don't share: block declaration never
// pauses on a CR 616 prompt the way replacement ordering can, so a
// BlockRule this walk hands to a check is used and discarded within
// the same locked call that built it — there is no later moment at
// which the record could have changed under it. Every such record
// still takes a Seq (appendScopedEffectLocked stamps one on any
// non-layer mod), but nothing here reads it back.

// blockRuleModProblem is registration's check on a block-rule mod's
// parameters, mirroring replacementModProblem. Registration panics on
// a problem: every caller is an engine function, so a bad parameter is
// a programming error in it.
func blockRuleModProblem(m Mod) string {
	switch m.Kind {
	case ModCantBeBlockedExceptBy:
		if m.Text == "" {
			return "a cantBeBlockedExceptBy rule needs Text, the printed allowed set"
		}
		if len(m.Keywords) == 0 && len(m.Subtypes) == 0 {
			return "a cantBeBlockedExceptBy rule needs at least one keyword or subtype"
		}
	case ModLimitBlockersPerDefender:
		if m.Amount < 1 {
			return fmt.Sprintf("a limitBlockersPerDefender rule needs an amount of at least 1, got %d", m.Amount)
		}
	}
	return ""
}

// ---------------------------------------------------------------
// Constructors — what card-facing effect code writes
// ---------------------------------------------------------------

// CantBeBlockedExceptByMod is "<creature> can't be blocked except by
// <keyword-or-subtype>" (Gingerbrute, Departed Deckhand). `keywords`
// and `subtypes` are each an any-of — a blocker matching either is
// allowed; `text` is the allowed set as the card prints it, read by
// the refusal sentence.
func CantBeBlockedExceptByMod(keywords, subtypes []string, text string) Mod {
	return Mod{Kind: ModCantBeBlockedExceptBy, Keywords: copyStrings(keywords), Subtypes: copyStrings(subtypes), Text: text}
}

// LimitBlockersPerDefenderMod is "each opponent can't block with more
// than N creatures this combat" (Mirri, Weatherlight Duelist), read
// per DEFENDING PLAYER against the record's affected scope
// (ScopeOpponentsCreatures).
func LimitBlockersPerDefenderMod(n int) Mod {
	return Mod{Kind: ModLimitBlockersPerDefender, Amount: n}
}

// ---------------------------------------------------------------
// The block-rule walk's adapter
// ---------------------------------------------------------------

// forEachScopedBlockRuleLocked is forEachBlockRuleLocked's third walk:
// every live ScopedEffect record whose mod is a block-rule kind,
// adapted into the BlockRule the pair, count and limit checks already
// consume. Walked after the battlefield and (today) nothing else, so a
// permanent's printed rule is reported first.
//
// Caller must hold g.mu (read or write) with fresh layers. Reads only.
func (g *Game) forEachScopedBlockRuleLocked(fn func(rule BlockRule, source *Card) bool) {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if modKinds[m.Kind].reader != readerBlockRule {
				continue
			}
			if !fn(blockRuleFromScopedMod(*e, m), nil) {
				return
			}
		}
	}
}

// blockRuleFromScopedMod builds the BlockRule the walk hands to a
// check, from one mod of one record. See the file comment for why the
// closures below capture their inputs by value rather than reading the
// record back by Seq.
func blockRuleFromScopedMod(e ScopedEffect, m Mod) BlockRule {
	switch m.Kind {
	case ModCantBeBlockedExceptBy:
		affected := affectedPredicate(e.Affected)
		allowed := blockRuleAllowedPredicate(m.Keywords, m.Subtypes)
		return BlockRule{
			Reason: BlockReasonCantBeBlockedExceptBy,
			Label:  m.Text,
			Pair: func(g *Game, attacker, blocker, _ *Card) bool {
				return affected(attacker, g, nil) && !allowed(blocker)
			},
		}
	case ModLimitBlockersPerDefender:
		// scopePredicate, not a hand-rolled "blocker.Controller != you":
		// the record's Scope (ScopeOpponentsCreatures) is what
		// RegisterScopedRuleEffectForEffect demanded at registration,
		// so the walk reads it back rather than re-deriving the same
		// answer a second way.
		inScope, n := scopePredicate(e.Scope, e.Controller), m.Amount
		return BlockRule{
			Limit: func(g *Game, blocker, _ *Card) int {
				if blocker == nil || !inScope(blocker, g, nil) {
					return 0
				}
				return n
			},
			LimitPerDefender: true,
			Label:            e.Label,
		}
	}
	return BlockRule{}
}

// blockRuleAllowedPredicate is ModCantBeBlockedExceptBy's "allowed"
// test on the BLOCKER: true when it carries one of the keywords or one
// of the subtypes. Mirrors effects.HasKeyword / effects.OfCreatureType
// (cards/effects/targets.go), which the catalog-printed twin of this
// rule (CatalogBlockRules) uses.
func blockRuleAllowedPredicate(keywords, subtypes []string) func(c *Card) bool {
	return func(c *Card) bool {
		if c == nil {
			return false
		}
		for _, kw := range keywords {
			if HasKeyword(c, kw) {
				return true
			}
		}
		if c.IsCreature() {
			for _, st := range subtypes {
				if c.HasSubtype(st) {
					return true
				}
			}
		}
		return false
	}
}
