package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minn, Wily Illusionist — Legendary Creature — Gnome Wizard {1}{U}{U}, 1/3
// (Adventures in the Forgotten Realms Commander, S82 batch 15, #5499):
//
//	"Whenever you draw your second card each turn, create a 1/1 blue
//	 Illusion creature token with 'This token gets +1/+0 for each other
//	 Illusion you control.'
//	 Whenever an Illusion you control dies, you may put a permanent card
//	 with mana value less than or equal to that creature's power from your
//	 hand onto the battlefield."
//
// The first trigger is Alandra's draw tally; the Illusion is the printed
// token template. The second reads the dead Illusion's power as it last
// existed (CR 603.10a), counted at resolution as the ability's "that
// creature's power", and offers the permanent cards in hand with a mana
// value no greater than it (a mana value of 0 qualifies at power 0; a
// split or double-faced card is judged on its front face as it sits in
// hand, CR 202.3). "You may" is a real decline.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4212d44-d6a0-4c5b-9c54-bc90f36828dd",
		Name:         "Minn, Wily Illusionist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Minn — create a 1/1 blue Illusion creature token with \"This token gets +1/+0 for each other Illusion you control.\"",
				Do(CreateToken{Template: AfrB15IllusionToken(), N: 1})),
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && leftUnderControlOf(ev, dead) == source.Controller &&
					leftAsSubtype(ev, dead, source.TextSubtype("Illusion"))
			}, "Minn — you may put a permanent card with mana value less than or equal to that creature's power from your hand onto the battlefield",
				minnWilyIllusionistPut),
		},
	})
}

// minnWilyIllusionistPut offers the permanent cards in hand whose mana value
// is at most the dead Illusion's last-known power.
func minnWilyIllusionistPut(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	power := b17LastKnownPowerOffBattlefield(g, ctx.Trigger().Event.CardID)
	return PutFromHandOntoBattlefield{
		Optional: true,
		Label:    "Minn, Wily Illusionist — put a permanent card with mana value " + strconv.Itoa(power) + " or less from your hand onto the battlefield",
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.ManaValue() <= power
		},
	}.Apply(ctx)
}
