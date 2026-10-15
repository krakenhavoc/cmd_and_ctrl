package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// temporary_body.go — #3802: a creature cast for evoke, warp, dash or
// blitz is a body that does not stay.
//
//	evoke  sacrificed as it enters, so its enters effect and nothing else
//	warp   exiled at the next end step, then castable from exile later
//	dash   haste, back to its owner's hand at the next end step
//	blitz  haste, sacrificed at the next end step, draws a card when it dies
//
// The heuristic priced every cast as the permanent it puts on the board
// (resolvedValueFor), which made a blitzed Workshop Warchief worth the
// same as a hard-cast one. This prices what the cast really buys:
//
//	value = the cast as priced
//	      − the lasting body (permanentValue × SickCreature)
//	      + the attack its haste makes this turn, when it has any and the
//	        attack is still ahead (dash, blitz, and a warped creature that
//	        has haste of its own)
//	      + the card back: blitz's draw, and the card dash and warp keep
//
// Its enters effect is not touched: it rides the card's own purpose, which
// castPurpose already reads on a cast that names an alternative cost with
// no purpose of its own. So the printed cost wins whenever it is
// affordable and the body is the point, and the alternative wins when the
// attack is lethal or the enters effect is the point (Mulldrifter).
//
// The key is read off the cast move, never the card: the same card is a
// permanent when hard-cast and a temporary body when evoked.

// temporaryBodyKind is the keyword a cast's alternative-cost key names, ""
// for any other. A bound offer (blitz from the graveyard, "blitz-graveyard")
// is the same keyword.
func temporaryBodyKind(key string) string {
	for _, k := range [...]string{"evoke", "warp", "dash", "blitz"} {
		if key == k || strings.HasPrefix(key, k+"-") {
			return k
		}
	}
	return ""
}

// temporaryBodyDelta is the change to a creature cast's price for the
// body it will not keep, and the keyword it was priced as ("" when the
// cast claims none of the four, the card is not a creature, or the Config
// does not price them).
func (p *Policy) temporaryBodyDelta(st *state, card *protocol.CardView, cp castParams) (float64, string) {
	kind := temporaryBodyKind(cp.AlternativeCost)
	if !p.cfg.PriceTemporaryBodies || kind == "" || card == nil || !isCreature(card) {
		return 0, ""
	}
	// What resolvedValueFor counted as the lasting body.
	delta := -st.w.permanentValue(card) * st.w.SickCreature

	haste := kind == "dash" || kind == "blitz" || (kind == "warp" && hasKeyword(card, "haste"))
	if haste {
		delta += p.hasteAttackValue(st, card)
	}
	switch kind {
	case "blitz":
		// "When this creature dies, draw a card."
		delta += st.w.Hand
	case "dash", "warp":
		// The card is not spent: dash returns it to hand, warp lets it be
		// cast from exile on a later turn. valueOfCast charged it as
		// leaving the hand, so give that back.
		delta += st.w.Hand
	}
	return delta, kind
}

// hasteAttackValue is what the card gains by attacking the turn it is
// cast: the best attack, over the opponents, of the card as a creature
// that can attack, counted only where it is positive. Zero unless the
// attack is still ahead on the bot's own turn.
func (p *Policy) hasteAttackValue(st *state, card *protocol.CardView) float64 {
	if !st.myTurn || (st.step != "precombat_main" && st.step != "begin_combat") || hasKeyword(card, "defender") {
		return 0
	}
	best := 0.0
	for _, o := range st.opps {
		declared, declaredPower := declaredAgainst(st, o.ID, p.cfg.FocusNeedsValue)
		if v, _ := p.attackValue(st, card, o, declared, declaredPower); v > best {
			best = v
		}
	}
	return best
}
