package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b10_cards_test.go — S82 NEO batch 10 (#5493): modified matters I.
// Every test plays the real catalog card through the engine.

const (
	neoB10FlameDischarge = "fa4755c2-e573-45c3-bd6a-61b2b35fbd24"
	neoB10LethalExploit  = "61dfb656-1cdd-4ab3-b533-93718cc66b15"
	neoB10KamisFlare     = "d76fc3aa-3962-4d64-a082-765e09bcdf25"
	neoB10AmbitiousAsslt = "3fb399fa-123f-448e-a473-573a1fab2c78"
	neoB10AkkiEmberKeep  = "8df42f4a-022c-4fcb-a953-fb165fdbfc13"
	neoB10UpriserRen     = "9f9b3ef8-e0ce-4906-96b3-a7727d7cc662"
	neoB10HeirAncientFng = "6bcdbded-46f0-44b9-86ad-edd6414e7ef6"
	neoB10OrochiMerge    = "53e39011-3d77-4883-9c8e-4707f6fc5c47"
	neoB10WalkingSky     = "90c763b9-6173-46ff-b31c-7dcafc774170"
	neoB10GuardiansOb    = "42b28dfb-de4d-4005-b0d7-1db3917cf159"
	neoB10GuideBot       = "28bb9a58-f45a-4452-a1f4-c7609373085f"
	neoB10HotSpring      = "a972630e-651e-439d-9412-5be50da9b934"
	neoB10Raiju          = "3d713c11-36a7-4b19-8c42-a9b61ef5902c"
	neoB10Bear           = "Creature — Bear"
)

// neoB10Modify makes a creature modified with a +1/+1 counter.
func neoB10Modify(g *game.Game, id uuid.UUID) { addCountersForTest(g, id, "+1/+1", 1) }

// neoB10Unmodify takes the counter back off.
func neoB10Unmodify(g *game.Game, id uuid.UUID) {
	setCounters(g, id, map[string]int{})
}

func neoB10Alive(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

func TestNeoB10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		neoB10FlameDischarge: "Flame Discharge", neoB10LethalExploit: "Lethal Exploit",
		neoB10KamisFlare: "Kami's Flare", neoB10AmbitiousAsslt: "Ambitious Assault",
		neoB10AkkiEmberKeep: "Akki Ember-Keeper", neoB10UpriserRen: "Upriser Renegade",
		neoB10HeirAncientFng: "Heir of the Ancient Fang", neoB10OrochiMerge: "Orochi Merge-Keeper",
		neoB10WalkingSky: "Walking Skyscraper", neoB10GuardiansOb: "Guardians of Oboro",
		neoB10GuideBot: "Towashi Guide-Bot", neoB10HotSpring: "Invigorating Hot Spring",
		neoB10Raiju: "Thundering Raiju",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestNeoB10FlameDischargeAddsTwoOnlyWithAModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	plain := b12Creature(g, opp.ID, "Plain", neoB10Bear, 2, 4)
	paCast(t, g, "Flame Discharge", "Instant", "{X}{R}", neoB10FlameDischarge, game.CastSpellParams{XValue: 2, Targets: b16TargetCard(plain)})
	passPriorityAroundTable(t, g)
	if !neoB10Alive(g, plain) {
		t.Error("X=2 with no modified creature should not kill a 4-toughness creature")
	}
	mine := b12Creature(g, me.ID, "Mine", neoB10Bear, 2, 2)
	neoB10Modify(g, mine)
	paCast(t, g, "Flame Discharge", "Instant", "{X}{R}", neoB10FlameDischarge, game.CastSpellParams{XValue: 2, Targets: b16TargetCard(plain)})
	passPriorityAroundTable(t, g)
	if neoB10Alive(g, plain) {
		t.Error("X=2 with a modified creature is 4 damage and should kill")
	}
}

func TestNeoB10FlameDischargeReadsTheBoardAsYouCastIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mine := b12Creature(g, me.ID, "Mine", neoB10Bear, 2, 2)
	neoB10Modify(g, mine)
	victim := b12Creature(g, opp.ID, "Victim", neoB10Bear, 2, 4)
	paCast(t, g, "Flame Discharge", "Instant", "{X}{R}", neoB10FlameDischarge, game.CastSpellParams{XValue: 2, Targets: b16TargetCard(victim), HoldPriority: true})
	neoB10Unmodify(g, mine) // in response
	passPriorityAroundTable(t, g)
	if neoB10Alive(g, victim) {
		t.Error("the creature was modified as the spell was cast; losing it in response must not shrink the damage")
	}
	// And the other direction: modified only after the cast does not count.
	victim2 := b12Creature(g, opp.ID, "Victim2", neoB10Bear, 2, 4)
	paCast(t, g, "Flame Discharge", "Instant", "{X}{R}", neoB10FlameDischarge, game.CastSpellParams{XValue: 2, Targets: b16TargetCard(victim2), HoldPriority: true})
	neoB10Modify(g, mine)
	passPriorityAroundTable(t, g)
	if !neoB10Alive(g, victim2) {
		t.Error("a modified creature that appeared after the cast must not add damage")
	}
}

func TestNeoB10LethalExploitShrinksOnePerModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "A", neoB10Bear, 2, 2)
	b := b12Creature(g, me.ID, "B", neoB10Bear, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", neoB10Bear, 2, 2)
	neoB10Modify(g, theirs) // an opponent's modified creature is not counted
	big := b12Creature(g, opp.ID, "Big", neoB10Bear, 9, 9)
	castCatalogSpell(t, g, "Lethal Exploit", "Instant", neoB10LethalExploit, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, big); p != 7 {
		t.Errorf("no modified creature of mine: power %d, want 7", p)
	}
	neoB10Modify(g, a)
	neoB10Modify(g, b)
	castCatalogSpell(t, g, "Lethal Exploit", "Instant", neoB10LethalExploit, b16TargetCard(big))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 7-4 || tg != 9-2-4 {
		t.Errorf("Big is %d/%d, want 3/3 after -2/-2 and -4/-4 (two modified creatures)", p, tg)
	}
}

func TestNeoB10KamisFlareHitsTheControllerOnlyWithAModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Victim", neoB10Bear, 2, 3)
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Kami's Flare", "Instant", neoB10KamisFlare, b16TargetCard(victim))
	passPriorityAroundTable(t, g)
	if neoB10Alive(g, victim) {
		t.Error("3 damage should kill a 3-toughness creature")
	}
	if got := lifeOf(g, opp.ID); got != life {
		t.Errorf("no modified creature: life %d → %d, want unchanged", life, got)
	}
	mine := b12Creature(g, me.ID, "Mine", neoB10Bear, 2, 2)
	neoB10Modify(g, mine)
	victim2 := b12Creature(g, opp.ID, "Victim2", neoB10Bear, 2, 3)
	castCatalogSpell(t, g, "Kami's Flare", "Instant", neoB10KamisFlare, b16TargetCard(victim2))
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("modified creature: life %d → %d, want -2 for the dead creature's controller", life, got)
	}
	if err := castCatalogSpellErrStrict(t, g, "Kami's Flare", "Instant", neoB10KamisFlare, b16TargetPlayer(opp.ID)); err == nil {
		t.Error("targeted a player")
	}
}

func TestNeoB10AmbitiousAssaultPumpsAndDrawsOnlyWithAModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	bear := b12Creature(g, me.ID, "Bear", neoB10Bear, 2, 2)
	hand := handSize(me)
	castCatalogSpell(t, g, "Ambitious Assault", "Instant", neoB10AmbitiousAsslt, nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if got := handSize(me); got != hand {
		t.Errorf("no modified creature: hand %d → %d, want no draw", hand, got)
	}
	neoB10Modify(g, bear)
	hand = handSize(me)
	castCatalogSpell(t, g, "Ambitious Assault", "Instant", neoB10AmbitiousAsslt, nil)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Errorf("modified creature: hand %d → %d, want one draw", hand, got)
	}
}

func TestNeoB10AkkiEmberKeeperMakesASpiritWhenANontokenModifiedCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Akki Ember-Keeper", "Enchantment Creature — Goblin Warrior", neoB10AkkiEmberKeep, false)
	spirits := func() int { return tokensNamedControlledBy(g, "Spirit", me.ID) }
	plain := b12Creature(g, me.ID, "Plain", neoB10Bear, 2, 2)
	killCreature(t, g, me.ID, plain)
	passPriorityAroundTable(t, g)
	if spirits() != 0 {
		t.Error("an unmodified creature dying makes no Spirit")
	}
	modified := b12Creature(g, me.ID, "Modified", neoB10Bear, 2, 2)
	neoB10Modify(g, modified)
	killCreature(t, g, me.ID, modified)
	passPriorityAroundTable(t, g)
	if spirits() != 1 {
		t.Errorf("a nontoken modified creature died: %d Spirits, want 1", spirits())
	}
	// The Spirit token, once modified, dying makes no more.
	var spirit uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == me.ID {
			spirit = c.InstanceID
		}
	}
	neoB10Modify(g, spirit)
	killCreature(t, g, me.ID, spirit)
	passPriorityAroundTable(t, g)
	if spirits() != 0 {
		t.Error("a token dying makes no Spirit")
	}
	theirs := b12Creature(g, opp.ID, "Theirs", neoB10Bear, 2, 2)
	neoB10Modify(g, theirs)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if spirits() != 0 {
		t.Error("an opponent's creature dying makes no Spirit")
	}
}

func TestNeoB10AkkiEmberKeeperSeesItselfDie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	akki := pushCatalogPermanent(g, me.ID, "Akki Ember-Keeper", "Enchantment Creature — Goblin Warrior", neoB10AkkiEmberKeep, false)
	neoB10Modify(g, akki)
	killCreature(t, g, me.ID, akki)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Spirit", me.ID); n != 1 {
		t.Errorf("a modified Akki dying: %d Spirits, want 1", n)
	}
}

func TestNeoB10UpriserRenegadeGrowsPerOtherModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	up := pushCatalogPermanent(g, me.ID, "Upriser Renegade", "Creature — Human Samurai", neoB10UpriserRen, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == up {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 3
			}
		}
	})
	if p := effectivePower(t, g, up); p != 1 {
		t.Errorf("alone: power %d, want 1", p)
	}
	a := b12Creature(g, me.ID, "A", neoB10Bear, 2, 2)
	neoB10Modify(g, a)
	b := b12Creature(g, me.ID, "B", neoB10Bear, 2, 2)
	neoB10Modify(g, b)
	theirs := b12Creature(g, opp.ID, "Theirs", neoB10Bear, 2, 2)
	neoB10Modify(g, theirs)
	if p := effectivePower(t, g, up); p != 5 {
		t.Errorf("two other modified creatures: power %d, want 5", p)
	}
	// Its own counter does not count: "other".
	neoB10Modify(g, up)
	// effectivePower leaves counters out; the Renegade's own counter must not add to the +2s.
	if p := effectivePower(t, g, up); p != 5 {
		t.Errorf("with its own counter: power %d, want 5 (its own modification is not \"other\")", p)
	}
}

func TestNeoB10HeirEntersWithACounterOnlyIfYouControlAModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	heir := b12PlayFromHand(t, g, "Heir of the Ancient Fang", "Creature — Snake Samurai", neoB10HeirAncientFng, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := counterCount(g, heir, "+1/+1"); n != 0 {
		t.Errorf("no modified creature: %d counters, want 0", n)
	}
	bear := b12Creature(g, me.ID, "Bear", neoB10Bear, 2, 2)
	neoB10Modify(g, bear)
	heir2 := b12PlayFromHand(t, g, "Heir of the Ancient Fang", "Creature — Snake Samurai", neoB10HeirAncientFng, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := counterCount(g, heir2, "+1/+1"); n != 1 {
		t.Errorf("a modified creature: %d counters, want 1", n)
	}
}

func TestNeoB10OrochiMergeKeeperTapsForTwoWhileModified(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	orochi := pushCatalogPermanent(g, me.ID, "Orochi Merge-Keeper", "Creature — Snake Druid", neoB10OrochiMerge, false)
	tapForMana(t, g, me.ID, orochi)
	if got := sortedPool(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("unmodified tap: pool %v, want [G]", got)
	}
	me.ManaPool.EmptyPool()
	if err := g.TapCard(orochi, false); err != nil {
		t.Fatalf("untap: %v", err)
	}
	neoB10Modify(g, orochi)
	// The granted ability is a second mana ability; either may be used.
	if err := g.ActivateManaAbility(me.ID, orochi, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("modified: the granted {T}: Add {G}{G} should exist: %v", err)
	}
	if got := sortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("modified tap: pool %v, want [G G]", got)
	}
	neoB10Unmodify(g, orochi)
	if err := g.ActivateManaAbility(me.ID, orochi, 1, game.ManaAbilityParams{}); err == nil {
		t.Error("the granted ability must go with the last modification")
	}
}

func TestNeoB10WalkingSkyscraperCostsLessAndHasHexproofWhileUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	for range 2 {
		neoB10Modify(g, b12Creature(g, me.ID, "Mod", neoB10Bear, 2, 2))
	}
	b12Creature(g, me.ID, "Plain", neoB10Bear, 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Walking Skyscraper", TypeLine: "Artifact Creature — Construct",
		ManaCost: "{8}", OracleID: neoB10WalkingSky, Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("two modified creatures make it cost {6}; five mana must not pay for it")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("six mana pays for {8} with two modified creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("lacks trample")
	}
	if !hasEffectiveKeyword(t, g, id, "hexproof") {
		t.Error("an untapped Skyscraper has hexproof")
	}
	if err := g.TapCard(id, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	if hasEffectiveKeyword(t, g, id, "hexproof") {
		t.Error("a tapped Skyscraper has no hexproof")
	}
}

func TestNeoB10GuardiansOfOboroLetsModifiedCreaturesAttackDespiteDefender(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Guardians of Oboro", "Creature — Moonfolk Samurai", neoB10GuardiansOb, false)
	modified := pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	plain := pushTestPermanent(g, me.ID, game.Card{Name: "Plain Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	theirs := pushTestPermanent(g, opp.ID, game.Card{Name: "Their Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	neoB10Modify(g, modified)
	neoB10Modify(g, theirs)
	advanceToMainOf(t, g, 0)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := declareResult(g, modified, opp.ID); err != nil {
		t.Errorf("a modified creature with defender may attack: %v", err)
	}
	wantDefenderRefused(t, g, plain, opp.ID)
	if err := declareResult(g, theirs, me.ID); err == nil {
		t.Error("an opponent's modified defender gets no waiver")
	}
}

func TestNeoB10TowashiGuideBotCountersOnEnterAndDrawsForLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", neoB10Bear, 2, 2)
	bot := b12PlayFromHand(t, g, "Towashi Guide-Bot", "Artifact Creature — Construct", neoB10GuideBot, game.CastSpellParams{})
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Fatalf("the enter trigger put %d counters on the Bear, want 1", n)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bot {
				g.Battlefield.Cards[i].SummonedThisTurn = false
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	seedLibrary(me, "A", "B", "C")
	// One modified creature (the Bear): the ability costs {3}.
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bot, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{2} must not pay for a {3} ability")
	}
	floatMana(t, g, me, "{C}")
	hand := handSize(me)
	if err := g.ActivateCatalogAbility(me.ID, bot, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("one modified creature makes it cost {3}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand %d → %d, want one draw", hand, got)
	}
}

func TestNeoB10InvigoratingHotSpringCountersHasteAndOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	spring := b12PlayFromHand(t, g, "Invigorating Hot Spring", "Enchantment", neoB10HotSpring, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := counterCount(g, spring, "+1/+1"); n != 4 {
		t.Fatalf("enters with %d counters, want 4", n)
	}
	bear := b12Creature(g, me.ID, "Bear", neoB10Bear, 2, 2)
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("an unmodified creature has no haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, spring, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, spring, "+1/+1"); n != 3 {
		t.Errorf("the Spring has %d counters, want 3", n)
	}
	if n := counterCount(g, bear, "+1/+1"); n != 1 {
		t.Errorf("the Bear has %d counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("a modified creature has haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, spring, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Error("only once each turn")
	}
}

func TestNeoB10ThunderingRaijuCountsAfterItsCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raiju := pushCatalogPermanent(g, me.ID, "Thundering Raiju", "Creature — Spirit", neoB10Raiju, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == raiju {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
			}
		}
	})
	other := b12Creature(g, me.ID, "Other", neoB10Bear, 2, 2)
	modified := b12Creature(g, me.ID, "Modified", neoB10Bear, 2, 2)
	neoB10Modify(g, modified)
	theirs := b12Creature(g, opp.ID, "Theirs", neoB10Bear, 2, 2)
	neoB10Modify(g, theirs) // not "you control"
	life := lifeOf(g, opp.ID)
	advanceToMainOf(t, g, 0)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raiju, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b04WaitForPick(t, g, me.ID)
	// The counter goes on the unmodified Other: it becomes modified first,
	// so X is 2 (Modified and Other), and the Raiju is not counted.
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, other, "+1/+1"); n != 1 {
		t.Errorf("Other has %d counters, want 1", n)
	}
	if got := lifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent life %d → %d, want 2 damage (Modified and Other, not the Raiju)", life, got)
	}
}
