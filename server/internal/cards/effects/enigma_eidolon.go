package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enigma Eidolon — Creature — Spirit {3}{U}, 2/2 (Dissension, #5598):
//
//	"{U}, Sacrifice this creature: Target player mills three cards.
//	 Whenever you cast a multicolored spell, you may return this card from
//	 your graveyard to your hand."
//
// The mill is TargetPlayerMills.
// The return is disEidolonReturn, a graveyard trigger (InGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4489a526-0e07-4d37-9e83-628f7eb256c6",
		Name:         "Enigma Eidolon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Target player mills three cards.",
			Cost:    Plus(ManaCost("{U}"), SacrificeThis()),
			Targets: TargetPlayer("target player"),
			Effect:  TargetPlayerMills(3),
		}},
		Triggered: []game.TriggeredAbility{disEidolonReturn("Enigma Eidolon")},
	})
}
