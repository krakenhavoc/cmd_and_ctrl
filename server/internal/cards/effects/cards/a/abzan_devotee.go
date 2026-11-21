package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abzan Devotee — Creature — Dog Cleric {1}{B}, 2/1 (Tarkir: Dragonstorm,
// #4463):
//
//	"{1}: Add {W}, {B}, or {G}. Activate only once each turn.
//	 {2}{B}: Return this card from your graveyard to your hand."
//
// The mana ability is a Signet-style filter (`ManaAbilityCost.Mana`) with
// no tap, limited by the declarative OncePerTurn bit, which the engine
// counts per object and per label: a Devotee that leaves and comes back is
// a new object with a fresh use (CR 400.7). The second ability works only
// from the graveyard, Sanitarium Skeleton's shape; if the card has left the
// graveyard by the time it resolves, nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d02ee2f-b7bd-40b0-b9be-2420991b2be0",
		Name:         "Abzan Devotee",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:        effects.ManaAbilityCost{Mana: "{1}"},
			Produced:    "{W|B|G}",
			Label:       "{1}: Add {W}, {B}, or {G}",
			OncePerTurn: true,
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
