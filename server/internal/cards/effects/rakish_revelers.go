package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rakish Revelers — Creature — Elf Druid Rogue {2}{R}{G}{W}, 5/3 (Streets
// of New Capenna):
//
//	"When this creature enters, create a 1/1 green and white Citizen
//	 creature token.
//	 {2}, Exile this card from your hand: Target land gains "{T}: Add {R},
//	 {G}, or {W}" until this card is cast from exile. You may cast this
//	 card for as long as it remains exiled."
//
// The hand ability is the cycle's shared one (sncB19ExileForLand).
//
// No simplification.
func init() {
	grant, ability := sncB19ExileForLand("Rakish Revelers", "rakish-revelers/tap-rgw", "{R|G|W}", "{R}, {G}, or {W}")
	Register(Spec{
		OracleID:     "7db1402e-a434-45d2-9d21-9e0faeeaff10",
		Name:         "Rakish Revelers",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{grant},
		Activated:    []ActivatedAbility{ability},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rakish Revelers — create a 1/1 green and white Citizen",
				Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1})),
		},
	})
}
