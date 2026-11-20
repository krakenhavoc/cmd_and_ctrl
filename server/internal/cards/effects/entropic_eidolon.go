package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Entropic Eidolon — Creature — Spirit {3}{B}, 2/2 (Dissension, #5598):
//
//	"{B}, Sacrifice this creature: Target player loses 1 life and you gain
//	 1 life.
//	 Whenever you cast a multicolored spell, you may return this card from
//	 your graveyard to your hand."
//
// The drain is Death Cultist's drainTargetOpponentOne.
// The return is disEidolonReturn, a graveyard trigger (InGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be5f788a-8a2e-447c-9e65-f9fc197ddfc9",
		Name:         "Entropic Eidolon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice this creature: Target player loses 1 life and you gain 1 life.",
			Cost:    Plus(ManaCost("{B}"), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect:  drainTargetOpponentOne,
		}},
		Triggered: []game.TriggeredAbility{disEidolonReturn("Entropic Eidolon")},
	})
}
