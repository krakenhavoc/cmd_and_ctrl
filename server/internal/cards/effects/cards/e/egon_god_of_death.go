package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Egon, God of Death // Throne of Death — modal double-faced card,
// {2}{B} // {B} (Kaldheim).
//
// Front, Legendary Creature — God, 6/6:
//
//	"Deathtouch
//	 At the beginning of your upkeep, exile two cards from your graveyard. If
//	 you can't, sacrifice Egon and draw a card."
//
// With two or more cards in the graveyard the controller names two (a
// mandatory pick) and they are exiled; with fewer, "you can't" and Egon is
// sacrificed and a card is drawn. The graveyard is counted as the trigger
// resolves, and the sacrifice only happens while it is still this Egon.
//
// Back, Legendary Artifact (registered under "<oracle_id>#1", ADR 0034):
//
//	"At the beginning of your upkeep, mill a card.
//	 {2}{B}, {T}, Exile a creature card from your graveyard: Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8fa7bd8f-f712-4dd1-9ffb-e30dfe744389",
		Name:            "Egon, God of Death",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Egon, God of Death — exile two cards from your graveyard; if you can't, sacrifice Egon and draw a card", egonUpkeep),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "8fa7bd8f-f712-4dd1-9ffb-e30dfe744389#1",
		Name:         "Throne of Death",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Throne of Death — mill a card", effects.Do(effects.MillCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}, Exile a creature card from your graveyard: Draw a card.",
			Purpose: game.Purpose{Draws: 1},
			Cost: effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost(),
				effects.ExileFromGraveyard(1, "a creature card", func(c game.Card) bool { return c.IsCreature() })),
			Effect: effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

// egonUpkeep is the upkeep trigger: two cards of the graveyard are exiled, or
// Egon is sacrificed and a card drawn.
func egonUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := g.PlayerByIDForEffect(item.Controller)
	if p != nil && p.Graveyard != nil && p.Graveyard.Size() >= 2 {
		ids := make([]uuid.UUID, 0, p.Graveyard.Size())
		for _, c := range p.Graveyard.Cards {
			ids = append(ids, c.InstanceID)
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    item.Controller,
			FromPlayer: item.Controller,
			Source:     item.SourceCardID,
			Question:   "Egon, God of Death — exile two cards from your graveyard",
			Cards:      ids,
			Min:        2,
			Max:        2,
			LiveZone:   game.ZoneGraveyard,
			ThenKey:    egonExiled,
			Asked:      game.NewAskedItem(item),
		})
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield || ctx.IsNewSourceObjectAsThis(item.SourceCardID) {
		return effects.DrawCards{N: 1}.Apply(ctx)
	}
	if err := (effects.SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
}

// egonExiled exiles the two cards chosen. An on-disk key: never renamed,
// never reused.
var egonExiled = effects.AskedPick("choose-cards/s82-old-b062-egon-exile-two", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) != 2 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(r.Picked, game.KeyedTail{})
})
