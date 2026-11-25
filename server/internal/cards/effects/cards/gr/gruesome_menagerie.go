package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gruesome Menagerie — Sorcery {3}{B}{B} (Guilds of Ravnica, #5525):
//
//	"Choose a creature card with mana value 1 in your graveyard, then do the
//	 same for creature cards with mana value 2 and 3. Return those cards to the
//	 battlefield."
//
// Three choices, one after another, each its own keyed choose_cards prompt over
// the creature cards of that mana value in the caster's graveyard (a value with
// no card is skipped, not asked); the number on the prompt is the value just
// asked and the carry is what was chosen so far. Only after the last do the
// chosen cards enter, all together as one entry (CR 614.12, 603.6a), and one
// that has left the graveyard in the meantime is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f95fda8c-0737-447f-9212-1c459062b9e7",
		Name:         "Gruesome Menagerie",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return gruesomeMenagerieAsk(ctx.Game, ctx.Controller(), ctx.Source(), item, 1, nil)
		},
	})
}

// gruesomeMenagerieAsk asks for the creature card of mana value `mv` or moves
// on to the next value; past 3 it returns what was chosen.
func gruesomeMenagerieAsk(g *game.Game, player, source uuid.UUID, item *game.StackItem, mv int, chosen []uuid.UUID) error {
	ctx, _ := effects.AskedContext(g, item, player, source, nil)
	for ; mv <= 3; mv++ {
		candidates := effects.GraveyardCardIDs(ctx, player, func(c game.Card) bool {
			return c.IsCreature() && !c.IsToken() && c.ManaValue() == mv
		})
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    player,
			FromPlayer: player,
			Source:     source,
			Question:   "Gruesome Menagerie — choose a creature card with mana value " + string(rune('0'+mv)) + " in your graveyard",
			Cards:      candidates,
			Min:        1,
			Max:        1,
			ThenKey:    gruesomeMenagerieChosen,
			LiveZone:   game.ZoneGraveyard,
			Carry:      append([]uuid.UUID(nil), chosen...),
			CarryNums:  []int{mv},
			Asked:      game.NewAskedItem(item),
		})
		return nil
	}
	if len(chosen) == 0 {
		return nil
	}
	return effects.ReturnFromGraveyardTogether{Targets: chosen}.Apply(ctx)
}

// gruesomeMenagerieChosen records one answer and asks for the next value. An
// on-disk key: never renamed, never reused.
var gruesomeMenagerieChosen game.ChooseCardsThen

func init() {
	// Registered here, not in the var, because the continuation asks the next
	// question and so refers back to this variable.
	gruesomeMenagerieChosen = game.RegisterChooseCardsThen("choose-cards/gruesome-menagerie",
		func(g *game.Game, r game.CardsChosen) error {
			chosen := append([]uuid.UUID(nil), r.Carry...)
			if r.Answered {
				chosen = append(chosen, r.Picked...)
			}
			mv := 1
			if len(r.Nums) > 0 {
				mv = r.Nums[0]
			}
			return gruesomeMenagerieAsk(g, r.Chooser, r.Source, r.Asked, mv+1, chosen)
		})
}
