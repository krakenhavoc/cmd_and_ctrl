package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b141_test.go — Throne of Eldraine, S82 older-sets batch 141
// (#5516): pump and combat tricks. Each test plays the real catalog entry.

const (
	oldB141RaidersOracle   = "05c42b13-5125-477c-991f-692c4e22bc53"
	oldB141SquireOracle    = "2a95ebaa-543e-4407-8050-adf6d8319fc1"
	oldB141BargeInOracle   = "32eb5d30-f4e7-48ef-b5f6-051ddddd4175"
	oldB141TrackerOracle   = "4ca31946-72ef-4c45-a7bb-494a99dc2982"
	oldB141ShinechaserOrcl = "6267ad33-34a6-47c8-ae09-f8d1499a761a"
	oldB141BelleOracle     = "6bc4a078-4d30-452d-9e33-ff09f943903e"
	oldB141NaughtyOracle   = "74303ddb-d3b6-433d-a299-e0c4a7267461"
	oldB141VengeanceOracle = "7b0d9139-b3ad-4065-9b55-aeeda8950bee"
	oldB141FerocityOracle  = "c75bdc5f-22f2-4558-bb8f-5a53ff7ab1ac"
	oldB141BattleguardOrcl = "ce07104c-8573-4dc0-a492-f9548d4317d2"
)

func TestOldB141EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB141RaidersOracle: "Redcap Raiders", oldB141SquireOracle: "Garenbrig Squire",
		oldB141BargeInOracle: "Barge In", oldB141TrackerOracle: "Wildwood Tracker",
		oldB141ShinechaserOrcl: "Shinechaser", oldB141BelleOracle: "Belle of the Brawl",
		oldB141NaughtyOracle: "Bog Naughty", oldB141VengeanceOracle: "Witch's Vengeance",
		oldB141FerocityOracle: "Ferocity of the Wilds", oldB141BattleguardOrcl: "Rowan's Battleguard",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB141Attack(t *testing.T, g *game.Game, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, attackers...)
}

// --- Redcap Raiders -------------------------------------------------------

func TestOldB141RedcapRaidersTapsANonHumanForPlusOneAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raiders := effectstest.B12Push(g, me.ID, "Redcap Raiders", "Creature — Goblin Warrior", oldB141RaidersOracle, 3, 2)
	human := effectstest.B12Creature(g, me.ID, "Peasant", "Creature — Human Peasant", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB141Attack(t, g, raiders)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	pc := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pc == nil {
		t.Fatal("no choose-cards prompt")
	}
	for _, id := range pc.ChooseCards {
		if id == human {
			t.Error("a Human was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, bear) {
		t.Error("the Bear was not tapped")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, raiders); p != 4 || tough != 3 {
		t.Errorf("Raiders is %d/%d, want 4/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, raiders, "trample") {
		t.Error("no trample")
	}
}

func TestOldB141RedcapRaidersDecliningGivesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raiders := effectstest.B12Push(g, me.ID, "Redcap Raiders", "Creature — Goblin Warrior", oldB141RaidersOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB141Attack(t, g, raiders)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the Bear was tapped after declining")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, raiders); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, raiders, "trample") {
		t.Error("trample without tapping")
	}
}

func TestOldB141RedcapRaidersOnlyHumansAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	raiders := effectstest.B12Push(g, me.ID, "Redcap Raiders", "Creature — Goblin Warrior", oldB141RaidersOracle, 3, 2)
	effectstest.B12Creature(g, me.ID, "Peasant", "Creature — Human Peasant", 1, 1)
	oldB141Attack(t, g, raiders)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked to tap although only a Human was untapped")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, raiders); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
}

// --- Garenbrig Squire -----------------------------------------------------

func TestOldB141GarenbrigSquireGrowsForACreatureSpellWithAnAdventureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	squire := effectstest.B12Push(g, me.ID, "Garenbrig Squire", "Creature — Human Soldier", oldB141SquireOracle, 2, 2)
	effectstest.B13PlayAs(t, g, 0, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, squire); p != 2 {
		t.Fatalf("a plain creature spell grew it to %d", p)
	}
	_, id := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if p, tough := effectstest.EffectivePTOf(t, g, squire); p != 3 || tough != 3 {
		t.Errorf("after an Adventure creature spell: %d/%d, want 3/3", p, tough)
	}
}

func TestOldB141GarenbrigSquireIgnoresTheAdventureHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	squire := effectstest.B12Push(g, me.ID, "Garenbrig Squire", "Creature — Human Soldier", oldB141SquireOracle, 2, 2)
	_, id := effectstest.HobA6HandAdventure(t, g, "adventurer-oracle",
		effectstest.HobA6Face("Knight", "Creature — Human Knight", "{W}", 1, 1, "W"),
		effectstest.HobA6Face("Trick", "Instant — Adventure", "{W}", 0, 0, "W"))
	effectstest.B10AddMana(me, "W")
	effectstest.HobA6MustCast(t, g, me, id, 1, nil)
	if p, _ := effectstest.EffectivePTOf(t, g, squire); p != 2 {
		t.Errorf("casting the Adventure half grew it to %d, want 2", p)
	}
}

// --- Barge In -------------------------------------------------------------

func oldB141CastInstant(t *testing.T, g *game.Game, p *game.Player, name, oracle string, targets []game.TargetRef, strict bool) error {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets, Strict: strict})
}

func TestOldB141BargeInPumpsTheTargetAndGivesEveryAttackingNonHumanTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Peasant", "Creature — Human Peasant", 1, 1)
	home := effectstest.B12Creature(g, me.ID, "Stay Home", "Creature — Wolf", 1, 1)
	oldB141Attack(t, g, bear, human)
	if err := oldB141CastInstant(t, g, me, "Barge In", oldB141BargeInOracle, effectstest.CardTarget(human), true); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, human); p != 3 || tough != 3 {
		t.Errorf("target is %d/%d, want 3/3", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, human, "trample") {
		t.Error("an attacking Human gained trample")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the attacking Bear did not gain trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, home, "trample") {
		t.Error("a creature that is not attacking gained trample")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the untargeted Bear is %d power, want 2", p)
	}
}

func TestOldB141BargeInRefusesANonAttackingTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	oldB141Attack(t, g, bear)
	if err := oldB141CastInstant(t, g, me, "Barge In", oldB141BargeInOracle, effectstest.CardTarget(idle), true); err == nil {
		t.Error("a creature that is not attacking was a legal target")
	}
}

// --- Wildwood Tracker -----------------------------------------------------

func TestOldB141WildwoodTrackerAttacksWithAnotherNonHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tracker := effectstest.B12Push(g, me.ID, "Wildwood Tracker", "Creature — Elf Warrior", oldB141TrackerOracle, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB141Attack(t, g, tracker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, tracker); p != 2 || tough != 2 {
		t.Errorf("Tracker is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB141WildwoodTrackerNeedsAnotherNonHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tracker := effectstest.B12Push(g, me.ID, "Wildwood Tracker", "Creature — Elf Warrior", oldB141TrackerOracle, 1, 1)
	effectstest.B12Creature(g, me.ID, "Peasant", "Creature — Human Peasant", 1, 1)
	oldB141Attack(t, g, tracker)
	effectstest.PassPriorityAroundTable(t, g)
	// The Tracker itself is a non-Human, but "another" excludes it.
	if p, _ := effectstest.EffectivePTOf(t, g, tracker); p != 1 {
		t.Errorf("Tracker is %d power with only Humans beside it, want 1", p)
	}
}

func TestOldB141WildwoodTrackerAlsoTriggersOnBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	tracker := effectstest.B12Push(g, opp.ID, "Wildwood Tracker", "Creature — Elf Warrior", oldB141TrackerOracle, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, tracker)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, tracker); p != 2 {
		t.Errorf("blocking Tracker is %d power, want 2", p)
	}
}

// --- Shinechaser ----------------------------------------------------------

func TestOldB141ShinechaserGrowsForEachOfAnArtifactAndAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chaser := effectstest.B12Push(g, me.ID, "Shinechaser", "Creature — Faerie", oldB141ShinechaserOrcl, 1, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, chaser); p != 1 {
		t.Fatalf("bare: %d, want 1", p)
	}
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if p, _ := effectstest.EffectivePTOf(t, g, chaser); p != 1 {
		t.Errorf("an opponent's artifact counted: %d", p)
	}
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if p, tough := effectstest.EffectivePTOf(t, g, chaser); p != 2 || tough != 2 {
		t.Errorf("with an artifact: %d/%d, want 2/2", p, tough)
	}
	effectstest.B12Permanent(g, me.ID, "Second Rock", "Artifact")
	if p, _ := effectstest.EffectivePTOf(t, g, chaser); p != 2 {
		t.Errorf("a second artifact stacked: %d, want 2", p)
	}
	effectstest.B12Permanent(g, me.ID, "Aura", "Enchantment")
	if p, tough := effectstest.EffectivePTOf(t, g, chaser); p != 3 || tough != 3 {
		t.Errorf("with both: %d/%d, want 3/3", p, tough)
	}
}

// --- Belle of the Brawl ---------------------------------------------------

func TestOldB141BelleOfTheBrawlPumpsOtherKnightsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	belle := effectstest.B12Push(g, me.ID, "Belle of the Brawl", "Creature — Human Knight", oldB141BelleOracle, 3, 2)
	knight := effectstest.B12Creature(g, me.ID, "Knight", "Creature — Human Knight", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", "Creature — Human Knight", 2, 2)
	oldB141Attack(t, g, belle)
	effectstest.PassPriorityAroundTable(t, g)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"Belle": {belle, 3}, "Knight": {knight, 3}, "Bear": {bear, 2}, "their Knight": {theirs, 2}} {
		if p, _ := effectstest.EffectivePTOf(t, g, c.id); p != c.want {
			t.Errorf("%s power %d, want %d", name, p, c.want)
		}
	}
}

// --- Bog Naughty ----------------------------------------------------------

func TestOldB141BogNaughtySacrificesAFoodForMinusThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	naughty := effectstest.B12Push(g, me.ID, "Bog Naughty", "Creature — Faerie", oldB141NaughtyOracle, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	var foods []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Food") {
			foods = append(foods, c.InstanceID)
		}
	}
	if len(foods) != 1 {
		t.Fatalf("Foods = %d, want 1", len(foods))
	}
	// No mana: refused.
	if err := g.ActivateCatalogAbility(me.ID, naughty, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim), SacrificeIDs: foods}); err == nil {
		t.Error("activated without paying {2}{B}")
	}
	effectstest.B10AddMana(me, "B", "B", "B")
	// No Food: refused.
	if err := g.ActivateCatalogAbility(me.ID, naughty, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Error("activated without sacrificing a Food")
	}
	if err := g.ActivateCatalogAbility(me.ID, naughty, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim), SacrificeIDs: foods}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foods[0]) {
		t.Error("the Food was not sacrificed")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("target is %d/%d, want 1/1", p, tough)
	}
}

// --- Witch's Vengeance ----------------------------------------------------

func TestOldB141WitchsVengeanceShrinksEveryCreatureOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Druid", 4, 4)
	theirElf := effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Warrior", 2, 3)
	goblin := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 2, 2)
	changeling := effectstest.PushTribalCreature(g, opp.ID, "Shapeshifter", "Creature — Shapeshifter", 5, 5, game.KeywordChangeling)
	effectstest.CastCatalogSpell(t, g, "Witch's Vengeance", "Sorcery", oldB141VengeanceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
	if p, tough := effectstest.EffectivePTOf(t, g, elf); p != 1 || tough != 1 {
		t.Errorf("my Elf is %d/%d, want 1/1", p, tough)
	}
	if g.Battlefield.Contains(theirElf) {
		t.Error("their 2/3 Elf survived -3/-3")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, goblin); p != 2 {
		t.Errorf("a Goblin is %d power, want 2", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, changeling); p != 2 {
		t.Errorf("a changeling is %d power, want 2", p)
	}
	late := effectstest.B12Creature(g, me.ID, "Late Elf", "Creature — Elf", 4, 4)
	if p, _ := effectstest.EffectivePTOf(t, g, late); p != 4 {
		t.Errorf("an Elf that arrived afterwards is %d power, want 4", p)
	}
}

// --- Ferocity of the Wilds ------------------------------------------------

func TestOldB141FerocityOfTheWildsPumpsAttackingNonHumansYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ferocity of the Wilds", "Enchantment", oldB141FerocityOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Peasant", "Creature — Human Peasant", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Wolf", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wolf", 2, 2)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Fatalf("before attacking: %d, want 2", p)
	}
	oldB141Attack(t, g, bear, human)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Errorf("attacking Bear power %d trample=%v, want 3 and trample", p, effectstest.HasEffectiveKeyword(t, g, bear, "trample"))
	}
	if p, _ := effectstest.EffectivePTOf(t, g, human); p != 2 || effectstest.HasEffectiveKeyword(t, g, human, "trample") {
		t.Errorf("attacking Human power %d, want 2 and no trample", p)
	}
	for name, id := range map[string]uuid.UUID{"idle": idle, "theirs": theirs} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != 2 || effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Errorf("%s creature changed: power %d", name, p)
		}
	}
}

// --- Rowan's Battleguard --------------------------------------------------

func TestOldB141RowansBattleguardGetsPlusThreeWithARowanPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guard := effectstest.B12Push(g, me.ID, "Rowan's Battleguard", "Creature — Human Knight", oldB141BattleguardOrcl, 3, 3)
	effectstest.B12Permanent(g, me.ID, "Other Walker", "Legendary Planeswalker — Jace")
	if p, _ := effectstest.EffectivePTOf(t, g, guard); p != 3 {
		t.Errorf("a non-Rowan walker pumped it: %d", p)
	}
	effectstest.B12Permanent(g, opp.ID, "Their Rowan", "Legendary Planeswalker — Rowan")
	if p, _ := effectstest.EffectivePTOf(t, g, guard); p != 3 {
		t.Errorf("an opponent's Rowan pumped it: %d", p)
	}
	effectstest.B12Permanent(g, me.ID, "Rowan, Fearless Sparkmage", "Legendary Planeswalker — Rowan")
	if p, tough := effectstest.EffectivePTOf(t, g, guard); p != 6 || tough != 3 {
		t.Errorf("with a Rowan: %d/%d, want 6/3", p, tough)
	}
}
