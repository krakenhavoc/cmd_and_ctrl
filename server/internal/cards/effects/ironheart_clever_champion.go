package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ironheart, Clever Champion — Legendary Artifact Creature — Human Hero
// {4}{U}, 3/4:
//
//	"Improvise ... Flying
//	 Noncreature spells you cast have improvise."
//
// Improvise is Spec.TapCost; the grant is a GrantedTapCost over
// noncreature spells (SpellsYouCastHaveImprovise, convoke_grant.go).
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "df3d5b86-7b75-4df9-8985-7159cde12261",
		Name:            "Ironheart, Clever Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         Improvise(),
		GrantedTapCosts: []game.GrantedTapCost{SpellsYouCastHaveImprovise(
			"Noncreature spells you cast have improvise.", SpellIs(Noncreature()))},
	})
}
