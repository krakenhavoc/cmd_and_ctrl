package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toph, the First Metalbender — Legendary Creature — Human Warrior Ally
// {1}{R}{G}{W}, 3/3:
//
//	"Nontoken artifacts you control are lands in addition to their other
//	 types. (They don't gain the ability to {T} for mana.)
//	 At the beginning of your end step, earthbend 2."
//
// The first line is Ashaya's layer-4 type change (CR 613.1d) without a
// subtype: the card type Land is added and nothing else, so no mana
// ability arrives with it (CR 305.6 grants one only to basic land
// types). It changes permanents only; it does not let an artifact card
// be played as a land (the ruling), which is the battlefield-only
// AppliesTo. An artifact made a land this way is a legal earthbend
// target.
//
// The end-step trigger earthbends 2 ("target land you control").
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "fbfef028-b10c-4ade-9503-d9af5b0c211b",
		Name:         "Toph, the First Metalbender",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer4Type,
			Label: "Nontoken artifacts you control are lands in addition to their other types.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsArtifact() && !target.IsToken() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Types, "Land") {
					c.Types = append(c.Types, "Land")
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Toph, the First Metalbender — earthbend 2", avatarB22Earthbend(2)),
				EarthbendTargets()),
		},
	})
}
