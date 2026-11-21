package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duskmantle Seer — Creature — Vampire Wizard {2}{U}{B}, 4/4 (Ravnica:
// Clue Edition):
//
//	"Flying
//	 At the beginning of your upkeep, each player reveals the top card of
//	 their library, loses life equal to that card's mana value, then puts
//	 it into their hand."
//
// Darkstar Augur's reveal for every player in turn order from the active
// player (CR 101.4), run as the three steps of the text across all players:
// everyone reveals, everyone loses the life (a land or a card with no mana
// cost costs nothing, and no life change of zero is announced), then
// everyone puts the revealed card into their hand. The life is read as the
// card sits revealed on top of the library, so a player at 0 life or less
// still takes the card before the state check removes them (CR 704.5a). A
// player with an empty library reveals nothing and loses nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8d9c7e2-8ae7-4e1c-9859-bb3b85a48117",
		Name:            "Duskmantle Seer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(duskmantleSeerLabel, duskmantleSeerUpkeep),
		},
	})
}

const duskmantleSeerLabel = "Duskmantle Seer — each player reveals the top card of their library, loses life equal to its mana value, then puts it into their hand"

var duskmantleSeerTake = effects.RegisterTake("choose-cards/duskmantle-seer-take", effects.TakeDef{})

func duskmantleSeerUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	players := effects.ApnapPlayers(g)
	revealed := make([][]uuid.UUID, len(players))
	for i, p := range players {
		revealed[i] = append(revealed[i], g.RevealTopOfLibraryForEffect(p, ctx.Source(), 1, duskmantleSeerLabel)...)
	}
	for i, p := range players {
		for _, id := range revealed[i] {
			c, ok := g.LookupCardForEffect(id)
			if !ok || c.ManaValue() <= 0 {
				continue
			}
			if err := g.ChangePlayerLifeForEffect(ctx.Source(), p, -c.ManaValue()); err != nil {
				return err
			}
		}
	}
	for i, p := range players {
		if len(revealed[i]) == 0 {
			continue
		}
		if err := (effects.TakeFromLibraryToHand{
			Player: p, Cards: revealed[i], All: true, Label: duskmantleSeerLabel, Key: duskmantleSeerTake,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
