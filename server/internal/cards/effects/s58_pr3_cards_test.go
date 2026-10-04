package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_pr3_cards_test.go — S58 PR 3, "Valgavoth: Rakdos punishers and
// damage": sixteen cards through the real catalog.

const (
	s58ValgavothHarrower = "5401439c-4068-4490-8d53-8ffd0026dd7d"
	s58BrashTaunter      = "3e41648f-c5a9-4b26-b97e-8176fe5e9c85"
	s58DoOrDie           = "f25af42d-f1bf-4bd3-aded-ecadafdbf6e6"
	s58HissingMiasma     = "e257d8e0-06e9-433d-a750-1962db399388"
	s58MaddeningHex      = "5d96170f-14dd-496e-a485-bd3b9eedfe02"
	s58Mai               = "953a2bd3-5bca-41fc-8785-66b2d7fa381a"
	s58Mogis             = "e5912d21-e188-4b9e-8e92-a6e8196353c6"
	s58Smoke             = "8aa97d25-cd51-4ceb-b7eb-af64f0914a8c"
	s58SootImp           = "c867e31a-90ea-4d20-80db-3977ec68b16b"
	s58UnstoppableSlash  = "6662968e-ca9e-454c-be44-24ad10e43d3e"
	s58PainForAll        = "4eedf21c-0ad1-48da-a9fe-2ffdbf8371e9"
	s58BurningAnger      = "a2c76ac2-5ab1-450c-b64e-e9a6c4dad1ca"
	s58ChainReaction     = "086b2564-9114-4ba2-94fd-b490f98f38a7"
	s58RequiemMonolith   = "13a8b19b-23bd-4e45-81ff-23371665e4fc"
	s58Simulacrum        = "20d69989-7250-40c7-a064-8ed78ccbe556"
)

func s58LoseLife(g *game.Game, source, player uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(source, player, -n) })
}

func TestS58PR3Registered(t *testing.T) {
	want := map[string]string{
		s58ValgavothHarrower: "Valgavoth, Harrower of Souls",
		s58BrashTaunter:      "Brash Taunter",
		s58DoOrDie:           "Do or Die",
		s58HissingMiasma:     "Hissing Miasma",
		s58MaddeningHex:      "Maddening Hex",
		s58Mai:               "Mai, Scornful Striker",
		s58Mogis:             "Mogis, God of Slaughter",
		s58Smoke:             "Smoke",
		s58SootImp:           "Soot Imp",
		s58UnstoppableSlash:  "Unstoppable Slasher",
		s58PainForAll:        "Pain for All",
		s58BurningAnger:      "Burning Anger",
		s58ChainReaction:     "Chain Reaction",
		s58RequiemMonolith:   "Requiem Monolith",
		s58Simulacrum:        "Simulacrum",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v name=%q", name, ok, spec.Name)
		}
	}
	nemesis, _ := Lookup("fcb7c93c-46ab-49b5-a6e0-35d73f3be8f0")
	for _, oracle := range []string{s58BrashTaunter, s58PainForAll} {
		spec, _ := Lookup(oracle)
		if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
			t.Errorf("%s: want one caveat, got %v / %v", spec.Name, spec.Completeness, spec.Caveats)
		}
		if oracle == s58BrashTaunter && spec.Caveats[0] != nemesis.Caveats[0] {
			t.Errorf("Brash Taunter's caveat should be Screaming Nemesis's: %q", spec.Caveats[0])
		}
	}
}

// Valgavoth: the first life an opponent loses on THEIR turn grows it
// and draws; a second loss that turn, or a loss on someone else's turn,
// does not.
func TestS58ValgavothHarrower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	val := b27Push(g, me.ID, "Valgavoth, Harrower of Souls", "Legendary Creature — Elder Demon", s58ValgavothHarrower, "{2}{B}{R}", 4, 4, "B", "R")
	assertKeywords(t, g, val, "flying")
	advanceToUpkeepOf(t, g, 1)
	hand := me.Hand.Size()
	s58LoseLife(g, val, third.ID, 1) // not their turn
	passPriorityAroundTable(t, g)
	if counterCount(g, val, game.CounterPlusOne) != 0 || me.Hand.Size() != hand {
		t.Fatal("a loss on another player's turn does nothing")
	}
	s58LoseLife(g, val, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if counterCount(g, val, game.CounterPlusOne) != 1 || me.Hand.Size() != hand+1 {
		t.Fatalf("first loss on their turn: counters %d hand %d→%d", counterCount(g, val, game.CounterPlusOne), hand, me.Hand.Size())
	}
	s58LoseLife(g, val, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if counterCount(g, val, game.CounterPlusOne) != 1 || me.Hand.Size() != hand+1 {
		t.Error("a second loss the same turn does nothing")
	}
}

// Hissing Miasma: each creature attacking its controller's victim, and
// only attacks aimed at the enchantment's controller, costs 1 life.
func TestS58HissingMiasma(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Hissing Miasma", "Enchantment", s58HissingMiasma, "{1}{B}{B}", 0, 0, "B")
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	life := opp.Life
	declareAttack(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("two attackers: attacker's controller at %d, want %d", opp.Life, life-2)
	}
}

func TestS58HissingMiasmaIgnoresAttacksOnOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	b27Push(g, me.ID, "Hissing Miasma", "Enchantment", s58HissingMiasma, "{1}{B}{B}", 0, 0, "B")
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	life := opp.Life
	declareAttack(t, g, third.ID, a)
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("an attack on somebody else cost %d life", life-opp.Life)
	}
}

func s58CastFromHand(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

// Mai: any player's noncreature spell costs that player 2 life.
func TestS58MaiScornfulStriker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mai := b27Push(g, me.ID, "Mai, Scornful Striker", "Legendary Creature — Human Noble Ally", s58Mai, "{1}{B}", 2, 2, "B")
	assertKeywords(t, g, mai, "first strike")
	life := me.Life
	s58CastFromHand(t, g, me, "Bear", "Creature — Bear", "{1}{G}")
	if me.Life != life {
		t.Errorf("a creature spell cost %d life", life-me.Life)
	}
	s58CastFromHand(t, g, me, "Shock", "Instant", "{R}")
	if me.Life != life-2 {
		t.Errorf("a noncreature spell: life %d, want %d", me.Life, life-2)
	}
}

// Soot Imp: a nonblack spell, colourless included, costs its caster 1.
func TestS58SootImp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	imp := b27Push(g, me.ID, "Soot Imp", "Creature — Imp", s58SootImp, "{1}{B}{B}", 1, 2, "B")
	assertKeywords(t, g, imp, "flying")
	life := me.Life
	s58CastFromHand(t, g, me, "Dark Spell", "Sorcery", "{1}{B}")
	if me.Life != life {
		t.Errorf("a black spell cost %d life", life-me.Life)
	}
	s58CastFromHand(t, g, me, "Red Spell", "Sorcery", "{1}{R}")
	if me.Life != life-1 {
		t.Errorf("a red spell: life %d, want %d", me.Life, life-1)
	}
	s58CastFromHand(t, g, me, "Rock", "Artifact", "{2}")
	if me.Life != life-2 {
		t.Errorf("a colourless spell: life %d, want %d", me.Life, life-2)
	}
	s58CastFromHand(t, g, me, "Gold Spell", "Sorcery", "{B}{R}")
	if me.Life != life-2 {
		t.Errorf("a black-and-red spell cost life: %d", me.Life)
	}
}

// Smoke: one creature untaps; a tapped land and artifact are not
// counted.
func TestS58SmokeCapsCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Smoke", s58Smoke, "Enchantment")
	bear := pushTappedForTest(g, owner.ID, "Bear", "", "Creature — Bear")
	wolf := pushTappedForTest(g, owner.ID, "Wolf", "", "Creature — Wolf")
	forest := pushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	choice := advanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 || len(choice.ChooseCards) != 2 {
		t.Fatalf("Smoke: bounds %d..%d over %d permanents, want 1..1 over the two creatures", choice.ChooseMin, choice.ChooseMax, len(choice.ChooseCards))
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{bear, wolf}); err == nil {
		t.Fatal("two creatures under Smoke was accepted")
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{wolf}); err != nil {
		t.Fatalf("one creature was refused: %v", err)
	}
	if tappedForTest(t, g, wolf) || !tappedForTest(t, g, bear) {
		t.Error("only the chosen creature untaps")
	}
	if tappedForTest(t, g, forest) {
		t.Error("a land is not a creature and untaps as normal")
	}
}

// Mogis: not a creature below seven devotion to black and red, and the
// upkeep trigger is damage unless a creature is sacrificed.
func TestS58MogisDevotion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mogis := b27Push(g, me.ID, "Mogis, God of Slaughter", "Legendary Enchantment Creature — God", s58Mogis, "{2}{B}{R}", 7, 5, "B", "R")
	if b39IsCreature(t, g, mogis) {
		t.Fatal("devotion 2: Mogis isn't a creature")
	}
	assertKeywords(t, g, mogis, "indestructible")
	b27Push(g, me.ID, "Black Thing", "Enchantment", "", "{B}{B}{B}", 0, 0, "B")
	b27Push(g, me.ID, "Hybrid Thing", "Enchantment", "", "{B/R}{B/R}", 0, 0, "B", "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !b39IsCreature(t, g, mogis) {
		t.Fatal("devotion 2 + 3 + 2 = 7: each hybrid symbol of both colours counts once, so Mogis is a creature")
	}
}

func TestS58MogisDevotionBelowSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mogis := b27Push(g, me.ID, "Mogis, God of Slaughter", "Legendary Enchantment Creature — God", s58Mogis, "{2}{B}{R}", 7, 5, "B", "R")
	b27Push(g, me.ID, "Black Thing", "Enchantment", "", "{B}{B}{B}", 0, 0, "B")
	b27Push(g, me.ID, "Green Thing", "Enchantment", "", "{G}{G}{G}{R}", 0, 0, "G", "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if b39IsCreature(t, g, mogis) {
		t.Fatal("devotion 2 + 3 + 1 = 6: still not a creature")
	}
	b27Push(g, me.ID, "Red Thing", "Enchantment", "", "{R}", 0, 0, "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !b39IsCreature(t, g, mogis) {
		t.Fatal("devotion 7: Mogis is a creature")
	}
}

func TestS58MogisUpkeepPunisher(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Mogis, God of Slaughter", "Legendary Enchantment Creature — God", s58Mogis, "{2}{B}{R}", 7, 5, "B", "R")
	// No creature: just the damage.
	life := opp.Life
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("no creature to sacrifice: life %d, want %d", opp.Life, life-2)
	}
	// A creature: the opponent chooses, and the sacrifice spares them.
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 2)
	advanceToUpkeepOf(t, g, 1)
	life = opp.Life
	passPriorityAroundTable(t, g)
	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("the opponent chooses between damage and a sacrifice: %+v", g.PendingChoices)
	}
	answerOptionPick(t, g, opp.ID, 1)
	cards := latestChooseCardsFor(g, opp.ID)
	if cards == nil {
		t.Fatal("they choose which creature to sacrifice")
	}
	if err := g.ResolveChooseCards(cards.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Life != life {
		t.Errorf("sacrificed: bear on battlefield %v, life %d → %d", g.Battlefield.Contains(bear), life, opp.Life)
	}
}

// Unstoppable Slasher: halves the damaged player's life, and returns
// once, stunned, only when it died with no counters.
func TestS58UnstoppableSlasherHalvesLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slasher := b27Push(g, me.ID, "Unstoppable Slasher", "Creature — Zombie Assassin", s58UnstoppableSlash, "{2}{B}", 2, 3, "B")
	assertKeywords(t, g, slasher, "deathtouch")
	attackWith(t, g, opp.ID, slasher)
	passPriorityAroundTable(t, g)
	if opp.Life != 19 {
		t.Errorf("40 - 2 combat damage = 38, then half rounded up (19) lost: want 19, got %d", opp.Life)
	}
}

func TestS58UnstoppableSlasherReturnsOnceWithStunCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slasher := b27Push(g, me.ID, "Unstoppable Slasher", "Creature — Zombie Assassin", s58UnstoppableSlash, "{2}{B}", 2, 3, "B")
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	b27Damage(g, foe, slasher, 5)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(slasher) {
		t.Fatal("it returns to the battlefield")
	}
	if n := counterCount(g, slasher, game.CounterStun); n != 2 {
		t.Errorf("returns with %d stun counters, want 2", n)
	}
	if !b16Tapped(t, g, slasher) {
		t.Error("it returns tapped")
	}
	b27Damage(g, foe, slasher, 5)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(slasher) {
		t.Error("it had counters this time: it stays dead")
	}
}

func TestS58UnstoppableSlasherWithACounterStaysDead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slasher := b27Push(g, me.ID, "Unstoppable Slasher", "Creature — Zombie Assassin", s58UnstoppableSlash, "{2}{B}", 2, 3, "B")
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(slasher, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	b27Damage(g, foe, slasher, 9)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(slasher) {
		t.Error("a +1/+1 counter on it means it does not return")
	}
}

// Brash Taunter: damage to it goes to a chosen opponent, and it fights.
func TestS58BrashTaunter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taunter := b27Push(g, me.ID, "Brash Taunter", "Creature — Goblin", s58BrashTaunter, "{4}{R}", 1, 1, "R")
	assertKeywords(t, g, taunter, "indestructible")
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	life := opp.Life
	b27Damage(g, foe, taunter, 3)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("3 damage reflected at the opponent: %d → %d", life, opp.Life)
	}
	if !g.Battlefield.Contains(taunter) {
		t.Error("indestructible: it survives the damage")
	}
	// The fight: it deals 1, takes 2, and reflects the 2.
	advanceToMain(t, g)
	b06AddMana(me, "R", "R", "R")
	life = opp.Life
	if err := g.ActivateCatalogAbility(me.ID, taunter, 0, game.ActivateAbilityParams{Targets: dgCardRef(foe)}); err != nil {
		t.Fatalf("fight: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, foe); got != 1 {
		t.Errorf("the foe took %d from the fight, want 1", got)
	}
	if opp.Life != life-2 {
		t.Errorf("the 2 damage it took in the fight is reflected: %d → %d", life, opp.Life)
	}
}

// Pain for All: the enters bite goes to any other target; damage to the
// enchanted creature then goes to each opponent.
func TestS58PainForAll(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 5)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 4)
	castCatalogSpell(t, g, "Pain for All", "Enchantment — Aura", s58PainForAll, dgCardRef(host))
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("the enters trigger asks for a target: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: host}); err == nil {
		t.Fatal("the enchanted creature is not a legal \"other\" target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: foe}); err != nil {
		t.Fatalf("pick the foe: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, foe); got != 3 {
		t.Errorf("the host's power 3 hit the foe: marked %d", got)
	}
	lives := b27Lives(g)
	b27Damage(g, foe, host, 2)
	passPriorityAroundTable(t, g)
	after := b27Lives(g)
	if after[0] != lives[0] {
		t.Errorf("the controller is not hit: %d → %d", lives[0], after[0])
	}
	for i := 1; i < len(after); i++ {
		if after[i] != lives[i]-2 {
			t.Errorf("opponent %d: %d → %d, want -2", i, lives[i], after[i])
		}
	}
}

// Burning Anger: the enchanted creature taps to deal its power to any
// target.
func TestS58BurningAnger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	aura := pushCatalogPermanent(g, me.ID, "Burning Anger", "Enchantment — Aura", s58BurningAnger, false)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
	advanceToMain(t, g)
	idx, ref := grantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature has the granted ability")
	}
	life := opp.Life
	b16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-3 {
		t.Errorf("power 3 to the target: %d → %d", life, opp.Life)
	}
	if !b16Tapped(t, g, host) {
		t.Error("it taps")
	}
}

// Chain Reaction: X is the number of creatures, counted once.
func TestS58ChainReaction(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 3)
	other := b12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Chain Reaction", "Sorcery", s58ChainReaction, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, other} {
		if g.Battlefield.Contains(id) {
			t.Error("three creatures: 3 damage kills a 1/3")
		}
	}
	if !g.Battlefield.Contains(big) || damageMarkedOn(g, big) != 3 {
		t.Errorf("the 4/4 survives with 3 marked, got %d", damageMarkedOn(g, big))
	}
}

// Requiem Monolith: the creature gains the draw-and-lose trigger; its
// controller may take the 1 damage.
func TestS58RequiemMonolith(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mono := pushCatalogPermanent(g, me.ID, "Requiem Monolith", "Artifact", s58RequiemMonolith, false)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 5)
	advanceToMain(t, g)
	hand, life := opp.Hand.Size(), opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mono, 0, game.ActivateAbilityParams{Targets: dgCardRef(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if ask == nil {
		t.Fatalf("the creature's controller is asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, foe); got != 1 {
		t.Errorf("the Monolith dealt %d, want 1", got)
	}
	if opp.Hand.Size() != hand+1 || opp.Life != life-1 {
		t.Errorf("its controller draws and loses 1: hand %d → %d, life %d → %d", hand, opp.Hand.Size(), life, opp.Life)
	}
	// The grant lasts the turn: more damage, more cards.
	b27Damage(g, foe, foe, 2)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+3 || opp.Life != life-3 {
		t.Errorf("2 more damage: hand %d, life %d, want %d and %d", opp.Hand.Size(), opp.Life, hand+3, life-3)
	}
}

func TestS58RequiemMonolithCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mono := pushCatalogPermanent(g, me.ID, "Requiem Monolith", "Artifact", s58RequiemMonolith, false)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 5)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, mono, 0, game.ActivateAbilityParams{Targets: dgCardRef(foe)}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if ask == nil {
		t.Fatal("asked")
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, false); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, foe) != 0 {
		t.Error("declined: no damage")
	}
}

// Simulacrum: gain life and deal damage equal to the damage dealt to
// you this turn.
func TestS58Simulacrum(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 9)
	src := b12Creature(g, opp.ID, "Src", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	life := me.Life
	castCatalogSpell(t, g, "Simulacrum", "Instant", s58Simulacrum, dgCardRef(mine))
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("gained %d, want 5", me.Life-life)
	}
	if got := damageMarkedOn(g, mine); got != 5 {
		t.Errorf("the creature took %d, want 5", got)
	}
}

// Do or Die: the caster separates, the target player picks the pile
// that is destroyed.
func TestS58DoOrDie(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Do or Die", "Sorcery", s58DoOrDie, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	split := latestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("the caster separates the piles: %+v", g.PendingChoices)
	}
	if len(split.ChooseCards) != 3 {
		t.Fatalf("only the target's three creatures are separated, got %d", len(split.ChooseCards))
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatal(err)
	}
	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("the target player chooses a pile: %+v", g.PendingChoices)
	}
	// Pile 2 (B and C) is destroyed.
	answerOptionPick(t, g, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Errorf("pile {B, C} destroyed, {A} kept: a=%v b=%v c=%v", g.Battlefield.Contains(a), g.Battlefield.Contains(b), g.Battlefield.Contains(c))
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the caster's creature is not separated or destroyed")
	}
}

// Maddening Hex: the enchanted player's noncreature spell rolls a d6
// of damage to them, and the Curse moves to another opponent.
func TestS58MaddeningHex(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hex := pushCatalogPermanent(g, me.ID, "Maddening Hex", "Enchantment — Aura Curse", s58MaddeningHex, false)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, hex).AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	})
	advanceToUpkeepOf(t, g, 1)
	life := opp.Life
	s58CastFromHand(t, g, opp, "Shock", "Instant", "{R}")
	dealt := life - opp.Life
	if dealt < 1 || dealt > 6 {
		t.Fatalf("a d6 of damage, got %d", dealt)
	}
	at := findBattlefieldCardForTest(g, hex).AttachedTo
	if at.Kind != game.TargetPlayer || at.ID == opp.ID || at.ID == me.ID {
		t.Errorf("the Curse moves to another one of the controller's opponents, now on %+v", at)
	}
}
