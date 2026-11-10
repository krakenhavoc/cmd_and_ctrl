package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const aclazotzOracleID = "fcdfe9d5-2743-4d3e-ab57-bf0f96beaa15"

// Aclazotz, Deepest Betrayal — Legendary Creature — Bat God {3}{B}{B}, 4/4,
// the front of a transforming card whose back is Temple of the Dead (The
// Lost Caverns of Ixalan, #4470):
//
//	"Flying, lifelink
//	 Whenever Aclazotz attacks, each opponent discards a card. For each
//	 opponent who can't, you draw a card.
//	 Whenever an opponent discards a land card, create a 1/1 black Bat
//	 creature token with flying.
//	 When Aclazotz dies, return it to the battlefield tapped and
//	 transformed under its owner's control."
//
// The attack trigger asks each opponent for their own discard; an opponent
// with no card in hand can't discard, and you draw one card for each such
// opponent (CR 701.9b, 608.2c: both counted as the ability resolves). The
// Bat trigger watches an opponent discarding a land card, whatever made
// them; the card is read in the graveyard it was just discarded to. The
// dies trigger is Ojer Axonil's (CR 712.14a), and a copy that isn't a
// transforming card does not come back (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        aclazotzOracleID,
		Name:            "Aclazotz, Deepest Betrayal",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Aclazotz — each opponent discards a card; for each who can't, you draw a card", aclazotzAttack),
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller || ev.Actor == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand()
			}, "Aclazotz — create a 1/1 black Bat creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 black Bat with flying"), N: 1})),
			ojerDiesReturnTransformed("Aclazotz", nil),
		},
	})
	Register(Spec{
		OracleID:     aclazotzOracleID + "#1",
		Name:         "Temple of the Dead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "{T}: Add {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}{B}, {T}: Transform this land. Activate only if a player has one or fewer cards in hand and only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{B}"), TapCost()),
			SorcerySpeed: true,
			Condition:    aPlayerHasOneOrFewerCardsInHand,
			Effect:       Do(TransformThis{}),
		}},
	})
}

// aPlayerHasOneOrFewerCardsInHand is the activation instruction of Temple
// of the Dead (CR 602.5b): any player at the table, you included.
func aPlayerHasOneOrFewerCardsInHand(g *game.Game, _, _ uuid.UUID) bool {
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.Hand != nil && p.Hand.Size() <= 1 {
			return true
		}
	}
	return false
}

// aclazotzAttack makes each opponent discard a card of their choice, as one
// run, then draws a card for each opponent who didn't (aclazotzDraw).
func aclazotzAttack(g *game.Game, item *game.StackItem) error {
	return g.EachPlayerDiscardsThenKeyedForEffect(item.Controller,
		game.DiscardPrompt{Source: item.SourceCardID, N: 1},
		aclazotzDraw, []uuid.UUID{item.Controller})
}

// aclazotzDraw is "for each opponent who can't, you draw a card", run once
// after the last opponent's discard has landed: an opponent who discarded
// no card, because they had none to discard, is one. The controller is the
// carry. An on-disk key: never renamed, never reused.
var aclazotzDraw = game.RegisterPromptRunThen("prompt-run/aclazotz-draw", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	you := r.Carry[0]
	discards := r.Discards()
	cant := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if len(discards.By(p.ID)) == 0 {
			cant++
		}
	}
	if cant == 0 {
		return nil
	}
	return g.DrawNForEffect(you, cant)
})
