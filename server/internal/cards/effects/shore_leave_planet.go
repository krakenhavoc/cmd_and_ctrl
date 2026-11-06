package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shore Leave Planet — Land — Planet (Star Trek Commander, #4974):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 {4}{U}: This land becomes a copy of target creature you control
//	 until end of turn, except it isn't legendary and it's a Robot
//	 artifact in addition to its other types."
//
// Hall of Echoes' copy ability with an "except" clause: the legendary
// supertype is dropped and Artifact and Robot are added as the values
// are copied, so they are copiable values (CR 707.9b). As a copy it is
// no longer a land and loses {T}: Add {U} until the effect ends at
// cleanup; it is still the same permanent, so it can attack only if it
// has been under its controller's control since their turn began
// (CR 302.6). A target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5f50ad0-a9b3-4bfc-aaf4-55199e489265",
		Name:         "Shore Leave Planet",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}{U}: This land becomes a copy of target creature you control until end of turn, except it isn't legendary and it's a Robot artifact in addition to its other types.",
			Cost:    ManaCost("{4}{U}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  trekB43ShoreLeaveCopy,
		}},
	})
}
