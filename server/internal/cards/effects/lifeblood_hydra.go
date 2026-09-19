package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifeblood Hydra — Creature — Hydra {X}{G}{G}{G}, 0/0 (EDHREC rank
// 3036):
//
//	"Trample
//	 This creature enters with X +1/+1 counters on it.
//	 When this creature dies, you gain life and draw cards equal to
//	 its power."
//
// The Hydra that pays out on death. Trample rides PrintedKeywords;
// the X counters go on as the spell resolves (Goldvein Hydra's
// posture, declared below); the death trigger reads the Hydra's
// last-known power — the harvester's LKI characteristic carries the
// layer-computed P/T and the +1/+1 counters are read back off the
// log (b13LastKnownPower, Conclave Mentor's shape), so a pumped or
// grown Hydra pays out for what it was when it died — and gains that
// much life, then draws that many, in printed order.
//
// One declared simplification, weaker than printed: the X +1/+1
// counters are placed as the spell resolves, a beat before the card
// enters (an entry replacement cannot read the spell's X), so a
// "whenever you put counters on a permanent" payoff does not see
// them. Cast for X=0 the Hydra enters as the printed 0/0 it is and
// the next state-based check puts it into its owner's graveyard
// (CR 704.5f), paying out nothing, exactly as in paper. CR 601.2b
// allows the announcement; it simply does not survive it. That was an
// engine gap until #691 — the toughness check read every printed 0/0
// as the importer's stand-in — and what closed it is the printing
// behind the object (game.Card.ToughnessIsKnown).
func init() {
	Register(Spec{
		OracleID:        "b14d05c0-fe10-4079-a90e-0aea1a8fd375",
		Name:            "Lifeblood Hydra",
		XMatters:        true,
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The X +1/+1 counters are put on the Hydra as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them."},
		PrintedKeywords: []string{"trample"},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: ctx.X()}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				power := b13LastKnownPower(g, source.InstanceID, lki)
				return game.NewTriggeredItem(source, "Lifeblood Hydra — gain life and draw cards equal to its power",
					func(g *game.Game, item *game.StackItem) error {
						return b28GainLifeAndDraw(g, item, power)
					})
			},
		}},
	})
}
