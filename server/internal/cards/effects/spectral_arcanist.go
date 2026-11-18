package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectral Arcanist — Creature — Spirit Wizard {3}{U}, 3/2 (Innistrad:
// Crimson Vow):
//
//	"Flying
//	 When this creature enters, you may cast an instant or sorcery spell
//	 with mana value less than or equal to the number of Spirits you
//	 control from a graveyard without paying its mana cost. If that spell
//	 would be put into a graveyard, exile it instead."
//
// Nothing is targeted: the spell is chosen as the trigger resolves
// (CR 608.2g), from any player's graveyard. The Spirits are counted then,
// this Arcanist included if it is still there, and the instants and
// sorceries whose mana value fits are locked as the set (CR 611.2c; a card
// in a graveyard has X of 0, CR 202.3e). One permission names them all and
// is spent by the first cast (CastsLeft 1: "an instant or sorcery spell"),
// free ({0}, additional costs are still paid), at any timing, offered
// through the resolution prompt. The spell is exiled instead of going to a
// graveyard whether it resolves, is countered or fizzles. Declining leaves
// every card where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "aba60536-ffbd-480c-8e8f-9639bdc53d4b",
		Name:            "Spectral Arcanist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spectral Arcanist — you may cast an instant or sorcery spell with mana value less than or equal to the number of Spirits you control from a graveyard without paying its mana cost",
				spectralArcanistFreeCast),
		},
	})
}

// spectralArcanistFreeCast is the enters trigger's resolution.
func spectralArcanistFreeCast(g *game.Game, item *game.StackItem) error {
	// "Spirits" is a word of the Arcanist's text (CR 612.2).
	spiritWord := NewContext(g, item).TextSubtype("Spirit")
	spirits := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype(spiritWord) })
	var cards []game.Card
	var ids []uuid.UUID
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if !c.IsInstant() && !c.IsSorcery() {
				continue
			}
			if mv, ok := g.ManaValueForEffect(c); !ok || mv > spirits {
				continue
			}
			cards = append(cards, c)
			ids = append(ids, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:                  item.Controller,
		Zone:                    game.ZoneGraveyard,
		Duration:                g.UntilEndOfTurnDuration(),
		Cost:                    "{0}",
		AltCostKey:              game.AltCostKeyGrantedCost,
		Timing:                  game.TimingFlash,
		CastOnly:                true,
		CastsLeft:               1,
		ExileInsteadOfGraveyard: true,
		// A pass closes the window; nothing moves, the cards stay in the
		// graveyards.
		LapseOnPass: game.LapseStaysInHand,
		Source:      item.SourceCardID,
		SourceName:  "Spectral Arcanist",
		Label:       "Spectral Arcanist — cast an instant or sorcery spell from a graveyard without paying its mana cost",
	}, cards)
	offerResolutionCastOver(g, item.Controller, item.SourceCardID, "Spectral Arcanist",
		game.MayCastKeywordFreeCast, declineLeaveInGraveyard, ids...)
	return nil
}
