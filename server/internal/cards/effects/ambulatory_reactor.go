package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ambulatory Reactor — Artifact Creature — Borg {3}{U}, 2/5 (Star Trek
// Commander, #4974):
//
//	"Artifact creatures you control have "{T}: Add {C}. Spend this mana
//	 only to cast an artifact spell or activate an ability of an artifact
//	 source.""
//
// A layer-6 grant of a mana ability (ADR 0093, Elven Chorus's shape), read
// on post-layer types: each artifact creature you control has its own
// copy, the Reactor included, and an artifact that stops being a creature
// loses it with the layer pass. The {T} in the cost is the host's, so a
// summoning-sick host cannot use it (CR 302.6).
//
// The restriction is one ManaRestrictAnyOf tag (CR 106.6): a cast whose
// spell is an artifact, or an activation whose source is an artifact (a
// spend context reads the ability's source for an activation, as Eldrazi
// Temple's does). Anything else, a nonartifact spell, a nonartifact
// source's ability, an unlock or a face-up turn, can't spend it.
//
// No simplification.
const ambulatoryReactorGrant = "ambulatory-reactor/artifact-colorless"

func init() {
	const text = "{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source."
	Register(Spec{
		OracleID:     "fea853c7-7dc8-45c7-829b-d857c65b6ee8",
		Name:         "Ambulatory Reactor",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  ambulatoryReactorGrant,
			Text: text,
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    text,
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
				)},
			}},
		}},
		Static: []game.StaticAbility{Labelled(
			GrantAbilities(trekB34ArtifactCreaturesYouControl, ambulatoryReactorGrant),
			"Artifact creatures you control have \""+text+"\"")},
	})
}
