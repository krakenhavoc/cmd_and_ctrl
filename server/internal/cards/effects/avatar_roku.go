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
//
// The Dragon token DOES carry its own firebending 4 since ADR 0083
// (#1248): it is a catalog template under `token:dragon-firebending`,
// and its attack trigger is found through `game.CatalogKey`'s
// token-key fallback like a printed card's. Before that this card
// declared a second caveat for it — the "Triggered and static
// abilities on non-copy tokens" seam. The token inherits the mana
// caveat above, because it is the same firebending.
func init() {
	Register(Spec{
		OracleID:     theLegendOfRokuOracleID + "#1",
		Name:         "Avatar Roku",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Firebending's mana empties at the end of the declare attackers step instead of lasting until end of combat, both on Avatar Roku and on the Dragon token it creates.",
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Avatar Roku — firebending 4", Do(AddMana{Produced: "{R}{R}{R}{R}"})),
		},
		Activated: []ActivatedAbility{{
			Label:  "{8}: Create a 4/4 red Dragon creature token with flying and firebending 4.",
			Cost:   ManaCost("{8}"),
			Effect: Do(CreateToken{Template: firebendingDragonToken(), N: 1}),
		}},
	})
}

// firebendingDragonToken is Avatar Roku's 4/4 red Dragon with flying
// and firebending 4.
func firebendingDragonToken() game.Card { return tokenFromCatalog(printedFirebendingDragonToken) }

// printedFirebendingDragonToken is that Dragon as PRINTED. Flying is
// a printed keyword and rides the Card as plain data; firebending is
// a triggered ability and rides the catalog slot, which is the whole
// distinction ADR 0083 draws.
func printedFirebendingDragonToken() tokenTemplate {
	return tokenTemplate{
		Slug: "dragon-firebending",
		Card: game.Card{
			Name:      "Dragon",
			TypeLine:  "Token Creature — Dragon",
			Power:     4,
			Toughness: 4,
			Colors:    []string{"R"},
			Keywords:  []string{"flying"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Dragon — firebending 4", Do(AddMana{Produced: "{R}{R}{R}{R}"})),
		},
		Text: "Flying\nFirebending 4 (Whenever this creature attacks, add {R}{R}{R}{R}. " +
			"This mana lasts until end of combat.)",
	}
}
