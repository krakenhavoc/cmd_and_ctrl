package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call to the Kindred — Enchantment — Aura {3}{U} (Dark Ascension, S82 OLD,
// #5571):
//
//	"Enchant creature
//	 At the beginning of your upkeep, you may look at the top five cards of
//	 your library. If you do, you may put a creature card that shares a
//	 creature type with enchanted creature from among them onto the
//	 battlefield, then you put the rest of those cards on the bottom of your
//	 library in any order."
//
// Aethermage's Touch's put-and-order, over a private look (only you see the
// five) instead of a reveal. The "you may look" is the trigger's own question;
// the creature must share a type with the enchanted creature as it is as the
// ability resolves (a changeling shares every type, CR 702.73), and the put is
// optional, so the five go to the bottom in your chosen order whether or not
// one comes out. If the Aura has left, no creature type is shared, so nothing
// is put.
//
// No simplification.
func init() {
	callToTheKindredTail = effects.PutFromLibraryTail("put-from-library/call-to-the-kindred-bottom",
		func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
			return ctx.Game.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
				Chooser:   res.Player,
				Source:    res.Source,
				Cards:     res.Rest,
				From:      game.ZoneLibrary,
				Placement: game.LibraryPlaceBottom,
				Reason:    effects.LibraryOrderLabel(ctx.Game, res.Source, "put the rest on the bottom of your library in any order"),
			})
		})
	effects.Register(effects.Spec{
		OracleID:     "c2fb8d4c-b973-42a0-9e9c-7abd5cc2c22e",
		Name:         "Call to the Kindred",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Call to the Kindred — look at the top five cards of your library and may put a creature card that shares a creature type with enchanted creature onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					var host game.Card
					if ref, ok := ctx.SourceRef(); ok {
						if aura, ok := g.PermanentForEffect(ref); ok && aura.AttachedTo.Kind == game.TargetCard {
							host, _ = g.LookupCardForEffect(aura.AttachedTo.ID)
						}
					}
					player := ctx.Controller()
					return effects.PutFromLibraryOntoBattlefield{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 5),
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return host.InstanceID != uuid.Nil && c.IsCreature() && game.SharesCreatureType(&host, &c)
						},
						Max:      1,
						Optional: true,
						Label:    "Call to the Kindred — you may put a creature card that shares a creature type with enchanted creature onto the battlefield",
						ThenTail: callToTheKindredTail,
					}.Apply(ctx)
				}),
				"Call to the Kindred — look at the top five cards of your library?"),
		},
	})
}

var callToTheKindredTail game.TailRef
