package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mysidian Elder — Creature — Human Wizard {2}{R}, 1/3 (batch 41,
// #448):
//
//	"When this creature enters, create a 0/1 black Wizard creature
//	 token with "Whenever you cast a noncreature spell, this token
//	 deals 1 damage to each opponent.""
//
// Skipped on batch 41 for the "Triggered and static abilities on
// non-copy tokens" seam; it ships with ADR 0083 (#1248).
//
// The Wizard is the same printed token Cornered by Black Mages makes,
// so it is the same catalog template (`token:wizard`, tokens.go) and
// not a second one — a token's identity is its template, and two
// cards that print the same token print the same object.
//
// Worth having as a proof because the token's trigger watches an
// EVENT SOMEWHERE ELSE, not its own death or its own attack: the
// Wizard pings whenever its controller casts a noncreature spell, so
// the harvester has to find the ability by walking the battlefield
// and reading a token's catalog entry, which is the path a dies
// trigger never exercises.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "10039992-d51a-47e7-9a70-02fe2227c163",
		Name:         "Mysidian Elder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mysidian Elder — create a 0/1 black Wizard",
				Do(CreateToken{Template: NoncreatureCastWizardToken(), N: 1})),
		},
	})
}
