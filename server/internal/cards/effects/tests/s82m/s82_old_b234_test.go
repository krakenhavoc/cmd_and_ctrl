package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b234_test.go — Commander 2017, batch 234 (#5533): each test plays the
// real catalog entry through the real cast, activation and trigger paths.

const (
	oldB234KheruOracle     = "dfea918b-5702-4c12-ba36-e6aa4ace09c7"
	oldB234CrosisOracle    = "f1ae1dcd-76cd-4714-9df2-1ab0ebd2ca41"
	oldB234OjutaiOracle    = "27d7f737-59a4-4948-9b2e-599ae848e054"
	oldB234DrinkerOracle   = "705d89f6-13a8-4f40-a01a-3d5d0ce4f34e"
	oldB234VampireOracle   = "7ed8a69b-fba0-4d4d-98b2-311a92373669"
	oldB234ArbiterOracle   = "a33f6070-590c-4be8-9180-b1ac2608cbc4"
	oldB234DemonicTutorOID = "82004860-e589-4e38-8d61-8c0210e4ea39"
)

func TestOldB234CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB234KheruOracle:   "Kheru Mind-Eater",
		oldB234CrosisOracle:  "Crosis, the Purger",
		oldB234OjutaiOracle:  "Ojutai, Soul of Winter",
		oldB234DrinkerOracle: "Vein Drinker",
		oldB234VampireOracle: "Skeletal Vampire",
		oldB234ArbiterOracle: "Leonin Arbiter",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Leonin Arbiter ---

func oldB234Searches(t *testing.T, g *game.Game, who *game.Player) bool {
	t.Helper()
	id := uuid.New()
	who.Hand.PushTop(game.Card{InstanceID: id, Name: "Demonic Tutor", TypeLine: "Sorcery", OracleID: oldB234DemonicTutorOID, Owner: who.ID, Controller: who.ID})
	if err := g.CastSpell(who.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Demonic Tutor: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return effectstest.SearchPromptFor(g, who.ID) != nil
}

func TestOldB234LeoninArbiterStopsEveryPlayersSearchUntilTheyPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Leonin Arbiter", "Creature — Cat Cleric", oldB234ArbiterOracle, false)
	effectstest.PushLibraryCard(me, game.Card{Name: "Needle", TypeLine: "Creature — Bear"})
	if oldB234Searches(t, g, me) {
		t.Fatal("a player searched despite Leonin Arbiter")
	}
	var arbiter uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Leonin Arbiter" {
				arbiter = c.InstanceID
			}
		}
	})
	if err := g.PerformSpecialAction(me.ID, arbiter, game.SpecialActionIgnoreEffect, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("the ignore was taken with no mana to pay {2}")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.PerformSpecialAction(me.ID, arbiter, game.SpecialActionIgnoreEffect, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("paying {2}: %v", err)
	}
	if !oldB234Searches(t, g, me) {
		t.Error("the player who paid {2} still could not search")
	}
}

func TestOldB234LeoninArbiterPayingDoesNotFreeOtherPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	arbiter := effectstest.PushCatalogPermanent(g, me.ID, "Leonin Arbiter", "Creature — Cat Cleric", oldB234ArbiterOracle, false)
	effectstest.PushLibraryCard(opp, game.Card{Name: "Needle", TypeLine: "Creature — Bear"})
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.PerformSpecialAction(me.ID, arbiter, game.SpecialActionIgnoreEffect, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("paying {2}: %v", err)
	}
	// The opponent's own search is still stopped: the ignore was for me only.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	if oldB234Searches(t, g, opp) {
		t.Error("another player searched because someone else paid")
	}
}

// --- Ojutai, Soul of Winter ---

func TestOldB234OjutaiTapsAndFreezesWhenADragonAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	whelp := effectstest.B12Creature(g, me.ID, "Whelp", "Creature — Dragon", 2, 2)
	effectstest.B12Push(g, me.ID, "Ojutai, Soul of Winter", "Legendary Creature — Dragon", oldB234OjutaiOracle, 5, 6)
	rock := effectstest.B12Push(g, opp.ID, "Rock Golem", "Artifact Creature — Golem", "", 3, 3)
	land := effectstest.B12Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AttackWithTrigger(t, g, whelp, opp.ID, rock)
	if !effectstest.AlliTapped(t, g, rock) {
		t.Fatal("the target was not tapped")
	}
	if effectstest.AlliTapped(t, g, land) {
		t.Error("a land was tapped")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.AlliTapped(t, g, rock) {
		t.Error("the target untapped during its controller's next untap step")
	}
}

func TestOldB234OjutaiIgnoresANonDragonAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, me.ID, "Ojutai, Soul of Winter", "Legendary Creature — Dragon", oldB234OjutaiOracle, 5, 6)
	rock := effectstest.B12Creature(g, opp.ID, "Rock Golem", "Artifact Creature — Golem", 3, 3)
	effectstest.AttackWithTrigger(t, g, bear, opp.ID, uuid.Nil)
	if effectstest.AlliTapped(t, g, rock) {
		t.Error("a non-Dragon attacker triggered Ojutai")
	}
}

// --- Crosis, the Purger ---

func oldB234CrosisHit(t *testing.T, pay bool) (*game.Game, *game.Player, *game.Player, map[string]uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crosis := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Crosis, the Purger", OracleID: oldB234CrosisOracle,
		TypeLine: "Legendary Creature — Dragon", Power: 6, Toughness: 6, Keywords: []string{"flying"}})
	_ = crosis
	hand := map[string]uuid.UUID{
		"red":     effectstest.HandCardFull(opp, "Red Spell", "Instant", "{R}", "", []string{"R"}),
		"redblue": effectstest.HandCardFull(opp, "Izzet Spell", "Instant", "{U}{R}", "", []string{"U", "R"}),
		"blue":    effectstest.HandCardFull(opp, "Blue Spell", "Instant", "{U}", "", []string{"U"}),
		"clr":     effectstest.HandCardFull(opp, "Rock", "Artifact", "{2}", "", nil),
	}
	effectstest.AttackWith(t, g, opp.ID, crosis)
	effectstest.ApaMana(me, "B", "B", "B")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, pay)
	return g, me, opp, hand
}

func TestOldB234CrosisDiscardsEveryCardOfTheChosenColor(t *testing.T) {
	g, me, opp, hand := oldB234CrosisHit(t, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"red", "redblue"} {
		if opp.Hand.Contains(hand[name]) {
			t.Errorf("the %s card was not discarded", name)
		}
	}
	for _, name := range []string{"blue", "clr"} {
		if !opp.Hand.Contains(hand[name]) {
			t.Errorf("the %s card was discarded", name)
		}
	}
	if opp.Life != 40-6 {
		t.Errorf("the opponent is at %d, want 34", opp.Life)
	}
}

func TestOldB234CrosisDeclinedDiscardsNothing(t *testing.T) {
	g, _, opp, hand := oldB234CrosisHit(t, false)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range hand {
		if !opp.Hand.Contains(id) {
			t.Errorf("the %s card was discarded after declining", name)
		}
	}
}

// --- Vein Drinker ---

func oldB234Drinker(t *testing.T, power, toughness int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	d := effectstest.PushCatalogPermanent(g, me.ID, "Vein Drinker", "Creature — Vampire", oldB234DrinkerOracle, false)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, d)
		c.Power, c.Toughness = power, toughness
	})
	return g, me, opp, d
}

func TestOldB234VeinDrinkerKillsASmallCreatureAndGrows(t *testing.T) {
	g, me, opp, d := oldB234Drinker(t, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear survived 4 damage")
	}
	if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 1 {
		t.Errorf("Vein Drinker has %d +1/+1 counters, want 1", got)
	}
	// The two blows are one resolution: no state-based action is checked
	// between them, so the Bear still hits back for its power.
	if marked := effectstest.DamageOn(t, g, d); marked != 2 {
		t.Errorf("Vein Drinker has %d damage, want 2 from the doomed Bear", marked)
	}
}

func TestOldB234VeinDrinkerTakesDamageBackFromABiggerCreature(t *testing.T) {
	g, me, opp, d := oldB234Drinker(t, 4, 4)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 9)
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ogre}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) {
		t.Fatal("a 3/9 died to 4 damage")
	}
	if got := effectstest.DamageOn(t, g, d); got != 3 {
		t.Errorf("Vein Drinker has %d damage, want 3 (the Ogre's power)", got)
	}
}

func TestOldB234VeinDrinkerGrowsWhenACreatureItDamagedInCombatDies(t *testing.T) {
	g, me, opp, d := oldB234Drinker(t, 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 4)
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, d, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// 4 damage on a 2/4 already killed it; a second creature dies later.
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear should be dead")
	}
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 1, 1)
	if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters = %d, want 1", got)
	}
	effectstest.P1Destroy(t, g, other)
	if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 1 {
		t.Errorf("a creature Vein Drinker never damaged gave a counter: %d", got)
	}
}

// --- Skeletal Vampire ---

func TestOldB234SkeletalVampireMakesTwoBatsAndSacrificesThemForMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vamp := effectstest.CastCatalogSpell(t, g, "Skeletal Vampire", "Creature — Vampire Skeleton", oldB234VampireOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Bat"); n != 2 {
		t.Fatalf("Bats after entering = %d, want 2", n)
	}
	var bats []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Bat" && c.Controller == me.ID {
				bats = append(bats, c.InstanceID)
			}
		}
	})
	effectstest.ApaMana(me, "B", "B", "B", "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, vamp, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: bats[:1]}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Bat"); n != 3 {
		t.Errorf("Bats after sacrificing one for two = %d, want 3", n)
	}
}

func TestOldB234SkeletalVampireSacrificingABatRegeneratesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vamp := effectstest.CastCatalogSpell(t, g, "Skeletal Vampire", "Creature — Vampire Skeleton", oldB234VampireOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var bat uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Bat" && c.Controller == me.ID {
				bat = c.InstanceID
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, vamp, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bat}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bat) {
		t.Error("the Bat was not sacrificed")
	}
	effectstest.P1Destroy(t, g, vamp)
	if !g.Battlefield.Contains(vamp) {
		t.Error("the regeneration shield did not save Skeletal Vampire")
	}
}

func TestOldB234SkeletalVampireRefusesANonBatSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vamp := effectstest.CastCatalogSpell(t, g, "Skeletal Vampire", "Creature — Vampire Skeleton", oldB234VampireOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, vamp, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a Bear was sacrificed as a Bat")
	}
}

// --- Kheru Mind-Eater ---

func TestOldB234KheruExilesAHandCardFaceDownAndYouMayCastIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kheru := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Kheru Mind-Eater", OracleID: oldB234KheruOracle,
		TypeLine: "Creature — Vampire", Power: 1, Toughness: 3, Keywords: []string{"menace"}})
	prize := effectstest.HandCardFull(opp, "Prize Bear", "Creature — Bear", "", "", []string{"G"})
	effectstest.HandCardFull(opp, "Other Bear", "Creature — Bear", "", "", []string{"G"})
	effectstest.AttackWith(t, g, opp.ID, kheru)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, prize)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(prize) {
		t.Fatalf("the chosen card is in %v, want exile", effectstest.ZoneOf(g, prize))
	}
	// Foes may not cast it; its link holder may, and it comes in under their control.
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, prize, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err == nil {
		t.Error("the card's owner cast the card exiled with Kheru Mind-Eater")
	}
	if err := g.CastSpell(me.ID, prize, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("casting the exiled card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(prize) {
		t.Errorf("the cast card is in %v, want the battlefield", effectstest.ZoneOf(g, prize))
	}
}

func TestOldB234KheruEmptyHandExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kheru := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Kheru Mind-Eater", OracleID: oldB234KheruOracle,
		TypeLine: "Creature — Vampire", Power: 1, Toughness: 3, Keywords: []string{"menace"}})
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	effectstest.AttackWith(t, g, opp.ID, kheru)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != 0 {
		t.Errorf("exile holds %d cards, want 0", len(g.Exile.Cards))
	}
}
