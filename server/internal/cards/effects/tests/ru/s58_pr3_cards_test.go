package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func s58LoseLife(g *game.Game, source, player uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(source, player, -n) })
}

func TestS58PR3Registered(t *testing.T) {
	want := map[string]string{
		effectstest.S58ValgavothHarrower: "Valgavoth, Harrower of Souls",
		effectstest.S58BrashTaunter:      "Brash Taunter",
		effectstest.S58DoOrDie:           "Do or Die",
		effectstest.S58HissingMiasma:     "Hissing Miasma",
		effectstest.S58MaddeningHex:      "Maddening Hex",
		effectstest.S58Mai:               "Mai, Scornful Striker",
		effectstest.S58Mogis:             "Mogis, God of Slaughter",
		effectstest.S58Smoke:             "Smoke",
		effectstest.S58SootImp:           "Soot Imp",
		effectstest.S58UnstoppableSlash:  "Unstoppable Slasher",
		effectstest.S58PainForAll:        "Pain for All",
		effectstest.S58BurningAnger:      "Burning Anger",
		effectstest.S58ChainReaction:     "Chain Reaction",
		effectstest.S58RequiemMonolith:   "Requiem Monolith",
		effectstest.S58Simulacrum:        "Simulacrum",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v name=%q", name, ok, spec.Name)
		}
	}
	// #3172 cav-d: both read the total of simultaneous damage now, so
	// neither carries the one-event-per-source caveat any more.
	for _, oracle := range []string{effectstest.S58BrashTaunter, effectstest.S58PainForAll} {
		spec, _ := effects.Lookup(oracle)
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want full, got %v / %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// Valgavoth: the first life an opponent loses on THEIR turn grows it
// and draws; a second loss that turn, or a loss on someone else's turn,
// does not.
func TestS58ValgavothHarrower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	val := effectstest.B27Push(g, me.ID, "Valgavoth, Harrower of Souls", "Legendary Creature — Elder Demon", effectstest.S58ValgavothHarrower, "{2}{B}{R}", 4, 4, "B", "R")
	effectstest.AssertKeywords(t, g, val, "flying")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	hand := me.Hand.Size()
	s58LoseLife(g, val, third.ID, 1) // not their turn
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, val, game.CounterPlusOne) != 0 || me.Hand.Size() != hand {
		t.Fatal("a loss on another player's turn does nothing")
	}
	s58LoseLife(g, val, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, val, game.CounterPlusOne) != 1 || me.Hand.Size() != hand+1 {
		t.Fatalf("first loss on their turn: counters %d hand %d→%d", effectstest.CounterCount(g, val, game.CounterPlusOne), hand, me.Hand.Size())
	}
	s58LoseLife(g, val, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, val, game.CounterPlusOne) != 1 || me.Hand.Size() != hand+1 {
		t.Error("a second loss the same turn does nothing")
	}
}

// Hissing Miasma: each creature attacking its controller's victim, and
// only attacks aimed at the enchantment's controller, costs 1 life.
func TestS58HissingMiasma(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Hissing Miasma", "Enchantment", effectstest.S58HissingMiasma, "{1}{B}{B}", 0, 0, "B")
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	life := opp.Life
	effectstest.DeclareAttack(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("two attackers: attacker's controller at %d, want %d", opp.Life, life-2)
	}
}

func TestS58HissingMiasmaIgnoresAttacksOnOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B27Push(g, me.ID, "Hissing Miasma", "Enchantment", effectstest.S58HissingMiasma, "{1}{B}{B}", 0, 0, "B")
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	life := opp.Life
	effectstest.DeclareAttack(t, g, third.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("an attack on somebody else cost %d life", life-opp.Life)
	}
}

func s58CastFromHand(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// Mai: any player's noncreature spell costs that player 2 life.
func TestS58MaiScornfulStriker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mai := effectstest.B27Push(g, me.ID, "Mai, Scornful Striker", "Legendary Creature — Human Noble Ally", effectstest.S58Mai, "{1}{B}", 2, 2, "B")
	effectstest.AssertKeywords(t, g, mai, "first strike")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	imp := effectstest.B27Push(g, me.ID, "Soot Imp", "Creature — Imp", effectstest.S58SootImp, "{1}{B}{B}", 1, 2, "B")
	effectstest.AssertKeywords(t, g, imp, "flying")
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Smoke", effectstest.S58Smoke, "Enchantment")
	bear := effectstest.PushTappedForTest(g, owner.ID, "Bear", "", "Creature — Bear")
	wolf := effectstest.PushTappedForTest(g, owner.ID, "Wolf", "", "Creature — Wolf")
	forest := effectstest.PushTappedForTest(g, owner.ID, "Forest", "", "Basic Land — Forest")
	choice := effectstest.AdvanceToUntapChoiceOf(t, g, 1)
	if choice.ChooseMin != 1 || choice.ChooseMax != 1 || len(choice.ChooseCards) != 2 {
		t.Fatalf("Smoke: bounds %d..%d over %d permanents, want 1..1 over the two creatures", choice.ChooseMin, choice.ChooseMax, len(choice.ChooseCards))
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{bear, wolf}); err == nil {
		t.Fatal("two creatures under Smoke was accepted")
	}
	if err := g.ResolveUntapChoice(choice.ID, owner.ID, []uuid.UUID{wolf}); err != nil {
		t.Fatalf("one creature was refused: %v", err)
	}
	if effectstest.TappedForTest(t, g, wolf) || !effectstest.TappedForTest(t, g, bear) {
		t.Error("only the chosen creature untaps")
	}
	if effectstest.TappedForTest(t, g, forest) {
		t.Error("a land is not a creature and untaps as normal")
	}
}

// Mogis: not a creature below seven devotion to black and red, and the
// upkeep trigger is damage unless a creature is sacrificed.
func TestS58MogisDevotion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mogis := effectstest.B27Push(g, me.ID, "Mogis, God of Slaughter", "Legendary Enchantment Creature — God", effectstest.S58Mogis, "{2}{B}{R}", 7, 5, "B", "R")
	if effectstest.B39IsCreature(t, g, mogis) {
		t.Fatal("devotion 2: Mogis isn't a creature")
	}
	effectstest.AssertKeywords(t, g, mogis, "indestructible")
	effectstest.B27Push(g, me.ID, "Black Thing", "Enchantment", "", "{B}{B}{B}", 0, 0, "B")
	effectstest.B27Push(g, me.ID, "Hybrid Thing", "Enchantment", "", "{B/R}{B/R}", 0, 0, "B", "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !effectstest.B39IsCreature(t, g, mogis) {
		t.Fatal("devotion 2 + 3 + 2 = 7: each hybrid symbol of both colours counts once, so Mogis is a creature")
	}
}

func TestS58MogisDevotionBelowSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mogis := effectstest.B27Push(g, me.ID, "Mogis, God of Slaughter", "Legendary Enchantment Creature — God", effectstest.S58Mogis, "{2}{B}{R}", 7, 5, "B", "R")
	effectstest.B27Push(g, me.ID, "Black Thing", "Enchantment", "", "{B}{B}{B}", 0, 0, "B")
	effectstest.B27Push(g, me.ID, "Green Thing", "Enchantment", "", "{G}{G}{G}{R}", 0, 0, "G", "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if effectstest.B39IsCreature(t, g, mogis) {
		t.Fatal("devotion 2 + 3 + 1 = 6: still not a creature")
	}
	effectstest.B27Push(g, me.ID, "Red Thing", "Enchantment", "", "{R}", 0, 0, "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !effectstest.B39IsCreature(t, g, mogis) {
		t.Fatal("devotion 7: Mogis is a creature")
	}
}

func TestS58MogisUpkeepPunisher(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Mogis, God of Slaughter", "Legendary Enchantment Creature — God", effectstest.S58Mogis, "{2}{B}{R}", 7, 5, "B", "R")
	// No creature: just the damage.
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("no creature to sacrifice: life %d, want %d", opp.Life, life-2)
	}
	// A creature: the opponent chooses, and the sacrifice spares them.
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	life = opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("the opponent chooses between damage and a sacrifice: %+v", g.PendingChoices)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	cards := effectstest.LatestChooseCardsFor(g, opp.ID)
	if cards == nil {
		t.Fatal("they choose which creature to sacrifice")
	}
	if err := g.ResolveChooseCards(cards.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Life != life {
		t.Errorf("sacrificed: bear on battlefield %v, life %d → %d", g.Battlefield.Contains(bear), life, opp.Life)
	}
}

// Unstoppable Slasher: halves the damaged player's life, and returns
// once, stunned, only when it died with no counters.
func TestS58UnstoppableSlasherHalvesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slasher := effectstest.B27Push(g, me.ID, "Unstoppable Slasher", "Creature — Zombie Assassin", effectstest.S58UnstoppableSlash, "{2}{B}", 2, 3, "B")
	effectstest.AssertKeywords(t, g, slasher, "deathtouch")
	effectstest.AttackWith(t, g, opp.ID, slasher)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 19 {
		t.Errorf("40 - 2 combat damage = 38, then half rounded up (19) lost: want 19, got %d", opp.Life)
	}
}

func TestS58UnstoppableSlasherReturnsOnceWithStunCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slasher := effectstest.B27Push(g, me.ID, "Unstoppable Slasher", "Creature — Zombie Assassin", effectstest.S58UnstoppableSlash, "{2}{B}", 2, 3, "B")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	effectstest.B27Damage(g, foe, slasher, 5)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(slasher) {
		t.Fatal("it returns to the battlefield")
	}
	if n := effectstest.CounterCount(g, slasher, game.CounterStun); n != 2 {
		t.Errorf("returns with %d stun counters, want 2", n)
	}
	if !effectstest.B16Tapped(t, g, slasher) {
		t.Error("it returns tapped")
	}
	effectstest.B27Damage(g, foe, slasher, 5)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(slasher) {
		t.Error("it had counters this time: it stays dead")
	}
}

// CR 400.7 (#2156): the return only finds the object that died. A
// Slasher exiled and put back into the graveyard in response is a new
// object and stays there.
func TestS58UnstoppableSlasherReturnsNothingOnceTheCardChangedZones(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slasher := effectstest.B27Push(g, me.ID, "Unstoppable Slasher", "Creature — Zombie Assassin", effectstest.S58UnstoppableSlash, "{2}{B}", 2, 3, "B")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	effectstest.B27Damage(g, foe, slasher, 5)
	g.RunStateChecksForTest()
	yard := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}
	exile := game.ZoneRef{Kind: game.ZoneExile}
	if err := g.MoveCardByID(yard, exile, slasher); err != nil {
		t.Fatalf("exile from graveyard: %v", err)
	}
	if err := g.MoveCardByID(exile, yard, slasher); err != nil {
		t.Fatalf("put back: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(slasher) {
		t.Error("a Slasher that left the graveyard and came back is a new object and must not return")
	}
}

func TestS58UnstoppableSlasherWithACounterStaysDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	slasher := effectstest.B27Push(g, me.ID, "Unstoppable Slasher", "Creature — Zombie Assassin", effectstest.S58UnstoppableSlash, "{2}{B}", 2, 3, "B")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(slasher, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.B27Damage(g, foe, slasher, 9)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(slasher) {
		t.Error("a +1/+1 counter on it means it does not return")
	}
}

// Brash Taunter: damage to it goes to a chosen opponent, and it fights.
func TestS58BrashTaunter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	taunter := effectstest.B27Push(g, me.ID, "Brash Taunter", "Creature — Goblin", effectstest.S58BrashTaunter, "{4}{R}", 1, 1, "R")
	effectstest.AssertKeywords(t, g, taunter, "indestructible")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.B27Damage(g, foe, taunter, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("3 damage reflected at the opponent: %d → %d", life, opp.Life)
	}
	if !g.Battlefield.Contains(taunter) {
		t.Error("indestructible: it survives the damage")
	}
	// The fight: it deals 1, takes 2, and reflects the 2.
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R")
	life = opp.Life
	if err := g.ActivateCatalogAbility(me.ID, taunter, 0, game.ActivateAbilityParams{Targets: effectstest.DgCardRef(foe)}); err != nil {
		t.Fatalf("fight: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, foe); got != 1 {
		t.Errorf("the foe took %d from the fight, want 1", got)
	}
	if opp.Life != life-2 {
		t.Errorf("the 2 damage it took in the fight is reflected: %d → %d", life, opp.Life)
	}
}

// Pain for All: the enters bite goes to any other target; damage to the
// enchanted creature then goes to each opponent.
func TestS58PainForAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 5)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 4)
	effectstest.CastCatalogSpell(t, g, "Pain for All", "Enchantment — Aura", effectstest.S58PainForAll, effectstest.DgCardRef(host))
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("the enters trigger asks for a target: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: host}); err == nil {
		t.Fatal("the enchanted creature is not a legal \"other\" target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: foe}); err != nil {
		t.Fatalf("pick the foe: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, foe); got != 3 {
		t.Errorf("the host's power 3 hit the foe: marked %d", got)
	}
	lives := effectstest.B27Lives(g)
	effectstest.B27Damage(g, foe, host, 2)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	after := effectstest.B27Lives(g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Ogre", 3, 3)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Burning Anger", "Enchantment — Aura", effectstest.S58BurningAnger, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
	effectstest.AdvanceToMain(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the enchanted creature has the granted ability")
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-3 {
		t.Errorf("power 3 to the target: %d → %d", life, opp.Life)
	}
	if !effectstest.B16Tapped(t, g, host) {
		t.Error("it taps")
	}
}

// Chain Reaction: X is the number of creatures, counted once.
func TestS58ChainReaction(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 3)
	other := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Chain Reaction", "Sorcery", effectstest.S58ChainReaction, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, other} {
		if g.Battlefield.Contains(id) {
			t.Error("three creatures: 3 damage kills a 1/3")
		}
	}
	if !g.Battlefield.Contains(big) || effectstest.DamageMarkedOn(g, big) != 3 {
		t.Errorf("the 4/4 survives with 3 marked, got %d", effectstest.DamageMarkedOn(g, big))
	}
}

// Requiem Monolith: the creature gains the draw-and-lose trigger; its
// controller may take the 1 damage.
func TestS58RequiemMonolith(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mono := effectstest.PushCatalogPermanent(g, me.ID, "Requiem Monolith", "Artifact", effectstest.S58RequiemMonolith, false)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 5)
	effectstest.AdvanceToMain(t, g)
	hand, life := opp.Hand.Size(), opp.Life
	if err := g.ActivateCatalogAbility(me.ID, mono, 0, game.ActivateAbilityParams{Targets: effectstest.DgCardRef(foe)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if ask == nil {
		t.Fatalf("the creature's controller is asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, foe); got != 1 {
		t.Errorf("the Monolith dealt %d, want 1", got)
	}
	if opp.Hand.Size() != hand+1 || opp.Life != life-1 {
		t.Errorf("its controller draws and loses 1: hand %d → %d, life %d → %d", hand, opp.Hand.Size(), life, opp.Life)
	}
	// The grant lasts the turn: more damage, more cards.
	effectstest.B27Damage(g, foe, foe, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+3 || opp.Life != life-3 {
		t.Errorf("2 more damage: hand %d, life %d, want %d and %d", opp.Hand.Size(), opp.Life, hand+3, life-3)
	}
}

func TestS58RequiemMonolithCanBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mono := effectstest.PushCatalogPermanent(g, me.ID, "Requiem Monolith", "Artifact", effectstest.S58RequiemMonolith, false)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 5)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, mono, 0, game.ActivateAbilityParams{Targets: effectstest.DgCardRef(foe)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if ask == nil {
		t.Fatal("asked")
	}
	if err := g.ResolveConfirm(ask.ID, opp.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, foe) != 0 {
		t.Error("declined: no damage")
	}
}

// Simulacrum: gain life and deal damage equal to the damage dealt to
// you this turn.
func TestS58Simulacrum(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 9)
	src := effectstest.B12Creature(g, opp.ID, "Src", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 3); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToPlayerForEffect(src, me.ID, 2); err != nil {
			t.Fatal(err)
		}
	})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Simulacrum", "Instant", effectstest.S58Simulacrum, effectstest.DgCardRef(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("gained %d, want 5", me.Life-life)
	}
	if got := effectstest.DamageMarkedOn(g, mine); got != 5 {
		t.Errorf("the creature took %d, want 5", got)
	}
}

// Do or Die: the caster separates, the target player picks the pile
// that is destroyed.
func TestS58DoOrDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Do or Die", "Sorcery", effectstest.S58DoOrDie, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("the caster separates the piles: %+v", g.PendingChoices)
	}
	if len(split.ChooseCards) != 3 {
		t.Fatalf("only the target's three creatures are separated, got %d", len(split.ChooseCards))
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, []uuid.UUID{a}); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("the target player chooses a pile: %+v", g.PendingChoices)
	}
	// Pile 2 (B and C) is destroyed.
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hex := effectstest.PushCatalogPermanent(g, me.ID, "Maddening Hex", "Enchantment — Aura Curse", effectstest.S58MaddeningHex, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, hex).AttachedTo = game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	life := opp.Life
	s58CastFromHand(t, g, opp, "Shock", "Instant", "{R}")
	dealt := life - opp.Life
	if dealt < 1 || dealt > 6 {
		t.Fatalf("a d6 of damage, got %d", dealt)
	}
	at := effectstest.FindBattlefieldCardForTest(g, hex).AttachedTo
	if at.Kind != game.TargetPlayer || at.ID == opp.ID || at.ID == me.ID {
		t.Errorf("the Curse moves to another one of the controller's opponents, now on %+v", at)
	}
}
