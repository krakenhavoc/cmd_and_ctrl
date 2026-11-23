package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Tolvada — Battle — Siege, defense 5, for {3}{W}{B} (March of
// the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, return target nonbattle permanent card from
//	 your graveyard to the battlefield."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the reanimation and the defeated
// trigger that hands over The Broken Sky. The target is chosen as the
// trigger goes on the stack; a Siege that is itself in the graveyard by
// then is never a legal target for it, being a battle card.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Invasion of Tolvada — return target nonbattle permanent card from your graveyard to the battlefield",
		ReturnFirstLegalGraveyardTargetToBattlefield)
	enters.Targets = TargetCardInGraveyard("target nonbattle permanent card from your graveyard", Permanent(), Not(Battle()), YouOwn())
	Register(Spec{
		OracleID:     InvasionOfTolvadaOracleID,
		Name:         "Invasion of Tolvada",
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 5,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			enters,
			DefeatedTrigger("Invasion of Tolvada — defeated: exile it, then cast The Broken Sky", SiegeDefeated()),
		},
	})
}

// invasionOfTolvadaOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const InvasionOfTolvadaOracleID = "7f8dad7f-a573-469f-9095-4ba3886b4f0b"
