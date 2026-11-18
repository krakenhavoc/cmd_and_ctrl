package deck

import (
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// companion_zirda_test.go — Zirda, the Dawnwaker's companion condition
// (companion_zirda.go, ADR 0172 §9, #5909).

// zirda is Zirda, the Dawnwaker as the Scryfall record has it.
func zirda() cards.Card {
	c := basicLegal("Zirda, the Dawnwaker", "Legendary Creature — Elemental Fox", "R", "W")
	c.OracleID = uuid.MustParse(zirdaOracle)
	c.ManaCost = "{1}{R/W}{R/W}"
	c.Keywords = []string{"Companion"}
	c.OracleText = "Companion — Each permanent card in your starting deck has an activated ability. (If this card is your chosen companion, you may put it into your hand from outside the game for {3} as a sorcery.)\nAbilities you activate that aren't mana abilities cost {2} less to activate. This effect can't reduce the mana in that cost to less than one mana.\n{1}, {T}: Target creature can't block this turn."
	return c
}

// texted is a legal card with a type line and oracle text.
func texted(name, typeLine, text string) cards.Card {
	c := costed(name, typeLine, "{1}{R}", "R")
	c.OracleText = text
	return c
}

// faced is a multi-faced card of the given layout.
func faced(layout string, faces ...cards.CardFace) cards.Card {
	name, typeLine := faces[0].Name, faces[0].TypeLine
	for _, f := range faces[1:] {
		name += " // " + f.Name
		typeLine += " // " + f.TypeLine
	}
	c := costed(name, typeLine, "", "R")
	c.Layout = layout
	c.CardFaces = faces
	return c
}

func TestZirdaConditionReadsEachPermanentCardsText(t *testing.T) {
	cases := []struct {
		name  string
		card  cards.Card
		fails bool
	}{
		// A "[Cost]: [Effect]" line (CR 602.1), whatever the cost is.
		{"mana and tap cost", texted("Sower", "Creature — Devil", "{2}{R}: Target creature can't block this turn."), false},
		{"non-mana cost", texted("Ghoul", "Creature — Zombie", "Sacrifice a creature: This creature gets +2/+2 until end of turn."), false},
		{"loyalty ability", texted("Walker", "Legendary Planeswalker — Chandra", "+1: Chandra deals 2 damage to any player.\n−3: Chandra deals 4 damage to target creature."), false},
		{"ability word before the cost", texted("Boaster", "Creature — Dwarf", "Boast — {1}{R}: This creature deals 1 damage to any target. (Activate only if this creature attacked this turn and only once each turn.)"), false},
		{"exhaust", texted("Exhauster", "Artifact — Vehicle", "Exhaust — {3}: Put two +1/+1 counters on this Vehicle."), false},
		{"class level", texted("Classy", "Enchantment — Class", "When this Class enters, draw a card.\n{1}{G}: Level 2\nLEVEL 2\nCreatures you control get +1/+1."), false},
		// A mana ability is an activated ability (CR 605.1a).
		{"mana ability", texted("Llanowar Elves", "Creature — Elf Druid", "{T}: Add {G}."), false},
		// The activated keywords, whose colon is only in the reminder.
		{"equip", texted("Sword", "Artifact — Equipment", "Equipped creature gets +2/+2.\nEquip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)"), false},
		{"equip with a quality", texted("Blade", "Legendary Artifact — Equipment", "Equipped creature gets +1/+1.\nEquip legendary creature {3}\nEquip {7}"), false},
		{"equip with a non-mana cost", texted("Tools", "Artifact — Equipment", "Equip—Sacrifice a creature."), false},
		{"cycling", texted("Cycler", "Creature — Beast", "Trample\nCycling {2} ({2}, Discard this card: Draw a card.)"), false},
		{"cycling after another keyword", texted("Flier", "Creature — Bird", "Flying, cycling {2}"), false},
		{"typecycling", texted("Yeti", "Creature — Yeti", "Mountaincycling {2} ({2}, Discard this card: Search your library for a Mountain card, reveal it, put it into your hand, then shuffle.)"), false},
		{"basic landcycling", texted("Strider", "Creature — Elemental", "Basic landcycling {1}{G}"), false},
		{"level up", texted("Leveler", "Creature — Human", "Level up {1}{W} ({1}{W}: Put a level counter on this. Level up only as a sorcery.)\nLEVEL 2-6\n3/3"), false},
		{"ninjutsu", texted("Ninja", "Creature — Ninja", "Ninjutsu {U}{B}"), false},
		{"commander ninjutsu", texted("Commander Ninja", "Legendary Creature — Ninja", "Commander ninjutsu {2}{U}{B}"), false},
		{"crew", texted("Vehicle", "Artifact — Vehicle", "Trample\nCrew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)"), false},
		{"unearth", texted("Unearther", "Creature — Zombie", "Unearth {B}"), false},
		{"reconfigure", texted("Reconfigurer", "Artifact Creature — Equipment Lizard", "Reconfigure {2}"), false},
		{"craft", texted("Crafter", "Artifact", "Craft with artifact {4}"), false},
		{"station", texted("Station", "Artifact — Spacecraft", "Station (Tap another creature you control: Put charge counters equal to its power on this Spacecraft. Station only as a sorcery.)\nSTATION 7+\nFlying"), false},
		{"saddle", texted("Mount", "Creature — Horse Mount", "Saddle 2"), false},
		{"outlast", texted("Outlaster", "Creature — Human", "Outlast {1}{W}"), false},

		// Text that looks activated but isn't the card's own.
		{"vanilla", texted("Bear", "Creature — Bear", ""), true},
		{"colon only in reminder text", texted("Treasure Maker", "Creature — Goblin", "When this creature enters, create a Treasure token. (It's an artifact with \"{T}, Sacrifice this token: Add one mana of any color.\")"), true},
		{"granted ability in quotation marks", texted("Cradle", "Enchantment", "Creatures you control have \"{T}: Add {G}.\""), true},
		{"granted to the equipped creature", texted("Kunai", "Artifact", "Equipped creature has \"{1}, {T}: Draw a card.\""), true},
		{"keyword it grants, not has", texted("Astor", "Legendary Creature — Human", "Equipment you control have equip {1}."), true},
		{"sentence about a keyword", texted("Fighter", "Enchantment", "Equip abilities you activate cost {1} less to activate."), true},
		{"equipped, not equip", texted("Pike", "Artifact", "Equipped creature has first strike."), true},
		{"triggered ability", texted("Trigger", "Creature — Human", "Whenever this creature attacks, draw a card."), true},
		{"static ability", texted("Anthem", "Enchantment", "Creatures you control get +1/+1."), true},
		{"morph is a special action", texted("Morpher", "Creature — Beast", "Morph {2}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)"), true},
		{"companion is a special action", jegantha(), false},
		{"companion alone", texted("Companion Only", "Creature — Cat", "Companion — Each creature card in your starting deck is a Cat. (If this card is your chosen companion, you may put it into your hand from outside the game for {3} as a sorcery.)"), true},
		{"name with a colon", texted("Vault 9: Party", "Enchantment — Saga", "Vault 9: Party enters tapped.\nI — Draw a card."), true},

		// Lands.
		{"basic land, reminder text only", texted("Forest", "Basic Land — Forest", "({T}: Add {G}.)"), false},
		{"dual with basic land types", texted("Tropical Island", "Land — Forest Island", ""), false},
		{"land with a tap ability", texted("Wastes", "Basic Land", "{T}: Add {C}."), false},
		{"land creature with a basic type", texted("Dryad Arbor", "Land Creature — Forest Dryad", ""), false},
		{"land with no ability", texted("Lonely Land", "Land", "This land enters tapped."), true},
		{"nonbasic land type", texted("Lone Desert", "Land — Desert", "This land enters tapped."), true},

		// Not permanent cards: nothing asked.
		{"instant", texted("Shock", "Instant", "Shock deals 2 damage to any target."), false},
		{"sorcery", texted("Divination", "Sorcery", "Draw two cards."), false},
		{"kindred instant", texted("Kindred Spell", "Kindred Instant — Goblin", "Draw a card."), false},
	}
	for _, tc := range cases {
		if got := permanentWithoutActivatedAbility(tc.card); got != tc.fails {
			t.Errorf("%s (%q): fails = %v, want %v", tc.name, tc.card.OracleText, got, tc.fails)
		}
	}
}

// Each face of a multi-faced card: outside the game a double-faced card
// is its front face (CR 712.8a), an adventurer card or a flip card its
// normal characteristics (CR 715.4, 710.2), and a split card both
// halves combined (CR 709.4).
func TestZirdaConditionReadsTheFacesACardHasOutsideTheGame(t *testing.T) {
	frontAbility := faced("transform",
		cards.CardFace{Name: "Front", TypeLine: "Creature — Human", OracleText: "{T}: Transform this creature."},
		cards.CardFace{Name: "Back", TypeLine: "Creature — Werewolf", OracleText: "Trample"})
	backOnly := faced("transform",
		cards.CardFace{Name: "Delver", TypeLine: "Creature — Human Wizard", OracleText: "At the beginning of your upkeep, look at the top card of your library. You may reveal that card. If an instant or sorcery card is revealed this way, transform this creature."},
		cards.CardFace{Name: "Aberration", TypeLine: "Creature — Human Insect", OracleText: "Flying\n{1}: Untap this creature."})
	mdfcSpellFront := faced("modal_dfc",
		cards.CardFace{Name: "Restoration", TypeLine: "Sorcery", OracleText: "Draw cards."},
		cards.CardFace{Name: "Reborn", TypeLine: "Land", OracleText: "This land enters tapped."})
	mdfcLandFront := faced("modal_dfc",
		cards.CardFace{Name: "Cliffs", TypeLine: "Land", OracleText: "As this land enters, you may pay 3 life. If you don't, it enters tapped.\n{T}: Add {R}."},
		cards.CardFace{Name: "Back Spell", TypeLine: "Instant", OracleText: "Deal 3 damage."})
	adventurer := faced("adventure",
		cards.CardFace{Name: "Giant", TypeLine: "Creature — Giant", OracleText: "Whenever this creature becomes the target of a spell, it deals 2 damage to that spell's controller."},
		cards.CardFace{Name: "Stomp", TypeLine: "Instant — Adventure", OracleText: "Stomp deals 2 damage to any target."})
	flip := faced("flip",
		cards.CardFace{Name: "Eggwatcher", TypeLine: "Creature — Snake Shaman", OracleText: "When this creature enters, flip it."},
		cards.CardFace{Name: "Broodmistress", TypeLine: "Legendary Creature — Snake Shaman", OracleText: "{G}, Sacrifice a creature: Target creature gets +3/+3 until end of turn."})
	roomOneHalf := faced("split",
		cards.CardFace{Name: "Hall", TypeLine: "Enchantment — Room", OracleText: "When you unlock this door, draw a card."},
		cards.CardFace{Name: "Vault", TypeLine: "Enchantment — Room", OracleText: "{2}: Scry 1."})
	roomNeither := faced("split",
		cards.CardFace{Name: "Foyer", TypeLine: "Enchantment — Room", OracleText: "When you unlock this door, draw a card."},
		cards.CardFace{Name: "Attic", TypeLine: "Enchantment — Room", OracleText: "Creatures you control get +1/+1."})
	spellSplit := faced("split",
		cards.CardFace{Name: "Fire", TypeLine: "Instant", OracleText: "Fire deals 2 damage divided as you choose."},
		cards.CardFace{Name: "Ice", TypeLine: "Instant", OracleText: "Tap target permanent."})

	cases := []struct {
		name  string
		card  cards.Card
		fails bool
	}{
		{"transform card, front face has one", frontAbility, false},
		{"transform card, only the back face has one", backOnly, true},
		{"modal card with a spell front", mdfcSpellFront, false},
		{"modal card with a land front", mdfcLandFront, false},
		{"adventurer card is its creature", adventurer, true},
		{"flip card is its unflipped half", flip, true},
		{"split permanent, one half has one", roomOneHalf, false},
		{"split permanent, neither half", roomNeither, true},
		{"split spell", spellSplit, false},
	}
	for _, tc := range cases {
		if got := permanentWithoutActivatedAbility(tc.card); got != tc.fails {
			t.Errorf("%s: fails = %v, want %v", tc.name, got, tc.fails)
		}
	}
}

// zirdaDeck is a legal 100-card deck that fulfils Zirda's condition: a
// commander with an activated ability and 99 instants, with Zirda as
// the companion. extra replaces that many of the 99.
func zirdaDeck(commander cards.Card, extra ...cards.Card) *List {
	list := &List{Commanders: []cards.Card{commander}}
	for i := range 99 - len(extra) {
		list.Mainboard = append(list.Mainboard, costed(fmt.Sprintf("Instant %d", i), "Instant", "{R}", "R"))
	}
	list.Mainboard = append(list.Mainboard, extra...)
	z := zirda()
	list.Companion = &z
	return list
}

func boros(name, text string) cards.Card {
	c := costed(name, "Legendary Creature — Human Soldier", "{2}{R}{W}", "R", "W")
	c.OracleText = text
	return c
}

func TestZirdaCompanionDeckValidates(t *testing.T) {
	list := zirdaDeck(boros("Tapper", "{T}: Tap target creature."),
		texted("Llanowar Elves", "Creature — Elf Druid", "{T}: Add {G}."),
		texted("Mountain", "Basic Land — Mountain", "({T}: Add {R}.)"),
		texted("Sword", "Artifact — Equipment", "Equip {2}"),
	)
	if err := Validate(list, game.FormatCommander); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

// The commander is in the starting deck (CR 702.139b, ADR 0172 §9), so
// a commander with no activated ability breaks the condition, and every
// offending card is named.
func TestZirdaCompanionNamesEachOffenderAndCountsTheCommander(t *testing.T) {
	list := zirdaDeck(boros("Vanilla Commander", "Vigilance"),
		texted("Bear", "Creature — Bear", ""),
		texted("Treasure Maker", "Creature — Goblin", "When this creature enters, create a Treasure token. (It's an artifact with \"{T}, Sacrifice this token: Add one mana of any color.\")"),
		texted("Sword", "Artifact — Equipment", "Equip {2}"),
	)
	err := Validate(list, game.FormatCommander)
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Validate: %v, want a ValidationError", err)
	}
	var offenders []string
	for _, v := range ve.Violations {
		if v.Code != CodeCompanionConditionNotMet {
			t.Errorf("unexpected violation %s: %s", v.Code, v.Message)
			continue
		}
		offenders = append(offenders, v.Card)
	}
	want := []string{"Vanilla Commander", "Bear", "Treasure Maker"}
	if !slices.Equal(offenders, want) {
		t.Errorf("offenders %v, want %v", offenders, want)
	}
}
