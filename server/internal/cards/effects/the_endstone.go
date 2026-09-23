package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Endstone — Legendary Artifact {7} (EDHREC rank 3315):
//
//	"Whenever you play a land or cast a spell, draw a card.
//	 At the beginning of your end step, your life total becomes half
//	 your starting life total, rounded up."
//
// Seven mana for a card off every land and every spell, paid for in
// life every end step. Two triggers:
//
//   - The draw is ONE printed ability with two trigger conditions,
//     watching the land play and the cast on one declaration
//     (b31YouPlayedALandOrCastASpell). A cast is EventCast by the
//     controller; a land play is b20LandPlayed's read of the zone
//     move — a play from exile or a graveyard is told from a
//     returned land by the land-drop tally, with the declared
//     ambiguity that helper documents. The Endstone itself is on the
//     stack when it is cast, so it never draws for its own casting.
//   - The end step sets the controller's life to half the format's
//     starting total rounded up — 20 in Commander — the CR 119.5
//     way: the player gains or loses the difference, so a lifegain
//     or life-loss payoff sees the change, and a total already at
//     20 changes nothing (b31LifeBecomes).
//
// One gap, shared with Horn of Greed and not previously declared
// here: b20LandPlayed still reads every hand-origin battlefield entry
// as a play, which was exact until #654 shipped
// PutFromHandOntoBattlefield (Eureka Moment, Spelunking, Chulane,
// Broken Bond). A land those effects PUT from hand still draws a
// card here, which is stronger than printed. The fix needs a
// "played" marker on the entry itself (#1326).
func init() {
	Register(Spec{
		OracleID:     "57ee0730-3ac5-4f17-ac82-bd4a6673c2bb",
		Name:         "The Endstone",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A land put onto the battlefield from hand by another effect (not played) still draws a card, as though it had been played.",
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b31YouPlayedALandOrCastASpell(ev, source, g)
			}, "The Endstone — draw a card", Do(DrawCards{N: 1})),
			AtYourEndStep("The Endstone — your life total becomes half your starting life total", func(g *game.Game, item *game.StackItem) error {
				return b31LifeBecomes(NewContext(g, item), item.Controller, (game.StartingLife+1)/2)
			}),
		},
	})
}
