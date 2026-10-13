package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magda, the Hoardmaster — Legendary Creature — Dwarf Berserker {1}{R},
// 2/2 (EDHREC rank 2000):
//
//	"Whenever you commit a crime, create a tapped Treasure token.
//	 This ability triggers only once each turn. (Targeting opponents,
//	 anything they control, and/or cards in their graveyards is a
//	 crime.)
//	 Sacrifice three Treasures: Create a 4/4 red Scorpion Dragon
//	 creature token with flying and haste. Activate only as a
//	 sorcery."
//
// Crimes use the once-per-announcement CrimeCommitted annotation (CR 700.13).
// The trigger goes above the announcing spell or ability even if it is
// later countered. Copies and target changes commit no new crime.
// "Only once each turn" is the Exemplar of Light tally: the per-turn
// tally is asked for an earlier announce of this same ability
// (b11TriggeredThisTurn). The Treasure enters tapped
// through the shared tapped template.
//
// "Sacrifice three Treasures: Create a 4/4 red Scorpion Dragon" is a
// sacrifice cost with a count of three (#747, SacrificeN) on a
// sorcery-speed ability. Until #747 it was omitted, because a cost
// could sacrifice only one permanent.
//
// No simplification.
const b18MagdaTreasureLabel = "Magda, the Hoardmaster — create a tapped Treasure"

func init() {
	Register(Spec{
		OracleID:     "7fd2beb8-f823-4723-beec-e59b62127490",
		Name:         "Magda, the Hoardmaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return YouCommittedCrime(ev, source) &&
					!b11TriggeredThisTurn(g, source.InstanceID, b18MagdaTreasureLabel)
			}, b18MagdaTreasureLabel, Do(CreateToken{
				Template: tappedTreasureToken(),
				N:        1,
			})),
		},
		Activated: []ActivatedAbility{{
			Label:        "Sacrifice three Treasures: Create a 4/4 red Scorpion Dragon creature token with flying and haste. Activate only as a sorcery.",
			Cost:         SacrificeN(3, "three Treasures", isTreasure),
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: TokenCard("4/4 red Scorpion Dragon with flying and haste"), N: 1}),
		}},
	})
}
