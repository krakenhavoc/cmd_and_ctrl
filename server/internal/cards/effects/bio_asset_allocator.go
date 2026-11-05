package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bio-Asset Allocator — Artifact Creature — Borg {3}{B}{B}, 3/4 (Star
// Trek Commander, #4974):
//
//	"Artifact creatures you control have "Whenever this creature
//	 attacks, each opponent loses 1 life and you gain 1 life.""
//
// A layer-6 grant of a catalog bundle (ADR 0093), so the trigger is each
// recipient's own: "this creature" is the artifact creature that
// attacked and "you" is its controller. The Allocator is an artifact
// creature itself, so it has the ability too. The grant reads post-layer
// types, so an artifact that is a creature only for now (a crewed
// Vehicle) has it while it is one; two Allocators grant two instances,
// which trigger separately. The drain is one point from each opponent
// and one point of life, whatever each opponent actually lost.
//
// No simplification.
const bioAssetAllocatorGrant = "bio-asset-allocator/attack-drain-one"

func init() {
	const text = "Whenever this creature attacks, each opponent loses 1 life and you gain 1 life."
	Register(Spec{
		OracleID:     "1ca22ca0-eca3-4578-9746-47d443d5fb73",
		Name:         "Bio-Asset Allocator",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: bioAssetAllocatorGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Bio-Asset Allocator — each opponent loses 1 life and you gain 1 life",
					Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(trekB47ArtifactCreaturesYouControl, bioAssetAllocatorGrant), "Artifact creatures you control have \""+text+"\"")},
	})
}
