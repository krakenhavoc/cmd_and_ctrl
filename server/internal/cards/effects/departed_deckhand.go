package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Departed Deckhand — Creature — Spirit Pirate {1}{U}, 2/2:
//
//	"When this creature becomes the target of a spell, sacrifice it.
//	 This creature can't be blocked except by Spirits.
//	 {3}{U}: Another target creature you control can't be blocked this
//	 turn except by Spirits."
//
// Only the first line ships, and it is the DRAWBACK — which is the
// right direction for a partial card (#259) and worth stating plainly:
// what is on the battlefield today is a vanilla 2/2 that falls over to
// any spell that looks at it.
//
// # What ships
//
// "Becomes the target of A SPELL" is narrower than the "spell or
// ability" every other card in the family prints, and the difference
// is real: an equip, a Rhystic Study trigger or a fight ability that
// names this creature does NOT kill it. EventBecomesTarget is emitted
// for both and carries no discriminator, so the spell half is read off
// the stack (SelfTargetedByASpell) — a spell's stack item shares its ID
// with the source card, an ability's does not.
//
// The trigger fires at ANNOUNCE (CR 115.7), so it goes on the stack
// above the spell that targeted and resolves first. The Deckhand is
// therefore already in its owner's graveyard when that spell tries to
// resolve, and the spell is countered on resolution for having no
// legal target (CR 608.2b) — which is the whole point of the card and
// falls out for free.
//
// # What does not ship, and why
//
// Both evasion clauses are "can't be blocked EXCEPT BY <a kind of
// creature>", and that is a conditional block restriction: a rule
// about the pair, parameterised by a predicate on the BLOCKER.
// game.Restriction is a flat bit set (ADR 0045 §2) and its
// CantBeBlocked bit says "nothing may block this" — shipping the bit
// here would make the Deckhand unconditionally unblockable, i.e.
// STRONGER than printed, which is the one direction a simplification
// must never take.
//
// The engine half of the real shape landed with #705's game-aware
// block check: game.BlockRule (server/internal/game/block_rules.go) is
// slot 4 of BlockPairRefusalLocked and takes exactly this predicate.
// What is missing is the CATALOG half — there is no Spec.BlockRules
// field, no CardDef slot and nothing sets game.CatalogBlockRules, so a
// card has no way to declare one. That is issue #750's PR 4 (its own
// first cards are Prowler's Helm and Hungering Hydra, plus the
// enumerator-agreement tests a new block rule needs), not a three-card
// batch's to invent.
//
// The activated ability waits on a second piece even after that lands:
// it grants the rule to ANOTHER creature until end of turn, and
// Game.TurnScopedBlockRules — the field a rule with no permanent of
// its own hangs off — is named in block_rules.go as deliberately
// unbuilt. Both clauses are appended to docs/engine-seams.md under
// "Conditional blocking restrictions".
//
// When #750's card half lands, this file gains a Spec.BlockRules
// entry refusing any blocker that is not a Spirit, and the first
// caveat goes.
func init() {
	Register(Spec{
		OracleID:     "a620a765-97ba-4687-acfd-4dec7da75d9f",
		Name:         "Departed Deckhand",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"\"Can't be blocked except by Spirits\" isn't implemented — any creature can block this one.",
			"Its {3}{U} ability isn't implemented — there is no way to activate it.",
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, SelfTargetedByASpell,
				"Departed Deckhand — sacrifice it",
				SacrificeThisIfStillOnBattlefield),
		},
	})
}
