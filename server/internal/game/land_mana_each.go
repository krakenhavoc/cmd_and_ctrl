package game

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// land_mana_each.go — ADR 0200, #3952: "<player> activates a mana ability
// of each land they control" (Drain Power, Pygmy Hippo), and "loses all
// unspent mana".
//
// The instruction makes a player activate mana abilities while a spell or
// ability resolves, which CR 605.3a allows ("whenever a rule or effect asks
// for a mana payment, even if it's in the middle of ... resolving"), through
// the ordinary activation path (activateManaAbilityLocked). Nothing here
// pays a cost: the effect says which player activates, and that player
// chooses which mana ability of each land, and how to answer the ability's
// own choices (Drain Power's 2004-10-04 rulings). A land with no mana
// ability that player can activate right now, a tapped land among them,
// adds nothing (CR 101.3).
//
// The colours a picking ability adds are named up front
// (ManaAbilityParams.Colors, #1443), so each (ability, colours) pair is one
// option and no mana_pick is queued behind it. The card side
// (effects.ActivateAManaAbilityOfEachLand) asks the player one option_pick
// per land that has more than one option, then calls
// LoseAllUnspentManaForEffect.

// LandManaOption is one way to activate a mana ability of one land: the
// ability row, the ref it had when the option was built, and the colours
// its picking slots add, named up front.
type LandManaOption struct {
	Land   uuid.UUID
	Index  int
	Ref    string
	Colors []string
	// Label is the option as the chooser reads it: "Plains: {T}: Add {W}."
	// with the named colours after it.
	Label string
}

// maxLandManaOptions bounds the options one land offers: a land whose
// ability adds several mana of any colour multiplies them, and a prompt
// with hundreds of rows helps nobody. The first ones in the slot order
// are kept (each slot's list is identity-first, #843).
const maxLandManaOptions = 64

// LandManaOptionsForEffect lists the ways `player` can activate a mana
// ability of `land` right now, and whether the land has a mana ability
// whose mana cost the pool cannot pay yet (a filter land's "{W/U}, {T}"),
// which a later land's mana might pay.
//
// The gates are the ones activateManaAbilityLocked checks first: the
// land is on the battlefield under `player`'s control, its mana abilities
// can be activated, the ability functions there, is not exhausted, meets
// its condition and the activation gate, its {T} can be paid, and its
// life, energy and mill components can be paid. An ability whose cost asks
// for another choice (sacrifice, tap, discard or exile another object) is
// left out: none of the lands that print one is a land the effect is
// about, and asking it would need a second prompt per land. The
// activation itself remains the judge; a refused option is dropped by the
// caller.
//
// Caller must hold g.mu.
func (g *Game) LandManaOptionsForEffect(player, land uuid.UUID) (opts []LandManaOption, waitingOnMana bool) {
	g.RecomputeLayersIfStaleLocked()
	card, ok := g.battlefieldCardLocked(land)
	p := g.playerByIDLocked(player)
	if !ok || p == nil || card.Controller != player || !card.IsLand() || !CanActivateManaAbilities(card) {
		return nil, false
	}
	abilities, origins := manaAbilityRows(card, true)
	identity := commanderIdentityFor(g, p)
	for idx := range abilities {
		ab := abilities[idx]
		if !ManaAbilityFunctionsFromZone(ab, ZoneBattlefield) {
			continue
		}
		if ab.SacrificeOther != nil || !ab.TapOthers.Empty() || ab.DiscardCards != nil || ab.ExileCards != nil ||
			!ab.ExilePermanents.Empty() || ab.ExileSelf {
			continue
		}
		if ab.TapCost && (card.Tapped || g.ActivationSummoningSickLocked(card, player)) {
			continue
		}
		if ab.SacrificeCost && g.CantBeSacrificedForEffect(land) {
			continue
		}
		if g.ManaAbilityExhausted(player, land, ab) || g.BorrowedRowSpentLocked(land, origins.At(idx)) {
			continue
		}
		if ab.Condition != nil && !ab.Condition(g, player, land) {
			continue
		}
		if g.ActivationGateLocked(player, *card, ZoneBattlefield, ActivationAbility{Label: ab.Label, Mana: true}) != nil {
			continue
		}
		if !g.CanPayLifeLocked(p, ab.LifeCost) || EnergyShortfall(p, ab.EnergyCost) != nil || g.MillCostShortfall(player, ab.MillCost) != nil {
			continue
		}
		if ab.ManaCost != "" {
			priced, err := g.ManaAbilityManaCostForEffect(player, *card, ab)
			if err != nil {
				continue
			}
			spend := ManaSpendForAbilityRow(*card, origins.At(idx))
			if !p.ManaPool.CanPayFor(g.costAsPaidByLocked(player, spend, priced, 0), 0, spend) {
				waitingOnMana = true
				continue
			}
		}
		slots, err := manaAbilitySlotsLocked(g, player, land, &ab, g.maxCounterPaymentLocked(player, land, ab.RemoveCounters))
		if err != nil {
			continue
		}
		base := card.Name + ": " + ab.Label
		var lists [][]string
		var groups []int
		if hasMultiOptionSlot(slots) {
			lists, groups = pickingSlotGroups(slots, identity, ab.NarrowToCommanderIdentity)
		}
		for _, colors := range colorCombinations(lists, groups, maxLandManaOptions-len(opts)) {
			label := base
			if len(colors) > 0 {
				label += " (" + manaSymbols(colors) + ")"
			}
			opts = append(opts, LandManaOption{
				Land: land, Index: idx, Ref: origins.Ref(idx), Colors: colors, Label: label,
			})
		}
		if len(opts) >= maxLandManaOptions {
			break
		}
	}
	return opts, waitingOnMana
}

// colorCombinations is every way to answer the picking slots `lists`, one
// colour per list, never naming one colour twice inside a different-colours
// group (groups[i] >= 0, #2558). No lists is one answer with no colours. At
// most `limit` answers, in slot order.
func colorCombinations(lists [][]string, groups []int, limit int) [][]string {
	if limit <= 0 {
		return nil
	}
	if len(lists) == 0 {
		return [][]string{nil}
	}
	var out [][]string
	cur := make([]string, 0, len(lists))
	var walk func(i int)
	walk = func(i int) {
		if len(out) >= limit {
			return
		}
		if i == len(lists) {
			out = append(out, append([]string(nil), cur...))
			return
		}
		for _, c := range lists[i] {
			if groups[i] >= 0 && repeatsInGroup(cur, groups, groups[i], c) {
				continue
			}
			cur = append(cur, c)
			walk(i + 1)
			cur = cur[:len(cur)-1]
		}
	}
	walk(0)
	return out
}

func repeatsInGroup(cur []string, groups []int, group int, c string) bool {
	for j, prev := range cur {
		if groups[j] == group && prev == c {
			return true
		}
	}
	return false
}

func manaSymbols(colors []string) string {
	var b strings.Builder
	for _, c := range colors {
		b.WriteString("{" + c + "}")
	}
	return b.String()
}

// ActivateLandManaForEffect activates the option's mana ability for
// `player`, with its colours named up front. It is not a priority action
// (CR 117.3c does not apply: the effect instructed it). An option whose
// ability moved since it was built is refused with ErrStaleAbilityRef, and
// every other refusal comes before anything is paid. Caller must hold
// g.mu.
func (g *Game) ActivateLandManaForEffect(player uuid.UUID, opt LandManaOption) error {
	if opt.Land == uuid.Nil {
		return fmt.Errorf("%w: no land", ErrInvalidParam)
	}
	return g.activateManaAbilityLocked(player, opt.Land, opt.Index, ManaAbilityParams{
		Ref:    opt.Ref,
		Colors: opt.Colors,
	})
}

// LandsControlledForEffect is the instance IDs of the lands `player`
// controls, in battlefield order. Caller must hold g.mu.
func (g *Game) LandsControlledForEffect(player uuid.UUID) []uuid.UUID {
	g.RecomputeLayersIfStaleLocked()
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// LoseAllUnspentManaForEffect is "<player> loses all unspent mana": the
// whole pool is emptied, mana that was there before the effect included,
// and what was lost is returned so the effect can count it (Pygmy Hippo)
// or add it to another pool (Drain Power). No "you don't lose this mana"
// effect applies: those keep mana as steps and phases end (CR 106.4,
// mana_keep.go), and this is an instruction to lose it.
//
// Emits EventManaPoolEmptied for a pool that held anything. Caller must
// hold g.mu.
func (g *Game) LoseAllUnspentManaForEffect(player uuid.UUID) []ManaToken {
	p := g.playerByIDLocked(player)
	if p == nil || len(p.ManaPool) == 0 {
		return nil
	}
	lost := append([]ManaToken(nil), p.ManaPool...)
	n := p.ManaPool.EmptyPool()
	g.EmitEvent(Event{Kind: EventManaPoolEmptied, Actor: player, Amount: n})
	return lost
}

// NextMainPhaseThisTurnForEffect is "your next main phase this turn":
// the first step of the next main phase still to begin in this turn's
// plan, and that phase's id, or false when the turn has none left. A main
// phase in progress is not "next". Caller must hold g.mu (write).
func (g *Game) NextMainPhaseThisTurnForEffect() (Step, int, bool) {
	g.ensureTurnPlanLocked()
	for _, ps := range g.TurnPlan {
		if PhaseKindOf(PhaseOf(ps.Step)) != PhaseKindMain || ps.PhaseID == g.Turn.PhaseID {
			continue
		}
		return ps.Step, ps.PhaseID, true
	}
	return "", 0, false
}
