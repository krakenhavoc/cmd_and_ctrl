package heuristic

import (
	"fmt"
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// face_up_pump.go — #3776: the one-turn pump a turn-face-up trigger gives
// its source, priced by the attack it enables.
//
// Unstable Hulk's "it gets +6/+6 and gains trample until end of turn" rides
// Purpose.FaceUp.Pump. The skipped turn it also declares is a flat charge
// (Weights.Turn); the pump is worth something only if the creature attacks
// this turn, so it is priced the way an exert's pump is (exert.go,
// pumped): the attack with the pumped creature, less the attack without
// it, plus LethalBonus when the pump is what makes the swing kill a seat
// (lethalPushAs). A flip for lethal then outweighs the turn it costs, and
// a flip with nothing to attack into, or after combat, is worth nothing.

// faceUpPumpValue is what the pump of turning `c` up is worth now, and a
// note for the reason. Zero when c cannot attack this turn.
func (p *Policy) faceUpPumpValue(st *state, c *protocol.CardView, pump *protocol.PumpView) (float64, string) {
	if !p.cfg.PriceFaceUpPump || pump == nil {
		return 0, ""
	}
	best, lethal := p.pumpAttackGain(st, c, pump)
	if best <= 0 {
		return 0, ""
	}
	if lethal {
		return best, fmt.Sprintf(" (pump +%d/+%d for lethal)", pump.Power, pump.Toughness)
	}
	return best, fmt.Sprintf(" (pump +%d/+%d attack)", pump.Power, pump.Toughness)
}

// pumpAttackGain is what a pump until end of turn on the bot's creature
// `c` is worth through the attack it makes this turn: the attack with the
// pumped creature less the attack without it, and LethalBonus when the
// pump is what makes the swing kill a seat. Zero, false when c cannot
// attack this turn. Shared by the turn-face-up pump (#3776) and a
// targeted pump of the bot's own (#4302, temporary_effect.go). A pump that
// grants haste lets a summoning-sick creature attack.
func (p *Policy) pumpAttackGain(st *state, c *protocol.CardView, pump *protocol.PumpView) (float64, bool) {
	if c == nil || pump == nil || !st.myTurn || !isCreature(c) || hasKeyword(c, "defender") {
		return 0, false
	}
	hasty := slices.Contains(pump.Keywords, "haste")
	// Still to attack with: untapped and not summoning-sick, in a step
	// before attackers; or already attacking, before damage.
	switch {
	case c.AttackingTarget != "":
		if st.step != "declare_attackers" && st.step != "declare_blockers" {
			return 0, false
		}
	case c.Tapped || (c.SummoningSick && !hasty):
		return 0, false
	case st.step != "precombat_main" && st.step != "begin_combat":
		return 0, false
	}
	up := pumped(c, []*protocol.PurposeView{{Pump: pump}})
	if hasty {
		up.SummoningSick = false
	}
	best, lethal := 0.0, false
	for _, o := range st.opps {
		if c.AttackingTarget != "" && c.AttackingTarget != o.ID {
			continue
		}
		declared, declaredPower := declaredAgainst(st, o.ID, p.cfg.FocusNeedsValue)
		if c.AttackingTarget == o.ID {
			// declaredAgainst counts c itself once it is attacking.
			declared--
			declaredPower -= combatPower(c)
		}
		before, _ := p.attackValue(st, c, o, declared, declaredPower)
		if c.SummoningSick && hasty {
			// Without the haste there is no attack to compare with.
			before = 0
		}
		after, _ := p.attackValue(st, up, o, declared, declaredPower)
		gain := max(after, 0) - max(before, 0)
		// The kill is the pump's only if the swing without it was short.
		kills := !p.lethalPushAs(st, o, nil) && p.lethalPushAs(st, o, up)
		if kills {
			gain += p.cfg.LethalBonus
		}
		if gain > best {
			best, lethal = gain, kills
		}
	}
	return best, lethal
}
