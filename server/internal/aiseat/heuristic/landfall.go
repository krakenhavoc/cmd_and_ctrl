package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// landfall.go — #3072: a land the bot puts onto the battlefield is worth
// what its own landfall payoffs pay for it.
//
// #2480 found Burnished Hart's activation offered 346 times on
// simic-ramp and taken once. Its two lands were priced as ramp
// (ManaSource and the ramp premium), which is worth little once Simic's
// mana deficit has closed; but in 141 of the 150 windows of its own main
// phase it controlled a landfall payoff, 127 of them Tatyova, Benthic
// Druid, so each of those lands was also a card and a life. Nothing
// priced that.
//
// A landfall row now declares `landfall_payoff`: what it pays for each
// land that enters under its controller's control. Every place the
// policy prices a land it would put onto the battlefield adds what the
// bot's battlefield pays for it:
//
//   - the Lands a declared purpose puts there (purposeValue): Burnished
//     Hart's and Sakura-Tribe Elder's activations, Harrow, Cultivate,
//     Explosive Vegetation, Solemn Simulacrum's enters effect;
//   - the lands an extra land drop lets the bot play this turn
//     (extraLandDropValue): Oracle of Mul Daya, Exploration;
//   - a land put from hand by a choose_cards prompt (putGain): Uro's
//     land, Eureka Moment.
//
// The bot's ordinary land drop is not repriced: it is already taken
// before any cast (LandValue), and a payoff changes nothing about that.
//
// One payoff is priced in the units the policy already uses for the
// same thing: a card at Weights.Hand, a point of life at MarginalLife, a
// token at TokenWeight, a +1/+1 counter at Weights.Power +
// Weights.Toughness, energy at Weights.Energy, and a point of damage to
// each live opponent at DamageToOpponent, as the discard payoffs do. A
// mana it adds is landfallManaShare of a token: it is one-shot mana like
// a Treasure's, but it empties at the end of the step.

// landfallManaShare is the share of TokenWeight one mana a landfall row
// adds is worth (Lotus Cobra, Omnath, Locus of the Void): a Treasure is
// a mana the bot can keep, and this one is gone at the end of the step.
const landfallManaShare = 0.5

// landfallPayoff is what one land entering under the bot's control pays
// it: the sum, over the triggered rows of the permanents it controls, of
// each declared landfall payoff, times the triggers each land causes
// (landfallTriggers: two with Ancient Greenwarden out). Zero when the
// Config does not price landfall payoffs.
func (st *state) landfallPayoff(cfg Config) float64 {
	if !cfg.PriceLandfallPayoffs || st.view == nil {
		return 0
	}
	return st.boardLandfall(cfg) * float64(st.landfallTriggers(cfg))
}

// boardLandfall is the sum of the bot's own permanents' landfall rows
// for one trigger each (landfallRows), before any doubler.
func (st *state) boardLandfall(cfg Config) float64 {
	var v float64
	cards := st.view.Battlefield.Cards
	for i := range cards {
		if src := &cards[i]; src.Controller == st.me {
			v += st.landfallRows(cfg, src)
		}
	}
	return v
}

// landfallRows is what card c's own landfall rows pay its controller
// for one land, one trigger each.
func (st *state) landfallRows(cfg Config, c *protocol.CardView) float64 {
	var v float64
	for _, r := range c.AbilityRows {
		if r.Kind != "triggered" || r.Purpose == nil || r.Purpose.LandfallPayoff == nil {
			continue
		}
		v += st.landfallRowValue(cfg, r.Purpose.LandfallPayoff)
	}
	return v
}

// landfallTriggers is how many times each landfall row of the bot's
// triggers for one land (#3146): one, plus each declared
// `landfall_extra_triggers` on a permanent it controls (Ancient
// Greenwarden, Elesh Norn, Mother of Machines). One when the Config does
// not price doublers.
func (st *state) landfallTriggers(cfg Config) int {
	n := 1
	if !cfg.PriceLandfallDoublers || st.view == nil {
		return n
	}
	cards := st.view.Battlefield.Cards
	for i := range cards {
		if c := &cards[i]; c.Controller == st.me {
			n += landfallExtraTriggers(c)
		}
	}
	return n
}

// landfallExtraTriggers is the doubling card c declares, zero for none.
func landfallExtraTriggers(c *protocol.CardView) int {
	if c == nil || c.Purpose == nil {
		return 0
	}
	return c.Purpose.LandfallExtraTriggers
}

// landfallRowValue prices one row's payoff for one land.
func (st *state) landfallRowValue(cfg Config, l *protocol.LandfallPayoffView) float64 {
	v := st.w.Hand * float64(l.Draws)
	if l.LifeGain > 0 && st.myEval != nil {
		v += float64(l.LifeGain) * st.w.MarginalLife(st.myEval.Life)
	}
	v += cfg.TokenWeight * float64(l.Tokens)
	v += landfallManaShare * cfg.TokenWeight * float64(l.Mana)
	v += (st.w.Power + st.w.Toughness) * float64(l.Counters)
	v += st.w.Energy * float64(l.Energy)
	v += cfg.DamageToOpponent * float64(l.DamageEachOpponent*len(st.opps))
	return v
}

// landfallValue is what `n` lands entering under the bot's control pay
// it through its landfall payoffs (landfallPayoff per land).
func (p *Policy) landfallValue(st *state, n int) float64 {
	if n <= 0 {
		return 0
	}
	return float64(n) * st.landfallPayoff(p.cfg)
}

// #3146: the three gaps #3072 left open.
//
// A payoff's own cast. Tireless Tracker, Lotus Cobra or a second
// Tatyova cast before a turn of land drops is worth what it will pay for
// them, which its body alone does not price. landfallCastValue adds, to
// a permanent's cast price, what it adds to the bot's landfall per land
// (its own rows, and its doubling of the board's, Ancient Greenwarden's)
// times the lands the bot expects to put down while it is out
// (expectedLandfalls). The commander is cast from the command zone,
// whose cards carry no ability rows on the wire, so it is not repriced.
//
// Fetch lands. A fetch's activation keeps the flat ActivateBase it has
// always had, and each land it puts down adds what the bot's payoffs pay
// for it (fetchLandfallValue). A land beyond the one it sacrifices
// (Myriad Landscape's second) is ramp, as a declared purpose's lands
// are. The fetch's declared `lands` is read nowhere else, so with the
// switch off a fetch is priced exactly as before.

// landfallCastTurns is how many of the bot's turns after this one
// landfallCastValue counts the lands of.
const landfallCastTurns = 2

// landfallCastValue is what casting permanent c adds to the bot's
// landfall: the per-land payoff with c on the battlefield less the one
// without it, times expectedLandfalls. Zero for a card with no landfall
// row and no doubling, and when the Config does not price it.
func (p *Policy) landfallCastValue(st *state, c *protocol.CardView) float64 {
	cfg := p.cfg
	if !cfg.PriceLandfallCasts || !cfg.PriceLandfallPayoffs || st.view == nil || c == nil ||
		isLand(c) || !(isCreature(c) || isPermanentSpell(c)) {
		return 0
	}
	own := st.landfallRows(cfg, c)
	extra := 0
	if cfg.PriceLandfallDoublers {
		extra = landfallExtraTriggers(c)
	}
	if own == 0 && extra == 0 {
		return 0
	}
	board, k := st.boardLandfall(cfg), st.landfallTriggers(cfg)
	gain := (board+own)*float64(k+extra) - board*float64(k)
	if gain <= 0 {
		return 0
	}
	return gain * st.expectedLandfalls(c)
}

// expectedLandfalls is how many lands the bot expects to put onto the
// battlefield from now to the end of its landfallCastTurns-th turn after
// this one, with card `self` cast: its land drops left this turn, met
// from the lands in hand, then its land drops on each later turn (with
// any `self` adds), met from the lands left in hand and a land drawn in
// millLandShare of its draws.
func (st *state) expectedLandfalls(self *protocol.CardView) float64 {
	if st.seat == nil {
		return 0
	}
	hand := 0
	for i := range st.seat.Hand.Cards {
		if c := &st.seat.Hand.Cards[i]; c.InstanceID != self.InstanceID && isLand(c) {
			hand++
		}
	}
	drops := max(1, st.seat.LandDropsPerTurn)
	if self.Purpose != nil {
		drops += self.Purpose.ExtraLandDrops
	}
	now := 0
	if st.myTurn {
		now = min(hand, max(0, st.seat.LandDropsPerTurn-st.seat.LandsPlayedThisTurn))
	}
	later := min(float64(hand-now)+millLandShare*landfallCastTurns, float64(drops*landfallCastTurns))
	return float64(now) + later
}

// fetchRow reports whether an activated row of src is a fetch: src is a
// land of the bot's own, the row sacrifices it, and what the row declares
// is lands alone (Misty Rainforest's 1, Myriad Landscape's 2).
func (st *state) fetchRow(src *protocol.CardView, row *protocol.ActivatedAbilityView, ps purposeSet) bool {
	if src == nil || row == nil || !row.SacrificeSelf || src.Controller != st.me || !isLand(src) || ps.lands <= 0 {
		return false
	}
	rest := ps
	rest.lands, rest.landsUntapped = 0, 0
	return !rest.hasAmounts() && len(ps.sweeps) == 0 && ps.impulse == 0 && ps.extraLands == 0
}

// fetchLandfallValue is what a fetch's activation is worth beyond the
// flat ActivateBase: each land it puts down pays the bot's landfall
// payoffs, and the lands beyond the one it sacrifices are ramp
// (ManaSource and the ramp premium). Zero when the Config does not
// price it.
func (p *Policy) fetchLandfallValue(st *state, ps purposeSet) float64 {
	if !p.cfg.PriceFetchLandfall {
		return 0
	}
	v := p.landfallValue(st, ps.lands)
	if net := ps.lands - 1; net > 0 && p.cfg.PricePurposes {
		v += st.w.ManaSource*float64(net) + p.rampFor(st, nil, net)
	}
	return v
}
