package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// poolAHand is one hand with something for every pool A filter, in
// this order:
//
//	0 Forest              basic land
//	1 Snow-Covered Swamp  basic land (snow is a supertype, not "nonbasic")
//	2 Sacred Foundry      nonbasic land with basic land types (CR 205.4c)
//	3 Lightning Bolt      instant, mana value 1
//	4 Grizzly Bears       creature, mana value 2
//	5 Jace                legendary planeswalker, mana value 4
//	6 Sol Ring            artifact, mana value 1
//	7 Venser              legendary creature, mana value 5
//	8 Spirit              creature — Spirit, mana value 2
//	9 Arcane instant      instant — Arcane, mana value 2
//	10 Changeling Outcast creature with changeling, every creature type (CR 702.73a)
//	11 Fireball           sorcery {X}{R}, mana value 1 in a hand (CR 202.3e)
//	12 Rite               sorcery, mana value 4
//	13 Pelakka Predation  MDFC, a nonland sorcery by its front face (CR 712.8a)
func PoolAHand() []game.Card {
	return []game.Card{
		RhForest(),
		{Name: "Snow-Covered Swamp", TypeLine: "Basic Snow Land — Swamp"},
		{Name: "Sacred Foundry", TypeLine: "Land — Mountain Plains"},
		RhBolt(),
		{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}},
		{Name: "Jace Beleren", TypeLine: "Legendary Planeswalker — Jace", ManaCost: "{1}{U}{U}{U}", Colors: []string{"U"}},
		{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
		{Name: "Venser, Shaper Savant", TypeLine: "Legendary Creature — Human Wizard", ManaCost: "{2}{U}{U}{U}", Colors: []string{"U"}},
		{Name: "Kami of Ancient Law", TypeLine: "Creature — Spirit", ManaCost: "{1}{W}", Colors: []string{"W"}},
		{Name: "Peer Through Depths", TypeLine: "Instant — Arcane", ManaCost: "{1}{U}", Colors: []string{"U"}},
		{Name: "Changeling Outcast", TypeLine: "Creature — Shapeshifter Rogue", ManaCost: "{B}", Colors: []string{"B"}, Keywords: []string{game.KeywordChangeling}},
		RhFireball(),
		{Name: "Rite of Replication", TypeLine: "Sorcery", ManaCost: "{2}{U}{U}", Colors: []string{"U"}},
		RhPelakka(),
	}
}
