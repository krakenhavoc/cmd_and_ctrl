package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b536_cards_test.go — Shadowmoor: Creatures with keywords, batch 536
// (#5591, S82 OLD). Every test plays the real catalog entry through the engine.

const (
	oldB536Horde      = "1229b82f-fc50-4259-83c6-31f706ac2303"
	oldB536Rabble     = "17386f88-dbf0-47ae-91d1-6b1aaceee05f"
	oldB536Drowner    = "19a4010f-f5bf-48be-a1d7-eacc3bbeeb59"
	oldB536Godhead    = "1ed5751a-4b45-4cd2-b19c-ac84508de4f8"
	oldB536Harpoonist = "8ee98d6a-6a7e-411d-818c-816d84cfb7cd"
	oldB536Swarm      = "d7178feb-29ce-45d9-a970-2c6003779a00"
	oldB536Crowd      = "fc071117-08aa-47c3-ae4a-4b6c630befa3"
)

func TestOldB536CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB536Horde:      "Horde of Boggarts",
		oldB536Rabble:     "Kithkin Rabble",
		oldB536Drowner:    "Drowner Initiate",
		oldB536Godhead:    "Godhead of Awe",
		oldB536Harpoonist: "Kinscaer Harpoonist",
		oldB536Swarm:      "Faerie Swarm",
		oldB536Crowd:      "Crowd of Cinders",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

// oldB536Colored puts a coloured permanent onto controller's battlefield.
func oldB536Colored(g *game.Game, controller uuid.UUID, name, typeLine string, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Colors: colors, Owner: controller, Controller: controller})
}

// oldB536Real puts the real catalog card onto the battlefield with the colour and
// size its printing has, which a bare catalog push leaves out.
func oldB536Real(g *game.Game, controller uuid.UUID, name, typeLine, oracle, cost string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		ManaCost: cost, Power: power, Toughness: toughness, Colors: colors, Owner: controller, Controller: controller})
}

// oldB536ColourCreature is the shared check of the four "equal to the number of
// <colour> permanents you control" creatures: it counts itself, your coloured
// permanents of any type, and nobody else's.
func oldB536ColourCreature(t *testing.T, name, typeLine, oracle, cost, color, keyword string) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := oldB536Real(g, me.ID, name, typeLine, oracle, cost, 0, 0, color)
	if p, tough := effectstest.PtOf(t, g, id); p != 1 || tough != 1 {
		t.Errorf("%s alone is %d/%d, want 1/1 (it counts itself)", name, p, tough)
	}
	oldB536Colored(g, me.ID, "My Enchantment", "Enchantment", color)
	oldB536Colored(g, me.ID, "My Creature", "Creature — Bear", color)
	oldB536Colored(g, me.ID, "My Other Colour", "Creature — Bear", "G", "R", "W", "U", "B")
	oldB536Colored(g, opp.ID, "Their Creature", "Creature — Bear", color)
	oldB536Colored(g, me.ID, "My Colourless", "Artifact")
	// 1 itself + 2 of mine + the five-coloured creature.
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("%s with four of your %s permanents is %d/%d, want 4/4", name, color, p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, keyword) {
		t.Errorf("%s lacks %s", name, keyword)
	}
}

func TestOldB536HordeOfBoggartsCountsRedPermanents(t *testing.T) {
	oldB536ColourCreature(t, "Horde of Boggarts", "Creature — Goblin", oldB536Horde, "{3}{R}", "R", "menace")
}

func TestOldB536KithkinRabbleCountsWhitePermanents(t *testing.T) {
	oldB536ColourCreature(t, "Kithkin Rabble", "Creature — Kithkin", oldB536Rabble, "{3}{W}", "W", "vigilance")
}

func TestOldB536FaerieSwarmCountsBluePermanents(t *testing.T) {
	oldB536ColourCreature(t, "Faerie Swarm", "Creature — Faerie", oldB536Swarm, "{3}{U}", "U", "flying")
}

func TestOldB536CrowdOfCindersCountsBlackPermanents(t *testing.T) {
	oldB536ColourCreature(t, "Crowd of Cinders", "Creature — Elemental", oldB536Crowd, "{3}{B}", "B", "fear")
}

func TestOldB536GodheadOfAweMakesEveryOtherCreatureOneOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 4, 6)
	boosted := effectstest.B12Creature(g, opp.ID, "Boosted", "Creature — Ogre", 4, 6)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(boosted, game.CounterPlusOne, 2) })
	god := oldB536Real(g, me.ID, "Godhead of Awe", "Creature — Spirit Avatar", oldB536Godhead, "{W/U}{W/U}{W/U}{W/U}{W/U}", 4, 4, "W", "U")
	for id, want := range map[uuid.UUID]int{mine: 1, theirs: 1, boosted: 3, god: 4} {
		if p, tough := effectstest.PtOf(t, g, id); p != want || tough != want {
			t.Errorf("a creature is %d/%d, want %d/%d", p, tough, want, want)
		}
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, god, "flying") {
		t.Error("Godhead of Awe lacks flying")
	}
}

func TestOldB536KinscaerHarpoonistTakesFlyingFromATargetUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harpoonist := effectstest.PushCatalogPermanent(g, me.ID, "Kinscaer Harpoonist", "Creature — Kithkin Soldier", oldB536Harpoonist, false)
	drake := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	if !effectstest.EffectiveAbilitiesContain(t, g, drake, "flying") {
		t.Fatal("test setup: the Drake does not fly")
	}
	effectstest.DeclareAttack(t, g, opp.ID, harpoonist)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, drake)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, drake, "flying") {
		t.Error("the targeted creature still has flying")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, drake, "flying") {
		t.Error("the targeted creature did not get flying back at end of turn")
	}
}

func TestOldB536KinscaerHarpoonistMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harpoonist := effectstest.PushCatalogPermanent(g, me.ID, "Kinscaer Harpoonist", "Creature — Kithkin Soldier", oldB536Harpoonist, false)
	drake := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Drake", TypeLine: "Creature — Drake",
		Power: 2, Toughness: 2, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.DeclareAttack(t, g, opp.ID, harpoonist)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, drake)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, drake, "flying") {
		t.Error("the creature lost flying though the controller declined")
	}
}

func TestOldB536DrownerInitiateMillsTwoWhenAnyPlayerCastsABlueSpellAndYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Drowner Initiate", "Creature — Merfolk Wizard", oldB536Drowner, false)
		effectstest.SeedLibrary(opp, "A", "B", "C", "D", "E")
		library := opp.Library.Size()
		effectstest.TdCastFromHand(t, g, game.Card{Name: "Blue Spell", TypeLine: "Sorcery", ManaCost: "{U}", Colors: []string{"U"}}, game.CastSpellParams{})
		effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("the trigger did not offer the {1}")
		}
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := library
		if pay {
			want = library - 2
		}
		if opp.Library.Size() != want {
			t.Errorf("pay=%v: library %d, want %d", pay, opp.Library.Size(), want)
		}
	}
}

func TestOldB536DrownerInitiateIgnoresASpellOfAnotherColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Drowner Initiate", "Creature — Merfolk Wizard", oldB536Drowner, false)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Red Spell", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}}, game.CastSpellParams{})
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a red spell triggered Drowner Initiate")
	}
}
