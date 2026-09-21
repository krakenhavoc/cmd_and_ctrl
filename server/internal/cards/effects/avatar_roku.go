package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avatar Roku — Legendary Creature — Avatar, 4/4 (colorless):
//
//	"Firebending 4 (Whenever this creature attacks, add {R}{R}{R}{R}.
//	 This mana lasts until end of combat.)
//	 {8}: Create a 4/4 red Dragon creature token with flying and
//	 firebending 4."
//
// The BACK face of The Legend of Roku ("<oracle_id>#1", ADR 0034),
// reached only through chapter III's exile-and-return
// (the_legend_of_roku.go).
//
// SANDBOX SIMPLIFICATIONS, and this is why the card is not Full:
//
//   - Firebending's mana does not persist through combat. AddMana
//     puts the four red mana in the pool exactly as Dark Ritual does,
//     and it empties with the pool at the end of the step it was
//     added in (CR 106.4) — the declare-attackers step here. "This
//     mana lasts until end of combat" needs a mana-pool exception this
//     engine has nowhere to hang, so the mana is real and spendable
//     immediately but gone before blockers are declared, rather than
//     lasting the rest of combat as printed. Weaker than printed
//     (#259): nothing spends mana it never received.
//   - The Dragon token this ability creates does not carry its own
//     firebending 4 — a token that isn't a copy has no oracle ID, so
//     no catalog key exists to hang a trigger off (#521, the same gap
//     this card's own front face's Spirit-token sibling on Avatar
//     Kuruk ships against, and Fable of the Mirror-Breaker's Goblin
//     Shaman token before it). It is a plain 4/4 flying Dragon.
func init() {
	Register(Spec{
		OracleID:     theLegendOfRokuOracleID + "#1",
		Name:         "Avatar Roku",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Firebending's mana empties at the end of the declare attackers step instead of lasting until end of combat.",
			"The Dragon token it creates doesn't have its own firebending 4 ability.",
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Avatar Roku — firebending 4", Do(AddMana{Produced: "{R}{R}{R}{R}"})),
		},
		Activated: []ActivatedAbility{{
			Label:  "{8}: Create a 4/4 red Dragon creature token with flying and firebending 4.",
			Cost:   ManaCost("{8}"),
			Effect: Do(CreateToken{Template: TokenCard("4/4 red Dragon with flying"), N: 1}),
		}},
	})
}
