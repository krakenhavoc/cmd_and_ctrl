package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One, Borg Evolved — Legendary Artifact Creature — Borg Artificer
// {4}{W}{U}{B}, 3/3 (Star Trek Commander, #4974):
//
//	"Artifact creatures you control have flying, double strike,
//	 vigilance, menace, lifelink, and ward {2}."
//
// A layer-6 grant (CR 613.1f) of the five plain keywords to every
// artifact creature its controller controls, One included, read through
// the source's text (CR 612.2) and never added twice (Krang, Utrom
// Warlord's shape without "other"). Ward {2} is the shared granted ward
// trigger (WardGranted) over the same set, so an opponent's spell or
// ability targeting any of them is countered unless its controller pays
// {2} (CR 702.21a).
//
// No simplification.
func init() {
	artifactCreatureYouControl := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsArtifact() && target.IsCreature()
	}
	Register(Spec{
		OracleID:     "eecdf18b-7a85-4b29-9a8a-26ee40da9e7a",
		Name:         "One, Borg Evolved",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Artifact creatures you control have flying, double strike, vigilance, menace, lifelink, and ward {2}.",
			Layer:     game.Layer6Ability,
			AppliesTo: artifactCreatureYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				for _, kw := range []string{"flying", "double strike", "vigilance", "menace", "lifelink"} {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, source.TextKeyword(kw))
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			WardGranted(WardMana("{2}"), "One, Borg Evolved — artifact creatures you control have ward {2}", artifactCreatureYouControl),
		},
	})
}
