package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Karsus — Battle — Siege, defense 4, for {2}{R}{R}:
//
//	"When this Siege enters, it deals 3 damage to each creature and
//	 each planeswalker.
//	 (As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)"
//
// A four-mana sweeper that leaves a 4/4 behind if anyone can be
// bothered to attack it. The bracketed reminder text is not
// card-effect data — the defense counters, the protector prompt and
// the CR 704.5v/w sweep are engine behaviour keyed on the card TYPE
// (server/internal/game/battle.go) and happen to a battle the catalog
// has never heard of. What this Spec carries is the sweep, and the
// defeated trigger that hands over Refraction Elemental.
//
// "each creature and each planeswalker", not "each other creature":
// the damage goes to every one on the battlefield including the
// caster's, which is most of why the card is symmetrical enough to
// be interesting. The Siege itself is neither a creature nor a
// planeswalker, so it does not damage itself off the board — battles
// lose defense counters to damage (CR 120.3h) and a self-sweep would
// have taken four off its own four.
//
// Damage rather than destruction is observable and is the reason the
// card sees play: an indestructible 2/2 dies to this, a regenerating
// one does not come back, and a 4/4 shrugs.
func init() {
	Register(Spec{
		OracleID:     invasionOfKarsusOracleID,
		Name:         "Invasion of Karsus",
		Completeness: CompletenessCaveats,
		Caveats:      []string{SiegeTransformedCastCaveat},
		Battle: &BattleSpec{
			Defense: 4,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Invasion of Karsus — 3 damage to each creature and planeswalker", func(g *game.Game, item *game.StackItem) error {
				return damageEachMatching(NewContext(g, item), Or(Creature(), Planeswalker()), 3)
			}),
			DefeatedTrigger("Invasion of Karsus — defeated: exile it, then cast Refraction Elemental", SiegeDefeated()),
		},
	})
}

// invasionOfKarsusOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const invasionOfKarsusOracleID = "2d125ef5-074c-44eb-a1d0-c308d48bfbc2"
