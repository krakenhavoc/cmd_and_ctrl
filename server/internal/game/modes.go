package game

import "fmt"

// modes.go — modal spells, triggers and activated abilities
// (CR 700.2). A card with "Choose one —" / "Choose two —" text
// declares a ModeSpec: the option labels, how many must be chosen,
// whether the same one may be chosen more than once, and — per
// option — the target clause list that governs the announcement when
// that option is picked.
//
// ONE ModeSpec, THREE owners (#764, ADR 0065 §3): Spec.Modes for a
// spell, TriggeredAbility.Modes for a trigger, and
// effects.ActivatedAbility.Modes for an activated ability. The
// difference between them is only WHEN the choice is made — CR
// 601.2b at announce for a spell, CR 603.3c as the trigger is put on
// the stack (through the mode_pick prompt), CR 602.2b at activation
// — never what is chosen or how it is validated.
//
// Modes ride CastSpellParams.Modes / ActivateAbilityParams.Modes as
// option indexes and land on StackItem.Modes, which since #764 is a
// MULTISET IN ANNOUNCE ORDER: Mystic Confluence choosing its draw
// mode three times is [0, 0, 0], and each occurrence gets its own
// target group through TargetRef.Mode. The catalog reads the choice
// back through effects.Context.HasMode / ModeCount / Modes, or lets
// the engine dispatch each chosen bullet's ModeOption.Effect in
// announce order (CR 608.2c).

// ModeOption is one bullet of a modal spell or ability.
type ModeOption struct {
	// Label is the oracle bullet, shown verbatim in the picker:
	// "Exile target player's graveyard."
	Label string

	// Targets is the target clause LIST this option adds to the
	// announcement. Nil for untargeted options. A bullet with two
	// clauses ("destroy target artifact and target enchantment") is
	// one spec with a Rest, exactly as a card-level clause list is.
	Targets *TargetSpec

	// Effect is this bullet's body, run at resolution in announce
	// order (CR 608.2c), once per OCCURRENCE — a mode chosen twice
	// under CR 700.2d runs twice. `occurrence` is the index into
	// StackItem.Modes, so the effect reads its own target group
	// (effects.Context.ModeTargets) rather than the item's whole
	// list.
	//
	// Optional. A modal SPELL may leave it nil and branch inside its
	// OnResolve on ctx.HasMode(i) instead — the older shape, still
	// supported, still reading the same data. A modal trigger or
	// activated ability has no OnResolve to branch in, so declaring
	// the bullet's body here is what lets those two be modal without
	// each card hand-writing a switch.
	//
	// Runs under g.mu held in write mode: MUST NOT call public
	// locking mutators. Added by #764.
	Effect func(g *Game, item *StackItem, occurrence int) error

	// Cost is CR 702.172a's Spree: "As an additional cost to cast
	// this spell, pay the costs associated with those modes chosen
	// this way." Brace notation ("{1}{U}"), paid IF AND ONLY IF this
	// bullet is chosen, on top of the spell's own cost and every
	// OTHER chosen bullet's. Empty for an ordinary modal bullet —
	// every modal card before S45 leaves it unset and pays nothing
	// extra for choosing it.
	//
	// Mana only, on purpose: every printed Spree card (Three Steps
	// Ahead, Explosive Derailment, Insatiable Avarice, Caught in the
	// Crossfire and the rest of the cycle) prices its bullets in mana
	// alone. A card-shaped mode cost ("discard a card" per bullet)
	// would need the enumerator's cost-payment search widened to a
	// THIRD axis beside modes and targets, and the flat discard /
	// sacrifice wire lists extended to carry a per-mode slice —
	// neither of which any catalog card asks for yet. Left as a
	// string rather than an *AdditionalCost for the same reason: a
	// mostly-empty struct enforced-empty by a boot panic is worse
	// than the honest field, and swapping the type is the whole of
	// what the future PR would do. Added by ADR 0065's 2026-09-23
	// amendment.
	Cost string
}

// ModeSpec declares a modal spell's or ability's choice.
type ModeSpec struct {
	// Prompt is the header line — "Choose one", "Choose two".
	Prompt string

	Options []ModeOption

	// Min / Max bound the number of options chosen. "Choose one" is
	// 1 / 1; "choose one or both" is 1 / 2; "choose two" is 2 / 2;
	// "choose one or more" is 1 / len(Options); "choose up to one"
	// is 0 / 1.
	Min, Max int

	// Repeatable is CR 700.2d — "you may choose the same mode more
	// than once" (Mystic Confluence, Rootcast Apprenticeship). With
	// it set the announced list is a multiset and each occurrence
	// carries its own targets; without it, a repeated index is
	// ErrInvalidParam at announce, which is what every printed modal
	// card that does not say the words expects. Added by #764.
	Repeatable bool
}

// CatalogModeSpec is the catalog hook the effects package wires at
// init. Nil, or a nil return, means the card isn't modal.
var CatalogModeSpec func(oracleID string) *ModeSpec

// ModeSpecFor returns the structured modes for a card, or nil.
func ModeSpecFor(oracleID string) *ModeSpec {
	if CatalogModeSpec == nil || oracleID == "" {
		return nil
	}
	return CatalogModeSpec(oracleID)
}

// validateModes is the announce-time gate for a modal announcement:
// every index must name an option, none may repeat unless the spec
// is Repeatable (CR 700.2d), and the count must fall within
// Min..Max. A nil spec (non-catalog card, or a catalog card that
// isn't modal) keeps the S13.1 free-form behaviour: any indexes the
// client sends are recorded for opponents to see and players resolve
// by hand.
func validateModes(spec *ModeSpec, modes []int) error {
	if spec == nil {
		return nil
	}
	seen := make(map[int]bool, len(modes))
	for _, m := range modes {
		if m < 0 || m >= len(spec.Options) {
			return ErrInvalidParam
		}
		if seen[m] && !spec.Repeatable {
			return ErrInvalidParam
		}
		seen[m] = true
	}
	if len(modes) < spec.Min || (spec.Max > 0 && len(modes) > spec.Max) {
		return ErrInvalidParam
	}
	return nil
}

// ModeCount is how many times option i was chosen (CR 700.2d).
func ModeCount(modes []int, i int) int {
	n := 0
	for _, m := range modes {
		if m == i {
			n++
		}
	}
	return n
}

// castClauseSources is the (card-level spec, mode spec) pair that
// governs a cast of this card. Exactly one is non-nil for a catalog
// card that targets: effects.Register refuses a card declaring both,
// and the modal path derives its clause list per chosen occurrence
// through AnnouncedClauses.
func castClauseSources(oracleID string) (*TargetSpec, *ModeSpec) {
	if spec := TargetSpecFor(oracleID); spec != nil {
		return spec, nil
	}
	return nil, ModeSpecFor(oracleID)
}

// castTargetSpecForItem is the clause list a SPELL item on the stack
// was announced under, for the callers that still want one spec
// rather than the step list: the CR 707.10 copy re-target (which
// re-targets the first clause only — ADR 0065 "Out of scope") and
// the S22 alternative-cost rewrite.
//
// For a modal item it returns the first chosen option that targets,
// which is what the copy path did before per-mode targets existed.
func castTargetSpecForItem(oracleID string, item *StackItem) *TargetSpec {
	if item == nil {
		return TargetSpecFor(oracleID)
	}
	spec, ms := castClauseSources(oracleID)
	if spec == nil && ms != nil {
		for _, m := range item.Modes {
			if m >= 0 && m < len(ms.Options) && ms.Options[m].Targets != nil {
				spec = ms.Options[m].Targets
				break
			}
		}
	}
	spec = TargetSpecUnderAlternativeCost(spec, AlternativeCostByKey(oracleID, item.AltCost))
	// ADR 0089 §3: the same rewrite the announce path applied, so a
	// restored or copied promised Long River's Pull is still judged
	// under "target spell".
	return TargetSpecUnderOptionalCosts(spec, OptionalCostsFor(oracleID), item.Paid.OptionalCosts)
}

// runChosenModeEffectsLocked runs each chosen bullet's ModeOption
// Effect in announce order, once per occurrence (CR 608.2c, and CR
// 700.2d for a repeated mode). A nil Effect means the card resolves
// its modes inside its own OnResolve instead, which is the older and
// still-supported shape.
//
// Caller must hold g.mu in write mode.
func (g *Game) runChosenModeEffectsLocked(item *StackItem, ms *ModeSpec) {
	if item == nil || ms == nil {
		return
	}
	for occ, opt := range item.Modes {
		if opt < 0 || opt >= len(ms.Options) {
			continue
		}
		fn := ms.Options[opt].Effect
		if fn == nil {
			continue
		}
		if err := fn(g, item, occ); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				Actor:    item.Controller,
				Source:   item.SourceCardID,
				ErrorMsg: err.Error(),
			})
		}
	}
}

// choosableModeOptionsLocked lists the option indexes a chooser may
// pick right now: every option, minus those whose clause list cannot
// be filled from the current board (CR 603.3d — an option with no
// legal target is not on offer). `src` names the spell or ability
// doing the choosing, because whether a clause is fillable depends on
// the source under CR 702.16b as well as on the chooser (#662).
//
// Caller must hold g.mu.
func (g *Game) choosableModeOptionsLocked(src TargetSource, ms *ModeSpec) []int {
	if ms == nil {
		return nil
	}
	out := make([]int, 0, len(ms.Options))
	for i, o := range ms.Options {
		if o.Targets != nil && g.anyClauseUnfillableLocked(src, AnnouncedClauses(o.Targets, nil, nil)) {
			continue
		}
		out = append(out, i)
	}
	return out
}

// ModeLabels is the oracle bullet of each chosen mode of this item,
// in announce order and with repeats — what the stack overlay shows
// instead of a row of indexes. Empty for a non-modal item and for
// one whose ModeSpec could not be re-derived after a restore.
func (s *StackItem) ModeLabels() []string {
	if s == nil || s.modeSpec == nil || len(s.Modes) == 0 {
		return nil
	}
	out := make([]string, 0, len(s.Modes))
	for _, m := range s.Modes {
		if m < 0 || m >= len(s.modeSpec.Options) {
			continue
		}
		out = append(out, s.modeSpec.Options[m].Label)
	}
	return out
}

// ChoosableModeOptionsForEffect is choosableModeOptionsLocked on the
// *ForEffect surface: which of a ModeSpec's options the spell or
// ability `src` could take right now, for callers already under g.mu
// — the bot's move enumerator and the protocol projection.
func (g *Game) ChoosableModeOptionsForEffect(src TargetSource, ms *ModeSpec) []int {
	return g.choosableModeOptionsLocked(src, ms)
}

// EnoughChoosableModes reports whether `n` takeable options can fill a
// selection of Min. A Repeatable spec (CR 700.2d) needs only ONE — it
// may take the same bullet Min times, which is the whole point of
// Mystic Confluence's "choose three, you may choose the same mode
// more than once" with two of its bullets unfillable.
func EnoughChoosableModes(n int, ms *ModeSpec) bool {
	if ms == nil || n == 0 {
		return false
	}
	if ms.Repeatable {
		return true
	}
	return n >= ms.Min
}

// AddModeCostMana is CR 702.172a's Spree, the mana half: the sum of
// every CHOSEN mode's own Cost, added into `cost` at CR 601.2f beside
// AddOptionalCostMana (ADR 0073 §3) — the same point in the
// precedence and the same reason. Thalia taxes a Spree spell once for
// the whole announced total, and the cast path, the bot enumerator
// and the auto-tap preview must price one mode selection identically
// (#544), so there is exactly one walk of this arithmetic.
//
// A nil ModeSpec, or one where no chosen option carries a Cost,
// changes nothing — every modal card that predates S45 reprices to
// the exact number it always did.
//
// `modes` is a MULTISET in announce order (CR 700.2d): a repeated
// option pays its Cost once per occurrence, which is correct by
// construction since the loop walks the multiset rather than a set of
// distinct indices.
func AddModeCostMana(cost ParsedCost, ms *ModeSpec, modes []int) (ParsedCost, error) {
	if ms == nil {
		return cost, nil
	}
	for _, m := range modes {
		if m < 0 || m >= len(ms.Options) || ms.Options[m].Cost == "" {
			continue
		}
		add, err := ParseCost(ms.Options[m].Cost)
		if err != nil {
			return cost, fmt.Errorf("%w for %s: %w", ErrUnparseableCost, ms.Options[m].Label, err)
		}
		cost.Generic += add.Generic
		cost.Required = append(cost.Required, add.Required...)
		cost.XSlots += add.XSlots
		cost.HasPhyrexian = cost.HasPhyrexian || add.HasPhyrexian
		cost.HasSnow = cost.HasSnow || add.HasSnow
	}
	return cost, nil
}
