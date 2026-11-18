package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glamorous Outlaw — Creature — Vampire Rogue {3}{U}{B}{R}, 4/5 (Streets
// of New Capenna):
//
//	"When this creature enters, it deals 2 damage to each opponent and you
//	 scry 2.
//	 {2}, Exile this card from your hand: Target land gains "{T}: Add {U},
//	 {B}, or {R}" until this card is cast from exile. You may cast this
//	 card for as long as it remains exiled."
//
// The damage is dealt by the creature itself, then the scry is the
// controller's. The hand ability is the cycle's shared one
// (sncB19ExileForLand).
//
// No simplification.
func init() {
	grant, ability := sncB19ExileForLand("Glamorous Outlaw", "glamorous-outlaw/tap-ubr", "{U|B|R}", "{U}, {B}, or {R}")
	Register(Spec{
		OracleID:     "cd4f7912-81c0-44dd-87c2-a98e955fe5a8",
		Name:         "Glamorous Outlaw",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 2, DamageEachOpponent: 2},
		Grants:       []AbilityGrant{grant},
		Activated:    []ActivatedAbility{ability},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glamorous Outlaw — 2 damage to each opponent and scry 2",
				Do(DamageEachOpponent{N: 2}, Scry{N: 2})),
		},
	})
}
