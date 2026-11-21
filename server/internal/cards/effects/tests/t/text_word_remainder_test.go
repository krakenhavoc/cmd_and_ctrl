package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_word_remainder_test.go — #3984, the readers #3872's two slices
// left (ADR 0160's 2026-10-16 follow-ups, "Still not reached"): reveal,
// discard and exile costs, colour- and type-setting effects, a shield's
// protected set, creature-type words in closures, and mana restrictions.
// Each reads its words through the text of the object whose clause it
// is (CR 612.2). The set rule on a sacrifice clause (Jarad's "a Swamp
// and a Forest") is game.TestCostSetRuleReadsTheSourcesWords: Jarad's
// own clause works only from the graveyard, where no text change
// reaches.

const (
	twxSlumberingToraOracle = "9317e70d-3d01-4652-b9ea-b4015a0f7b80"
	twxAnchorageOracle      = "91320daf-f69c-4350-b0fc-4bb37a6904b1"
	twxImperiousOracle      = "3fa71348-fa4d-4f39-a451-cf1570591991"
)

// Martyr of Bones' "Reveal X black cards from your hand" under a Sleight
// of Mind (black → red) is "X red cards": the shared candidate walk lists
// the red card and not the black ones, and the validator agrees.
func TestTextWordRemainderMartyrOfBonesRevealsRedCards(t *testing.T) {
	g, me, martyr, hand := effectstest.MartyrTable(t, "Martyr of Bones", effectstest.MartyrOfBonesOracle, 0)
	twhSleight(t, g, martyr, effectstest.TwBlack, effectstest.TwRed)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	spec, ok := effects.Lookup(effectstest.MartyrOfBonesOracle)
	if !ok {
		t.Fatal("Martyr of Bones is not in the catalog")
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() {
		opts = g.RevealCardsOptionsForEffect(me.ID, martyr, spec.Activated[0].Cost.RevealCards)
	})
	if !twhHas(opts, hand["red"]) || twhHas(opts, hand["black1"]) || twhHas(opts, hand["black2"]) {
		t.Errorf("reveal options %v: want the red card and no black one", opts)
	}

	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue: 1, RevealIDs: effectstest.Ids(hand["black1"]),
	}); err == nil {
		t.Fatal("a black card paid \"Reveal X red cards\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, martyr, 0, game.ActivateAbilityParams{
		XValue: 1, RevealIDs: effectstest.Ids(hand["red"]),
	}); err != nil {
		t.Fatalf("a red card pays \"Reveal X red cards\": %v", err)
	}
}

// Slumbering Tora's "Discard a Spirit or Arcane card" under an Artificial
// Evolution (Spirit → Elf) takes an Elf card and not a Spirit card; the
// animation's "Cat" is a word of the same text.
func TestTextWordRemainderSlumberingToraDiscardsAnElfCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tora := effectstest.PushCatalogPermanent(g, me.ID, "Slumbering Tora", "Artifact", twxSlumberingToraOracle, false)
	spirit := effectstest.CcHandCard(me, "Kami", "Creature — Spirit", "{2}")
	elf := effectstest.CcHandCard(me, "Llanowar Elves", "Creature — Elf Druid", "{G}")
	effectstest.TwxEvolve(t, g, tora, "Spirit", "Elf")

	spec, ok := effects.Lookup(twxSlumberingToraOracle)
	if !ok {
		t.Fatal("Slumbering Tora is not in the catalog")
	}
	var opts []uuid.UUID
	g.ReadSnapshot(func() {
		opts = g.DiscardCostOptionsForEffect(me.ID, tora, spec.Activated[0].Cost.DiscardCards)
	})
	if !twhHas(opts, elf) || twhHas(opts, spirit) {
		t.Errorf("discard options %v: want the Elf card and not the Spirit card", opts)
	}

	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, tora, 0, game.ActivateAbilityParams{DiscardIDs: effectstest.Ids(spirit)}); err == nil {
		t.Fatal("a Spirit card paid \"Discard an Elf or Arcane card\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, tora, 0, game.ActivateAbilityParams{DiscardIDs: effectstest.Ids(elf)}); err != nil {
		t.Fatalf("an Elf card pays it: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, tora); !c.IsCreature() {
		t.Fatal("the Tora is animated")
	}
	// "Cat" is not one of the changed words, so it is a Cat still.
	if c := effectstest.GaLayered(t, g, tora); !c.HasSubtype("Cat") {
		t.Errorf("the Tora is a Cat: subtypes %v", c.Effective().Subtypes)
	}
}

// Restless Anchorage's "becomes a 2/3 white and blue Bird" under a
// Sleight of Mind (white → green) and an Artificial Evolution (Bird →
// Elf) makes a green and blue Elf: SetColorsMod and AddSubtypesMod read
// the ability's text.
func TestTextWordRemainderRestlessAnchorageBecomesAGreenAndBlueElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	anchorage := effectstest.PushCatalogPermanent(g, me.ID, "Restless Anchorage", "Land", twxAnchorageOracle, false)
	twhSleight(t, g, anchorage, effectstest.TwWhite, effectstest.TwGreen)
	effectstest.TwxEvolve(t, g, anchorage, "Bird", "Elf")
	for _, c := range []string{"C", "W", "U"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if err := g.ActivateCatalogAbility(me.ID, anchorage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, anchorage)
	if !c.IsCreature() {
		t.Fatal("the land is a creature")
	}
	if !c.HasColor("G") || !c.HasColor("U") || c.HasColor("W") {
		t.Errorf("colours %v: want green and blue, not white", c.Effective().Colors)
	}
	if !c.HasSubtype("Elf") || c.HasSubtype("Bird") {
		t.Errorf("subtypes %v: want an Elf and not a Bird", c.Effective().Subtypes)
	}
}

// Eye of Nidhogg's "enchanted creature is a black Dragon" (the attached
// statics SetAttachedColors and SetAttachedTypes) reads the Aura's text:
// changed to "red" and to "Elf", the creature is a red Elf.
func TestTextWordRemainderEyeOfNidhoggMakesARedElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	eye := effectstest.CastCatalogSpell(t, g, "Eye of Nidhogg", "Legendary Enchantment — Aura", effectstest.YsBEyeOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.GaLayered(t, g, bear); !v.HasColor("B") || !v.HasSubtype("Dragon") {
		t.Fatal("setup: the printed Eye makes a black Dragon")
	}
	twhSleight(t, g, eye, effectstest.TwBlack, effectstest.TwRed)
	effectstest.TwxEvolve(t, g, eye, "Dragon", "Elf")
	v := effectstest.GaLayered(t, g, bear)
	if !v.HasColor("R") || v.HasColor("B") {
		t.Errorf("colours %v: want red and not black", v.Effective().Colors)
	}
	if !v.HasSubtype("Elf") || v.HasSubtype("Dragon") {
		t.Errorf("subtypes %v: want an Elf and not a Dragon", v.Effective().Subtypes)
	}
}

// Pack Leader's "prevent all combat damage that would be dealt this turn
// to Dogs you control" under an Artificial Evolution (Dog → Cat) protects
// Cats: the shield's recipient set is read through its text as it is
// made. Pack Leader's own type line is a Cat too.
func TestTextWordRemainderPackLeaderShieldsCats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leader := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Pack Leader", TypeLine: "Creature — Dog", OracleID: effectstest.RsPackLeader,
		Power: 2, Toughness: 5, Colors: []string{"W"}})
	dog := effectstest.SfPush(g, me.ID, "Dog", "Creature — Dog", 1, []string{"W"})
	cat := effectstest.SfPush(g, me.ID, "Cat", "Creature — Cat", 1, []string{"W"})
	effectstest.TwxEvolve(t, g, leader, "Dog", "Cat")
	b1 := effectstest.SfPush(g, opp.ID, "B1", "Creature — Test", 3, []string{"R"})
	b2 := effectstest.SfPush(g, opp.ID, "B2", "Creature — Test", 3, []string{"R"})
	effectstest.Pr7bAttack(t, g, opp.ID, leader, dog, cat)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: dog, b2: cat})
	effectstest.SfCombat(t, g)
	if effectstest.Pr6Marked(g, cat) != 0 {
		t.Errorf("cat %d: combat damage to Cats you control is prevented", effectstest.Pr6Marked(g, cat))
	}
	if effectstest.Pr6Marked(g, dog) != 3 {
		t.Errorf("dog %d, want 3: the shield no longer names Dogs", effectstest.Pr6Marked(g, dog))
	}
}

// A creature-type word in a static's closure (Imperious Perfect's "Other
// Elves you control get +1/+1") reads its source: changed to Goblins, it
// pumps a Goblin and not an Elf.
func TestTextWordRemainderImperiousPerfectPumpsGoblins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	perfect := effectstest.PushCatalogPermanent(g, me.ID, "Imperious Perfect", "Creature — Elf Warrior", twxImperiousOracle, false)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	goblin := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 1)
	if effectstest.GaLayered(t, g, elf).CurrentPower() != 2 || effectstest.GaLayered(t, g, goblin).CurrentPower() != 1 {
		t.Fatal("setup: the printed Perfect pumps Elves")
	}
	effectstest.TwxEvolve(t, g, perfect, "Elf", "Goblin")
	if p := effectstest.GaLayered(t, g, goblin).CurrentPower(); p != 2 {
		t.Errorf("Goblin power %d, want 2", p)
	}
	if p := effectstest.GaLayered(t, g, elf).CurrentPower(); p != 1 {
		t.Errorf("Elf power %d, want 1", p)
	}
}

// Eldrazi Temple's "Spend this mana only to cast colorless Eldrazi
// spells" under an Artificial Evolution (Eldrazi → Elf) mints mana for
// colourless Elves: a declared mana restriction reads its source's text.
func TestTextWordRemainderEldraziTempleManaFollowsTheChange(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	temple := effectstest.SeedPermanentWithOracle(g, me.ID, "Eldrazi Temple", "Land", effectstest.EldraziTempleOracle)
	effectstest.TwxEvolve(t, g, temple, "Eldrazi", "Elf")
	if err := g.ActivateManaAbility(me.ID, temple, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	cost, _ := game.ParseCost("{C}{C}")
	elf := game.Card{Name: "Colourless Elf", TypeLine: "Artifact Creature — Elf", ManaCost: "{2}"}
	eldrazi := game.Card{Name: "Ulamog", TypeLine: "Legendary Creature — Eldrazi", ManaCost: "{10}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(elf)) {
		t.Error("the mana pays for a colourless Elf")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(eldrazi)) {
		t.Error("the mana no longer pays for an Eldrazi")
	}
}
