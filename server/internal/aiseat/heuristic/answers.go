package heuristic

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// answers.go is ADR 0142 decision 6: the bot reads the answers an
// activated row declares (`activated_abilities[].purpose.answers`) and
// prices them in a window with something on the stack. With
// PriceAnswerOnlyRows (#4318), a purely defensive row with nothing
// to answer is held, and a self-sacrifice in response pays for its source.
//
// Before it, every non-tap activation without a priced purpose was the
// flat ActivateBase (0.50), under InstantThreshold (1.50) in a response
// window, so the bot let a creature die with its regeneration mana up.
// Three prices change, all behind Config.PriceAnswers:
//
//   - SAVING A CREATURE. When an item on the stack would remove one of
//     the bot's creatures (an opponent's item that targets it, or a
//     declared sweep that takes it), an untargeted row on that creature
//     is priced by what it saves: the creature's value times the chance
//     the bot loses it (dyingAnyway: RemovalConfidence for a target, the
//     sweep's share for a sweep). `survive` (regenerate, indestructible,
//     a blink) saves it from anything; `protect` (hexproof, shroud,
//     protection) only from what targets it, so it is no answer to a
//     declared sweep (ADR 0142, amendment 2026-10-15).
//     `prevent` saves it only when every threat is declared damage that
//     can be prevented. `pump` saves it only when the row's declared
//     Pump keeps it alive against the declared damage and −N/−N. A row
//     already activated from that creature above the threats has done
//     its job, so the bot does not stack a second shield (a pump is the
//     exception: see repeatedPumpSave).
//   - A `value` ROW WAITS. A row declared `value` answers nothing, so
//     it is priced below passing while anything is on the stack: what
//     it buys is the same once the stack resolves. The exception is a
//     row whose cost spends a permanent the bot is about to lose, which
//     is ADR 0126 §7's use-it-or-lose-it move. Layer A's value-only
//     rule (rules.RuleValueOnly) passes those windows without asking.
//   - DYING ANYWAY, FOR EVERY OUTLET. ADR 0126 §7 priced only the
//     permanents a move's sacrifice_ids name. A purpose-priced row that
//     sacrifices its own source (Sakura-Tribe Elder), or a `sac_outlet`
//     row that exiles a creature (The Soul Stone), now pays the same
//     keep-chance price, and clears the same leftover bar when that is
//     all it costs besides mana and taps.
//
// Two more prices sit behind their own switches: a pump that needs
// several activations (Config.PriceRepeatedPumps, #2939 gap 2,
// repeatedPumpSave), and a row on one permanent that saves others, such
// as Selfless Spirit (Config.PriceSaveOthers, gap 3, answerSaveOthers).
//
// What it does not price: a combat grant (the combat planner's
// business, ADR 0142 decision 6 "Not covered").

// valueInResponse is what a declared-value activation is worth while
// the stack is not empty: below passing, so the bot waits for the stack.
const valueInResponse = -0.5

// defensiveAnswerOnly is a row whose entire declared benefit is a
// defensive answer. Priced excludes rows with a draw, pump, targeted
// effect or other payoff that can be useful without a stack item.
func defensiveAnswerOnly(pv *protocol.PurposeView) bool {
	if pv == nil || pv.Priced() || pv.Answers == nil || len(*pv.Answers) == 0 {
		return false
	}
	for _, answer := range *pv.Answers {
		switch answer {
		case "survive", "protect", "prevent", "restrict", "graveyard":
		default:
			return false
		}
	}
	return true
}

// threatKind is how a stack item would take a creature off the
// battlefield.
type threatKind uint8

const (
	// threatRemoval is an item that targets the creature with no
	// declared damage, or a destroy, exile, sacrifice or bounce sweep:
	// presumed removal, answered only by `protect`.
	threatRemoval threatKind = iota
	// threatDamage is declared damage: a damage entry on the target's
	// slot, or a damage sweep.
	threatDamage
	// threatMinus is a declared −N/−N sweep. Not damage, so it cannot
	// be prevented; a pump can still outgrow it.
	threatMinus
)

// threats is what the stack would do to one of the bot's creatures.
type threats struct {
	// chance is the largest chance any one threat removes it.
	chance float64
	// removal is set when any threat is not declared damage or −N/−N.
	removal bool
	// damage and minus sum the declared amounts.
	damage, minus int
	// unpreventable is set when some declared damage can't be
	// prevented (CR 615.12).
	unpreventable bool
	// source is the damage's source, for the kill test's deathtouch
	// and protection; nil when the view does not show it.
	source *protocol.CardView
	// top is the highest stack index of any threat (StackItems runs
	// bottom to top).
	top int
	// sweep is set when any threat is a declared sweep, which names no
	// target: hexproof, shroud and protection do nothing against it.
	sweep bool
	// fromSpell is set when some threat is a spell: declared damage, a
	// declared sweep, or a spell that targets the creature. A targeted
	// ability alone (a planeswalker's phase out, "can't be blocked") is
	// presumed removal like any targeted item, but a row that costs a
	// card to use (Selfless Spirit) does not take it for one.
	fromSpell bool
	// beyondDestroy is set when some declared sweep takes the creature
	// by something other than destroy or damage: exile, sacrifice,
	// bounce, −N/−N. Indestructible and persist are no answer to those.
	beyondDestroy bool
}

// threatsTo reads the stack for what would remove creature c, with conf
// the chance an opponent's targeted item is removal
// (Config.RemovalConfidence). ok is false when nothing on it would.
func (st *state) threatsTo(c *protocol.CardView, conf float64) (threats, bool) {
	var t threats
	found := false
	add := func(i int, kind threatKind, chance float64, amount int, src *protocol.CardView, unpreventable bool) {
		found = true
		t.top = max(t.top, i)
		t.chance = max(t.chance, chance)
		switch kind {
		case threatRemoval:
			t.removal = true
		case threatDamage:
			t.damage += amount
			t.unpreventable = t.unpreventable || unpreventable
			if t.source == nil {
				t.source = src
			}
		case threatMinus:
			t.minus += amount
		}
	}
	cantPrevent := len(st.view.DamageCantBePrevented) > 0
	for i := range st.view.StackItems {
		it := &st.view.StackItems[i]
		card := st.stack[it.SourceCardID]
		src := card
		if src == nil {
			src = st.bf[it.SourceCardID]
		}
		if it.Controller != st.me {
			for _, ref := range it.Targets {
				if ref.ID != c.InstanceID {
					continue
				}
				dmg := 0
				if it.Kind == "spell" && card != nil {
					cp := castParams{AlternativeCost: it.AltCost, Modes: it.Modes}
					if e := castEntryFor(card, cp, targetRef{Slot: ref.Slot, Mode: ref.Mode}); e != nil {
						dmg = e.Damage
					}
				}
				if dmg > 0 {
					t.fromSpell = true
					add(i, threatDamage, conf, dmg, src, cantPrevent || it.DamageCantBePrevented)
				} else {
					t.fromSpell = t.fromSpell || it.Kind == "spell"
					add(i, threatRemoval, conf, 0, nil, false)
				}
			}
		}
		if it.Kind != "spell" || card == nil {
			continue
		}
		for _, p := range stackItemPurposes(card, it) {
			s := p.Sweep
			if s == nil {
				continue
			}
			share := st.sweepRemoves(c, s, it.Controller, it.XValue)
			if share <= 0 {
				continue
			}
			t.sweep = true
			t.fromSpell = true
			n := s.Amount
			if s.AmountIsX {
				n = it.XValue
			}
			switch s.How {
			case "damage":
				add(i, threatDamage, share, n, src, cantPrevent || it.DamageCantBePrevented)
			case "minus":
				t.beyondDestroy = true
				add(i, threatMinus, share, n, nil, false)
			default:
				t.beyondDestroy = t.beyondDestroy || s.How != "destroy"
				add(i, threatRemoval, share, 0, nil, false)
			}
		}
	}
	if !found {
		return t, false
	}
	// Declared damage and −N/−N alone that the creature survives are no
	// threat at all.
	if !t.removal && !dies(c, t, nil) {
		return t, false
	}
	return t, true
}

// dies reports whether creature c dies to the declared damage and
// −N/−N in t, with pump applied first when it is not nil. A removal
// threat is not read here.
func dies(c *protocol.CardView, t threats, pump *protocol.PumpView) bool {
	b := *c
	if pump != nil {
		b.Power += pump.Power
		b.Toughness += pump.Toughness
		if len(pump.Keywords) > 0 {
			b.Abilities = append(append([]string(nil), c.Abilities...), pump.Keywords...)
		}
	}
	// CR 704.5f: −N/−N that takes its toughness to 0 kills it,
	// indestructible or not.
	b.Toughness -= t.minus
	if b.Toughness <= 0 {
		return true
	}
	return t.damage > 0 && damageKills(t.source, &b, t.damage)
}

// answeredFrom reports whether the bot already has an item on the stack
// from source id above index top: the answer it made to these threats.
func (st *state) answeredFrom(id string, top int) bool {
	for i := top + 1; i < len(st.view.StackItems); i++ {
		it := &st.view.StackItems[i]
		if it.Controller == st.me && it.SourceCardID == id {
			return true
		}
	}
	return false
}

// answerSave prices an untargeted row on the bot's creature src by what
// it saves from the stack (ADR 0142 decision 6). ok is false when the
// row saves nothing here, and the move keeps its other price.
func (p *Policy) answerSave(st *state, src *protocol.CardView, row *protocol.ActivatedAbilityView, cp activateParams) (float64, string, bool) {
	if !p.cfg.PriceAnswers || len(st.view.StackItems) == 0 || src == nil || row == nil ||
		src.Controller != st.me || !isCreature(src) || len(cp.Targets) > 0 || row.SacrificeSelf {
		return 0, "", false
	}
	pv := row.Purpose
	if !pv.Declares("survive") && !pv.Declares("protect") && !pv.Declares("prevent") && !pv.Declares("pump") {
		return 0, "", false
	}
	t, ok := st.threatsTo(src, p.cfg.RemovalConfidence)
	if !ok {
		return 0, "", false
	}
	// #2939 gap 2: a plain pump row is answered by its own earlier
	// activations only in part, so it is counted rather than refused.
	repeats := p.cfg.PriceRepeatedPumps && repeatablePump(pv)
	if !repeats && st.answeredFrom(src.InstanceID, t.top) {
		return 0, "", false
	}
	saved := t.chance * st.permanentValue(src)
	switch {
	case pv.Declares("survive"):
		return saved, "answer: survive it", true
	case pv.Declares("protect") && !t.sweep:
		return saved, "answer: protect it from the targeted item", true
	case t.removal:
		return 0, "", false
	case pv.Declares("prevent") && t.minus == 0 && !t.unpreventable:
		return saved, "answer: prevent the damage", true
	case repeats:
		return p.repeatedPumpSave(st, src, row, t, saved)
	case pv.Declares("pump") && pv.Pump != nil && pv.Pump.Toughness > 0 && !dies(src, t, pv.Pump):
		return saved, "answer: pump it out of range", true
	}
	return 0, "", false
}

// maxPumpSteps caps how many activations of one pump row the bot
// plans for, so a free or very cheap row cannot be read as unlimited.
const maxPumpSteps = 8

// repeatablePump reports whether a row's declared purpose is a plain
// self pump that adds toughness: the only kind a second activation
// stacks on (#2939 gap 2). A row that also shields or prevents is
// priced as it was.
func repeatablePump(pv *protocol.PurposeView) bool {
	return pv.Declares("pump") && pv.Pump != nil && pv.Pump.Toughness > 0 &&
		!pv.Declares("survive") && !pv.Declares("protect") && !pv.Declares("prevent")
}

// repeatedPumpSave prices one more activation of a pump row on creature
// src whose first activation may not be enough (#2939 gap 2). The
// activations it takes are the fewest whose summed pump keeps src alive
// through the declared damage and −N/−N; those already on the stack from
// src count toward them, and so do the ones its mana can still pay.
// When all of that adds up, this activation is priced by what saving the
// creature is worth, less a flat ActivateBase for every activation still
// to come after it. ok is false when the creature is out of reach, or
// the pumps on the stack already cover it.
func (p *Policy) repeatedPumpSave(st *state, src *protocol.CardView, row *protocol.ActivatedAbilityView, t threats, saved float64) (float64, string, bool) {
	pump := row.Purpose.Pump
	need := 0
	for k := 1; k <= maxPumpSteps; k++ {
		if !dies(src, t, &protocol.PumpView{Power: pump.Power * k, Toughness: pump.Toughness * k, Keywords: pump.Keywords}) {
			need = k
			break
		}
	}
	pending := st.activationsFrom(src.InstanceID, t.top)
	if need == 0 || need <= pending || need > pending+p.pumpActivations(st, row) {
		return 0, "", false
	}
	if need == 1 {
		return saved, "answer: pump it out of range", true
	}
	steps := need - pending
	return saved - float64(steps-1)*p.cfg.ActivateBase, "answer: pump it out of range (" + strconv.Itoa(need) + " activations)", true
}

// activationsFrom counts the bot's own stack items from source id above
// index top: the activations it has already made in answer. Every item
// counts, whichever row made it, so the bot errs toward stopping.
func (st *state) activationsFrom(id string, top int) int {
	n := 0
	for i := top + 1; i < len(st.view.StackItems); i++ {
		it := &st.view.StackItems[i]
		if it.Controller == st.me && it.SourceCardID == id {
			n++
		}
	}
	return n
}

// pumpActivations is how many times the bot's mana can pay for row now:
// at least the one on offer. A row that costs anything but mana (a tap,
// energy, a sacrifice, a discard) is counted once.
func (p *Policy) pumpActivations(st *state, row *protocol.ActivatedAbilityView) int {
	if !rowCostsOnlyManaAndTaps(row) || row.TapCost || row.EnergyCost > 0 || row.MillCost > 0 || row.Exert ||
		row.CrewCost > 0 || row.RevealCostN > 0 || row.RevealHandCost {
		return 1
	}
	cost := row.ManaCost
	if row.ChargedManaCost != nil {
		cost = *row.ChargedManaCost
	}
	c := parseManaCost(cost, 0)
	units := st.manaAvailable(p.manaModel(st))
	n := 0
	for n < maxPumpSteps {
		var ok bool
		if units, ok = payMana(units, c, nil); !ok {
			break
		}
		n++
	}
	return max(n, 1)
}

// destroyShare is the chance a targeted removal spell with no declared
// damage destroys its target, and so is answered by a grant of
// indestructible (answerSaveOthers). Half: a Murder, not a Swords to
// Plowshares.
const destroyShare = 0.5

// answerSaveOthers prices an untargeted `survive` or `protect` row of
// the bot's whose declared Saves covers permanents besides its own
// source (#2939 gap 3, Config.PriceSaveOthers): Selfless Spirit's
// "creatures you control gain indestructible" in answer to a Wrath of
// God. Every permanent of the bot's in the declared class that the stack
// would remove is worth what saving it is worth (the chance the bot
// loses it, times its value; half of that for a class the row only
// partly covers). A row that sacrifices its source pays for it at the
// dying-anyway price, and does not count itself among the saved. ok is
// false when it saves nothing here, or the bot has already answered.
func (p *Policy) answerSaveOthers(st *state, src *protocol.CardView, row *protocol.ActivatedAbilityView, cp activateParams) (float64, string, bool) {
	if !p.cfg.PriceAnswers || !p.cfg.PriceSaveOthers || len(st.view.StackItems) == 0 || src == nil || row == nil ||
		src.Controller != st.me || len(cp.Targets) > 0 || row.Purpose == nil || row.Purpose.Saves == nil {
		return 0, "", false
	}
	pv := row.Purpose
	survive := pv.Declares("survive")
	if !survive && !pv.Declares("protect") {
		return 0, "", false
	}
	saves := pv.Saves
	var total float64
	covered, top := 0, -1
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller != st.me || !sweepClassHolds(c, saves.Matches) || (row.SacrificeSelf && c.InstanceID == src.InstanceID) {
			continue
		}
		t, ok := st.threatsTo(c, p.cfg.RemovalConfidence)
		if !ok || !t.fromSpell || t.beyondDestroy || (!survive && t.sweep) {
			continue
		}
		share := 1.0
		if saves.Partial {
			share = 0.5
		}
		// A targeted spell with no declared damage is presumed removal,
		// and indestructible answers it only when it destroys: no wire
		// field tells a Murder from a Swords to Plowshares. Hexproof
		// answers it either way.
		if !t.sweep && t.damage == 0 && !pv.Declares("protect") {
			share *= destroyShare
		}
		total += share * t.chance * st.permanentValue(c)
		covered++
		top = max(top, t.top)
	}
	if covered == 0 || st.answeredOthers(src, row, top) {
		return 0, "", false
	}
	if row.SacrificeSelf {
		total -= p.selfSacrificeCost(st, src)
	}
	return total, "answer: save " + strconv.Itoa(covered) + " other permanents", true
}

// answeredOthers reports whether the bot already has an ability above
// index top on the stack that answers it: one from src itself, or, for a
// row that sacrifices its source, one from a permanent that has left the
// battlefield (the Spirit that paid for it). A guess, so the bot errs
// toward not stacking a second Selfless Spirit on the same Wrath.
func (st *state) answeredOthers(src *protocol.CardView, row *protocol.ActivatedAbilityView, top int) bool {
	if st.answeredFrom(src.InstanceID, top) {
		return true
	}
	if !row.SacrificeSelf {
		return false
	}
	for i := top + 1; i < len(st.view.StackItems); i++ {
		it := &st.view.StackItems[i]
		if it.Controller == st.me && it.Kind == "ability" && st.bf[it.SourceCardID] == nil {
			return true
		}
	}
	return false
}

// valueWaits reports whether a row declared `value` should wait for the
// stack: something is on it, and nothing the row spends is a permanent
// the bot is about to lose.
func (p *Policy) valueWaits(st *state, src *protocol.CardView, row *protocol.ActivatedAbilityView, cp activateParams) bool {
	if !p.cfg.PriceAnswers || len(st.view.StackItems) == 0 || row == nil || !row.Purpose.AnswersNothing() {
		return false
	}
	if row.SacrificeSelf && p.dyingAnyway(st, src) > 0 {
		return false
	}
	for _, ids := range [][]string{cp.SacrificeIDs, cp.ExilePermanentIDs} {
		for _, id := range ids {
			if p.dyingAnyway(st, st.bf[id]) > 0 {
				return false
			}
		}
	}
	return true
}

// selfSacrificeCost is what a targeted or purpose-priced row pays for
// sacrificing its own source: the whole source, and under PriceAnswers
// its value times the chance the bot would have kept it. An untargeted
// row with no priced purpose, except a declared defensive answer under
// PriceAnswerOnlyRows (#4318), still pays nothing, as before ADR 0142: its
// payoff is the flat ActivateBase, which cannot cover an undeclared
// benefit when charged the full source. Declared draw rows already
// price their draw and pay this cost.
func (p *Policy) selfSacrificeCost(st *state, src *protocol.CardView) float64 {
	if !p.cfg.PriceAnswers || !p.cfg.SacrificeDyingAnyway {
		return st.permanentValue(src)
	}
	return (1 - p.dyingAnyway(st, src)) * st.permanentValue(src)
}

// exiledOutletCost is what a `sac_outlet` row's exile-a-permanent cost
// (The Soul Stone) pays for one permanent under PriceAnswers, priced as
// sacrifice.go prices a sacrifice. ok is false for every other row,
// which keeps the full board value.
func (p *Policy) exiledOutletCost(st *state, row *protocol.ActivatedAbilityView, c *protocol.CardView) (float64, bool) {
	if !p.cfg.PriceAnswers || !p.cfg.SacrificeDyingAnyway || row == nil || !row.Purpose.Declares("sac_outlet") {
		return 0, false
	}
	return (1 - p.dyingAnyway(st, c)) * st.permanentValue(c), true
}
