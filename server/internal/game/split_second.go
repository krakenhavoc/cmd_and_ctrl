package game

import "github.com/google/uuid"

// split_second.go — CR 702.61, read off the card (#1519).
//
// The rule was built in S13.1 and never switched on. Every consumer
// already existed and already asked one cache, Game.SplitSecondActive:
// CastSpell, both activation paths and ActivateLoyalty refuse with
// ErrSplitSecondActive, the enumerator's castMoves and activatedMoves
// return early, suspend's special-action window shuts, and the wire
// carries `split_second_active`. What was missing was the WRITER: the
// only thing that ever stamped StackItem.SplitSecond was the sandbox
// `split_second` flag on cast_spell, which no client sends. So a
// Krosan Grip on the stack could be answered like any other spell.
//
// The keyword now joins canonicalKeywords, so the deck importer stops
// filtering Scryfall's "Split second" out of Card.Keywords, and the
// cast path stamps the stack item from the spell's own keywords
// through castHasSplitSecond below. The sandbox flag is routed through
// the same function rather than beside it, so there is one writer and
// one answer.
//
// What split second does NOT stop (CR 702.61b), and the engine
// already honoured before any card could turn it on:
//
//   - mana abilities (ActivateManaAbility has no split-second check,
//     and ActivationTimingOpenLocked returns before its own);
//   - special actions — foretell, turning a face-down permanent face
//     up (SpecialActionTimingOKLocked; suspend is the exception,
//     CR 702.62c, because its window is the card's casting window);
//   - triggered abilities, which trigger and go on the stack as
//     usual. A trigger whose resolution would CAST a spell cannot
//     (Gatherer ruling on Sudden Shock, 2006-09-25), which CastSpell's
//     own check answers for free.
//
// The cache is refreshed by recomputeSplitSecondLocked after every
// stack exit, so it ends the moment the spell resolves, is countered,
// fizzles or is exiled off the stack. See ADR 0007's amendment of
// 2026-09-24.

// KeywordSplitSecond is CR 702.61's canonical token — Scryfall's
// "Split second", lowercased.
const KeywordSplitSecond = "split second"

// castHasSplitSecond reports whether a spell being cast has split
// second, which is the fact StackItem.SplitSecond records. `card` is
// CastSpell's value copy with the chosen face already materialised,
// so a multi-face card answers for the face being cast.
//
// Two sources, one answer:
//
//   - the card's own keywords, read through HasKeyword: Card.Keywords
//     for a deck-imported card and CatalogPrintedKeywords for a
//     catalog entry that declares it (which is how a fixture with no
//     Scryfall data gets it). Off the battlefield there is no layer
//     engine, so the cast card carries no granted keyword; a GRANT
//     (Samut, Tyrant of Naktamun) is a battlefield static and is read
//     live by spellGrantedSplitSecondLocked, not stamped here (ADR
//     0007's amendment of 2026-10-14);
//   - the sandbox `split_second` flag on cast_spell (S13.1), kept so
//     a table playing a card the catalog has never heard of can still
//     say so by hand.
//
// A spell cast FACE DOWN has neither: CR 708.2 makes it a 2/2 with no
// text, and CR 708.4 is explicit that the face-down spell is that
// object while it is on the stack. The flag is refused too, because a
// face-down spell announcing split second would tell the table which
// card it is.
func castHasSplitSecond(card *Card, faceDown FaceDownKind, sandboxFlag bool) bool {
	if faceDown != FaceDownNone {
		return false
	}
	return sandboxFlag || HasKeyword(card, KeywordSplitSecond)
}

// SplitSecondGrant is a static ability of a permanent that gives split
// second to the spells its controller controls: "Instant and sorcery
// spells you control have split second." (Samut, Tyrant of Naktamun,
// #3502). A catalog declaration (Spec.SplitSecondGrants).
type SplitSecondGrant struct {
	// Spells narrows which spells the grant reaches, judged against the
	// card the spell is on the stack as (so a spell cast face down, a
	// 2/2 creature with no text, matches nothing but "creature").
	Spells PermissionFilter

	// YouCast narrows "spells you control" to "spells you CAST" (Shadow
	// the Hedgehog's "each spell you cast"): the grant reaches a spell
	// whose caster controls the source, whoever controls the spell now,
	// and never a copy, which was not cast (CR 707.10). The caster is
	// the player the spell was put on the stack under
	// (StackItem.BaseController, else Controller). False is "spells you
	// control", the spell's current controller (Samut).
	YouCast bool

	// SpentFrom, when nonzero, narrows the grant to a spell that mana
	// from a source of one of these kinds was spent to cast (#4724,
	// Shadow the Hedgehog's "if mana from an artifact was spent to cast
	// it"). Read off the spell's payment record (StackItem.Paid), whose
	// tokens keep a snapshot of what each mana's source was when it was
	// made (#1212), so a Treasure that sacrificed itself to pay still
	// counts. The record is complete when the spell lands on the stack
	// and never changes, so the condition is fixed at the cast; whether
	// the source is on the battlefield to grant it is still read live.
	// A copy spent nothing (CR 707.10) and never matches. Zero asks
	// nothing.
	SpentFrom ManaSourceKinds

	// Label is the oracle sentence the card's ability row shows.
	Label string
}

// reaches reports whether this grant, on a permanent controlled by
// sourceController, reaches the spell `item` (on the stack as `spell`).
func (gr SplitSecondGrant) reaches(sourceController uuid.UUID, item *StackItem, spell Card) bool {
	if gr.YouCast {
		if item.IsCopy || spellCasterOf(item) != sourceController {
			return false
		}
	} else if item.Controller != sourceController {
		return false
	}
	if !gr.Spells.Matches(spell) {
		return false
	}
	return gr.SpentFrom == 0 || item.Paid.Spent().From(gr.SpentFrom)
}

// spellCasterOf is the player a spell was put on the stack under: its
// base controller, which the first control change stamps, else its
// controller (ADR 0104).
func spellCasterOf(item *StackItem) uuid.UUID {
	if item.BaseController != uuid.Nil {
		return item.BaseController
	}
	return item.Controller
}

// CatalogSplitSecondGrants returns the split-second grants a battlefield
// permanent with the given catalog ABILITY key makes. Installed at init
// to read CardDef.SplitSecondGrants; a game-package test may stub it.
var CatalogSplitSecondGrants func(abilityKey string) []SplitSecondGrant

// spellGrantedSplitSecondLocked reports whether a permanent on the
// battlefield gives split second to the spell `item` right now.
//
// LIVE, not stamped: CR 611.3a says a continuous effect from a static
// ability "isn't locked in; it applies at any given moment to whatever
// its text indicates", and CR 702.61a makes split second a static
// ability that functions only while the spell is on the stack. So the
// grant exists exactly while a source is on the battlefield, under the
// spell's controller, with its abilities: a Samut sacrificed to an
// altar in response (a mana ability, which split second does not stop)
// takes split second off the spell, and one that arrives with a spell
// already on the stack gives it. StackItem.SplitSecond stays the
// spell's OWN fact (its printed keyword or the sandbox flag, written
// once at announce). recomputeSplitSecondLocked folds the two together
// and refreshSplitSecondForBoardChangeLocked re-runs it when the board
// changes under a spell.
//
// "Spells you control" is the spell's current controller
// (StackItem.Controller, CR 112.2), the CounterShieldYouControl
// reading: a copy you control is a spell you control. "Each spell you
// cast" (SplitSecondGrant.YouCast) is its caster instead, and a copy
// is never one.
//
// Caller must hold g.mu.
func (g *Game) spellGrantedSplitSecondLocked(item *StackItem) bool {
	if item == nil || item.Kind != StackItemSpell || CatalogSplitSecondGrants == nil || g.Battlefield == nil {
		return false
	}
	spell := g.spellCardOnStackLocked(item.ID)
	if spell == nil {
		return false
	}
	caster := spellCasterOf(item)
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		// Only the spell's controller's or its caster's permanents can
		// grant it anything; skip the catalog read for the rest.
		if src.Controller != item.Controller && src.Controller != caster {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, gr := range CatalogSplitSecondGrants(key) {
			if gr.reaches(src.Controller, item, *spell) {
				return true
			}
		}
	}
	return false
}

// SpellHasSplitSecondForEffect reports whether the spell on the stack
// with this ID has split second from any source, for a caller that
// already holds the lock: the view projection, which shows the chip on
// a spell Samut gave it to. Anything that is not a spell answers false.
func (g *Game) SpellHasSplitSecondForEffect(id uuid.UUID) bool {
	item := g.StackMeta[id]
	if item == nil || item.Kind != StackItemSpell {
		return false
	}
	return item.SplitSecond || g.spellGrantedSplitSecondLocked(item)
}

// refreshSplitSecondForBoardChangeLocked re-derives SplitSecondActive
// when the battlefield changes while something is on the stack, which
// is the only way a granted split second can start or stop without the
// stack itself changing. Cheap: one length test when nothing is on the
// stack or no grant is declared anywhere.
//
// Caller must hold g.mu.
func (g *Game) refreshSplitSecondForBoardChangeLocked() {
	if CatalogSplitSecondGrants == nil || len(g.StackMeta) == 0 {
		return
	}
	g.recomputeSplitSecondLocked()
}
