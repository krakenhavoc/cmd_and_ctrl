package legal

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2853: the effect-text half of abilityInteracts, against printed rows.
// A false "interacts" is a pointless stop and a false "no" a missed
// window, so both lists matter.
func TestEffectTextInteracts(t *testing.T) {
	answers := []string{
		"{2}{G}: Regenerate this creature.",
		"{B}: This creature gets +1/+1 until end of turn.",
		"{R}: This creature gets +1/+0 until end of turn.",
		"Sacrifice this creature: Creatures you control gain indestructible until end of turn.",
		"Sacrifice this creature: Prevent all combat damage that would be dealt this turn.",
		"{W}: This creature gains protection from the color of your choice until end of turn.",
		"{G}: This creature gains hexproof until end of turn.",
		"{1}{U}: This creature phases out.",
		"{2}: Put a +1/+1 counter on this creature.",
		"{1}{W}: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
		"{2}{B}, Sacrifice this artifact: It deals 1 damage to each creature.",
		"{X}: This creature gets +X/+0 until end of turn.",
		"Pay {E}{E}: Exile this creature, then return it to the battlefield under its owner's control.",
		"Pay {E}{E}{E}{E}: Return this creature to its owner's hand.",
		"{2}{U}{U}: Return Arcanis to its owner's hand.",
		"{T}: The next time a source of your choice would deal damage this turn, that damage is dealt to Opal-Eye instead.",
		"{5}{R}{R}: Monstrosity 3.",
		"{2}{G}: Adapt 2",
		"{X}: Until end of turn, creatures you control have base power and toughness X/X and gain all creature types.",
		"{1}{B}, Sacrifice this creature: Each nontoken Vampire creature you control gains persist until end of turn.",
		"{B}: This creature gains deathtouch until end of turn.",
		"Sacrifice this creature: Your opponents can't cast noncreature spells this turn.",
		"{B}: Pestilence deals 1 damage to each creature and each player.",
		"{1}, {T}: Destroy all artifacts, creatures, and enchantments",
	}
	value := []string{
		"{1}, {T}, Sacrifice this artifact: Draw a card.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"{2}, Sacrifice this artifact: Draw a card.",
		"{2}, Sacrifice this artifact: You gain 3 life.",
		"Cycling {2} ({2}, Discard this card: Draw a card.)",
		"{T}: Scry 1.",
		"{3}, {T}: Create a 1/1 white Soldier creature token.",
		"{1}, {T}: Look at the top card of your library.",
		"{4}, {T}: Draw a card, then discard a card.",
		"{T}, Remove a +1/+1 counter from this creature: Draw a card.",
		"{T}: Put a charge counter on this artifact.",
		"{1}, {T}: Put a storage counter on this land.",
		"{T}: This creature deals 1 damage to each opponent.",
		"{2}{R}, {T}: Shivan Gorge deals 1 damage to each opponent.",
		"{3}{R}: Destroy this enchantment.",
		"{4}, {T}, Sacrifice this artifact: Return all cards exiled with this artifact to their owner's hand.",
		"{T}, Sacrifice two artifacts: Exile the top card of your library. You may play that card this turn.",
	}
	for _, l := range answers {
		if !effectTextInteracts(l) {
			t.Errorf("%q should interact", l)
		}
	}
	for _, l := range value {
		if effectTextInteracts(l) {
			t.Errorf("%q is pure value and should not interact", l)
		}
	}
}

// A sacrifice cost is an outlet when it can eat a creature, and a price
// when it names a land, a token kind or a plain artifact.
func TestSacrificeInteracts(t *testing.T) {
	outlets := []string{
		"a creature", "another creature", "an artifact or creature", "a permanent",
		"another black creature", "a Goblin", "two Eldrazi", "an artifact creature",
	}
	prices := []string{
		"a land", "a Food", "a Treasure", "three Clues", "two artifacts",
		"a noncreature artifact", "a Swamp and a Forest", "a Desert", "a token",
	}
	for _, l := range outlets {
		if !sacrificeInteracts(&game.TargetSpec{Label: l}) {
			t.Errorf("sacrifice %q should be an outlet", l)
		}
	}
	for _, l := range prices {
		if sacrificeInteracts(&game.TargetSpec{Label: l}) {
			t.Errorf("sacrifice %q is a price, not an outlet", l)
		}
	}
	if sacrificeInteracts(nil) {
		t.Error("no sacrifice cost is not an outlet")
	}
}
