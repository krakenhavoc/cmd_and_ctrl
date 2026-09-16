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
// The Outlaws of Thunder Junction crime commander. "Commit a crime"
// is EventBecomesTarget (S22), which fires once per target slot at
// announce with the targeting player in Actor: an opponent, a
// permanent or spell an opponent controls, or a card in an
// opponent's graveyard qualifies (b18CommittedCrime, CR 700.13).
// The trigger goes on the stack above the spell that targeted, as
// printed — Magda pays out even if the spell is then countered.
// "Only once each turn" is the Exemplar of Light tally: the log is
// walked back to the turn's upkeep for an earlier announce of this
// same ability (b11TriggeredThisTurn). The Treasure enters tapped
// through the shared tapped template.
//
// DECLARED SIMPLIFICATION: the Scorpion Dragon ability is not
// implemented. "Sacrifice three Treasures" is a cost the engine
// cannot express — AbilityCost.SacrificeOther sacrifices exactly ONE
// permanent, and three activations of a one-Treasure cost is not the
// printed card. Shipping it for one Treasure would be three times
// stronger than printed, the #259 direction; Magda makes Treasures
// and stops there until a multi-permanent sacrifice cost lands.
const b18MagdaTreasureLabel = "Magda, the Hoardmaster — create a tapped Treasure"

func init() {
	Register(Spec{
		OracleID:     "7fd2beb8-f823-4723-beec-e59b62127490",
		Name:         "Magda, the Hoardmaster",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Sacrificing three Treasures for a 4/4 Scorpion Dragon isn't implemented — only the Treasure-per-crime trigger works."},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18CommittedCrime(ev, source, g) &&
					!b11TriggeredThisTurn(g, source.InstanceID, b18MagdaTreasureLabel)
			}, b18MagdaTreasureLabel, Do(CreateToken{
				Template: tappedTreasureToken(),
				N:        1,
			})),
		},
	})
}
