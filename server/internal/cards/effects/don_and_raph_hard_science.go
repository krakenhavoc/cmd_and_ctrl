package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Don & Raph, Hard Science — Legendary Creature — Mutant Ninja Turtle
// {1}{U/R}{U/R}, 2/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Menace
//	 Whenever Don & Raph attack, the next noncreature spell you cast this
//	 turn has affinity for artifacts. (It costs {1} less to cast for each
//	 artifact you control.)"
//
// Saheeli, the Gifted's affinity promise (game.NextSpellPromise with
// AffinityFor) narrowed to noncreature spells by the promise's filter, so
// a creature spell cast first neither gets the reduction nor spends it.
// The artifacts are counted as the spell is priced (CR 702.41a), and the
// promise ends with the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "01758a4a-ac29-47ad-8e88-158374f1ae4c",
		Name:            "Don & Raph, Hard Science",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Don & Raph, Hard Science — the next noncreature spell you cast this turn has affinity for artifacts",
				Do(GrantNextSpellPromise{From: "Don & Raph, Hard Science", Promise: game.NextSpellPromise{
					Filter:      game.PermissionFilter{NoncreatureOnly: true},
					AffinityFor: "Artifact",
					Text:        "The next noncreature spell you cast this turn has affinity for artifacts.",
				}})),
		},
	})
}
