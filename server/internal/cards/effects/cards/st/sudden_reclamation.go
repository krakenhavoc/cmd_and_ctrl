package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Reclamation — {3}{G} Instant (Fate Reforged):
//
//	"Mill four cards, then return a creature card and a land card from your
//	 graveyard to your hand."
//
// Port of Karfell's ordering: the mill finishes first (its tail), then the
// graveyard is read, so a card the mill just put there is a legal pick. Each
// pick is mandatory when a card of its kind is there (Min 1) and the two
// prompts are the controller's, creature first. A card that is both a creature
// and a land can be chosen only once: it has left the graveyard by the second
// pick, which the prompt's live-zone check withdraws.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c34dec60-031a-4cd3-b2e5-db7a02a34d96",
		Name:         "Sudden Reclamation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.MillToZone{Player: ctx.Controller(), N: 4, ThenTail: ctx.ItemTail(s82OldB338ReclaimAfterMill)}.Apply(ctx)
		},
	})
}

// s82OldB338ReclaimAfterMill is the rest of the card, asked once the mill is
// done: the creature pick first, then the land pick from its answer, so the
// player is never asked both at once. An on-disk key: never renamed, never
// reused.
var s82OldB338ReclaimAfterMill = effects.CardTail("then/s82-old-b338-sudden-reclamation", func(ctx *effects.Context, _ game.TailOutcome) error {
	if s82OldB338AskReclaim(ctx.Game, ctx.Controller(), ctx.Source(), game.Card.IsCreature, s82OldB338ReclaimCreature,
		"Sudden Reclamation — return a creature card from your graveyard to your hand") {
		return nil
	}
	s82OldB338AskReclaim(ctx.Game, ctx.Controller(), ctx.Source(), s82OldB338IsLand, s82OldB338ReclaimLand,
		"Sudden Reclamation — return a land card from your graveyard to your hand")
	return nil
})

func s82OldB338IsLand(c game.Card) bool { return c.IsLand() }

// s82OldB338AskReclaim queues the mandatory pick of one matching card in the
// player's graveyard; false when there is none to pick.
func s82OldB338AskReclaim(g *game.Game, player, source uuid.UUID, match func(game.Card) bool, then game.ChooseCardsThen, question string) bool {
	ids := effects.GraveyardCardIDs(effects.KeyedContext(g, player, source), player, func(c game.Card) bool { return !c.IsToken() && match(c) })
	if len(ids) == 0 {
		return false
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     source,
		Question:   question,
		Cards:      ids,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneGraveyard,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    then,
	})
	return true
}

// s82OldB338ReturnPicked puts the picked card into its owner's hand when it is
// still in the graveyard it was picked from.
func s82OldB338ReturnPicked(g *game.Game, r game.CardsChosen) error {
	ctx := effects.KeyedContext(g, r.Chooser, r.Source)
	for _, id := range r.Picked {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// s82OldB338ReclaimCreature returns the picked creature card, then asks for the
// land card. An on-disk key: never renamed, never reused.
var s82OldB338ReclaimCreature = game.RegisterChooseCardsThen("choose-cards/s82-old-b338-reclaim-creature", func(g *game.Game, r game.CardsChosen) error {
	if err := s82OldB338ReturnPicked(g, r); err != nil {
		return err
	}
	s82OldB338AskReclaim(g, r.Chooser, r.Source, s82OldB338IsLand, s82OldB338ReclaimLand,
		"Sudden Reclamation — return a land card from your graveyard to your hand")
	return nil
})

// s82OldB338ReclaimLand returns the picked land card. An on-disk key: never
// renamed, never reused.
var s82OldB338ReclaimLand = game.RegisterChooseCardsThen("choose-cards/s82-old-b338-reclaim-land", s82OldB338ReturnPicked)
