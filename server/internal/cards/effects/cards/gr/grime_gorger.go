package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grime Gorger — Creature — Horror {2}{B}{G}, 3/3 (Streets of New
// Capenna):
//
//	"Menace
//	 Whenever this creature attacks, exile up to one card of each card type
//	 from defending player's graveyard. Put a +1/+1 counter on this
//	 creature for each card exiled this way."
//
// Not a targeting ability: the cards are chosen as it resolves, from the
// graveyard of the player attacked (or who controls the planeswalker or
// battle attacked, CR 506.3d). The pick is a choose-cards prompt carrying a
// set rule (sncB16OneCardOfEachType): every card picked must be able to
// stand for a different card type, and a card with two types counts for one
// of them. Each card is exiled on its own, and the counter goes on for each
// one that really reached exile, as the printed "for each card exiled this
// way" reads; one that left in response, or that a replacement sent
// elsewhere, adds none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7271c442-2bd1-4d78-a049-b28bd5ee6876",
		Name:            "Grime Gorger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Grime Gorger — exile up to one card of each card type from defending player's graveyard; put a +1/+1 counter on it for each card exiled this way",
				grimeGorgerAsk),
		},
	})
}

func grimeGorgerAsk(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	defender := effects.DefendingPlayerOfAttack(g, ctx.Trigger().Event)
	p := g.PlayerByIDForEffect(defender)
	if p == nil || p.Graveyard == nil || len(p.Graveyard.Cards) == 0 {
		return nil
	}
	cards := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for _, c := range p.Graveyard.Cards {
		cards = append(cards, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: defender,
		Source:     ctx.Source(),
		Question:   "Grime Gorger — exile up to one card of each card type from defending player's graveyard",
		Cards:      cards,
		Min:        0,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    grimeGorgerExile,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

// grimeGorgerExile exiles the picked cards one at a time; each that reaches
// exile adds a counter. An on-disk key: never renamed, never reused.
var grimeGorgerExile = effects.AskedPickRule("choose-cards/grime-gorger-exile", effects.SncB16OneCardOfEachType,
	func(ctx *effects.Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if err := (effects.ExileTarget{Target: id, ThenTail: ctx.ItemTail(grimeGorgerCounter)}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

var grimeGorgerCounter = effects.CardTail("grime-gorger/counter-per-exiled-card", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: len(o.IDs)}.Apply(ctx)
})
