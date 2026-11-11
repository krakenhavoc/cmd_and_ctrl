package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goldlust Triad — Creature — Dragon {4}{R}, 4/3 (Tarkir: Dragonstorm
// Commander):
//
//	"Flying
//	 Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and attacking
//	 that player or a planeswalker they control. Exile the tokens at end of
//	 combat.)
//	 Whenever this creature deals combat damage to a player, create a Treasure
//	 token."
//
// Myriad is the shared trigger (myriad.go, CR 702.116): the creature's own,
// one "you may" per other opponent, the copies exiled at end of combat. Each
// copy is a Goldlust Triad too, so its combat damage makes a Treasure of its
// own for its controller, and it does not myriad again (CR 508.4: it was never
// declared as an attacker). The second ability is the ordinary combat-damage
// trigger over the shared Treasure token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86d76078-9af3-45a7-90d6-72946afdc78b",
		Name:            "Goldlust Triad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			GrantedMyriad("Goldlust Triad"),
			WheneverThisDealsCombatDamageToAPlayer("Goldlust Triad — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
