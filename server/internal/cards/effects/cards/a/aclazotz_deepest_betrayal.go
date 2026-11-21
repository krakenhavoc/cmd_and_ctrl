package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        aclazotzOracleID,
		Name:            "Aclazotz, Deepest Betrayal",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Aclazotz — each opponent discards a card; for each who can't, you draw a card", effects.AclazotzAttack),
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor == source.Controller || ev.Actor == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand()
			}, "Aclazotz — create a 1/1 black Bat creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Bat with flying"), N: 1})),
			effects.OjerDiesReturnTransformed("Aclazotz", nil),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     aclazotzOracleID + "#1",
		Name:         "Temple of the Dead",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "{T}: Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{B}, {T}: Transform this land. Activate only if a player has one or fewer cards in hand and only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			SorcerySpeed: true,
			Condition:    aPlayerHasOneOrFewerCardsInHand,
			Effect:       effects.Do(effects.TransformThis{}),
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
