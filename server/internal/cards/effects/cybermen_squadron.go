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
// the resolving trigger creates no tokens.
//
// The shared myriad body asks separately for each opponent, settles token
// count replacements, then lets each copy attack that player or a walker.
func init() {
	Register(Spec{
		OracleID:     "8d035688-089f-4d5c-bcad-9140a6c1681b",
		Name:         "Cybermen Squadron",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{{Key: "cybermen-squadron/myriad", Triggered: []game.TriggeredAbility{GrantedMyriad("Cybermen Squadron")}, Text: "Myriad"}},
		Static: []game.StaticAbility{GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller && target.IsArtifact() && target.IsCreature() && !target.IsLegendary()
		}, "cybermen-squadron/myriad")},
	})
}
