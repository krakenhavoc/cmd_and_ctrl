package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// April O'Neil, Human Element — Legendary Creature — Human Detective
// {3}{U}, 2/5 (TMNT Commander, #5011):
//
//	"Whenever a player casts an artifact, instant, or sorcery spell, you
//	 create a Mutagen token."
//
// Any player's cast, April's controller's included. The trigger resolves
// above the spell, so the Mutagen arrives even if the spell is countered
// (the card's ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "020963ef-24c0-48cc-8776-bc257df684bc",
		Name:         "April O'Neil, Human Element",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, tmcB7APlayerCasts(tmcB7ArtifactInstantOrSorcery()),
				"April O'Neil, Human Element — create a Mutagen token",
				Do(CreateToken{Template: MutagenToken(), N: 1})),
		},
	})
}
