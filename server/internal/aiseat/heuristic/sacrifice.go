package heuristic

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sacrifice.go is ADR 0126 §7's second half: sacrifice outlets.
//
// Before it, a permanent sacrificed to pay a cost (Viscera Seer,
// Carrion Feeder, Warren Soultrader, Vampiric Rites, Deadly Dispute's
// additional cost) cost its whole board value, so an outlet was used
// only when what it bought was worth more than the creature: in
// practice, never. Two things change that, and both are read off the
// view (ADR 0033 §3):
//
//   - DYING ANYWAY. A permanent the bot is about to lose is cheap to
//     sacrifice. It costs its value times the chance the bot would have
//     kept it: an opponent's spell or ability on the stack that targets
//     it (kept with 1 − RemovalConfidence, because a spell that may
//     target a creature is presumed, not known, to be removal), a
//     declared sweep on the stack that would remove it (kept with 1 −
//     the share the sweep takes), or, in the declare-blockers step, a
//     combat it loses by the planner's own trade arithmetic (kept with
//     0).
//   - DEATH PAYOFFS. Each `death_payoff` row (ADR 0126 §6: Blood Artist,
//     Zulaport Cutthroat, Bastion of Remembrance, Syr Konrad) on a
//     permanent the bot controls adds Config.DeathPayoff to every
//     creature it sacrifices.
//
// A sacrifice of a permanent that is dying anyway spends nothing the bot
// would otherwise keep, which is §5's premise for its two windows, so
// such a move clears the same bar (dyingAnywayEligible): "sacrifice it
// in response" is a use-it-or-lose-it move, and holding it has no option
// value for InstantThreshold to protect.
//
// The Altars (Ashnod's, Phyrexian) are mana abilities, and the bot does
// not activate a mana ability for floating mana (§7). The auto-tapper
// still sacrifices to them when a cast needs it, as before. A row that
// sacrifices its own source (Sakura-Tribe Elder) is not read here: PR 7
// charges the source's full value when the row declares a purpose
// (moves.go), and nothing at all when it does not.

// sacrificeCost prices sacrificing the permanents `ids` to pay a cost,
// as a positive number. Under BaselineConfig it is the sum of their
// board values, the pre-S66 price.
func (p *Policy) sacrificeCost(st *state, ids []string) float64 {
	var v float64
	payoffs := -1
	for _, id := range ids {
		c := st.bf[id]
		if c == nil {
			continue
		}
		keep := 1.0
		if p.cfg.SacrificeDyingAnyway {
			keep -= p.dyingAnyway(st, c)
		}
		v += keep * st.permanentValue(c)
		if p.cfg.DeathPayoff != 0 && isCreature(c) {
			if payoffs < 0 {
				payoffs = st.deathPayoffRows(ids)
			}
			v -= p.cfg.DeathPayoff * float64(payoffs)
		}
	}
	return v
}

// deathPayoffRows counts the `death_payoff` triggered rows on the
// permanents the bot controls, leaving out the permanents being
// sacrificed. Whether a payoff sees its own death is printed text the
// wire does not carry (Blood Artist does, Syr Konrad does not), so the
// count assumes it does not: a payoff is never sacrificed for its own
// trigger.
func (st *state) deathPayoffRows(sacrificed []string) int {
	n := 0
	for _, c := range st.bf {
		if c.Controller != st.me || c.Unimplemented || slices.Contains(sacrificed, c.InstanceID) {
			continue
		}
		for _, r := range c.AbilityRows {
			if r.Purpose != nil && r.Purpose.DeathPayoff {
				n++
			}
		}
	}
	return n
}

// dyingAnyway is the chance, between 0 and 1, that the bot loses
// permanent c whatever it does now: the largest of the three threats
// sacrifice.go describes. Cached per decision.
func (p *Policy) dyingAnyway(st *state, c *protocol.CardView) float64 {
	if c == nil || c.Controller != st.me {
		return 0
	}
	if d, ok := st.dying[c.InstanceID]; ok {
		return d
	}
	d := 0.0
	if st.targetedByAnOpponent(c.InstanceID) {
		d = p.cfg.RemovalConfidence
	}
	if s := st.stackSweepShare(c); s > d {
		d = s
	}
	if d < 1 && st.step == "declare_blockers" && st.losingCombat(c) {
		d = 1
	}
	d = min(max(d, 0), 1)
	if st.dying == nil {
		st.dying = map[string]float64{}
	}
	st.dying[c.InstanceID] = d
	return d
}

// targetedByAnOpponent reports whether a spell or ability on the stack
// that another player controls names permanent id as a target.
func (st *state) targetedByAnOpponent(id string) bool {
	for i := range st.view.StackItems {
		it := &st.view.StackItems[i]
		if it.Controller == st.me {
			continue
		}
		for _, t := range it.Targets {
			if t.ID == id {
				return true
			}
		}
	}
	return false
}

// stackSweepShare is how much of permanent c the declared sweeps on the
// stack would remove: 1 for all of it, a fraction for a bounce or a
// partial sweep, 0 when none would touch it. Only a spell's sweep is
// read, from the slot its stack item names: the alternative cost it was
// cast for (an overloaded Cyclonic Rift), the modes it chose (Austere
// Command), or else the card. An ability's item does not say which row
// it came from, so a planeswalker's or a saga's sweep is not seen here.
func (st *state) stackSweepShare(c *protocol.CardView) float64 {
	best := 0.0
	for i := range st.view.StackItems {
		it := &st.view.StackItems[i]
		if it.Kind != "spell" {
			continue
		}
		card := st.stack[it.SourceCardID]
		if card == nil {
			continue
		}
		for _, p := range stackItemPurposes(card, it) {
			if p.Sweep == nil {
				continue
			}
			if s := st.sweepRemoves(c, p.Sweep, it.Controller, it.XValue); s > best {
				best = s
			}
		}
	}
	return best
}

// stackItemPurposes is the purposes a spell on the stack declares for
// the way it was cast: the claimed alternative cost's, else the chosen
// modes', else the card's own.
func stackItemPurposes(card *protocol.CardView, it *protocol.StackItemView) []*protocol.PurposeView {
	if it.AltCost != "" {
		for i := range card.AlternativeCosts {
			if ac := &card.AlternativeCosts[i]; ac.Key == it.AltCost && ac.Purpose != nil {
				return []*protocol.PurposeView{ac.Purpose}
			}
		}
	}
	if card.Modes != nil && len(it.Modes) > 0 {
		var out []*protocol.PurposeView
		for _, m := range it.Modes {
			if m >= 0 && m < len(card.Modes.Options) && card.Modes.Options[m].Purpose != nil {
				out = append(out, card.Modes.Options[m].Purpose)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	if card.Purpose != nil {
		return []*protocol.PurposeView{card.Purpose}
	}
	return nil
}

// sweepRemoves is how much of permanent c one declared sweep, cast by
// `caster`, removes (ADR 0126 §4's reading of `sweep`): a destroy spares
// an indestructible permanent, damage and −N/−N spare a creature whose
// toughness is above the amount, a bounce takes half of a card (it comes
// back) and all of a token, and a partial sweep is presumed to take half
// of what its class holds.
func (st *state) sweepRemoves(c *protocol.CardView, s *protocol.SweepView, caster string, x int) float64 {
	if s.OpponentsOnly && c.Controller == caster {
		return 0
	}
	if !sweepClassHolds(c, s.Matches) {
		return 0
	}
	share := 1.0
	switch s.How {
	case "destroy":
		if hasKeyword(c, "indestructible") {
			return 0
		}
	case "exile", "sacrifice":
	case "bounce":
		if !c.IsToken {
			share = 0.5
		}
	case "damage", "minus":
		n := s.Amount
		if s.AmountIsX {
			n = x
		}
		if !isCreature(c) || effectiveToughness(c) > n {
			return 0
		}
		if s.How == "damage" && hasKeyword(c, "indestructible") {
			return 0
		}
	default:
		return 0
	}
	if s.Partial {
		share *= 0.5
	}
	return share
}

// sweepClassHolds reports whether permanent c is in a sweep's class
// (game.SweepMatches). An unknown class holds nothing, so a class a
// later catalog adds is read as no threat rather than as every
// permanent.
func sweepClassHolds(c *protocol.CardView, m string) bool {
	switch m {
	case "creatures":
		return isCreature(c)
	case "nonland_permanents":
		return !isLand(c)
	case "artifacts":
		return isType(c, "artifact")
	case "enchantments":
		return isType(c, "enchantment")
	case "artifacts_and_enchantments":
		return isType(c, "artifact") || isType(c, "enchantment")
	case "all_permanents":
		return true
	case "creatures_mana_value_3_or_less":
		return isCreature(c) && manaValue(c.ManaCost, 0) <= 3
	case "creatures_mana_value_4_or_greater":
		return isCreature(c) && manaValue(c.ManaCost, 0) >= 4
	}
	return false
}

// losingCombat reports whether the bot's creature c, once blockers are
// declared, is in a combat it loses: it dies, and sacrificing it first
// gives up nothing the combat would have bought. Read with the combat
// planner's own arithmetic (combat.go, gang.go).
//
//   - An attacker loses when its blockers kill it (gangKills) and it
//     kills none of them and puts no trample damage through. One that
//     trades is not losing: sacrificed, the trade is gone.
//   - A blocker loses when an attacker it blocks kills it (losses, the
//     defender's worst case) and, for every attacker it blocks, the
//     attacker dies or lives the same without it. The attacker stays
//     blocked when its blockers leave combat (CR 509.1h), so a chump
//     sacrificed after blocks still stops the damage, unless the
//     attacker has trample (CR 702.19d), which then counts as not
//     losing.
func (st *state) losingCombat(c *protocol.CardView) bool {
	if !isCreature(c) {
		return false
	}
	if c.AttackingTarget != "" {
		blockers := blockersOn(st, c.InstanceID)
		if len(blockers) == 0 {
			return false
		}
		return gangKills(c, blockers) && len(orderedKills(c, blockers)) == 0 && connects(c, blockers) == 0
	}
	blocked := c.BlockingTargets
	if len(blocked) == 0 && c.BlockingTarget != "" {
		blocked = []string{c.BlockingTarget}
	}
	if len(blocked) == 0 {
		return false
	}
	dies := false
	for _, id := range blocked {
		a := st.bf[id]
		if a == nil {
			continue
		}
		if hasKeyword(a, "trample") {
			return false
		}
		blockers := blockersOn(st, id)
		if slices.Contains(st.losses(a, blockers), c) {
			dies = true
		}
		without := slices.DeleteFunc(slices.Clone(blockers), func(b *protocol.CardView) bool { return b == c })
		if gangKills(a, blockers) && !gangKills(a, without) {
			return false
		}
	}
	return dies
}

// dyingAnywayEligible reports whether m gets the leftover bar because
// everything it costs besides mana and taps is permanents the bot is
// about to lose (ADR 0126 §7, by §5's reasoning): it sacrifices at
// least one permanent, every one of them is dying anyway, and nothing
// else non-mana is paid.
func (p *Policy) dyingAnywayEligible(st *state, m legal.Move) bool {
	if !p.cfg.SacrificeDyingAnyway {
		return false
	}
	if c := m.Cost; c != nil && (c.Life > 0 || c.Loyalty != 0 || len(c.Counters) > 0 || c.Hand > 0) {
		return false
	}
	var ids []string
	switch m.Kind {
	case legal.KindCast:
		if paramsSet(m.Params, without(nonManaCastKeys, "sacrifice_ids")) {
			return false
		}
		ids = decode[castParams](m.Params).SacrificeIDs
	case legal.KindActivate:
		if paramsSet(m.Params, without(nonManaActivateKeys, "sacrifice_ids")) {
			return false
		}
		ap := decode[activateParams](m.Params)
		src := st.bf[ap.SourceCardID]
		if src == nil {
			return false
		}
		row := activatedRow(src, ap.AbilityIndex)
		if row == nil || !rowCostsOnlyManaTapsAndSacrifice(row) {
			return false
		}
		ids = ap.SacrificeIDs
	default:
		return false
	}
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if p.dyingAnyway(st, st.bf[id]) <= 0 {
			return false
		}
	}
	return true
}

// rowCostsOnlyManaTapsAndSacrifice is rowCostsOnlyManaAndTaps with a
// "Sacrifice a creature" clause allowed: the outlet's own cost.
func rowCostsOnlyManaTapsAndSacrifice(r *protocol.ActivatedAbilityView) bool {
	cp := *r
	cp.SacrificeLabel = ""
	return rowCostsOnlyManaAndTaps(&cp)
}

// without returns keys less the one named.
func without(keys []string, drop string) []string {
	return slices.DeleteFunc(slices.Clone(keys), func(k string) bool { return k == drop })
}
