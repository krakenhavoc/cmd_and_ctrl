package heuristic

// land_swap.go prices a land swap by what it nets (#2469, ADR 0126 §6).
//
// Harrow sacrifices a land as an additional cost and puts two basic
// lands onto the battlefield untapped. Priced gross, its declared
// `purpose.lands: 2` was two lands of ramp (ManaSource and the ramp
// premium each), SpellFloor was applied to that, and the sacrificed
// land was charged afterwards. Late in a game that came to about
// −0.20, below a Rampant Growth (one land, tapped, at sorcery speed),
// which SpellFloor lifts to +0.10. A card that nets the same land,
// untapped and at instant speed, was priced below one that nets it
// tapped at sorcery speed.
//
// Priced net:
//
//   - each sacrificed land is replaced one for one by a land the spell
//     puts onto the battlefield: it costs its own value (sacrificeCost,
//     a tapped land less than an untapped one) and gives back an
//     untapped land, ManaSource;
//   - only the lands beyond those are ramp: ManaSource and the ramp
//     premium for `lands − sacrificed`, with SpellFloor under that net
//     amount, as it is under any other ramp spell's.
//
// The swap applies only when every sacrificed permanent is a land the
// bot controls and the purpose declares more lands than are sacrificed.
// Anything else (a creature sacrificed, one land for one land, no
// `lands` purpose) is priced as before. The purpose does not say
// whether the lands enter tapped, and Harrow's do not. A card whose
// lands enter tapped (Roiling Regrowth, Cycle of Renewal) declares no
// purpose today; declaring one would price its replacement as
// untapped, which is a reason to add a tapped flag to the purpose
// first.

// ownLandsSacrificed reports how many permanents `ids` names when every
// one of them is a land the bot controls, and false otherwise or when
// it names none.
func (st *state) ownLandsSacrificed(ids []string) (int, bool) {
	if len(ids) == 0 {
		return 0, false
	}
	for _, id := range ids {
		c := st.bf[id]
		if c == nil || c.Controller != st.me || !isLand(c) {
			return 0, false
		}
	}
	return len(ids), true
}

// landSwap is the number of lands a cast sacrifices that its own
// declared purpose replaces, when NetLandSwaps prices it as a swap:
// every sacrificed permanent is a land of the bot's own and the purpose
// puts more lands onto the battlefield than that. Zero otherwise.
func (p *Policy) landSwap(st *state, ps purposeSet, sacrificeIDs []string) int {
	if !p.cfg.NetLandSwaps || !p.cfg.PricePurposes {
		return 0
	}
	n, ok := st.ownLandsSacrificed(sacrificeIDs)
	if !ok || ps.lands <= n {
		return 0
	}
	return n
}
