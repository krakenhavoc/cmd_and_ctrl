package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const blexOracleID = "b2c53124-e422-4786-8135-67c350d71a36"

// Blex, Vexing Pest // Search for Blex — a modal double-faced card (Strixhaven:
// School of Mages, #5502), one file for both faces.
//
// Front, Legendary Creature — Pest {2}{G}, 3/2:
//
//	"Other Pests, Bats, Insects, Snakes, and Spiders you control get +1/+1.
//	 When Blex dies, you gain 4 life."
//
// Back, Sorcery {2}{B}{B} (key "<oracle>#1", the MDFC face keyspace):
//
//	"Look at the top five cards of your library. You may put any number of
//	 them into your hand and the rest into your graveyard. You lose 3 life for
//	 each card you put into your hand this way."
//
// The lord is Squirrel Sovereign's, over five types; it reads the creatures'
// effective subtypes, so a changeling is every one of them. The dies trigger
// is a plain "when this dies" (CR 603.10a).
//
// Search for Blex is one registered take (RegisterTake): up to five cards,
// "any number" with a floor of none, the rest into the graveyard first and
// then 3 life for each card that actually reached the hand, the order the
// sentences print.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     blexOracleID,
		Name:         "Blex, Vexing Pest",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Pest", "Bat", "Insect", "Snake", "Spider"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Blex, Vexing Pest — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     blexOracleID + "#1",
		Name:         "Search for Blex",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Max:      0,
				Optional: true,
				Label:    "Search for Blex — you may put any number of them into your hand and the rest into your graveyard",
				Key:      searchForBlexTake,
			}.Apply(ctx)
		},
	})
}

// searchForBlexTake is "the rest into your graveyard", then "you lose 3 life
// for each card you put into your hand this way". An on-disk key: never
// renamed, never reused.
var searchForBlexTake = effects.RegisterTake("choose-cards/search-for-blex-take",
	effects.TakeDef{Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		if err := effects.TakeRestIntoGraveyard(g, res); err != nil {
			return err
		}
		if len(res.Taken) == 0 {
			return nil
		}
		return g.ChangePlayerLifeForEffect(res.Source, res.Player, -3*len(res.Taken))
	}})
