package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Djeru and Hazoret — Legendary Creature — Human God {2}{R}{R}{W}, 5/4:
//
//	"As long as you have one or fewer cards in hand, Djeru and Hazoret has
//	 vigilance and haste.
//	 Whenever Djeru and Hazoret attacks, look at the top six cards of your
//	 library. You may exile a legendary creature card from among them. Put
//	 the rest on the bottom of your library in a random order. Until end of
//	 turn, you may cast the exiled card without paying its mana cost."
//
// The two statics are layer 6 keyword grants on the source, on while its
// controller has at most one card in hand (re-read every recompute).
//
// The attack trigger is Florian, Voldaren Scion's look-and-exile: the
// looker is offered only the legendary creature cards among the six
// (Min 0, since it is a "may"), the pick is exiled with a free,
// until-end-of-turn cast permission (an empty Duration is this turn, and
// the cast is an ordinary one, so it keeps its own timing, CR 601.3), and
// every other looked-at card still in the library goes to the bottom in a
// random order. With nothing to pick, all six go to the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c0061ab-4123-4b0c-8286-3207e8f9497d",
		Name:         "Djeru and Hazoret",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			djeruHandKeyword("vigilance"),
			djeruHandKeyword("haste"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Djeru and Hazoret — look at the top six cards; you may exile a legendary creature card from among them, put the rest on the bottom in a random order, and cast it free this turn", djeruLook),
		},
	})
}

func djeruHandKeyword(keyword string) game.StaticAbility {
	ab := KeywordGrant(whileYouHave(youHaveOneOrFewerCardsInHand, selfCreature), keyword)
	ab.Label = "As long as you have one or fewer cards in hand, has " + keyword + "."
	ab.DependsOnHandSize = true
	return ab
}

func youHaveOneOrFewerCardsInHand(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	return p != nil && p.Hand != nil && p.Hand.Size() <= 1
}

func djeruLook(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	looked := g.LookAtTopOfLibraryForEffect(player, 6)
	if len(looked) == 0 {
		return nil
	}
	var legends []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() && c.IsLegendary() {
			legends = append(legends, id)
		}
	}
	if len(legends) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Djeru and Hazoret — exile a legendary creature card from among them (you may cast it free this turn); the rest go on the bottom in a random order",
		Cards:    legends,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  djeruExile,
		Carry:    looked,
	})
	return nil
}

var djeruExile = game.RegisterChooseCardsThen("choose-cards/s82-mom-b21-djeru-exile", func(g *game.Game, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, r.Carry) {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if picked != uuid.Nil {
		if err := g.ExileCardWithPermissionForEffect(picked, game.CastPermission{
			Player:     r.Chooser,
			Cost:       "{0}",
			CastOnly:   true,
			Source:     r.Source,
			SourceName: "Djeru and Hazoret",
			Label:      "Djeru and Hazoret — cast the exiled card without paying its mana cost",
		}); err != nil {
			return err
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, rest)
})
