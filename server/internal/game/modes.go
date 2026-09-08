package game

// modes.go — S20 sub-PR 4: modal spells (CR 700.2). A catalog card
// with "Choose one —" / "Choose two —" text declares a ModeSpec:
// the option labels, how many must be chosen, and — per option —
// the TargetSpec that governs the cast when that option is picked
// (Rakdos Charm's first mode targets a player, its second an
// artifact, its third nothing).
//
// Modes are announce-time choices (CR 601.2b) that ride
// CastSpellParams.Modes as option indexes and land on
// StackItem.Modes; the catalog reads them back through
// effects.Context.HasMode at resolution. The engine validates the
// choice at announce (distinct, in range, count within Min..Max)
// and derives the effective TargetSpec for the cast from the chosen
// options, so the S20 target legality gate and the CR 608.2b
// re-check apply to modal spells with no card-file changes.
//
// Sub-PR 4 supports at most ONE targeted option among the chosen
// modes. "Choose two" cards whose options each carry their own
// target (Cryptic Command, Kolaghan's Command) need per-mode target
// slots on the wire and in StackItem.Targets — that's the multi-
// target work S20 defers alongside "up to N" and divided damage.

// ModeOption is one bullet of a modal spell.
type ModeOption struct {
	// Label is the oracle bullet, shown verbatim in the picker:
	// "Exile target player's graveyard."
	Label string

	// Targets is the target clause this option adds to the cast.
	// Nil for untargeted options.
	Targets *TargetSpec
}

// ModeSpec declares a modal spell's choice.
type ModeSpec struct {
	// Prompt is the header line — "Choose one", "Choose two".
	Prompt string

	Options []ModeOption

	// Min / Max bound the number of options chosen. "Choose one" is
	// 1 / 1; "choose one or both" is 1 / 2; "choose two" is 2 / 2.
	Min, Max int
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

// validateModes is the announce-time gate for a modal cast: every
// index must name an option, none may repeat, and the count must
// fall within Min..Max. A nil spec (non-catalog card, or a catalog
// card that isn't modal) keeps the S13.1 free-form behaviour: any
// indexes the client sends are recorded for opponents to see and
// players resolve by hand.
func validateModes(spec *ModeSpec, modes []int) error {
	if spec == nil {
		return nil
	}
	seen := make(map[int]bool, len(modes))
	for _, m := range modes {
		if m < 0 || m >= len(spec.Options) || seen[m] {
			return ErrInvalidParam
		}
		seen[m] = true
	}
	if len(modes) < spec.Min || (spec.Max > 0 && len(modes) > spec.Max) {
		return ErrInvalidParam
	}
	return nil
}

// castTargetSpec is the TargetSpec that governs a cast of the card
// with the given modes chosen: the card-level spec when the card
// declares one, otherwise the single chosen option's spec. Returns
// (nil, nil) for an untargeted cast. Two chosen options that both
// target is unsupported in sub-PR 4 and reports ErrInvalidParam so
// the cast is rejected rather than silently mis-validated.
func castTargetSpec(oracleID string, modes []int) (*TargetSpec, error) {
	if spec := TargetSpecFor(oracleID); spec != nil {
		return spec, nil
	}
	ms := ModeSpecFor(oracleID)
	if ms == nil {
		return nil, nil
	}
	var out *TargetSpec
	for _, m := range modes {
		if m < 0 || m >= len(ms.Options) || ms.Options[m].Targets == nil {
			continue
		}
		if out != nil {
			return nil, ErrInvalidParam
		}
		out = ms.Options[m].Targets
	}
	return out, nil
}

// castTargetSpecForItem is the resolution-time twin: the spec the
// item was announced under, read back from its stamped modes.
// Unsupported combinations were rejected at announce, so the error
// is dropped here.
func castTargetSpecForItem(oracleID string, item *StackItem) *TargetSpec {
	if item == nil {
		return TargetSpecFor(oracleID)
	}
	spec, _ := castTargetSpec(oracleID, item.Modes)
	return spec
}
