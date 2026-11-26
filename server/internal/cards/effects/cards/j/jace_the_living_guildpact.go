package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jace, the Living Guildpact — Legendary Planeswalker — Jace {2}{U}{U}, loyalty
// 5 (Magic 2015, #5556):
//
//	"+1: Look at the top two cards of your library. Put one of them into your
//	     graveyard.
//	 −3: Return another target nonland permanent to its owner's hand.
//	 −8: Each player shuffles their hand and graveyard into their library. You
//	     draw seven cards."
//
// The +1 is Soulcipher Board's look: the one picked goes to the graveyard and
// the other stays on top (a library of one puts that card there). The −3 is
// any permanent but Jace and not a land. The −8 is Echo of Eons' wheel: every
// player's hand and graveyard are tucked into their own library and shuffled,
// and only once every shuffle has landed do you draw seven.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "55259de3-561d-4a9e-bb28-63c6a7ac5c68",
		Name:            "Jace, the Living Guildpact",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Look at the top two cards of your library. Put one of them into your graveyard.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{SelfMillTutor: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					player := item.Controller
					look := g.LookAtTopOfLibraryForEffect(player, 2)
					switch len(look) {
					case 0:
						return nil
					case 1:
						return effects.RestIntoGraveyard(g, look)
					}
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  player,
						Source:   item.SourceCardID,
						Question: "Jace, the Living Guildpact — put one of them into your graveyard",
						Cards:    look,
						Min:      1,
						Max:      1,
						ThenKey:  jaceLivingGuildpactPicked,
						LiveZone: game.ZoneLibrary,
					})
					return nil
				},
			},
			{
				Label:   "−3: Return another target nonland permanent to its owner's hand.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.Another(effects.TargetPermanent("another target nonland permanent", effects.Not(effects.Land()))),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := effects.FirstLegalBattlefieldTarget(ctx)
					if id == uuid.Nil {
						return nil
					}
					return effects.BounceToHand{Target: id}.Apply(ctx)
				},
			},
			{
				Label: "−8: Each player shuffles their hand and graveyard into their library. You draw seven cards.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.EchoOfEonsShuffleEach(g, effects.TablePlayers(ctx), []uuid.UUID{ctx.Controller()})
				},
			},
		},
	})
}

// jaceLivingGuildpactPicked puts the chosen card of the look into the
// graveyard; the other stays on top. An on-disk key: never renamed, never
// reused.
var jaceLivingGuildpactPicked = game.RegisterChooseCardsThen("choose-cards/jace-living-guildpact-put-one-into-graveyard", func(g *game.Game, r game.CardsChosen) error {
	return effects.RestIntoGraveyard(g, r.Picked)
})
