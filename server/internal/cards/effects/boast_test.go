package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// boast_test.go — #2697, CR 702.142. The engine half is game/boast.go;
// the enumerator's and the wire's halves are in internal/legal and
// internal/protocol. Everything here is a real printed card.
//
// Fearless Liberator is the smallest boast card (a mana cost and a
// token), Birgi is the limit modifier, Frenzied Raider watches the
// activation, Dragonkin Berserker prices it, Broadside Bombardiers
// reads the sacrificed permanent.

const (
	fearlessLiberatorOracle  = "b0b93253-6432-401e-9a01-94c41fb72c10"
	frenziedRaiderOracle     = "1207881c-4c03-48e1-9ca7-a2330e8053d1"
	dragonkinBerserkerOracle = "06b77a80-4968-4385-960a-fb66bb4faa94"
	broadsideBombardiersOrcl = "e0c00a74-bdf8-4fee-8042-42c5878e9e3c"
	varragothOracle          = "68b41a04-8cb0-4edf-b488-a219494453ae"
	eradicatorValkyrieOracle = "06b7987d-c876-4369-9c2d-ce1d35a01097"
	tuskeriFirewalkerOracle  = "36dda03b-a3c1-4ef4-ae62-da318028a39e"
)

// boastPool pays one Liberator boast: {2}{R}.
func boastPool(p *game.Player) {
	fillPool(p, 2)
	fillPoolColored(p, "R", 1)
}

func pushLiberator(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Fearless Liberator",
		"Creature — Dwarf Berserker", fearlessLiberatorOracle, false)
}

func dwarfTokens(g *game.Game, controller uuid.UUID) int {
	n, _ := b14Tokens(g, controller, "Dwarf Berserker")
	return n
}

// TestBoastRefusedUntilTheCreatureAttacks is "Activate only if this
// creature attacked this turn": before the attack the activation is
// refused with its own error and nothing is paid.
func TestBoastRefusedUntilTheCreatureAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := pushLiberator(g, me.ID)
	advanceToMain(t, g)

	boastPool(me)
	err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrBoastNotAttacked) {
		t.Fatalf("before attacking: %v, want ErrBoastNotAttacked", err)
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool holds %d mana, want 3 — a refused activation pays nothing", len(me.ManaPool))
	}

	attackWith(t, g, opp.ID, dwarf)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("after attacking: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := dwarfTokens(g, me.ID); got != 1 {
		t.Errorf("Dwarf Berserker tokens = %d, want 1", got)
	}
}

// TestBoastIsOncePerTurnAndSpentAtTheAnnounce is "only once each turn":
// the second activation is refused whether the first has resolved or is
// still on the stack, because the activation is spent when it is
// announced (CR 602.2), not when it resolves.
func TestBoastIsOncePerTurnAndSpentAtTheAnnounce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := pushLiberator(g, me.ID)
	attackWith(t, g, opp.ID, dwarf)

	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first boast: %v", err)
	}
	// Still on the stack: not resolved, and already spent.
	boastPool(me)
	err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrBoastSpent) {
		t.Fatalf("second boast with the first on the stack: %v, want ErrBoastSpent", err)
	}
	passPriorityAroundTable(t, g)
	err = g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrBoastSpent) {
		t.Fatalf("second boast after the first resolved: %v, want ErrBoastSpent", err)
	}
	if got := dwarfTokens(g, me.ID); got != 1 {
		t.Errorf("tokens = %d, want 1", got)
	}
}

// TestBoastIsPerCreature: one creature spending its boast does not
// spend another's, and a creature that did not attack cannot boast
// however many others did.
func TestBoastIsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushLiberator(g, me.ID)
	b := pushLiberator(g, me.ID)
	stayHome := pushLiberator(g, me.ID)
	attackWith(t, g, opp.ID, a, b)

	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, a, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first creature: %v", err)
	}
	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, b, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("second creature: %v", err)
	}
	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, stayHome, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrBoastNotAttacked) {
		t.Fatalf("a creature that did not attack: %v, want ErrBoastNotAttacked", err)
	}
}

// TestBirgiLetsEveryCreatureBoastTwice is CR 702.142 as Birgi modifies
// it: a limit of two, not a free extra activation.
func TestBirgiLetsEveryCreatureBoastTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := pushLiberator(g, me.ID)
	birgi := pushCatalogPermanent(g, me.ID, "Birgi, God of Storytelling",
		"Legendary Creature — God", birgiOracle, false)
	attackWith(t, g, opp.ID, dwarf)

	for i := 0; i < 2; i++ {
		boastPool(me)
		if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("boast %d under Birgi: %v", i+1, err)
		}
		passPriorityAroundTable(t, g)
	}
	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrBoastSpent) {
		t.Fatalf("third boast: %v, want ErrBoastSpent", err)
	}
	if got := dwarfTokens(g, me.ID); got != 2 {
		t.Errorf("tokens = %d, want 2", got)
	}

	// Birgi is a creature too, and not exempt from the attack half: she
	// has not attacked, so she cannot boast.
	if limit := boastLimit(g, birgi); limit != 2 {
		t.Errorf("Birgi's own limit = %d, want 2", limit)
	}
}

// TestBirgiDoesNotStackAndLeavesWithHerLimit: two Birgis say "twice
// rather than once" and make two, not three; and with Birgi gone a
// creature that has boasted once is simply spent.
func TestBirgiDoesNotStackAndLeavesWithHerLimit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := pushLiberator(g, me.ID)
	one := pushCatalogPermanent(g, me.ID, "Birgi, God of Storytelling",
		"Legendary Creature — God", birgiOracle, false)
	// A second source of the same static, not legendary so the legend
	// rule leaves it alone (a copy of Birgi would be the printed way).
	pushCatalogPermanent(g, me.ID, "Birgi, Second Source",
		"Creature — God", birgiOracle, false)
	if limit := boastLimit(g, dwarf); limit != 2 {
		t.Fatalf("two Birgis: limit = %d, want 2 — they replace the number, they do not add", limit)
	}
	attackWith(t, g, opp.ID, dwarf)

	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first boast: %v", err)
	}
	passPriorityAroundTable(t, g)

	// One Birgi leaves; the other still counts.
	b14Kill(g, one)
	if limit := boastLimit(g, dwarf); limit != 2 {
		t.Fatalf("one Birgi left: limit = %d, want 2", limit)
	}
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == birgiOracle {
			b14Kill(g, c.InstanceID)
			break
		}
	}
	if limit := boastLimit(g, dwarf); limit != 1 {
		t.Fatalf("no Birgi: limit = %d, want 1", limit)
	}
	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrBoastSpent) {
		t.Fatalf("second boast with no Birgi: %v, want ErrBoastSpent", err)
	}
}

// TestBirgiIsYoursAndOnYourTurnOnly: "Creatures YOU control … during
// each of YOUR turns". An opponent's Birgi does nothing for my
// creatures, and mine does nothing for theirs.
func TestBirgiIsYoursAndOnYourTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushLiberator(g, me.ID)
	theirs := pushLiberator(g, opp.ID)
	pushCatalogPermanent(g, opp.ID, "Birgi, God of Storytelling",
		"Legendary Creature — God", birgiOracle, false)
	advanceToMain(t, g)

	if limit := boastLimit(g, mine); limit != 1 {
		t.Errorf("my creature under their Birgi: limit = %d, want 1", limit)
	}
	// It is MY turn, so their Birgi's controller is not the active
	// player and their creature gets nothing either.
	if limit := boastLimit(g, theirs); limit != 1 {
		t.Errorf("their creature during my turn: limit = %d, want 1 — it is not their turn", limit)
	}
}

// boastLimit reads game.Game.BoastLimitFor for a battlefield card.
func boastLimit(g *game.Game, id uuid.UUID) int {
	var limit int
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				limit = g.BoastLimitFor(c)
				return
			}
		}
	})
	return limit
}

// TestFrenziedRaiderWatchesAnyBoastYouActivate: the trigger reads the
// bit the announcement stamped, so another creature's boast counts and
// a non-boast activation does not.
func TestFrenziedRaiderWatchesAnyBoastYouActivate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dwarf := pushLiberator(g, me.ID)
	raider := pushCatalogPermanent(g, me.ID, "Frenzied Raider",
		"Creature — Demon Berserker", frenziedRaiderOracle, false)
	attackWith(t, g, opp.ID, dwarf)

	boastPool(me)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, raider, game.CounterPlusOne); got != 1 {
		t.Errorf("Frenzied Raider +1/+1 counters = %d, want 1", got)
	}
}

// TestDragonkinBerserkerDiscountsBoastPerDragon: "{1} less for each
// Dragon you control", on boast abilities only, and only for the
// controller. Its own {4}{R} boast creates the Dragon, so the second
// activation is not available (once per turn) — the discount is read
// through the price the engine charges.
func TestDragonkinBerserkerDiscountsBoastPerDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragonkin := pushCatalogPermanent(g, me.ID, "Dragonkin Berserker",
		"Creature — Human Berserker", dragonkinBerserkerOracle, false)
	dwarf := pushLiberator(g, me.ID)
	for i := 0; i < 2; i++ {
		dragon := TokenCard("5/5 red Dragon with flying")
		dragon.Owner, dragon.Controller, dragon.InstanceID = me.ID, me.ID, uuid.New()
		g.Battlefield.PushTop(dragon)
	}
	attackWith(t, g, opp.ID, dragonkin, dwarf)

	// Liberator's boast is {2}{R}; two Dragons take {2} off the generic.
	fillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, dwarf, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("a {2}{R} boast with two Dragons should cost {R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := dwarfTokens(g, me.ID); got != 1 {
		t.Errorf("tokens = %d, want 1", got)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool holds %d mana, want 0: %+v", len(me.ManaPool), me.ManaPool)
	}
}

// TestBroadsideBombardiersDamageReadsTheSacrificedManaValue: 2 plus the
// sacrificed permanent's mana value, to any target.
func TestBroadsideBombardiersDamageReadsTheSacrificedManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bombardiers := pushCatalogPermanent(g, me.ID, "Broadside Bombardiers",
		"Creature — Goblin Pirate", broadsideBombardiersOrcl, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Artifact", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fodder {
				g.Battlefield.Cards[i].ManaCost = "{3}"
			}
		}
	})
	attackWith(t, g, opp.ID, bombardiers)
	before := opp.Life

	err := g.ActivateCatalogAbility(me.ID, bombardiers, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{fodder},
	})
	if err != nil {
		t.Fatalf("boast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 5 {
		t.Errorf("damage = %d, want 5 (2 plus a mana value 3 artifact)", got)
	}
}

// TestEveryBoastCardIsMarkedAndBlockedBeforeAttack: each card this slice
// ships declares the bit through the constructor, and none can be
// activated by a creature that has not attacked.
func TestEveryBoastCardIsMarkedAndBlockedBeforeAttack(t *testing.T) {
	oracles := map[string]int{
		varragothOracle:                        0,
		broadsideBombardiersOrcl:               0,
		eradicatorValkyrieOracle:               0,
		dragonkinBerserkerOracle:               0,
		fearlessLiberatorOracle:                0,
		tuskeriFirewalkerOracle:                0,
		"a6eb06dc-a62d-4fdb-b336-e304d8d68c92": 0, // Usher of the Fallen
		"61295adf-9a58-479a-b6e5-403aa0876c22": 0, // Fearless Pup
		"b9227f38-7db7-4ea1-9482-94e5de3984fe": 0, // Duskwielder
		"f854ea0d-aae3-4b6a-9767-6ac1788f2190": 0, // Draugr Recruiter
		"e0610ab1-88d3-4502-84b7-3a08d0170167": 0, // Horizon Seeker
		"d0d4ab83-8b9b-49cd-86b8-720abd0550f4": 0, // Axgard Braggart
		"f8e8cdc6-5685-4986-a14b-20bb9e31d9cf": 0, // Battershield Warrior
	}
	for oracle := range oracles {
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: oracle})
		if len(abilities) != 1 || !abilities[0].Boast {
			t.Errorf("%s: want exactly one ability and it a boast ability, got %d", oracle, len(abilities))
			continue
		}
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := pushCatalogPermanent(g, me.ID, "Boaster", "Creature — Test", oracle, false)
		advanceToMain(t, g)
		fillPool(me, 8)
		fillPoolColored(me, "R", 2)
		fillPoolColored(me, "B", 2)
		fillPoolColored(me, "G", 2)
		fillPoolColored(me, "W", 2)
		err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{})
		if !errors.Is(err, game.ErrBoastNotAttacked) {
			t.Errorf("%s: activating before attacking: %v, want ErrBoastNotAttacked", oracle, err)
		}
	}
}

// TestBoastRegisterRefusesAMismatchedLabel is the boot contract: the
// printed keyword and the bit agree in both directions.
func TestBoastRegisterRefusesAMismatchedLabel(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	cases := map[string]ActivatedAbility{
		"bit without the keyword": {Label: "{1}: Do a thing.", Boast: true, Cost: ManaCost("{1}"), Effect: noop},
		"keyword without the bit": {Label: "Boast — {1}: Do a thing.", Cost: ManaCost("{1}"), Effect: noop},
	}
	for name, ab := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted it", name)
				}
			}()
			checkBoast(Spec{Name: "Test " + name, Activated: []ActivatedAbility{ab}})
		}()
	}
	// The constructor itself passes.
	checkBoast(Spec{Name: "Test ok", Activated: []ActivatedAbility{Boast("{1}: Do a thing.", ManaCost("{1}"), noop)}})
	// A limit below two changes nothing and is refused.
	func() {
		defer func() {
			if recover() == nil {
				t.Error("a boast limit of 1 was accepted")
			}
		}()
		checkBoast(Spec{Name: "Test limit", BoastLimits: []game.BoastLimit{YourCreaturesBoastTimes("x", 1)}})
	}()
}
