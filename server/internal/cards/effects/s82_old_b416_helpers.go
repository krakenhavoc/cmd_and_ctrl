package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b416_helpers.go — Gatecrash batch 416 (#5565): Deathpact Angel's
// Cleric token.

// OldB416ClericToken is Deathpact Angel's Cleric for CreateToken.
func OldB416ClericToken() game.Card { return TokenFromCatalog(PrintedOldB416Cleric) }

// PrintedOldB416Cleric is the Cleric as PRINTED (Gatecrash token): a 1/1
// white and black Cleric creature with "{3}{W}{B}{B}, {T}, Sacrifice this
// creature: Return a card named Deathpact Angel from your graveyard to the
// battlefield." Its slug is an on-disk identity: never renamed, never reused.
func PrintedOldB416Cleric() TokenTemplate {
	return TokenTemplate{
		Slug: "cleric-returns-deathpact-angel",
		Card: game.Card{
			Name:      "Cleric",
			TypeLine:  "Token Creature — Cleric",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W", "B"},
		},
		Text: "{3}{W}{B}{B}, {T}, Sacrifice this creature: Return a card named Deathpact Angel from your graveyard to the battlefield.",
		Activated: []game.ActivatedAbilityShape{{
			Label:   "{3}{W}{B}{B}, {T}, Sacrifice this creature: Return a card named Deathpact Angel from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{3}{W}{B}{B}", Tap: true, SacrificeSelf: true},
			Effect:  oldB416ReturnAngel,
		}},
	}
}

// oldB416ReturnAngel returns a card named Deathpact Angel from its
// controller's graveyard: the one card when there is one, the controller's
// pick among several, nothing when there is none (the cost is already paid).
func oldB416ReturnAngel(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.Name == "Deathpact Angel" })
	switch len(ids) {
	case 0:
		return nil
	case 1:
		return ReturnFromGraveyard{Target: ids[0], Dest: game.ZoneBattlefield}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Cleric — return a card named Deathpact Angel from your graveyard to the battlefield",
		Cards:    ids,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  oldB416AngelPicked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// oldB416AngelPicked returns the chosen Deathpact Angel. An on-disk key:
// never renamed, never reused.
var oldB416AngelPicked = AskedPick("choose-cards/s82-old-b416-cleric-return-angel", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if id == uuid.Nil {
			continue
		}
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(ctx)
	}
	return nil
})
