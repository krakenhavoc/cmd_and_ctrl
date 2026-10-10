package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// answers.go is ADR 0142 decision 6: the bot reads the answers an
// activated row declares (`activated_abilities[].purpose.answers`) and
// prices them in a window with something on the stack.
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
//     sweep's share for a sweep). `protect` saves it from anything.
//     `prevent` saves it only when every threat is declared damage that
//     can be prevented. `pump` saves it only when the row's declared
//     Pump keeps it alive against the declared damage and −N/−N. A row
//     already activated from that creature above the threats has done
//     its job, so the bot does not stack a second shield.
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
// What it does not price: a combat grant (the combat planner's
// business, ADR 0142 decision 6 "Not covered"), a row on another
// permanent that saves this one (Selfless Spirit), and a pump that needs
// more than one activation to cover the damage.

// valueInResponse is what a declared-value activation is worth while
// the stack is not empty: below passing, so the bot waits for the stack.
const valueInResponse = -0.5

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
					add(i, threatDamage, conf, dmg, src, cantPrevent || it.DamageCantBePrevented)
				} else {
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
			n := s.Amount
			if s.AmountIsX {
				n = it.XValue
			}
			switch s.How {
			case "damage":
				add(i, threatDamage, share, n, src, cantPrevent || it.DamageCantBePrevented)
			case "minus":
				add(i, threatMinus, share, n, nil, false)
			default:
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
	if !pv.Declares("protect") && !pv.Declares("prevent") && !pv.Declares("pump") {
		return 0, "", false
	}
	t, ok := st.threatsTo(src, p.cfg.RemovalConfidence)
	if !ok || st.answeredFrom(src.InstanceID, t.top) {
		return 0, "", false
	}
	saved := t.chance * st.permanentValue(src)
	switch {
	case pv.Declares("protect"):
		return saved, "answer: protect it from the stack", true
	case t.removal:
		return 0, "", false
	case pv.Declares("prevent") && t.minus == 0 && !t.unpreventable:
		return saved, "answer: prevent the damage", true
	case pv.Declares("pump") && pv.Pump != nil && pv.Pump.Toughness > 0 && !dies(src, t, pv.Pump):
		return saved, "answer: pump it out of range", true
	}
	return 0, "", false
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

// selfSacrificeCost is what a purpose-priced row that sacrifices its
// own source pays for it: the whole source, and under PriceAnswers its
// value times the chance the bot would have kept it. A row with no
// priced purpose still pays nothing for it, as before ADR 0142: its
// payoff is the flat ActivateBase, and charging a Mind Stone's whole
// value against that would never crack it.
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
