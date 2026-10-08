package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// daynight_b_cards_test.go — the second half of the day/night card list
// (#2586, slice daynight-b). The werewolves go through the importer
// (dnaWerewolf.row) so the per-face daybound / nightbound stamping is on
// the road under test; the rest are pushed catalogue permanents.

const (
	dnbOakshade   = "1b639537-a8fe-4615-b937-afa478fdec0f"
	dnbLiberator  = "9545f126-062f-4410-a362-e16255a128d6"
	dnbStormseek  = "ea5fd21a-c23a-49ee-aab8-0a9618d65c11"
	dnbShady      = "10be1b27-bc9f-4c6e-ac85-f1fa8b2a34d6"
	dnbPainter    = "4ff8c359-430f-434e-ac1e-822abdc28360"
	dnbPhoenix    = "94f14f42-ea66-4432-81cc-667ad62a4497"
	dnbStowaway   = "da412474-1da8-409b-bd7b-87cbfc239fb0"
	dnbRuffian    = "73a3b9a1-37a0-469a-9557-8c118a1ee78f"
	dnbHauler     = "c31e9db3-5d9d-470a-871a-b4b5b0536db5"
	dnbHuntmaster = "18563bc9-6090-4629-b304-89a67d93f635"
	dnbTovolar    = "45d49831-548a-4a0e-9a18-9f7397913895"
	dnbMoonrise   = "df33b298-9e4b-4522-96e5-c8d8d6af2bb4"
	dnbVadrik     = "a9ab17ce-ba07-4ce8-99f9-9c4c633957f7"
	dnbArsonist   = "215dfa88-b130-44df-9cfc-f1f0e4a36f4d"
	dnbPrisoner   = "bdcf0af3-3976-400d-a8b7-15e959e2b255"
	dnbWeaver     = "a64ecae7-0b09-48e1-8108-89442547ffda"
	dnbWolfStrike = "34237f39-1db3-4b3d-b87b-5fd1fbd5462a"
	dnbOutcast    = "ac3d07cd-88c4-4e22-9d86-f92d34f406d4"
)

var dnbWerewolves = []dnaWerewolf{
	{dnbOakshade, "Oakshade Stalker", "Moonlit Ambusher", "Creature — Human Ranger Werewolf", "Creature — Werewolf", "{2}{G}",
		"You may cast this spell as though it had flash if you pay {2} more to cast it.\n" + dnDay,
		dnNight, "3", "3", "6", "3", []string{"G"}, nil, nil, nil},
	{dnbLiberator, "Outland Liberator", "Frenzied Trapbreaker", "Creature — Human Werewolf", "Creature — Werewolf", "{1}{G}",
		"{1}, Sacrifice this creature: Destroy target artifact or enchantment.\n" + dnDay,
		"{1}, Sacrifice this creature: Destroy target artifact or enchantment.\nWhenever this creature attacks, destroy target artifact or enchantment defending player controls.\n" + dnNight,
		"2", "2", "3", "3", []string{"G"}, nil, nil, nil},
	{dnbStormseek, "Reckless Stormseeker", "Storm-Charged Slasher", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{R}",
		"At the beginning of combat on your turn, target creature you control gets +1/+0 and gains haste until end of turn.\n" + dnDay,
		"At the beginning of combat on your turn, target creature you control gets +2/+0 and gains trample and haste until end of turn.\n" + dnNight,
		"2", "3", "3", "4", []string{"R"}, nil, nil, nil},
	{dnbShady, "Shady Traveler", "Stalking Predator", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{B}",
		"Menace (This creature can't be blocked except by two or more creatures.)\n" + dnDay,
		"Menace (This creature can't be blocked except by two or more creatures.)\n" + dnNight,
		"2", "3", "4", "4", []string{"B"}, []string{"Menace"}, []string{"menace"}, []string{"menace"}},
	{dnbPainter, "Spellrune Painter", "Spellrune Howler", "Creature — Human Shaman Werewolf", "Creature — Werewolf", "{2}{R}",
		"Whenever you cast an instant or sorcery spell, this creature gets +1/+1 until end of turn.\n" + dnDay,
		"Whenever you cast an instant or sorcery spell, this creature gets +2/+2 until end of turn.\n" + dnNight,
		"2", "3", "3", "4", []string{"R"}, nil, nil, nil},
	{dnbStowaway, "Suspicious Stowaway", "Seafaring Werewolf", "Creature — Human Rogue Werewolf", "Creature — Werewolf", "{1}{U}",
		"This creature can't be blocked.\nWhenever this creature deals combat damage to a player, draw a card, then discard a card.\n" + dnDay,
		"This creature can't be blocked.\nWhenever this creature deals combat damage to a player, draw a card.\n" + dnNight,
		"1", "1", "2", "1", []string{"U"}, nil, nil, nil},
	{dnbRuffian, "Tavern Ruffian", "Tavern Smasher", "Creature — Human Warrior Werewolf", "Creature — Werewolf", "{3}{R}",
		dnDay, dnNight, "2", "5", "6", "5", []string{"R"}, nil, nil, nil},
	{dnbHauler, "Tireless Hauler", "Dire-Strain Brawler", "Creature — Human Werewolf", "Creature — Werewolf", "{4}{G}",
		"Vigilance\n" + dnDay, "Vigilance\n" + dnNight,
		"4", "5", "6", "6", []string{"G"}, []string{"Vigilance"}, []string{"vigilance"}, []string{"vigilance"}},
	{dnbHuntmaster, "Tovolar's Huntmaster", "Tovolar's Packleader", "Creature — Human Werewolf", "Creature — Werewolf", "{4}{G}{G}",
		"When this creature enters, create two 2/2 green Wolf creature tokens.\n" + dnDay,
		"Whenever this creature enters or attacks, create two 2/2 green Wolf creature tokens.\n{2}{G}{G}: Another target Wolf or Werewolf you control fights target creature you don't control.\n" + dnNight,
		"6", "6", "7", "7", []string{"G"}, nil, nil, nil},
	{dnbTovolar, "Tovolar, Dire Overlord", "Tovolar, the Midnight Scourge", "Legendary Creature — Human Werewolf", "Legendary Creature — Werewolf", "{1}{R}{G}",
		"Whenever a Wolf or Werewolf you control deals combat damage to a player, draw a card.\nAt the beginning of your upkeep, if you control three or more Wolves and/or Werewolves, it becomes night. Then transform any number of Human Werewolves you control.\nDaybound",
		"Whenever a Wolf or Werewolf you control deals combat damage to a player, draw a card.\n{X}{R}{G}: Target Wolf or Werewolf you control gets +X/+0 and gains trample until end of turn.\nNightbound",
		"3", "3", "4", "4", []string{"R", "G"}, nil, nil, nil},
	{dnbArsonist, "Volatile Arsonist", "Dire-Strain Anarchist", "Creature — Human Werewolf", "Creature — Werewolf", "{3}{R}{R}",
		"Menace, haste\nWhenever this creature attacks, it deals 1 damage to each of up to one target creature, up to one target player, and/or up to one target planeswalker.\n" + dnDay,
		"Menace, haste\nWhenever this creature attacks, it deals 2 damage to each of up to one target creature, up to one target player, and/or up to one target planeswalker.\n" + dnNight,
		"4", "4", "5", "5", []string{"R"}, []string{"Menace", "Haste"}, []string{"menace", "haste"}, []string{"menace", "haste"}},
	{dnbPrisoner, "Weary Prisoner", "Wrathful Jailbreaker", "Creature — Human Werewolf", "Creature — Werewolf", "{3}{R}",
		"Defender\n" + dnDay, "This creature attacks each combat if able.\n" + dnNight,
		"2", "6", "6", "6", []string{"R"}, []string{"Defender"}, []string{"defender"}, nil},
	{dnbWeaver, "Weaver of Blossoms", "Blossom-Clad Werewolf", "Creature — Human Werewolf", "Creature — Werewolf", "{2}{G}",
		"{T}: Add one mana of any color.\n" + dnDay, "{T}: Add two mana of any one color.\n" + dnNight,
		"2", "3", "3", "4", []string{"G"}, nil, nil, nil},
	{dnbOutcast, "Wolfkin Outcast", "Wedding Crasher", "Creature — Human Werewolf", "Creature — Werewolf", "{5}{G}",
		"This spell costs {2} less to cast if you control a Wolf or Werewolf.\n" + dnDay,
		"Whenever this creature or another Wolf or Werewolf you control dies, draw a card.\n" + dnNight,
		"5", "4", "6", "5", []string{"G"}, nil, nil, nil},
}

func dnbWerewolf(oracle string) dnaWerewolf {
	for _, w := range dnbWerewolves {
		if w.oracle == oracle {
			return w
		}
	}
	panic("no such werewolf " + oracle)
}

// Every werewolf in the slice turns over with the designation, its two
// faces are the right cards, and each face's printed keywords ride.
func TestDayNightBWerewolvesTurnOverAndKeepTheirKeywords(t *testing.T) {
	for _, w := range dnbWerewolves {
		t.Run(w.front, func(t *testing.T) {
			spec, ok := Lookup(w.oracle)
			if !ok || spec.Name != w.front {
				t.Fatalf("front face is not registered as %q: %+v", w.front, spec)
			}
			back, ok := Lookup(w.oracle + "#1")
			if !ok || back.Name != w.back {
				t.Fatalf("back face is not registered as %q: %+v", w.back, back)
			}
			if !slices.Contains(spec.PrintedKeywords, "daybound") || !slices.Contains(back.PrintedKeywords, "nightbound") {
				t.Errorf("keywords: front %v back %v, want daybound / nightbound", spec.PrintedKeywords, back.PrintedKeywords)
			}
			g, _, _, id := dnaTable(t, w)
			if c, _ := battlefieldCard(g, id); c.Name != w.front || c.ActiveFace != 0 {
				t.Fatalf("by day: %s face %d, want %s", c.Name, c.ActiveFace, w.front)
			}
			for _, k := range w.frontKeywords {
				if !slices.Contains(effectiveAbilities(t, g, id), k) {
					t.Errorf("front lacks %q: %v", k, effectiveAbilities(t, g, id))
				}
			}
			dnaNight(g)
			if c, _ := battlefieldCard(g, id); c.Name != w.back || c.ActiveFace != 1 {
				t.Fatalf("by night: %s face %d, want %s", c.Name, c.ActiveFace, w.back)
			}
			for _, k := range w.backKeys {
				if !slices.Contains(effectiveAbilities(t, g, id), k) {
					t.Errorf("back lacks %q: %v", k, effectiveAbilities(t, g, id))
				}
			}
			dnaDay(g)
			if c, _ := battlefieldCard(g, id); c.Name != w.front {
				t.Errorf("back by day: %s, want %s", c.Name, w.front)
			}
		})
	}
}

// The vanilla and keyword werewolves carry exactly their printed
// numbers on each face.
func TestDayNightBTableWerewolvesHaveTheirPrintedPower(t *testing.T) {
	for _, tc := range []struct {
		oracle        string
		front, back   int
		frontT, backT int
	}{
		{dnbRuffian, 2, 6, 5, 5},
		{dnbHauler, 4, 6, 5, 6},
		{dnbShady, 2, 4, 3, 4},
	} {
		w := dnbWerewolf(tc.oracle)
		g, _, _, id := dnaTable(t, w)
		if got := effectivePower(t, g, id); got != tc.front {
			t.Errorf("%s power = %d, want %d", w.front, got, tc.front)
		}
		if got := effectiveToughness(t, g, id); got != tc.frontT {
			t.Errorf("%s toughness = %d, want %d", w.front, got, tc.frontT)
		}
		dnaNight(g)
		if got := effectivePower(t, g, id); got != tc.back {
			t.Errorf("%s power = %d, want %d", w.back, got, tc.back)
		}
		if got := effectiveToughness(t, g, id); got != tc.backT {
			t.Errorf("%s toughness = %d, want %d", w.back, got, tc.backT)
		}
	}
}

// Oakshade Stalker's flash-for-more option is the declared gap.
func TestOakshadeStalkerShipsWithItsFlashCaveat(t *testing.T) {
	spec, ok := Lookup(dnbOakshade)
	if !ok || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Fatalf("Oakshade Stalker: %+v, want one caveat", spec)
	}
	if back, _ := Lookup(dnbOakshade + "#1"); back.Completeness != CompletenessFull {
		t.Errorf("Moonlit Ambusher is %v, want full", back.Completeness)
	}
}

// Outland Liberator sacrifices to destroy an artifact or enchantment.
func TestOutlandLiberatorSacrificesToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lib := b12Push(g, me.ID, "Outland Liberator", "Creature — Human Werewolf", dnbLiberator, 2, 2)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("the ability accepted a creature that is neither artifact nor enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, lib, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: relic}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := battlefieldCard(g, lib); ok {
		t.Error("the sacrifice was not paid at announce")
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, relic); ok {
		t.Error("the Relic survived")
	}
}

// Frenzied Trapbreaker's attack trigger reaches only what the defending
// player controls.
func TestFrenziedTrapbreakerDestroysADefendersArtifactOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	breaker := b12Push(g, me.ID, "Frenzied Trapbreaker", "Creature — Werewolf", dnbLiberator+"#1", 3, 3)
	mine := b12Permanent(g, me.ID, "My Relic", "Artifact")
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	other := b12Permanent(g, third.ID, "Third Relic", "Enchantment")
	declareAttack(t, g, opp.ID, breaker)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	for _, bad := range []uuid.UUID{mine, other} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Error("the trigger accepted a permanent the defending player does not control")
		}
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, theirs); ok {
		t.Error("the defending player's Relic survived")
	}
	for _, id := range []uuid.UUID{mine, other} {
		if _, ok := battlefieldCard(g, id); !ok {
			t.Error("a permanent the defending player does not control was destroyed")
		}
	}
}

// Reckless Stormseeker boosts and hastes a creature at the beginning of
// combat on its controller's turn; the back face adds trample.
func TestRecklessStormseekerBoostsAtBeginningOfCombat(t *testing.T) {
	for _, night := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		oracle, name, power := dnbStormseek, "Reckless Stormseeker", 1
		if night {
			oracle, name, power = dnbStormseek+"#1", "Storm-Charged Slasher", 2
		}
		b12Push(g, me.ID, name, "Creature — Werewolf", oracle, 2, 3)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		advanceTo(t, g, game.StepBeginCombat)
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
		if got := effectivePower(t, g, bear); got != 2+power {
			t.Errorf("night=%v: Bear power = %d, want %d", night, got, 2+power)
		}
		abil := effectiveAbilities(t, g, bear)
		if !slices.Contains(abil, "haste") {
			t.Errorf("night=%v: Bear lacks haste: %v", night, abil)
		}
		if night != slices.Contains(abil, "trample") {
			t.Errorf("night=%v: trample = %v", night, slices.Contains(abil, "trample"))
		}
	}
}

// Spellrune Painter grows for an instant or sorcery and for nothing else.
func TestSpellruneWerewolfGrowsForInstantsAndSorceries(t *testing.T) {
	for _, night := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		oracle, name, bonus := dnbPainter, "Spellrune Painter", 1
		if night {
			oracle, name, bonus = dnbPainter+"#1", "Spellrune Howler", 2
		}
		id := b12Push(g, me.ID, name, "Creature — Werewolf", oracle, 2, 3)
		castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
		passPriorityAroundTable(t, g)
		if got := effectivePower(t, g, id); got != 2 {
			t.Fatalf("night=%v: a creature spell pumped it to %d", night, got)
		}
		castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
		passPriorityAroundTable(t, g)
		if got := effectivePower(t, g, id); got != 2+bonus {
			t.Errorf("night=%v: power = %d, want %d", night, got, 2+bonus)
		}
		if got := effectiveToughness(t, g, id); got != 3+bonus {
			t.Errorf("night=%v: toughness = %d, want %d", night, got, 3+bonus)
		}
	}
}

// Suspicious Stowaway can't be blocked, and Seafaring Werewolf draws.
func TestSuspiciousStowawayCantBeBlockedAndSeafaringWerewolfDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := b12Push(g, me.ID, "Seafaring Werewolf", "Creature — Werewolf", dnbStowaway+"#1", 2, 1)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	handBefore := len(me.Hand.Cards)
	declareAttack(t, g, opp.ID, wolf)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, wolf); err == nil {
		t.Fatal("a creature blocked an unblockable werewolf")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d: one card drawn", got, handBefore+1)
	}
}

// The front face loots: draws one, discards one.
func TestSuspiciousStowawayLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	front := b12Push(g, me.ID, "Suspicious Stowaway", "Creature — Human Rogue Werewolf", dnbStowaway, 1, 1)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Drawn Card", TypeLine: "Sorcery", Owner: me.ID})
	handBefore := len(me.Hand.Cards)
	declareAttack(t, g, opp.ID, front)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) == handBefore+1 {
		answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
		passPriorityAroundTable(t, g)
	}
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("hand = %d, want %d after draw then discard", got, handBefore)
	}
}

// Weaver of Blossoms taps for one colour; Blossom-Clad Werewolf for two
// of ONE colour.
func TestWeaverOfBlossomsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	weaver := b12Push(g, me.ID, "Weaver of Blossoms", "Creature — Human Werewolf", dnbWeaver, 2, 3)
	clad := b12Push(g, me.ID, "Blossom-Clad Werewolf", "Creature — Werewolf", dnbWeaver+"#1", 3, 4)
	if err := g.ActivateManaAbility(me.ID, weaver, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Weaver: %v", err)
	}
	if b10ResolveAllManaPicks(t, g, me.ID, "G") != 1 {
		t.Fatal("Weaver raised no colour pick")
	}
	if got := len(me.ManaPool); got != 1 {
		t.Fatalf("pool after Weaver = %d, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, clad, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Blossom-Clad: %v", err)
	}
	if b10ResolveAllManaPicks(t, g, me.ID, "R") != 1 {
		t.Fatal("Blossom-Clad Werewolf raised no colour pick")
	}
	if got := len(me.ManaPool); got != 3 {
		t.Errorf("pool after both = %d, want 3", got)
	}
	reds := 0
	for _, m := range me.ManaPool {
		if m.Color == "R" {
			reds++
		}
	}
	if reds != 2 {
		t.Errorf("red mana = %d, want both of Blossom-Clad Werewolf's to be red", reds)
	}
}

// Sunstreak Phoenix comes back from the graveyard tapped when the
// designation flips and its controller pays {1}{R}.
func TestSunstreakPhoenixReturnsOnAFlipForTheFee(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phoenix := dnCast(t, g, "Sunstreak Phoenix", "Creature — Phoenix", dnbPhoenix)
	if !g.IsDay() {
		t.Fatal("entering did not make it day")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(phoenix); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if _, ok := battlefieldCard(g, phoenix); ok {
		t.Fatal("the Phoenix did not die")
	}
	dnaNight(g)
	monSettle(t, g)
	floatMana(t, g, me, "{R}{R}")
	answerPayUnless(t, g, me.ID, true)
	monSettle(t, g)
	c, ok := battlefieldCard(g, phoenix)
	if !ok {
		t.Fatal("the Phoenix did not return")
	}
	if !c.Tapped {
		t.Error("it returned untapped")
	}
}

func TestSunstreakPhoenixStaysIfYouDeclineTheFee(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	phoenix := dnCast(t, g, "Sunstreak Phoenix", "Creature — Phoenix", dnbPhoenix)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(phoenix); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	dnaNight(g)
	monSettle(t, g)
	answerPayUnless(t, g, me.ID, false)
	monSettle(t, g)
	if _, ok := battlefieldCard(g, phoenix); ok {
		t.Error("it returned without being paid for")
	}
}

// A Phoenix that is still on the battlefield offers nothing on a flip.
func TestSunstreakPhoenixOnTheBattlefieldIgnoresAFlip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dnCast(t, g, "Sunstreak Phoenix", "Creature — Phoenix", dnbPhoenix)
	dnaNight(g)
	monSettle(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a payment was offered for a Phoenix that is not in the graveyard")
	}
}

// Vadrik makes your instants and sorceries cheaper by its power, and
// grows on each flip.
func TestVadrikDiscountsByPowerAndGrowsOnAFlip(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vadrik := b12Push(g, me.ID, "Vadrik, Astral Archmage", "Legendary Creature — Human Wizard", dnbVadrik, 1, 2)
	g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationDay })
	if got := dnaPower(t, g, vadrik); got != 1 {
		t.Fatalf("Vadrik power = %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Bolt", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("an instant costs %d, want 1 with Vadrik at power 1", got)
	}
	if got := priceInHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("a creature spell costs %d, want 2: Vadrik discounts instants and sorceries", got)
	}
	if got := priceInHand(t, g, opp, "Their Bolt", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("an opponent's instant costs %d, want 2", got)
	}
	dnaNight(g)
	monSettle(t, g)
	if got := dnaPower(t, g, vadrik); got != 2 {
		t.Fatalf("Vadrik power = %d after a flip, want 2", got)
	}
	if got := priceInHand(t, g, me, "Big Sorcery", "Sorcery", "{3}{U}"); got != 2 {
		t.Errorf("a sorcery costs %d, want 2 with Vadrik at power 2", got)
	}
}

// Wolf Strike: the biter deals its power; at night it first gets +2/+0.
func TestWolfStrikeBitesAndPumpsAtNight(t *testing.T) {
	for _, night := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		biter := b12Creature(g, me.ID, "Biter", "Creature — Wolf", 2, 2)
		victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 0, 6)
		if night {
			g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationNight })
		} else {
			g.WithWriteLock(func() { g.DayNight.Designation = game.DesignationDay })
		}
		castCatalogSpell(t, g, "Wolf Strike", "Instant", dnbWolfStrike, []game.TargetRef{
			{Kind: game.TargetCard, ID: biter}, {Kind: game.TargetCard, ID: victim},
		})
		passPriorityAroundTable(t, g)
		wantPower, wantDamage := 2, 2
		if night {
			wantPower, wantDamage = 4, 4
		}
		if got := effectivePower(t, g, biter); got != wantPower {
			t.Errorf("night=%v: Biter power = %d, want %d", night, got, wantPower)
		}
		c, ok := battlefieldCard(g, victim)
		if !ok || c.DamageMarked != wantDamage {
			t.Errorf("night=%v: Victim damage = %d (present %v), want %d", night, c.DamageMarked, ok, wantDamage)
		}
		if c, _ := battlefieldCard(g, biter); c.DamageMarked != 0 {
			t.Errorf("night=%v: a one-sided bite damaged the biter (%d)", night, c.DamageMarked)
		}
	}
}

// Wolf Strike cannot name your own creature as the victim.
func TestWolfStrikeRefusesAnOwnVictim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", "Creature — Wolf", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	err := castCatalogSpellErr(t, g, "Wolf Strike", "Instant", dnbWolfStrike, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b},
	})
	if err == nil {
		t.Error("Wolf Strike accepted a creature you control as the second target")
	}
}

// Wolfkin Outcast costs {2} less with a Wolf or Werewolf, and only then.
func TestWolfkinOutcastCostsLessWithAWolfOrWerewolf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, dnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 6 {
		t.Fatalf("with no Wolf the Outcast costs %d, want 6", got)
	}
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := selfPricedMV(t, g, me, dnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 6 {
		t.Errorf("with a Bear the Outcast costs %d, want 6", got)
	}
	b12Creature(g, g.Seats[1].ID, "Their Wolf", "Creature — Wolf", 2, 2)
	if got := selfPricedMV(t, g, me, dnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 6 {
		t.Errorf("with only an opponent's Wolf the Outcast costs %d, want 6", got)
	}
	b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	if got := selfPricedMV(t, g, me, dnbOutcast, "Creature — Human Werewolf", "{5}{G}"); got != 4 {
		t.Errorf("with a Wolf the Outcast costs %d, want 4", got)
	}
}

// Wedding Crasher draws when it or another Wolf or Werewolf you control
// dies, and for nothing else.
func TestWeddingCrasherDrawsForWolvesAndWerewolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crasher := b12Push(g, me.ID, "Wedding Crasher", "Creature — Werewolf", dnbOutcast+"#1", 6, 5)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	kill := func(id uuid.UUID) {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
	}
	hand := len(me.Hand.Cards)
	kill(bear)
	kill(theirs)
	if got := len(me.Hand.Cards); got != hand {
		t.Fatalf("a Bear or an opponent's Wolf drew %d cards", got-hand)
	}
	kill(wolf)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("your Wolf dying drew %d cards, want 1", got-hand)
	}
	kill(crasher)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("Wedding Crasher dying drew %d more cards, want 1", got-hand-1)
	}
}

// Unnatural Moonrise: night, +1/+0, trample and a draw on combat damage.
func TestUnnaturalMoonriseMakesNightAndGrantsTheDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Unnatural Moonrise", "Sorcery", dnbMoonrise, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !g.IsNight() {
		t.Fatalf("designation = %q, want night", dnDesignation(g))
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power = %d, want 3", got)
	}
	if !slices.Contains(effectiveAbilities(t, g, bear), "trample") {
		t.Error("the Bear did not gain trample")
	}
	hand := len(me.Hand.Cards)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d, want %d: the granted trigger draws", got, hand+1)
	}
}

// The spell's only target is the creature, so when it is gone the spell
// does not resolve at all (CR 608.2b) and nothing becomes night.
func TestUnnaturalMoonriseFizzlesWhenTheTargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Unnatural Moonrise", "Sorcery", dnbMoonrise, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.IsNight() {
		t.Error("the spell resolved with no legal target")
	}
}

func TestUnnaturalMoonriseHasFlashback(t *testing.T) {
	spec, ok := Lookup(dnbMoonrise)
	if !ok || !slices.Contains(spec.CastableZones, game.ZoneGraveyard) || len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "flashback" {
		t.Fatalf("Unnatural Moonrise: %+v, want flashback from the graveyard", spec)
	}
}

// Tovolar's Huntmaster makes two Wolves; cast at night, Packleader's
// enters trigger makes two as well.
func TestTovolarsHuntmasterMakesTwoWolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	importAndCast(t, g, dnbWerewolf(dnbHuntmaster).row(), me)
	monSettle(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Wolf"); n != 2 {
		t.Errorf("Wolf tokens by day = %d, want 2", n)
	}
	g2 := newCatalogGame(t)
	g2.WithWriteLock(func() { g2.DayNight.Designation = game.DesignationNight })
	me2 := g2.Seats[0]
	id := importAndCast(t, g2, dnbWerewolf(dnbHuntmaster).row(), me2)
	monSettle(t, g2)
	if c, _ := battlefieldCard(g2, id); c.Name != "Tovolar's Packleader" {
		t.Fatalf("cast at night: %s, want Tovolar's Packleader", c.Name)
	}
	if n, _ := countTokensNamed(g2, me2.ID, "Wolf"); n != 2 {
		t.Errorf("Wolf tokens at night = %d, want 2 (the back face's enters trigger)", n)
	}
}

// Tovolar's Packleader's fight: another Wolf or Werewolf of yours fights
// a creature you don't control.
func TestTovolarsPackleaderFightsWithAnotherWolf(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	leader := b12Push(g, me.ID, "Tovolar's Packleader", "Creature — Werewolf", dnbHuntmaster+"#1", 7, 7)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 4)
	floatMana(t, g, me, "{G}{G}{G}{G}")
	// Not itself, and not a non-Wolf.
	for _, bad := range []uuid.UUID{leader, bear} {
		if err := g.ActivateCatalogAbility(me.ID, leader, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}, {Kind: game.TargetCard, ID: victim}},
		}); err == nil {
			t.Errorf("the fight accepted %s as the fighter", bad)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, leader, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wolf}, {Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, wolf); !ok || c.DamageMarked != 2 {
		t.Errorf("Wolf damage = %d (present %v), want 2", c.DamageMarked, ok)
	}
	if c, ok := battlefieldCard(g, victim); !ok || c.DamageMarked != 3 {
		t.Errorf("Victim damage = %d (present %v), want 3", c.DamageMarked, ok)
	}
}

// Tovolar draws once for each Wolf or Werewolf that connects, not for a
// Bear.
func TestTovolarDrawsForEachWolfOrWerewolfThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Tovolar, Dire Overlord", "Legendary Creature — Human Werewolf", dnbTovolar, 3, 3)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := b12Creature(g, me.ID, "Werewolf", "Creature — Werewolf", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := len(me.Hand.Cards)
	declareAttack(t, g, opp.ID, wolf, were, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("cards drawn = %d, want 2 (Wolf and Werewolf, not the Bear)", got-hand)
	}
}

// dnbAfterASpellTurn plays through to the last seat's turn and has it
// cast one spell, so the CR 502.2 check at the next untap step changes
// nothing and any change to the designation at seat 0's upkeep is the
// card's own.
func dnbAfterASpellTurn(t *testing.T, g *game.Game) {
	t.Helper()
	last := len(g.Seats) - 1
	advanceToMainOf(t, g, last)
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	// The turns on the way here cast nothing and ran the check, so it is
	// night and the werewolves are turned over; make it day again the way
	// the game does, so they turn back, for the turn under test.
	g.WithWriteLock(func() {
		g.DayNight.Designation = game.DesignationNight
		g.BecomeDayForEffect()
	})
	if c := g.CastTallyFor(g.Seats[last].ID); c.Total != 1 {
		t.Fatalf("the last seat has cast %d spells, want 1", c.Total)
	}
}

// At your upkeep with three Wolves and/or Werewolves it becomes night,
// and you may transform Human Werewolves that are not daybound.
func TestTovolarsUpkeepMakesNight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tov := importAndCast(t, g, dnbWerewolf(dnbTovolar).row(), me)
	monSettle(t, g)
	b12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	b12Creature(g, me.ID, "Wolf B", "Creature — Wolf", 2, 2)
	dnbAfterASpellTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	answerAllPrompts(t, g, me.ID)
	if !g.IsNight() {
		t.Fatalf("designation = %q, want night", dnDesignation(g))
	}
	if c, _ := battlefieldCard(g, tov); c.ActiveFace != 1 {
		t.Errorf("Tovolar is on face %d, want its night face", c.ActiveFace)
	}
}

// Two Wolves are not enough.
func TestTovolarsUpkeepNeedsThreeWolvesOrWerewolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	importAndCast(t, g, dnbWerewolf(dnbTovolar).row(), me)
	monSettle(t, g)
	b12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	dnbAfterASpellTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if !g.IsDay() {
		t.Errorf("designation = %q, want day: only two Wolves/Werewolves", dnDesignation(g))
	}
}

// The back face's X pump: +X/+0 and trample to a Wolf or Werewolf.
func TestTovolarTheMidnightScourgePumpsAWolf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tov := b12Push(g, me.ID, "Tovolar, the Midnight Scourge", "Legendary Creature — Werewolf", dnbTovolar+"#1", 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{R}{R}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tov, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("the pump accepted a Bear")
	}
	floatMana(t, g, me, "{R}{R}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, tov, 0, game.ActivateAbilityParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tov}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, tov); got != 6 {
		t.Errorf("power = %d, want 6", got)
	}
	if !slices.Contains(effectiveAbilities(t, g, tov), "trample") {
		t.Error("no trample")
	}
}

// Volatile Arsonist deals its damage to each chosen kind of target.
func TestVolatileArsonistDamagesACreatureAndAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arsonist := b12Push(g, me.ID, "Dire-Strain Anarchist", "Creature — Werewolf", dnbArsonist+"#1", 5, 5)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 4)
	life := opp.Life
	declareAttack(t, g, opp.ID, arsonist)
	// One prompt per clause: the creature, the player, no planeswalker.
	for i, answer := range [][]game.TargetRef{
		{{Kind: game.TargetCard, ID: victim}},
		{{Kind: game.TargetPlayer, ID: opp.ID}},
		nil,
	} {
		pick := latestPickTarget(g, me.ID)
		if pick == nil && answer == nil {
			continue // no planeswalker on the table: no prompt for that clause
		}
		if pick == nil {
			t.Fatalf("no target prompt for clause %d of the attack trigger", i)
		}
		if err := g.ResolvePickTargets(pick.ID, me.ID, answer); err != nil {
			t.Fatalf("clause %d: ResolvePickTargets: %v", i, err)
		}
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, victim); !ok || c.DamageMarked != 2 {
		t.Errorf("Victim damage = %d (present %v), want 2", c.DamageMarked, ok)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-2)
	}
}

// With nothing chosen the trigger does nothing and does not error.
func TestVolatileArsonistMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arsonist := b12Push(g, me.ID, "Volatile Arsonist", "Creature — Human Werewolf", dnbArsonist, 4, 4)
	life := opp.Life
	declareAttack(t, g, opp.ID, arsonist)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("declining every target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life = %d, want %d", opp.Life, life)
	}
}

// Wrathful Jailbreaker has to attack; Weary Prisoner is a defender.
func TestWrathfulJailbreakerAttacksEachCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jail := b12Push(g, me.ID, "Wrathful Jailbreaker", "Creature — Werewolf", dnbPrisoner+"#1", 6, 6)
	var reqs []game.AttackRequirement
	g.ReadSnapshot(func() {
		if c, ok := battlefieldCard(g, jail); ok {
			reqs = c.Effective().AttackRequirements
		}
	})
	if len(reqs) == 0 {
		t.Error("Wrathful Jailbreaker carries no attack requirement")
	}
	prisoner := b12Push(g, me.ID, "Weary Prisoner", "Creature — Human Werewolf", dnbPrisoner, 2, 6)
	if !slices.Contains(effectiveAbilities(t, g, prisoner), "defender") {
		t.Error("Weary Prisoner is not a defender")
	}
}
