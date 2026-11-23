package cardsn

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Niv-Mizzet Reborn — Legendary Creature — Dragon Avatar {W}{U}{B}{R}{G},
// 6/6:
//
//	"Flying
//	 When Niv-Mizzet enters, reveal the top ten cards of your library. For
//	 each color pair, choose a card that's exactly those colors from among
//	 them. Put the chosen cards into your hand and the rest on the bottom
//	 of your library in a random order."
//
// Atraxa, Grand Unifier's shape. The ten pairs are walked in turn; a pair
// with no card of exactly those two colours asks nothing, and a pair with
// at least one asks for one of them (a choice, not a "may"). A card is
// exactly two colours, so it can be chosen for one pair only. Each pick
// goes to hand as it is made, and the rest go to the bottom in a random
// order once the last pair is done.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10764b35-5c6c-4161-8579-1124d5963a0e",
		Name:            "Niv-Mizzet Reborn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Niv-Mizzet Reborn — reveal the top ten cards; for each color pair, choose a card of exactly those colors and put it into your hand; put the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 10,
						"Niv-Mizzet Reborn — reveal the top ten cards of your library")
					if len(revealed) == 0 {
						return nil
					}
					return nivRebornAsk(ctx, revealed, 0)
				}),
		},
	})
}

// nivRebornPairs are the ten two-colour pairs, in WUBRG order.
var nivRebornPairs = [][2]string{
	{"W", "U"}, {"W", "B"}, {"W", "R"}, {"W", "G"}, {"U", "B"},
	{"U", "R"}, {"U", "G"}, {"B", "R"}, {"B", "G"}, {"R", "G"},
}

var nivRebornPairNames = map[string]string{"W": "white", "U": "blue", "B": "black", "R": "red", "G": "green"}

// nivRebornExactly reports whether the card is exactly the two colours.
func nivRebornExactly(c game.Card, pair [2]string) bool {
	colors := c.EffectiveColors()
	return len(colors) == 2 && slices.Contains(colors, pair[0]) && slices.Contains(colors, pair[1])
}

func nivRebornAsk(ctx *effects.Context, revealed []uuid.UUID, from int) error {
	g, you := ctx.Game, ctx.Controller()
	left := effects.CardsStillInALibrary(g, revealed)
	for i := from; i < len(nivRebornPairs); i++ {
		var candidates []uuid.UUID
		for _, id := range left {
			if c, ok := g.LookupCardForEffect(id); ok && nivRebornExactly(c, nivRebornPairs[i]) {
				candidates = append(candidates, id)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    you,
			FromPlayer: you,
			Source:     ctx.Source(),
			Question:   "Niv-Mizzet Reborn — choose a " + nivRebornPairNames[nivRebornPairs[i][0]] + "-" + nivRebornPairNames[nivRebornPairs[i][1]] + " card from among the revealed cards",
			Cards:      candidates,
			Min:        1,
			Max:        1,
			LiveZone:   game.ZoneLibrary,
			ThenKey:    nivRebornPicked,
			Carry:      revealed,
			CarryNums:  []int{i},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, left)
}

var nivRebornPicked game.ChooseCardsThen

func init() {
	nivRebornPicked = effects.AskedPick("choose-cards/niv-mizzet-reborn-color-pair", func(ctx *effects.Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneLibrary {
				continue
			}
			if err := ctx.Game.BounceToHandForEffect(id); err != nil {
				return err
			}
		}
		next := 0
		if len(r.Nums) > 0 {
			next = r.Nums[0] + 1
		}
		return nivRebornAsk(ctx, r.Carry, next)
	})
}
