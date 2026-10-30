package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zarda, the Power Princess — Legendary Creature — Human Warrior Hero
// {3}{W}, 2/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Other Heroes you control have exalted."
//
// Captain America, Super-Soldier's layer-6 Hero grant (TribeFilter, others,
// yours) with the exalted keyword. Exalted is the engine's derived keyword
// trigger (CR 702.83, game/exalted.go): each granted instance is a trigger
// of its own, and the instance on a lone attacker counts for every exalted
// source among your permanents, not just the attacker's own.
//
// No simplification.
func init() {
	heroes := TribeFilter{Tribes: []string{"Hero"}, Others: true, YoursOnly: true}
	grant := KeywordGrant(heroes.Matches, game.KeywordExalted)
	grant.Label = "Other Heroes you control have exalted."
	Register(Spec{
		OracleID:        "c7b158b9-19f7-47b6-a97c-74023dfe73b9",
		Name:            "Zarda, the Power Princess",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{grant},
	})
}
