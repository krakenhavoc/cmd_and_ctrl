package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Three Seasons — Enchantment — Saga {G}{U} (Kaldheim, #5505):
//
//	"I — Mill three cards.
//	 II — Return up to two target snow permanent cards from your graveyard to
//	 your hand.
//	 III — Choose three cards in each graveyard. Their owners shuffle those
//	 cards into their libraries."
//
// Chapter II's clause is "up to two" snow permanent cards in your graveyard,
// each rechecked as the chapter resolves (CR 608.2b).
//
// Chapter III is not a target: as it resolves the Saga's controller chooses
// three cards in each graveyard in turn (all of them when a graveyard holds
// three or fewer, so nothing is asked), asking for one graveyard at a time and
// moving nothing until every graveyard has been chosen from. Then the chosen
// cards go to their owners' libraries and each owner whose library took a card
// shuffles it.
//
// No simplification.
func init() {
	const name = "The Three Seasons"
	effects.Register(effects.Spec{
		OracleID:     "a38ebf4a-064c-4517-a553-15f0a252c159",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "mill three cards"),
				effects.Do(effects.MillCards{N: 3})),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "return up to two target snow permanent cards from your graveyard to your hand"),
				effects.TargetCardInGraveyard("up to two target snow permanent cards from your graveyard", effects.YouOwn(), effects.Permanent(), effects.Snow()).WithCount(0, 2),
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(3, effects.SagaChapterLabel(name, 3, "choose three cards in each graveyard; their owners shuffle those cards into their libraries"),
				func(g *game.Game, item *game.StackItem) error {
					return threeSeasonsAsk(effects.NewContext(g, item), 0, nil)
				}),
		},
	})
}

// threeSeasonsSeats is every player still in the game, the Saga's controller
// first and the rest in seat order.
func threeSeasonsSeats(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID == controller {
			out = append(out, p.ID)
		}
	}
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.ID != controller {
			out = append(out, p.ID)
		}
	}
	return out
}

// threeSeasonsAsk chooses from the graveyard of seat `from` onward. `picked`
// is what earlier graveyards gave. When every graveyard has been chosen from,
// everything moves together.
func threeSeasonsAsk(ctx *effects.Context, from int, picked []uuid.UUID) error {
	g := ctx.Game
	seats := threeSeasonsSeats(g, ctx.Controller())
	for i := from; i < len(seats); i++ {
		p := g.PlayerByIDForEffect(seats[i])
		if p == nil || len(p.Graveyard.Cards) == 0 {
			continue
		}
		var ids []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			ids = append(ids, c.InstanceID)
		}
		if len(ids) <= 3 {
			picked = append(picked, ids...)
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: seats[i],
			Source:     ctx.Source(),
			Question:   "The Three Seasons — choose three cards in " + p.Name + "'s graveyard; their owner shuffles them into their library",
			Cards:      ids,
			Min:        3,
			Max:        3,
			Zone:       game.ZoneGraveyard,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    threeSeasonsChosen,
			Carry:      append([]uuid.UUID(nil), picked...),
			CarryNums:  []int{i},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return threeSeasonsShuffle(ctx, picked)
}

// threeSeasonsChosen records one graveyard's three and asks for the next. A
// prompt that ended unanswered takes the first three. An on-disk key: never
// renamed, never reused.
var threeSeasonsChosen game.ChooseCardsThen

func init() {
	threeSeasonsChosen = effects.ItemChoice("choose-cards/the-three-seasons-graveyard", threeSeasonsChosenBody)
}

func threeSeasonsChosenBody(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Nums) != 1 {
		return nil
	}
	picked := append([]uuid.UUID(nil), r.Carry...)
	if r.Answered {
		picked = append(picked, r.Picked...)
	} else if len(r.Candidates) >= 3 {
		picked = append(picked, r.Candidates[:3]...)
	}
	return threeSeasonsAsk(ctx, r.Nums[0]+1, picked)
}

// threeSeasonsShuffle puts every chosen card still in a graveyard into its
// owner's library, then each owner who received one shuffles.
func threeSeasonsShuffle(ctx *effects.Context, picked []uuid.UUID) error {
	var owners []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, id := range picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneLibrary}).Apply(ctx); err != nil {
			return err
		}
		if !seen[c.Owner] {
			seen[c.Owner] = true
			owners = append(owners, c.Owner)
		}
	}
	for _, o := range owners {
		if err := (effects.ShuffleLibrary{Player: o}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
