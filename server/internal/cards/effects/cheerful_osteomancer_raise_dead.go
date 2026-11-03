package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cheerful Osteomancer // Raise Dead — a preparation card (CR 722, ADR 0090),
// Secrets of Strixhaven (#4456):
//
//	Cheerful Osteomancer {3}{B}, Creature — Orc Warlock 4/2:
//	"This creature enters prepared. (While it's prepared, you may cast a copy
//	 of its spell. Doing so unprepares it.)"
//	Raise Dead {B}, Sorcery:
//	"Return target creature card from your graveyard to your hand."
//
// "Enters prepared" is SelfEntersPrepared. Raise Dead is Midnight Recovery's
// return without the cipher, so it shares that body
// (returnTargetFromGraveyardToHand): a target that left in response does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	const id = "fd01229f-1c4c-4ad1-bc6f-4023f24d8c71"
	Register(Spec{
		OracleID:     id,
		Name:         "Cheerful Osteomancer",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Raise Dead",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
