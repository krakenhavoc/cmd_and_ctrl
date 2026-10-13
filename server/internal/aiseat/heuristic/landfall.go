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
// each declared landfall payoff. Zero when the Config does not price
// landfall payoffs.
func (st *state) landfallPayoff(cfg Config) float64 {
	if !cfg.PriceLandfallPayoffs || st.view == nil {
		return 0
	}
	var v float64
	cards := st.view.Battlefield.Cards
	for i := range cards {
		src := &cards[i]
		if src.Controller != st.me {
			continue
		}
		for _, r := range src.AbilityRows {
			if r.Kind != "triggered" || r.Purpose == nil || r.Purpose.LandfallPayoff == nil {
				continue
			}
			v += st.landfallRowValue(cfg, r.Purpose.LandfallPayoff)
		}
	}
	return v
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
