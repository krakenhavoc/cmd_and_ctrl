package heuristic

import (
	"encoding/json"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// windows.go is ADR 0126 §5: spend what would be wasted.
//
// Mana empties between steps (CR 106.4), and a tapped permanent untaps
// in its controller's untap step (CR 502.3). So in two windows a move
// whose only costs are mana and tapping spends nothing the bot would
// otherwise keep:
//
//   - the bot's own last main phase, stack empty: the turn's last
//     sorcery-speed window;
//   - the end step of the turn just before the bot's, stack empty: the
//     last window before everything it controls untaps. This is when a
//     player loots, or casts an end-of-turn tutor or cantrip.
//
// In those windows such a move needs only Config.LeftoverThreshold
// (decideGeneral). The windows are computed in newState
// (state.leftover, state.beforeMyUntap); this file says which moves
// qualify and what tapping a creature costs.

// tappedBlocker is what tapping one of the bot's untapped creatures
// costs outside ADR 0126 §5's timing: a blocker it no longer has. The
// pre-S66 flat price, everywhere.
const tappedBlocker = 0.3

// tapCreatureCost prices tapping the bot's untapped creature c to pay
// a cost, as a positive number.
//
// With Config.TapByTiming off it is tappedBlocker, the pre-S66 price.
// With it on (ADR 0126 §5):
//
//   - in the end step just before the bot's turn it is nothing: the
//     creature untaps before any opponent can attack;
//   - in the bot's own first main phase, for a creature that could
//     attack, it is the blocker plus the attack it gives up, priced as
//     station prices it (#759): Weights.Power per point of power;
//   - elsewhere it is tappedBlocker.
func (p *Policy) tapCreatureCost(st *state, c *protocol.CardView) float64 {
	if !p.cfg.TapByTiming {
		return tappedBlocker
	}
	if st.beforeMyUntap {
		return 0
	}
	if st.myTurn && st.step == "precombat_main" && couldAttack(c) {
		return tappedBlocker + st.w.Power*float64(c.Power)
	}
	return tappedBlocker
}

// tapFuelValue prices the bot's permanent with this instance ID as
// something a cost would TAP (ADR 0135 §1: a tap alternative cost): what
// tapping it costs (tapCreatureCost), since the permanent stays. A
// permanent the seat cannot see on the battlefield is Weights.Unknown,
// as fuelValue prices an unreadable card.
func (p *Policy) tapFuelValue(st *state, id string) float64 {
	c := st.bf[id]
	if c == nil {
		return st.w.Unknown
	}
	return p.tapCreatureCost(st, c)
}

// altCostTaps reports whether the alternative cost keyed `key` on c taps
// its permanents rather than spending them: the view stamps such an
// offer's candidates on `tap_options` (ADR 0135 §1).
func altCostTaps(c *protocol.CardView, key string) bool {
	if c == nil || key == "" {
		return false
	}
	for i := range c.AlternativeCosts {
		if ac := &c.AlternativeCosts[i]; ac.Key == key {
			return ac.TapOptions != nil
		}
	}
	return false
}

// couldAttack reports whether c could be declared as an attacker this
// turn as far as the card itself says: untapped, no summoning sickness
// (the view's flag already allows for haste), some power to swing
// with, and no defender.
func couldAttack(c *protocol.CardView) bool {
	return c != nil && isCreature(c) && !c.Tapped && !c.SummoningSick && c.Power > 0 && !hasKeyword(c, "defender")
}

// nonManaCastKeys are the cast_spell params that pay a cost with
// something other than mana: a card, a permanent, life or a choice of
// cost the policy cannot price as mana. A cast that sends any of them
// keeps the normal threshold in a leftover window.
var nonManaCastKeys = []string{
	"alt_cost_ids", "discard_ids", "sacrifice_ids", "delve_ids",
	"teamwork_ids", "blight_ids", "reveal_ids", "optional_costs", "cost_branch",
	"phyrexian_life",
}

// nonManaCastKeysBesidesSacrifice is nonManaCastKeys without the
// sacrifice, for landSacrificeIsNetMana (#2469).
var nonManaCastKeysBesidesSacrifice = func() []string {
	var out []string
	for _, k := range nonManaCastKeys {
		if k != "sacrifice_ids" {
			out = append(out, k)
		}
	}
	return out
}()

// nonManaActivateKeys are the activate_ability params that pay a cost
// with something other than mana and tapping. Crew, waterbend and
// station's tap_ids are taps, and are allowed.
var nonManaActivateKeys = []string{
	"sacrifice_ids", "counter_source_ids", "counter_counts", "counter_kind",
	"counter_kinds", "discard_ids", "exile_ids", "top_ids", "return_ids",
	"exile_permanent_ids", "reveal_ids", "phyrexian_life",
}

// leftoverEligible reports whether m gets LeftoverThreshold in the
// window the bot is in (ADR 0126 §5).
//
// In the end step before the bot's turn that is every move that costs
// only mana and taps (costsOnlyManaAndTaps).
//
// The bot's own second main phase is "the last sorcery-speed window of
// the turn" (§5), and there the window is for sorcery-speed moves only.
// Mana the bot leaves untapped after combat stays untapped through every
// opponent's turn and is still there in the end step before its own, so
// an instant, a flash spell or an instant-speed ability loses nothing by
// waiting for that step, and it keeps the mana up meanwhile. That is the
// end-of-turn Entomb or Vampiric Tutor §5 describes. The same holds for
// tapping a creature: tapped after combat, it stays tapped through every
// opponent's turn, so a move that taps one waits too. The bot loots with
// Mary Read and Anne Bonny at the end of the turn before its own, as §5's
// worked example prices it, rather than giving up a blocker. This is the
// reading of §5 that keeps its premise, "costs nothing the bot would
// otherwise keep", true in both windows.
func (p *Policy) leftoverEligible(st *state, m legal.Move) bool {
	if !st.beforeMyUntap && (p.tapsACreature(st, m) || p.instantSpeed(st, m)) {
		return false
	}
	return p.costsOnlyManaAndTaps(st, m)
}

// instantSpeed reports whether m could equally be made in the end step
// before the bot's turn: the cast of an instant or of a card with flash,
// or an activated ability with no sorcery-speed clause.
func (p *Policy) instantSpeed(st *state, m legal.Move) bool {
	switch m.Kind {
	case legal.KindCast:
		c := st.castSource(decode[castParams](m.Params).InstanceID)
		return c != nil && (isType(c, "instant") || hasKeyword(c, "flash"))
	case legal.KindActivate:
		ap := decode[activateParams](m.Params)
		if src := st.bf[ap.SourceCardID]; src != nil {
			if row := activatedRow(src, ap.AbilityIndex); row != nil {
				return !row.SorcerySpeed
			}
		}
	}
	return false
}

// costsOnlyManaAndTaps reports whether m costs only mana and tapping,
// plus the card itself for a cast (ADR 0126 §5). A move that also costs
// life, a sacrifice, a discard, another card or a counter keeps the
// normal threshold, and so does anything that is not a cast or an
// activation.
//
// Read off the move's declared cost, its params and the ability row,
// all of it on the wire (ADR 0033 §3).
func (p *Policy) costsOnlyManaAndTaps(st *state, m legal.Move) bool {
	if c := m.Cost; c != nil {
		if c.Life > 0 || c.Loyalty != 0 || len(c.Counters) > 0 || c.Hand > 0 || c.Energy > 0 {
			return false
		}
	}
	switch m.Kind {
	case legal.KindCast:
		if !paramsSet(m.Params, nonManaCastKeys) {
			return true
		}
		return p.landSacrificeIsNetMana(st, m)
	case legal.KindActivate:
		if paramsSet(m.Params, nonManaActivateKeys) {
			return false
		}
		ap := decode[activateParams](m.Params)
		// An ability activated from anywhere but the battlefield —
		// cycling, channel — spends the card it is printed on.
		src := st.bf[ap.SourceCardID]
		if src == nil {
			return false
		}
		row := activatedRow(src, ap.AbilityIndex)
		return row != nil && rowCostsOnlyManaAndTaps(row)
	}
	return false
}

// landSacrificeIsNetMana reports whether a cast's only non-mana cost is
// sacrificing lands that the cast's own declared purpose more than
// replaces (#2469). Harrow sacrifices a land and puts two onto the
// battlefield untapped: the move ends with more lands than it started
// with, all of them untapped, so like a move that costs mana and taps
// it spends nothing the bot would otherwise keep, and it gets the same
// leftover bar. A sacrifice with no `lands` purpose behind it, or one
// that does not replace what it sacrifices, keeps the normal bar, as
// does a sacrifice of anything but a land of the bot's own.
func (p *Policy) landSacrificeIsNetMana(st *state, m legal.Move) bool {
	cp := decode[castParams](m.Params)
	if paramsSet(m.Params, nonManaCastKeysBesidesSacrifice) {
		return false
	}
	n, ok := st.ownLandsSacrificed(cp.SacrificeIDs)
	if !ok {
		return false
	}
	card := st.castSource(cp.InstanceID)
	if card == nil {
		return false
	}
	return castPurpose(card, cp).lands > n
}

// tapsACreature reports whether paying m's cost taps one of the bot's
// untapped creatures: the source of a {T} ability, or a creature named
// to crew, to waterbend or to station.
func (p *Policy) tapsACreature(st *state, m legal.Move) bool {
	if m.Kind != legal.KindActivate {
		return false
	}
	ap := decode[activateParams](m.Params)
	if src := st.bf[ap.SourceCardID]; src != nil && isCreature(src) && !src.Tapped {
		if row := activatedRow(src, ap.AbilityIndex); row != nil && row.TapCost {
			return true
		}
	}
	for _, ids := range [][]string{ap.CrewIDs, ap.WaterbendIDs, ap.TapIDs} {
		for _, id := range ids {
			if c := st.bf[id]; c != nil && isCreature(c) && !c.Tapped {
				return true
			}
		}
	}
	return false
}

// rowCostsOnlyManaAndTaps reports whether an activated ability's
// printed cost has no component but mana and tapping.
func rowCostsOnlyManaAndTaps(r *protocol.ActivatedAbilityView) bool {
	return !r.SacrificeSelf && !r.DiscardSelf && !r.ExileSelf && !r.ReturnSelf &&
		r.SacrificeLabel == "" && r.LifeCost == 0 && r.LoyaltyCost == nil &&
		r.DiscardCostN == 0 && r.TopCostN == 0 && r.ExileCostN == 0 &&
		r.LibraryExileCostN == 0 && r.ReturnLabel == "" && r.ExilePermanentLabel == ""
}

// activatedRow finds the activated ability with this index on c.
func activatedRow(c *protocol.CardView, index int) *protocol.ActivatedAbilityView {
	for i := range c.ActivatedAbilities {
		if c.ActivatedAbilities[i].Index == index {
			return &c.ActivatedAbilities[i]
		}
	}
	return nil
}

// paramsSet reports whether the params object carries any of keys with
// a value other than null.
func paramsSet(raw json.RawMessage, keys []string) bool {
	if len(raw) == 0 {
		return false
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		// Unreadable params are not provably mana-only.
		return true
	}
	for _, k := range keys {
		if v, ok := obj[k]; ok && string(v) != "null" {
			return true
		}
	}
	return false
}

// untargetedSpell reports whether an instant or sorcery takes no
// announce-time target, on the card or on any of its modes. SpellFloor
// prices only those: a targeted spell is priced by what it points at.
func untargetedSpell(c *protocol.CardView) bool {
	if c.TargetMode != "" {
		return false
	}
	if c.Modes != nil {
		for i := range c.Modes.Options {
			if c.Modes.Options[i].TargetMode != "" {
				return false
			}
		}
	}
	return true
}
