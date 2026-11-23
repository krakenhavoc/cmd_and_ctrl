package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattered Seraph — Creature — Angel Rogue {4}{W}{U}{B}, 4/4 (Streets of
// New Capenna):
//
//	"Flying
//	 When this creature enters, you gain 3 life.
//	 {2}, Exile this card from your hand: Target land gains "{T}: Add {W},
//	 {U}, or {B}" until this card is cast from exile. You may cast this
//	 card for as long as it remains exiled."
//
// The hand ability is the cycle's shared one (sncB19ExileForLand).
//
// No simplification.
func init() {
	grant, ability := effects.SncB19ExileForLand("Shattered Seraph", "shattered-seraph/tap-wub", "{W|U|B}", "{W}, {U}, or {B}")
	effects.Register(effects.Spec{
		OracleID:        "92eae6af-b915-4eb8-93c4-527f732bb712",
		Name:            "Shattered Seraph",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{LifeGain: 3},
		Grants:          []effects.AbilityGrant{grant},
		Activated:       []effects.ActivatedAbility{ability},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shattered Seraph — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
