package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cybermen Squadron — Artifact Creature — Cyberman {7}, 5/5:
//
//	"Nonlegendary artifact creatures you control have myriad.
//	 (Whenever a creature with myriad attacks, for each opponent other
//	 than defending player, you may create a token copy that's tapped
//	 and attacking that player or a planeswalker they control. Exile
//	 the tokens at end of combat.)"
//
// Legion Loyalty's grant (GrantedMyriad, myriad.go) narrowed to
// nonlegendary artifact creatures. The Squadron is one itself, so it
// has myriad too (the ruling). Whether an attacker has myriad is read
// as the attack is declared, after every layer: a creature that is an
// artifact only through an effect counts, and a legendary one never
// does. At a two-player table, or once every other opponent has left,
// the trigger is not queued (the ruling: no defending player but your
// only opponent means no tokens).
//
// Granted myriad's two shared simplifications (myriad.go) are this
// card's caveats. Both are weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "8d035688-089f-4d5c-bcad-9140a6c1681b",
		Name:         "Cybermen Squadron",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Myriad asks once per attacker whether to make the copies for every other opponent, rather than letting you choose per opponent.",
			"The copies always attack the player, not a planeswalker they control.",
		},
		Triggered: []game.TriggeredAbility{
			GrantedMyriad("Cybermen Squadron", func(attacker game.Card) bool {
				return attacker.IsArtifact() && attacker.IsCreature() && !attacker.IsLegendary()
			}),
		},
	})
}
