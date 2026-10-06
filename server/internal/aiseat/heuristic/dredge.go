package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// dredge.go — the bot's answer to a dredge offer (#2390, CR 702.52a).
//
// Since #2388 a dredge card in the bot's graveyard asks, on every draw,
// "mill N cards and return this card to your hand instead?" through the
// ordinary optional_replacement yes/no. The default "yes" every other
// "may" gets would have the bot dredge every draw of the game, needed
// card or not, until the library ran out. So the two answers are priced
// against each other in the units the rest of the policy uses:
//
//   - YES is the dredged card in hand (cardValue), plus DredgePlaySoon
//     when the bot could play it this turn or next, plus the N milled
//     cards at what a card in the bot's graveyard is worth to it
//     (millValue, the fuel pricer's number), less what the dredged card
//     was worth where it was (fuelValue).
//   - NO is the card it would draw (drawValue): unknown, and so priced
//     at the mean cardValue of the cards the bot has already seen from
//     that library.
//
// And one rule above both: a dredge that would leave fewer than
// DredgeLibraryFloor cards in the library is declined, whatever it
// returns.
//
// The dredged card is the prompt's Source. A dredge that names none is
// a grant over a class of cards, picked after the yes; the catalog's
// only one is The Necrobloom's "land cards in your graveyard have
// dredge 2" (TestOnlyTheNecrobloomGrantsDredgeWithoutACard pins that),
// so it is priced at the bot's best land card there.
//
// What it cannot see, and does not pretend to: which other cards in the
// graveyard have dredge (the wire says so only for the one being
// offered, so when two are offered in turn the first one worth more
// than a draw is taken, not the best of them), and what a card does
// beyond its type line, so a Life from the Loam is priced as a
// two-mana sorcery rather than as three lands.

// drawPrior is how many pseudo-samples at Config.ScryKeep — the scry
// branch's own price of an unknown card — the draw estimate starts
// from, so a bot that has seen two cards of its library does not take
// their mean as gospel.
const drawPrior = 3

// millLandShare is the share of lands millValue assumes of a library
// it knows nothing about: a Commander deck's thirty-seven or so lands
// in ninety-nine.
const millLandShare = 0.4

// dredgeValue scores one answer to a dredge offer. `yes` takes the
// dredge.
func (p *Policy) dredgeValue(st *state, ch *protocol.PendingChoiceView, yes bool) (float64, string) {
	draw := p.drawValue(st)
	if !yes {
		return draw, "dredge: draw instead"
	}
	if st.seat != nil && st.seat.Library.Count-ch.Dredge < p.cfg.DredgeLibraryFloor {
		return draw - 1, "dredge: the library is too short to mill"
	}
	card := st.dredgedCard(p.cfg, ch)
	v := st.cardValue(p.cfg, card)
	if card != nil {
		if st.playableSoon(p.cfg, card) {
			v += p.cfg.DredgePlaySoon
		}
		v -= p.fuelValue(st, card.InstanceID)
	}
	v += float64(ch.Dredge) * (p.millValue(st, card) - st.w.Library)
	if v > draw {
		return v, "dredge: the card and the mill beat a draw"
	}
	return v, "dredge: a draw is worth more"
}

// dredgedCard is the card a dredge returns: the prompt's Source, in
// the bot's graveyard, or for a grant that names none (The Necrobloom)
// the bot's most valuable land card there. Nil when there is none.
func (st *state) dredgedCard(cfg Config, ch *protocol.PendingChoiceView) *protocol.CardView {
	if ch.Source != "" {
		return st.graveyard[ch.Source]
	}
	if st.seat == nil {
		return nil
	}
	var best *protocol.CardView
	bestV := 0.0
	for i := range st.seat.Graveyard.Cards {
		c := &st.seat.Graveyard.Cards[i]
		if !isLand(c) {
			continue
		}
		if v := st.cardValue(cfg, c); best == nil || v > bestV {
			best, bestV = c, v
		}
	}
	return best
}

// playableSoon reports whether the bot could play a card in hand this
// turn or next: a nonland card it has the mana for, or one more land
// than it has (cardValue's own horizon, which discounts the rest); a
// land while it still wants lands and holds none.
func (st *state) playableSoon(cfg Config, c *protocol.CardView) bool {
	if isLand(c) {
		if st.myMana >= cfg.LandsWanted {
			return false
		}
		for _, h := range st.seatHand() {
			if isLand(&h) {
				return false
			}
		}
		return true
	}
	return manaValue(c.ManaCost, 0) <= st.myMana+1
}

// drawValue is what the bot expects the top card of its library to be
// worth in hand: the mean cardValue of the cards it has already seen
// from that library, shrunk toward ScryKeep by drawPrior.
//
// Every card the bot has drawn or milled came off the top of the same
// library, wherever it went next, so its hand, the permanents it owns,
// its graveyard and its face-up exile are a sample of what is still in
// there. The commander never was in the library, and a token never was
// a card, so neither counts; a card the bot cannot read counts as
// nothing. Each card is priced as a card in hand, not as the permanent
// it became: untapped, able to act, unrestricted.
//
// That sample is what makes the draw worth more exactly when the bot
// needs it: cardValue prices a land at its LandsWanted premium while
// the bot is short, so a library full of lands it needs outbids a
// dredge that returns a spell it cannot cast yet.
func (p *Policy) drawValue(st *state) float64 {
	sum, n := float64(drawPrior)*p.cfg.ScryKeep, float64(drawPrior)
	add := func(c *protocol.CardView) {
		if c.IsCommander || c.IsToken || c.Name == "" || (c.FaceDown && !c.KnownByYou) {
			return
		}
		fresh := *c
		fresh.Tapped, fresh.SummoningSick, fresh.Restrictions = false, false, nil
		sum += st.cardValue(p.cfg, &fresh)
		n++
	}
	if st.seat == nil {
		return p.cfg.ScryKeep
	}
	for i := range st.seat.Hand.Cards {
		add(&st.seat.Hand.Cards[i])
	}
	for i := range st.seat.Graveyard.Cards {
		add(&st.seat.Graveyard.Cards[i])
	}
	// The view's slices, not the state's maps: a sum taken in map
	// order differs in its last bits from run to run, and a lockstep
	// replay must decide the same way twice.
	for _, zone := range [][]protocol.CardView{st.view.Battlefield.Cards, st.view.Exile.Cards} {
		for i := range zone {
			if zone[i].Owner == st.me {
				add(&zone[i])
			}
		}
	}
	return sum / n
}

// millValue is what one card milled off the bot's library is worth to
// it in the graveyard: the mean fuelValue of the cards already in its
// graveyard (the dredged card aside), which is the fuel pricer's read
// of what that graveyard is worth — a flashback or escape card there
// raises it, a pile of dead lands keeps it at the floor. With nothing
// there to go on, the two fuel floors weighted by millLandShare.
func (p *Policy) millValue(st *state, dredged *protocol.CardView) float64 {
	var sum float64
	n := 0
	if st.seat != nil {
		for i := range st.seat.Graveyard.Cards {
			c := &st.seat.Graveyard.Cards[i]
			if dredged != nil && c.InstanceID == dredged.InstanceID {
				continue
			}
			sum += p.fuelValue(st, c.InstanceID)
			n++
		}
	}
	if n == 0 {
		return millLandShare*p.cfg.FuelFloor + (1-millLandShare)*p.cfg.FuelIdle
	}
	return sum / float64(n)
}
