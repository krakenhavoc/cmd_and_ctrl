package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ayesha Tanaka, Armorer — Legendary Creature — Human Artificer
// {3}{W}{U}, 2/4:
//
//	"Whenever Ayesha Tanaka attacks, look at the top four cards of your
//	 library. You may put any number of artifact cards with mana value less
//	 than or equal to Ayesha Tanaka's power from among them onto the
//	 battlefield tapped. Put the rest on the bottom of your library in a
//	 random order.
//	 Ayesha Tanaka can't be blocked as long as defending player controls
//	 three or more artifacts."
//
// The look is the top four of the library, and the cards chosen enter in
// one simultaneous batch, tapped (CR 614.12). Ayesha's power is read as the
// ability resolves, from the permanent as it is now or as it last existed
// if she has left (CR 608.2h, TriggeringPermanent), so a pump in response
// raises the ceiling and a negative power admits nothing. A put that
// leaves cards behind, or declines them all, sends the rest to the bottom
// in an order drawn from the game's seeded RNG (a keyed tail, so a table
// waiting on the pick is a restore point).
//
// The static is a block rule on Ayesha herself, read live as the block is
// declared (CR 509.1b): the defending player is the one she is attacking,
// or the controller of the planeswalker or battle she attacks (CR 506.3),
// and artifacts are counted through the layers, so an animated artifact
// land counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59fcd09c-4633-457e-b36a-488e94c1c941",
		Name:         "Ayesha Tanaka, Armorer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ayesha Tanaka, Armorer — look at the top four cards of your library; you may put any number of artifact cards with mana value at most her power onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok {
						return nil
					}
					you := ctx.Controller()
					return effects.PutFromLibraryOntoBattlefield{
						Player:   you,
						Cards:    g.LookAtTopOfLibraryForEffect(you, 4),
						Match:    effects.And(effects.Artifact(), effects.ManaValueLE(info.Power)),
						Optional: true,
						Tapped:   true,
						Label:    "Ayesha Tanaka, Armorer — put any number of artifact cards with mana value at most her power onto the battlefield tapped",
						ThenTail: effects.PutRestOnBottomInRandomOrderTail,
					}.Apply(ctx)
				}),
		},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedWhile(effects.OnSelf(), func(g *game.Game, _ uuid.UUID, attacker game.Card) bool {
				defender := g.DefendingPlayerForAttackerForEffect(attacker.InstanceID)
				return defender != uuid.Nil &&
					len(effects.PermanentsControlledByMatching(g, defender, effects.Artifact())) >= 3
			}),
		},
	})
}
