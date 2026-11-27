package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Nocturnus — Creature — Vampire {1}{B}{B}{B}, 3/3 (Magic 2013):
//
//	"Play with the top card of your library revealed.
//	 As long as the top card of your library is black, this creature and other
//	 Vampire creatures you control get +2/+1 and have flying."
//
// Oracle of Mul Daya's revealed top card (LibraryTopRevealed, CR 401.5) and two
// live layer statics that read that card: the +2/+1 in layer 7c and flying in
// layer 6, on this creature and every Vampire creature its controller controls.
// The top card's colour is its own as a card in a library (CR 202.2); an empty
// library has none. A change of the top card re-evaluates both at once.
//
// No simplification.
func init() {
	applies := func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.Controller != source.Controller || !target.IsCreature() {
			return false
		}
		if target.InstanceID != source.InstanceID && !target.HasSubtype(source.TextSubtype("Vampire")) {
			return false
		}
		return vampireNocturnusTopIsBlack(g, source.Controller)
	}
	flying := effects.KeywordGrant(applies, "flying")
	flying.Label = "As long as the top card of your library is black, this creature and other Vampire creatures you control have flying."
	effects.Register(effects.Spec{
		OracleID:          "2109c49e-dc81-4a10-8265-3cf8c90db523",
		Name:              "Vampire Nocturnus",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopRevealed,
		Static: []game.StaticAbility{{
			Label:     "As long as the top card of your library is black, this creature and other Vampire creatures you control get +2/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: applies,
			Apply:     effects.PumpBy(2, 1),
		}, flying},
	})
}

// vampireNocturnusTopIsBlack reads the top card of the player's library.
func vampireNocturnusTopIsBlack(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Library == nil {
		return false
	}
	top, err := p.Library.Top()
	if err != nil {
		return false
	}
	for _, col := range top.EffectiveColors() {
		if col == "B" {
			return true
		}
	}
	return false
}
