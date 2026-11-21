package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borg Scout Sphere — Artifact — Spacecraft {2}{U}, 5/5 (Star Trek
// Commander, #4974):
//
//	"When this Spacecraft enters, draw a card.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 6+.)
//	 6+ | Artifact creatures you control have flying."
//
// Suurok-Class Science Vessel's station shape (Station, SpacecraftAt;
// CR 702.184, CR 721) with Command Subadjunct's keyword grant behind the
// CR 721.2a gate: the grant exists while the Sphere has six or more charge
// counters (ActiveWhen is read on the source), and reaches every artifact
// creature you control on post-layer types, the stationed Sphere itself
// among them.
//
// No simplification.
func init() {
	fly := effects.KeywordGrant(effects.TrekB34ArtifactCreaturesYouControl, "flying")
	fly.Label = "6+ | Artifact creatures you control have flying."
	fly.ActiveWhen = effects.AtChargeCounters(6)
	effects.Register(effects.Spec{
		OracleID:     "4cfb6f06-137c-4b41-97ef-6710e899473b",
		Name:         "Borg Scout Sphere",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Borg Scout Sphere — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static:    append(effects.SpacecraftAt(6, 5, 5), fly),
	})
}
