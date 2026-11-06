package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// team_pump.go — #4320: an untargeted row that pumps every creature its
// controller controls until end of turn, priced by the attack it adds.
//
// Purphoros, God of the Forge's "{2}{R}: Creatures you control get +1/+0
// until end of turn." declared only that it answers (pump), so it kept
// the flat ActivateBase (0.50) in every window, over PassThreshold
// (0.25) in the bot's main phase. In 96 four-seat games the bot pumped
// 105 times before combat in turns it attacked with nothing, and 19
// times after combat damage. #4302 priced a change until end of turn by
// its target entry, and this row has no target.
//
// The row now declares the pump (Purpose.TeamPump, `team_pump` on the
// wire), and with Config.PriceTeamPumps it is worth, in place of
// ActivateBase:
//
//   - the attack the pump adds to each of the bot's creatures this turn
//     (pumpAttackGain, #4302's price for a pump on one creature: more
//     damage through, lifelink's life, evasion before blockers, haste on a
//     creature that could not swing), with LethalBonus once when it makes
//     a kill;
//   - for each of its creatures blocking before damage, the attacker the
//     pumped blocker now kills and the blocker that now survives
//     (blockGain), each at its permanent value;
//   - plus CombatGrant once while one of them is attacking or blocking
//     and damage is still to come, when the pump changes something
//     there: flat, it bought +1/+0 for every chump blocker with the
//     mana an opponent's turn leaves spare.
//
// So a pump in a main phase with nothing able to attack, or after combat
// damage, is worth nothing and the bot passes; with attackers declared it
// clears the response bar when its mana is spare (spareCombatGrant). A
// keyword grant adds nothing to a creature that has the keyword (CR
// 702.15f for lifelink, and so for the others), and nothing a second time
// while the first is still on the stack.

// teamPumpValue prices a row of the bot's own that declares a team pump
// until end of turn. False when the Config or the row does not price
// one, which keeps the row's other prices.
func (p *Policy) teamPumpValue(st *state, src *protocol.CardView, row *protocol.ActivatedAbilityView) (float64, bool) {
	if !p.cfg.PriceTeamPumps || src == nil || row == nil || row.Purpose == nil || row.Purpose.TeamPump == nil || src.Controller != st.me {
		return 0, false
	}
	pump := row.Purpose.TeamPump
	if pump.Power == 0 && pump.Toughness == 0 && st.activationsFrom(src.InstanceID, -1) > 0 {
		// A keyword grant still on the stack shows nothing on the
		// creatures yet; a second one adds nothing to the first.
		return redundantGrant, true
	}
	v, lethal, combat := 0.0, false, false
	for _, c := range st.teamPumped(pump) {
		gain, kills := p.pumpAttackGain(st, c, pump)
		if kills {
			gain -= p.cfg.LethalBonus
			lethal = true
		}
		v += gain + st.blockGain(c, pump)
		combat = combat || inCombatNow(st, c)
	}
	if lethal {
		v += p.cfg.LethalBonus
	}
	// CombatGrant rides on a pump that changes something in this combat.
	// Flat, it bought a +1/+0 for every chump blocker, again and again
	// while the mana lasted.
	if combat && v > 0 {
		v += p.cfg.CombatGrant
	}
	return v, true
}

// blockGain is what a pump is worth to the bot's creature c while it
// blocks, before damage: the attacker it now kills, and c itself when it
// now survives. Each exchange is read one on one, as kills reads it.
func (st *state) blockGain(c *protocol.CardView, pump *protocol.PumpView) float64 {
	if c.BlockingTarget == "" || !inCombatNow(st, c) {
		return 0
	}
	a := st.bf[c.BlockingTarget]
	if a == nil {
		return 0
	}
	up := pumped(c, []*protocol.PurposeView{{Pump: pump}})
	v := 0.0
	if !kills(c, a) && kills(up, a) {
		v += st.permanentValue(a)
	}
	if kills(a, c) && !kills(a, up) {
		v += st.permanentValue(c)
	}
	return v
}

// teamPumped is each creature of the bot's that a team pump adds
// something to.
func (st *state) teamPumped(pump *protocol.PumpView) []*protocol.CardView {
	var out []*protocol.CardView
	for i := range st.view.Battlefield.Cards {
		c := &st.view.Battlefield.Cards[i]
		if c.Controller == st.me && isCreature(c) && !grantsNothingNew(c, pump) {
			out = append(out, c)
		}
	}
	return out
}

// spareTeamPump reports whether m is a team pump of the bot's own with a
// creature it helps in combat, paid with mana and taps only, that leaves
// the bot the mana for the biggest card it could still cast this turn:
// spareCombatGrant's premise for a row with no target.
func (p *Policy) spareTeamPump(st *state, m legal.Move) bool {
	if !p.cfg.PriceTeamPumps || m.Kind != legal.KindActivate || !p.costsOnlyManaAndTaps(st, m) {
		return false
	}
	cp := decode[activateParams](m.Params)
	src := st.bf[cp.SourceCardID]
	row := rowAt(src, cp.AbilityIndex)
	if src == nil || src.Controller != st.me || row == nil || row.Purpose == nil || row.Purpose.TeamPump == nil {
		return false
	}
	inCombat := false
	for _, c := range st.teamPumped(row.Purpose.TeamPump) {
		inCombat = inCombat || inCombatNow(st, c)
	}
	return inCombat && st.manaToSpare(m)
}
