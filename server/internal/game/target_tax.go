package game

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// target_tax.go — "Spells your opponents cast that target this creature
// cost an additional 3 life to cast" (Terror of the Peaks, #3860), CR
// 601.2f's additional costs charged by a permanent OTHER than the spell.
//
// # Why it is not a CostModifier
//
// CostModifier (ADR 0048) prices GENERIC MANA: its kinds are increase,
// reduction and floor, all arithmetic on the mana cost. "An additional 3
// life" is a non-mana component, and it is not the spell's own
// AdditionalCost either (that belongs to the card being cast and is
// declared on it). It is the AttackTax family one verb over (attack_tax.go,
// ADR 0080): a COST someone else's permanent adds to a game action,
// keyed on what the action NAMES. An attack tax is keyed on the defending
// player; a target tax is keyed on what the spell targets, so it is
// asked after the targets are announced (CR 601.2c) and before they are
// paid for (CR 601.2h). Same shape, same reasons: a catalog-contributed
// static, a struct of hooks a card file writes as a literal, collected
// through a catalog key, gated by a designation in exactly one accessor.
//
// # One question, four askers
//
// castTargetTaxLocked answers "how much extra life does casting THIS
// spell at THESE targets cost, for this caster?" and four things read it,
// so none of them can disagree about the price (#544):
//
//   - CastSpell, which refuses a caster who cannot pay it (CR 119.4 --
//     they may pay exactly their life total, and a life total that can't
//     change pays nothing, so it is a refusal) and pays it with the rest
//     of the additional costs, the spell already on the stack, so a Blood
//     Artist watching the life loss triggers above it;
//   - the legal-move enumerator, which prices each target set and drops
//     a cast the caster cannot pay -- a player who can't pay can't target
//     it, so neither the picker nor the bot is offered it;
//   - the view, which names each legal target's tax (LegalTargetsView.
//     LifeTaxes) so the picker can say "3 life to target this" before the
//     click that commits the cast;
//   - the bot, through the enumerator's MoveCost.
//
// # What it taxes
//
// Spells only: "spells your opponents cast". An ability that targets the
// permanent pays nothing here (that is ward's business, ADR 0038), and a
// copy of a spell is not cast (CR 707.10), so it owes nothing. A source
// taxes a given spell once however many of its targets name it (the
// cost is added to the SPELL), and two sources each add their own.

// TargetTaxQuery is everything a target tax may look at. Passed by
// value, for CostQuery's reason: the tax is consulted inside the cast
// path under the write lock.
type TargetTaxQuery struct {
	// Game is the game the cast is happening in. Read-only.
	Game *Game

	// Source is the permanent contributing the tax, live off the
	// battlefield.
	Source Card

	// Caster is the player casting the spell -- "your opponents" is
	// measured against Source.Controller.
	Caster uuid.UUID

	// Spell is the spell being cast, as it stands at announce.
	Spell Card

	// Target is one permanent the spell targets. The tax is asked once
	// per distinct targeted permanent until one of them applies.
	Target Card
}

// TargetTax is one "spells <whom> cast that target <what> cost an
// additional <N> life to cast" static ability contributed by a permanent
// on the battlefield. Declared as a struct of hooks, as AttackTax and
// CostModifier are: a card file writes a literal, not a type.
type TargetTax struct {
	// Label is the clause as printed ("Spells your opponents cast that
	// target this creature cost an additional 3 life to cast."), for
	// the picker's note, the refusal and the move log.
	Label string

	// Life is the life the spell's caster pays, once, if the tax
	// applies. Must be positive; Register refuses otherwise.
	Life int

	// AppliesTo decides whether this tax applies to a spell `q.Spell`
	// cast by `q.Caster` at the permanent `q.Target`. Required: a tax
	// that names no whom/what is a tax on every spell, which no printed
	// card says. Read-only, under g.mu.
	AppliesTo func(q TargetTaxQuery) bool

	// ActiveWhen is the CR 716 / 719 / 721 / 709.5 designation gate:
	// this tax applies only while its source has the designation named.
	// The zero value is "no gate". Evaluated in TargetTaxesForCard and
	// NOWHERE ELSE, as AttackTax.ActiveWhen is.
	ActiveWhen Designation
}

// CatalogTargetTaxes returns the target taxes a catalog entry declares.
// Wired to the card catalog in carddef.go; nil in a test that builds
// cards by hand, which is the same contract every other catalog hook
// has.
var CatalogTargetTaxes func(key string) []TargetTax

// TargetTaxesForCard is the target taxes a permanent contributes right
// now: its catalog entry's taxes, minus those a CR 613.1f
// ability-removing effect took away (CatalogAbilityKey) and minus those
// whose designation gate is unsatisfied. The one accessor.
func TargetTaxesForCard(c Card) []TargetTax { return targetTaxesOf(&c) }

// targetTaxesOf is TargetTaxesForCard without the copy: Card is over a
// kilobyte and this is asked per permanent per cast (#1498).
func targetTaxesOf(c *Card) []TargetTax {
	if CatalogTargetTaxes == nil {
		return nil
	}
	key := catalogAbilityKeyOf(c)
	if key == "" {
		return nil
	}
	return activeOnly(c, CatalogTargetTaxes(key), func(t TargetTax) Designation {
		return t.ActiveWhen
	})
}

// TargetTaxLine is one tax charged by one permanent: the per-source
// breakdown behind a cast's total, for the refusal's sentence.
type TargetTaxLine struct {
	// Source is the permanent that charges it.
	Source uuid.UUID
	// SourceName is its name, as it stands now.
	SourceName string
	// Label is the clause as printed.
	Label string
	// Life is what this line charges.
	Life int
}

// castTargetTaxLinesLocked lists every tax charged to the spell `spell`
// cast by `caster` at `targets`: each battlefield permanent's target
// taxes, charged once when any distinct targeted permanent satisfies the
// tax's AppliesTo. Nil when nothing is charged, which is nearly every
// cast.
//
// Caller must hold g.mu (read or write) with fresh layers. Reads only.
func (g *Game) castTargetTaxLinesLocked(caster uuid.UUID, spell Card, targets []TargetRef) []TargetTaxLine {
	if g == nil || g.Battlefield == nil || CatalogTargetTaxes == nil || len(targets) == 0 {
		return nil
	}
	var targeted []*Card
	seen := make(map[uuid.UUID]bool, len(targets))
	for _, t := range targets {
		if t.Kind != TargetCard || seen[t.ID] {
			continue
		}
		seen[t.ID] = true
		if c := findBattlefieldCard(g, t.ID); c != nil {
			targeted = append(targeted, c)
		}
	}
	if len(targeted) == 0 {
		return nil
	}
	var lines []TargetTaxLine
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		for _, tax := range targetTaxesOf(src) {
			if tax.Life <= 0 || tax.AppliesTo == nil {
				continue
			}
			for _, tgt := range targeted {
				if tax.AppliesTo(TargetTaxQuery{Game: g, Source: *src, Caster: caster, Spell: spell, Target: *tgt}) {
					lines = append(lines, TargetTaxLine{Source: src.InstanceID, SourceName: src.Name, Label: tax.Label, Life: tax.Life})
					break
				}
			}
		}
	}
	return lines
}

// castTargetTaxLocked is the total life castTargetTaxLinesLocked
// charges. Caller must hold g.mu.
func (g *Game) castTargetTaxLocked(caster uuid.UUID, spell Card, targets []TargetRef) int {
	n := 0
	for _, l := range g.castTargetTaxLinesLocked(caster, spell, targets) {
		n += l.Life
	}
	return n
}

// CastTargetTaxForEffect is the extra life a spell cast at `targets`
// costs `caster` because of what it targets (CR 601.2f), on the
// *ForEffect surface for callers already under g.mu: the legal-move
// enumerator and the protocol projection. Zero for nearly every cast.
func (g *Game) CastTargetTaxForEffect(caster uuid.UUID, spell Card, targets []TargetRef) int {
	return g.castTargetTaxLocked(caster, spell, targets)
}

// TargetTaxOnTargetForEffect is the extra life a spell of `spell`'s kind
// cast by `caster` would cost for naming the permanent `target` and
// nothing else: the figure the picker shows beside that candidate. Zero
// when naming it is free. Caller must hold g.mu.
func (g *Game) TargetTaxOnTargetForEffect(caster uuid.UUID, spell Card, target uuid.UUID) int {
	return g.castTargetTaxLocked(caster, spell, []TargetRef{{Kind: TargetCard, ID: target}})
}

// TargetTaxesExistForEffect reports whether any permanent on the
// battlefield contributes a target tax right now, so a caller about to
// price a whole legal set (the view's picker, the enumerator's target
// expansion) can skip the work on the boards that have none -- which is
// nearly every board. Caller must hold g.mu.
func (g *Game) TargetTaxesExistForEffect() bool {
	if g == nil || g.Battlefield == nil || CatalogTargetTaxes == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		if len(targetTaxesOf(&g.Battlefield.Cards[i])) > 0 {
			return true
		}
	}
	return false
}

// targetTaxRefusal is the error a cast is refused with when its caster
// cannot pay the tax: it wraps ErrCantPayTargetTax with the sentence a
// player reads, naming the permanent and the life.
func targetTaxRefusal(lines []TargetTaxLine, total int) error {
	names := make([]string, 0, len(lines))
	for _, l := range lines {
		names = append(names, l.SourceName)
	}
	return fmt.Errorf("%w: targeting %s costs %d more life", ErrCantPayTargetTax, strings.Join(names, " and "), total)
}
