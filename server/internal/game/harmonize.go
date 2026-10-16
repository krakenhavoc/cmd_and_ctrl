package game

import "github.com/google/uuid"

// harmonize.go — #3931 (ADR 0135 amendment 2026-10-16): harmonize, CR
// 702.180.
//
// "Harmonize [cost]" means "You may cast this card from your graveyard
// by paying [cost] and tapping up to one untapped creature you control
// rather than paying this spell's mana cost", "If you cast this spell
// using its harmonize ability, its total cost is reduced by an amount
// of generic mana equal to the tapped creature's power" and "If the
// harmonize cost was paid, exile this card instead of putting it
// anywhere else any time it would leave the stack" (CR 702.180a).
//
// Three existing pieces, one new flag. The place and the exile are
// flashback's (FromZone graveyard, ExileOnLeavingStack). The tap is
// the alternative cost's TapOthers component (ADR 0135 §1), widened to
// "up to one" (TapOthersCost.UpTo), so the candidate walk, the
// validator and the payer are Orim's Cure's. The reduction is emerge's
// shape (ADR 0135 §4) with power for mana value: CR 702.180b chooses
// the creature as the cost is chosen (601.2b) and taps it as the total
// is paid (601.2h), so CR 601.2f totals the cost while it is still on
// the battlefield, and the pricer reads its power here.

// AltCostKeyHarmonize is harmonize's offer key: on the wire as
// `alternative_cost`, on StackItem.AltCost, and on a cast permission's
// AltCostKey when an effect grants harmonize (Songcrafter Mage). An
// on-disk identity: never renamed.
const AltCostKeyHarmonize = "harmonize"

// HarmonizeTapLabel is the clause harmonize taps, without the verb.
const HarmonizeTapLabel = "up to one untapped creature you control"

// HarmonizeTapCost is the card component every harmonize offer carries,
// printed or granted: up to one untapped creature the caster controls
// (CR 702.180a). A cost, not a target (CR 601.2h), so hexproof never
// narrows it, and not the {T} symbol, so a creature that arrived this
// turn may pay (CR 302.6). Untapped and "you control" are the tap-others
// validator's own checks.
func HarmonizeTapCost() *TapOthersCost {
	return &TapOthersCost{
		Count: 1,
		UpTo:  true,
		Filter: &TargetSpec{
			Mode:  "permanent",
			Label: HarmonizeTapLabel,
			Zones: []ZoneKind{ZoneBattlefield},
			CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 1,
		},
		Label: HarmonizeTapLabel,
	}
}

// altTappedPowerLocked is CostQuery.AltTappedPower for one announcement:
// the power of the one creature the claimed harmonize offer names in
// AltCostIDs, or 0 when the cast claims no such offer or names no
// creature (a legal payment that reduces nothing). More than one is
// refused by the validator; this only keeps a malformed announcement
// from being priced as a discount. Caller must hold g.mu.
func (g *Game) altTappedPowerLocked(playerID uuid.UUID, card Card, params CastSpellParams) int {
	if len(params.AltCostIDs) != 1 {
		return 0
	}
	alt := g.claimedAltCostLocked(playerID, card, params)
	if alt == nil || !alt.ReducedByTappedPower {
		return 0
	}
	return g.AltTappedPowerForEffect(params.AltCostIDs[0])
}

// AltTappedPowerForEffect is the generic mana harmonize reduces its
// cost by when `id` is the creature tapped (CR 702.180a): its power as
// it stands on the battlefield, with every layer and counter, clamped
// at zero (a reduction never adds mana). Zero when `id` names no
// permanent.
//
// Exported for the readers that price a payment before announcing it
// — the bot enumerator's per-payment price and the view's per-candidate
// one — so all three read the number the payment is charged with.
// Caller must hold g.mu.
func (g *Game) AltTappedPowerForEffect(id uuid.UUID) int {
	c := findBattlefieldCard(g, id)
	if c == nil {
		return 0
	}
	return c.CurrentPower()
}
