package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brass's Tunnel-Grinder — Legendary Artifact {2}{R}, the front face of
// Brass's Tunnel-Grinder // Tecutlan, the Searing Rift (#1112):
//
//	"When Brass's Tunnel-Grinder enters, discard any number of cards,
//	 then draw that many cards plus one.
//	 At the beginning of your end step, if you descended this turn, put
//	 a bore counter on Brass's Tunnel-Grinder. Then if there are three
//	 or more bore counters on it, remove those counters and transform
//	 it. (You descended if a permanent card was put into your graveyard
//	 from anywhere.)"
//
// The back face is tecutlan_the_searing_rift.go, registered under
// "<oracle_id>#1".
//
// The ETB is a prompted discard RUN (#1027) with the draw as its
// continuation, so "that many" is the number really discarded — read
// after the cards have moved, not off the hand size. "Any number" is an
// up-to over the whole hand as the trigger resolves; discarding none is
// a legal answer and still draws one.
//
// "If you descended this turn" is an intervening if (CR 603.4): checked
// as the end step begins, so the ability does not trigger at all
// without it, and again as it resolves. Descending is read off this
// turn's events (descendedThisTurn, graveyard_provenance.go).
//
// The bore counter goes through the counter window (a Doubling Season
// doubles it), so the three-counter check waits for the counter to
// land, as a continuation — Dawn of a New Age's shape (#1282). "Remove
// those counters and transform it" removes the bore counters it has
// and turns the permanent over IN PLACE (ADR 0079's Transform, not an
// exile and return): the same object becomes Tecutlan.
//
// No simplification on this face; the back face carries the caveat.
func init() {
	Register(Spec{
		OracleID:     brassTunnelGrinderOracleID,
		Name:         "Brass's Tunnel-Grinder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Brass's Tunnel-Grinder — discard any number of cards, then draw that many plus one",
				brassTunnelGrinderLoot),
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && descendedThisTurn(g, source.Controller)
			}, "Brass's Tunnel-Grinder — put a bore counter on it; at three, transform it", brassTunnelGrinderBore),
		},
	})
}

// brassTunnelGrinderOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const brassTunnelGrinderOracleID = "af1553eb-4f9f-4335-9078-56649bd8d8fc"

// brassTunnelGrinderBoreCounter is the printed counter kind.
const brassTunnelGrinderBoreCounter = "bore"

// brassTunnelGrinderLoot is the ETB: discard any number, then draw that
// many plus one.
func brassTunnelGrinderLoot(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	hand := 0
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Hand != nil {
		hand = p.Hand.Size()
	}
	return b39MayDiscardThenDraw(hand, false,
		"Brass's Tunnel-Grinder — discard any number of cards, then draw that many plus one",
		func(discarded int) int { return discarded + 1 })(ctx)
}

// brassTunnelGrinderBore is the end-step body, in printed order.
func brassTunnelGrinderBore(g *game.Game, item *game.StackItem) error {
	// CR 603.4's second check, and the source must still be this
	// object on the battlefield to take a counter.
	if !descendedThisTurn(g, item.Controller) || !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	source := item.SourceCardID
	return g.AddCounterThenForEffect(source, brassTunnelGrinderBoreCounter, 1, func(g *game.Game, _ int) error {
		n := b39CountersOn(g, source, brassTunnelGrinderBoreCounter)
		if n < 3 {
			return nil
		}
		return g.AddCounterThenForEffect(source, brassTunnelGrinderBoreCounter, -n, func(g *game.Game, _ int) error {
			return TransformThis{}.Apply(NewContext(g, item))
		})
	})
}
