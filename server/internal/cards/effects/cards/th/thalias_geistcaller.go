package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thalia's Geistcaller — Creature — Human Cleric {2}{W}, 3/1 (Commander
// 2019, #5517):
//
//	"Lifelink
//	 Whenever you cast a spell from your graveyard, create a 1/1 white
//	 Spirit creature token with flying.
//	 Sacrifice a Spirit: This creature gains indestructible until end of
//	 turn."
//
// The cast trigger is Devoted Grafkeeper's condition: the spell's announce
// record says it came from a graveyard and you own it, so flashback,
// escape, disturb and a granted graveyard cast all count, and a spell cast
// out of an opponent's graveyard does not. The token is Kykar's white
// flying Spirit. The sacrifice is a cost over any Spirit you control, a
// token or a Spirit card (CR 602.2b), and the grant is a layer 6 keyword on
// the Geistcaller until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee8cf4b4-2e23-4f01-a9fc-7b47ddda31b7",
		Name:            "Thalia's Geistcaller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastASpellFromYourGraveyard,
				"Thalia's Geistcaller — create a 1/1 white Spirit with flying",
				effects.Do(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Spirit: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(1, "a Spirit", effects.HasSubtype("Spirit")),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Thalia's Geistcaller — indestructible until end of turn"),
		}},
	})
}
