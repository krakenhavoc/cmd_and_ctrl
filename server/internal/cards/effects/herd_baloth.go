package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Herd Baloth — Creature — Beast {3}{G}{G}, 4/4 (EDHREC rank 4214):
//
//	"Whenever one or more +1/+1 counters are put on this creature,
//	 you may create a 4/4 green Beast creature token."
//
// A five-mana 4/4 that makes another 4/4 every time anything grows
// it: a Hardened Scales deck, a proliferate deck, or just a Rishkar
// turns it into a token engine. The rate is why it is played over a
// bigger body — one counter, one 4/4.
//
// "One or more … are put" is per PLACEMENT EVENT, not per counter: a
// single AddCounter of three counters is one trigger and one token,
// and three separate placements are three triggers and three tokens.
// That is the same reading Exemplar of Light takes, and it shares its
// helper.
//
// EventCounterPlaced is emitted for removals too and carries only the
// post-change total, so b11CountersWerePlaced reads the previous total
// back off the log — a counter REMOVED never triggers. Unlike
// Exemplar of Light, there is no "you put" clause here: a counter an
// OPPONENT puts on the Baloth (a Tempt with Vengeance-style gift, a
// Grumgully) triggers it too, exactly as printed.
//
// The token is optional ("you may"), so it is a real prompt to the
// Baloth's controller during resolution rather than an automatic
// creation. Declining is a legal and occasionally correct play — a
// token is a creature that a symmetrical sacrifice effect can take.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03873314-6e64-43ab-95c0-3d8692a57a03",
		Name:         "Herd Baloth",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11CountersWerePlaced(ev, source.InstanceID, game.CounterPlusOne, g)
			}, "Herd Baloth — you may create a 4/4 green Beast", Do(MayChoice{
				Question: "Herd Baloth — create a 4/4 green Beast creature token?",
				OnYes: func(ctx *Context) error {
					return CreateToken{Template: TokenCard("4/4 green Beast"), N: 1}.Apply(ctx)
				},
			})),
		},
	})
}
