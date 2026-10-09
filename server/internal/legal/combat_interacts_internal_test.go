package legal

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2871: the effect-text half of abilityCombatKind, against printed
// rows. A false "yes" is a pointless stop in combat and a false "no" a
// missed block, so both lists matter.
func TestCombatTextInteracts(t *testing.T) {
	combat := []string{
		"{2}: This creature gains flying until end of turn",
		"{B}: This creature gains menace until end of turn.",
		"{T}: Creatures you control gain haste until end of turn.",
		"{4}{W}{W}, {T}: Creatures you control gain flying and lifelink until end of turn.",
		"Pay {E}: This creature gains your choice of flying, vigilance, or lifelink until end of turn.",
		"{1}{G}: This creature can block an additional creature this turn.",
		"{1}: This creature can't be blocked this turn except by creatures with haste.",
		"{1}{W}{U}: Until end of turn, this land becomes a 2/3 white and blue Bird creature with flying. It's still a land.",
		"Pay {E}{E}{E}{E}: This Vehicle becomes an artifact creature until end of turn.",
		"{4}{R}: This artifact becomes a 4/4 red Giant artifact creature with trample until end of turn.",
		"{2}{W}, {T}: Whenever you attack this turn, create two 1/1 red Warrior tokens tapped and attacking; sacrifice them at the next end step",
		"{4}: Create a 2/2 colorless Wizard Soldier creature token named Cadet. Then creatures you control gain haste until end of turn.",
		"Channel — {3}{R}, Discard this card: Create two 1/1 colorless Spirit creature tokens. They gain haste until end of turn",
		"{3}{R}{R}{G}{G}: Anzrag must be blocked each combat this turn if able.",
		"{1}{U}, {T}: Creatures your opponents control attack this turn if able.",
		"{U}: Untap enchanted creature",
		"{2}: This creature loses flying until end of turn. Any player may activate this ability.",
	}
	value := []string{
		"{T}: Draw a card.",
		"{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		"{T}: Create a Treasure token.",
		"{1}{G}, {T}: Create a Food token.",
		"Pay {E}{E}{E}: Create a tapped 3/3 colorless Robot artifact creature token.",
		"{1}{B}, Discard a creature card: Draw a card. If the discarded card was a Zombie card, create a tapped 2/2 black Zombie creature token.",
		"Sacrifice this artifact: Choose a basic land type. Each land you control becomes that type until end of turn.",
		"{W}{U}{B}{R}{G}, {T}: Each creature you control becomes prepared.",
		"{2}, {T}: You gain 1 life for each colorless creature you control.",
		"{T}: Put a charge counter on this artifact.",
		"Cycling {2} ({2}, Discard this card: Draw a card.)",
	}
	// Owner answer (#2871): a token is a blocker, so these count only
	// for a player who is being attacked.
	defender := []string{
		"{1}{R}, {T}: Create a 0/1 red Kobold creature token named Kobolds of Kher Keep.",
		"{2}{W}{W}, {T}: Create a 1/1 white Human creature token.",
		"{T}: Create X 1/1 Goblins, where X is the number of Goblins you control",
		"{3}, {T}: Create a 2/2 blue Shapeshifter with changeling",
		"{8}: Create a 4/4 red Dragon creature token with flying and firebending 4.",
		"{1}{G}{W}, {T}: Populate.",
		"{X}{X}{B}, {T}: Amass Orcs X. Activate only if a creature died this turn.",
	}
	for _, l := range combat {
		if got := combatTextKind(l); got != combatAny {
			t.Errorf("combat row read as %v: %q", got, l)
		}
	}
	for _, l := range defender {
		if got := combatTextKind(l); got != combatDefender {
			t.Errorf("token row read as %v: %q", got, l)
		}
	}
	for _, l := range value {
		if got := combatTextKind(l); got != combatNone {
			t.Errorf("value row read as %v: %q", got, l)
		}
	}
}

// A crew cost is a combat ability whatever its label says, and a crew
// row whose cost is printed another way is read from its label.
func TestCrewIsACombatAbility(t *testing.T) {
	if abilityCombatKind(game.ActivatedAbilityShape{Label: "Crew 3", Cost: game.AbilityCost{Crew: 3}}) != combatAny {
		t.Error("Crew 3 is not a combat ability")
	}
	if abilityCombatKind(game.ActivatedAbilityShape{Label: "Crew — remove a loyalty counter from a planeswalker you control"}) != combatAny {
		t.Error("Heart of Kiran's loyalty crew is not a combat ability")
	}
}
