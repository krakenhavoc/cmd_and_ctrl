package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// K'Vort-Class Raider — Artifact — Spacecraft {4}, 5/3 (Star Trek
// Commander, #4974):
//
//	"During your turn, this Spacecraft and creatures you control have
//	 hexproof.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 4+.)
//	 4+ | Flying, menace, haste"
//
// Station is the ordinary keyword (station.go). The first line is one
// layer-6 grant with two receivers: the Spacecraft itself, creature or
// not, and every creature its controller controls, read on every
// recompute while it is its controller's turn (Gruul Spellbreaker's
// "during your turn"). Hexproof on a permanent is CR 702.11b.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a52d7ab2-a750-4c3a-9f0c-999121aa81af",
		Name:         "K'Vort-Class Raider",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Station()},
		Static: append(append(
			SpacecraftAt(4, 5, 3),
			ThresholdKeywords(4, "flying", "menace", "haste")),
			Labelled(KeywordGrant(kVortClassRaiderShielded, KeywordHexproof),
				"During your turn, this Spacecraft and creatures you control have hexproof.")),
	})
}

// kVortClassRaiderShielded is "during your turn, this Spacecraft and
// creatures you control".
func kVortClassRaiderShielded(target *game.Card, g *game.Game, source *game.Card) bool {
	if !isActivePlayer(g, source.Controller) {
		return false
	}
	return target.InstanceID == source.InstanceID ||
		(target.IsCreature() && target.Controller == source.Controller)
}
