package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Psionic Ritual — Sorcery {4}{U}{U}:
//
//	"Replicate—Tap an untapped Horror you control. (When you cast this
//	 spell, copy it for each time you paid its replicate cost. You may
//	 choose new targets for the copies.)
//	 Exile target instant or sorcery card from a graveyard and copy it.
//	 You may cast the copy without paying its mana cost.
//	 Exile Psionic Ritual."
//
// Replicate is the engine's keyword (replicate.go, CR 702.56a) with a
// cost that taps instead of paying energy: ReplicateTapping (#4092),
// one untapped Horror per payment, named on cast_spell's teamwork_ids
// and tapped with the spell on the stack. Tapping is not the {T} symbol,
// so a Horror that arrived this turn pays (CR 302.6), and a cost does
// not target, so a hexproof one does too. The cast trigger copies the
// spell once per payment and offers each copy a new target (CR 707.10c).
//
// The resolution is Mizzix's Mastery's (CopyAndCast, ADR 0066's card-copy
// amendment) over ANY graveyard: the card is exiled through the CR 614
// window, and only a card that really reached exile is copied (CR
// 707.12). The copy is offered to the caster for free, flash-timed
// (CR 608.2g); one not cast by their next priority pass ceases to exist,
// and one that is cast ceases to exist as it resolves (CR 707.10a). The
// card itself stays in exile.
//
// "Exile Psionic Ritual" is the last instruction: the card leaves the
// stack to exile rather than its graveyard. A replicate copy is not a
// card and simply ceases to exist (CR 707.10a). A cast whose target left
// in response fizzles (CR 608.2b) and goes to the graveyard as any
// fizzled spell does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c3e6980-97fe-4e15-93a0-368280da8608",
		Name:         "Psionic Ritual",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target instant or sorcery card from a graveyard",
			Or(Instant(), Sorcery())),
		OptionalCosts: []game.AdditionalCost{ReplicateTapping("an untapped Horror you control", 10, OfCreatureType("Horror"))},
		Triggered:     []game.TriggeredAbility{Replicate()},
		OnResolve:     nghaPsionicRitualResolve,
	})
}

// nghaPsionicRitualResolve exiles the still-legal target, copies what
// reached exile with a free-cast offer, then exiles the spell.
func nghaPsionicRitualResolve(item *game.StackItem, ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	return exileCopyCastThenExileSelf(item, ctx, ids)
}
