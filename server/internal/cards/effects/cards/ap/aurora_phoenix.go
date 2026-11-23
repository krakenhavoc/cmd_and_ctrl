package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aurora Phoenix — Creature — Phoenix {4}{R}{R}, 5/3 (Commander Legends: Battle
// for Baldur's Gate):
//
//	"Flying
//	 Cascade (When you cast this spell, exile cards from the top of your library
//	 until you exile a nonland card that costs less. You may cast it without
//	 paying its mana cost. Put the exiled cards on the bottom in a random order.)
//	 Whenever you cast a spell with cascade, return this card from your graveyard
//	 to your hand."
//
// Cascade is the engine keyword whole. The last ability works from the graveyard
// (Bloodghast's shape: the trigger watches from ZoneGraveyard, CR 113.6): it
// fires on a spell its controller casts that has a cascade trigger of its own,
// and as it resolves returns this card to hand only if it is still in the
// graveyard (CR 400.7).
//
// No simplification.
func init() {
	back := effects.WheneverYouCast(effects.GapB218SpellWithCascade,
		"Aurora Phoenix — return this card from your graveyard to your hand",
		effects.ReturnThisCardFromYourGraveyardToHand)
	back.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:        "a98cc4d0-a322-4142-aee0-83f39b137616",
		Name:            "Aurora Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.Cascade(), back},
	})
}
