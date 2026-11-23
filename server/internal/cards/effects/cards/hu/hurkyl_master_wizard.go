package cardshu

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hurkyl, Master Wizard — Legendary Creature — Human Wizard Advisor {1}{U}{U},
// 2/4 (The Brothers' War, S82):
//
//	"At the beginning of your end step, if you've cast a noncreature spell
//	 this turn, reveal the top five cards of your library. For each card
//	 type among noncreature spells you've cast this turn, you may put a card
//	 of that type from among the revealed cards into your hand. Put the rest
//	 on the bottom of your library in a random order."
//
// The intervening "if" is rechecked on resolution (CR 603.4). The types are
// read from the turn's noncreature-spell tally (CastTally.NoncreatureCardTypes:
// a creature spell adds nothing, an artifact enchantment spell adds both).
// The five are revealed and the controller is asked once per such type, in
// alphabetical order, for a card of that type still among them (Atraxa's
// pattern): a card taken goes to the hand at once and cannot fill a second
// type, so a two-type card fills one slot and leaves the other open. A type
// with no candidate left is skipped, and what is still on the library goes
// to the bottom in a random order once the last type is asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "957df9f4-73e9-4540-b838-52f65d0c62ee",
		Name:         "Hurkyl, Master Wizard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you've cast a noncreature spell this turn",
				Test: func(ctx *effects.Context) bool { return ctx.Game.CastTallyFor(ctx.Controller()).Noncreature > 0 },
			}, effects.AtYourEndStep("Hurkyl, Master Wizard — reveal the top five cards; for each card type among noncreature spells you've cast this turn, you may put a card of that type into your hand; put the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 5,
						"Hurkyl, Master Wizard — reveal the top five cards of your library")
					if len(revealed) == 0 {
						return nil
					}
					return hurkylAsk(ctx, revealed, 0)
				})),
		},
	})
}

// hurkylAsk asks for the card of the first noncreature-spell type at or after
// index `from` that has a candidate among the revealed cards still on the
// library, and bottoms what is left once no type is.
func hurkylAsk(ctx *effects.Context, revealed []uuid.UUID, from int) error {
	g, you := ctx.Game, ctx.Controller()
	left := effects.CardsStillInALibrary(g, revealed)
	types := g.CastTallyFor(you).NoncreatureCardTypesCast()
	for i := from; i < len(types); i++ {
		name := strings.ToLower(types[i])
		var candidates []uuid.UUID
		for _, id := range left {
			if c, ok := g.LookupCardForEffect(id); ok && c.HasCardType(name) {
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
			Question:   "Hurkyl, Master Wizard — you may put a " + name + " card from among the revealed cards into your hand",
			Cards:      candidates,
			Min:        0,
			Max:        1,
			LiveZone:   game.ZoneLibrary,
			ThenKey:    hurkylPicked,
			Carry:      revealed,
			CarryNums:  []int{i},
			Asked:      game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(you, game.ZoneLibrary, left)
}

// hurkylPicked is one type's answer: the card to the hand, then the next
// type. An on-disk key.
var hurkylPicked game.ChooseCardsThen

func init() {
	hurkylPicked = effects.AskedPick("choose-cards/hurkyl-master-wizard-card-type", func(ctx *effects.Context, r game.CardsChosen) error {
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
		return hurkylAsk(ctx, r.Carry, next)
	})
}
