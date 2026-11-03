package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b24_helpers.go — shared readers and tokens for Secrets of Strixhaven
// batch 24 (#4456).

// PestAttackToken is the Secrets of Strixhaven Pest: a 1/1 black and green
// Pest creature token with "Whenever this token attacks, you gain 1 life."
// (Essenceknit Scholar). It is not the Strixhaven Pest of PestToken, whose
// life comes when it dies.
func PestAttackToken() game.Card { return tokenFromCatalog(printedPestAttackToken) }

// printedPestAttackToken is that Pest as PRINTED. "You" in a token's own
// trigger is the token's controller (the item's Controller), so a Pest that
// changes hands gains life for its new controller.
func printedPestAttackToken() tokenTemplate {
	return tokenTemplate{
		Slug: "pest-attack",
		Card: game.Card{
			Name:      "Pest",
			TypeLine:  "Token Creature — Pest",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"B", "G"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Pest — you gain 1 life", Do(GainLife{Amount: 1})),
		},
		Text: "Whenever this token attacks, you gain 1 life.",
	}
}

// socB24CardsWerePutIntoExileThisTurn is "one or more cards were put into
// exile this turn" (Ennis, Debate Moderator): any card, any owner, from any
// zone, by any route that logs a zone move to exile. A token that was exiled
// is not a card (CR 111.7, it ceases to exist) and is not counted. The tally
// only rises within a turn, so a trigger that held as the step began still
// holds as it resolves (CR 603.4).
func socB24CardsWerePutIntoExileThisTurn(g *game.Game) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneExile {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && !c.IsToken() {
			return true
		}
	}
	return false
}
