package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ozai, the Phoenix King — Legendary Creature — Human Noble {2}{B}{B}{R}{R},
// 7/7 (Avatar: The Last Airbender, #4459):
//
//	"Trample, firebending 4, haste
//	 If you would lose unspent mana, that mana becomes red instead.
//	 Ozai has flying and indestructible as long as you have six or more
//	 unspent mana."
//
// Trample and haste are printed data; firebending 4 is Avatar Roku's
// (`firebending`), an attack trigger on the stack adding red that lasts
// until end of combat.
//
// The mana clause is Kruphix, God of Horizons' Spec.ManaPool conversion
// with red in place of colourless (game.ManaPoolBecomesRed): at every step
// and phase boundary the mana you would lose becomes red and stays, red
// mana included, so while Ozai is on the battlefield you keep it from step
// to step and turn to turn (the ruling). Its restrictions and riders stay
// with it (the White Lotus Hideout ruling). The static is derived from the
// battlefield, so once Ozai leaves the next boundary empties the pool as
// usual.
//
// The last line is a layer 6 grant that reads the controller's pool, so it
// declares DependsOnManaPool (Omnath, Locus of Mana's hint): tapping or
// spending mana changes it at once. "Unspent mana" is every mana in the
// pool, of any type.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7472af8e-2e93-43b2-bc32-a3741f6d7502",
		Name:            "Ozai, the Phoenix King",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		ManaPool:        []game.ManaPoolStatic{{Kind: game.ManaPoolBecomesRed}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Ozai, the Phoenix King — firebending 4", Do(firebending(4))),
		},
		Static: ozaiSixManaKeywords(),
	})
}

// ozaiSixManaKeywords is "Ozai has flying and indestructible as long as you
// have six or more unspent mana": one layer 6 grant per keyword, both on
// the same condition.
func ozaiSixManaKeywords() []game.StaticAbility {
	var out []game.StaticAbility
	for _, kw := range []string{"flying", "indestructible"} {
		ab := KeywordGrant(ozaiHasSixUnspentMana, kw)
		ab.Label = "Has " + kw + " as long as you have six or more unspent mana."
		ab.DependsOnManaPool = true
		out = append(out, ab)
	}
	return out
}

// ozaiHasSixUnspentMana is "this permanent, while its controller has six or
// more unspent mana".
func ozaiHasSixUnspentMana(target *game.Card, g *game.Game, source *game.Card) bool {
	if !selfOnly(target, g, source) {
		return false
	}
	p := g.PlayerByIDForEffect(source.Controller)
	return p != nil && p.ManaPool.Total() >= 6
}
