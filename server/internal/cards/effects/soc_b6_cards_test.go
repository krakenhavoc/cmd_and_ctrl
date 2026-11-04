package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b6_cards_test.go — Secrets of Strixhaven Commander, batch 6
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB6FurygaleOracle    = "943fc88d-c281-46bb-9dc7-1052f789362b"
	socB6GalazethOracle    = "f9217532-f7cf-4436-b476-72d97a3147dc"
	socB6GhoulishOracle    = "276eccd6-4a76-46e8-b604-94472e618115"
	socB6GormaOracle       = "ae5f30cc-a22e-4fbb-9510-e9429de3623c"
	socB6AugmenterOracle   = "d925e560-faa1-40a9-b25a-47065ee891e6"
	socB6FaithOracle       = "8e04a5e0-e22c-430b-a39a-c9f38e16ea86"
	socB6ScalelordOracle   = "d6c7d7a6-2d38-4c06-ab83-bee9cc18de52"
	socB6GyomeOracle       = "0f1b5cea-2035-444c-9f63-87dcb9782c74"
	socB6EidolonOracle     = "a3b39590-34b7-4fd0-877f-748e6df30e88"
	socB6HeraldOracle      = "6c25c6b5-d98c-4eb4-aacd-b6960a8779a9"
	socB6ElementalTokenKey = "3/3 blue and red Elemental with flying"
)

func TestSocB6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB6FurygaleOracle:  "Furygale Flocking",
		socB6GalazethOracle:  "Galazeth Prismari",
		socB6GhoulishOracle:  "Ghoulish Impetus",
		socB6GormaOracle:     "Gorma, the Gullet",
		socB6AugmenterOracle: "Guardian Augmenter",
		socB6FaithOracle:     "Guardian of Faith",
		socB6ScalelordOracle: "Guardian Scalelord",
		socB6GyomeOracle:     "Gyome, Master Chef",
		socB6EidolonOracle:   "Hateful Eidolon",
		socB6HeraldOracle:    "Herald of Amity",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB6Named lists the battlefield permanents with the name.
func socB6Named(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// socB6Graveyard puts a card of the type line and cost in p's graveyard.
func socB6Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Furygale Flocking ---

func TestSocB6FurygaleFlockingCostsLessPerInstantAndSorceryCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	for _, tl := range []string{"Instant", "Instant", "Sorcery", "Sorcery"} {
		socB6Graveyard(me, "Spent "+tl, tl, "{1}")
	}
	socB6Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	spell := handCardFull(me, "Furygale Flocking", "Sorcery", "{8}{R}{R}", socB6FurygaleOracle, []string{"R"})
	b06AddMana(me, "C", "C", "C", "C", "R", "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("four instant and sorcery cards should make it cost {4}{R}{R}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in the pool, want 0 (the creature card does not count)", me.ManaPool.Total())
	}
}

func TestSocB6FurygaleFlockingMakesTwoBoundHastyFliersPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Furygale Flocking", "Sorcery", socB6FurygaleOracle, nil)
	passPriorityAroundTable(t, g)
	toks := socB6Named(g, "Elemental")
	if len(toks) != 6 {
		t.Fatalf("%d Elementals, want two for each of three opponents", len(toks))
	}
	perOpp := map[uuid.UUID]int{}
	g.ReadSnapshot(func() {
		for _, tok := range toks {
			c, _ := g.LookupCardForEffect(tok)
			eff := c.Effective()
			if c.Controller != me.ID || eff.Power != 3 || eff.Toughness != 3 {
				t.Errorf("token is %d/%d under %v, want a 3/3 of yours", eff.Power, eff.Toughness, c.Controller)
			}
			for _, r := range eff.AttackRequirements {
				if r.MustAttack != uuid.Nil {
					perOpp[r.MustAttack]++
				}
			}
		}
	})
	for _, tok := range toks {
		if !hasEffectiveKeyword(t, g, tok, "flying") || !hasEffectiveKeyword(t, g, tok, "haste") {
			t.Error("an Elemental lacks flying or haste")
		}
	}
	if len(perOpp) != 3 {
		t.Fatalf("tokens bound to %d players, want 3", len(perOpp))
	}
	for p, n := range perOpp {
		if p == me.ID || n != 2 {
			t.Errorf("%d tokens bound to %v, want two per opponent", n, p)
		}
	}
	// The requirement holds: the declare-attackers step can't be left with
	// the hasty tokens idle.
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.AdvanceStep(); err == nil {
		t.Error("the bound tokens were allowed not to attack")
	}
}

// --- Galazeth Prismari ---

func TestSocB6GalazethMakesATreasureAndGivesArtifactsRestrictedMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castHoldCreature(t, g, "Galazeth Prismari", "Legendary Creature — Elder Dragon", socB6GalazethOracle, 3, 4)
	treasures := socB6Named(g, "Treasure")
	if len(treasures) != 1 {
		t.Fatalf("%d Treasures, want 1", len(treasures))
	}
	rock := gaPut(g, me.ID, "Mind Stone", "Artifact", "")
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := gaPut(g, g.Seats[1].ID, "Their Rock", "Artifact", "")
	if grantedManaRowsOf(g, rock) != 1 || grantedManaRowsOf(g, treasures[0]) != 1 {
		t.Fatal("an artifact you control lacks the granted mana ability")
	}
	if grantedManaRowsOf(g, bear) != 0 || grantedManaRowsOf(g, theirs) != 0 {
		t.Error("a nonartifact or an opponent's artifact got the ability")
	}
	if err := tapGrantedMana(t, g, me.ID, rock, "U"); err != nil {
		t.Fatalf("tap the rock: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 1 || units[0].Color != "U" || len(units[0].Restrictions) == 0 {
		t.Fatalf("pool %+v, want one restricted {U}", units)
	}
	creature := handCardFull(me, "Blue Bear", "Creature — Bear", "{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil && !me.Hand.Contains(creature) {
		t.Error("the restricted mana paid for a creature spell")
	}
	if me.ManaPool.Total() != 1 {
		t.Fatal("the restricted mana was spent on a creature spell")
	}
	instant := handCardFull(me, "Opt", "Instant", "{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the mana should pay for an instant: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the instant was not paid from the pool")
	}
}

// --- Ghoulish Impetus ---

func TestSocB6GhoulishImpetusPumpsGrantsDeathtouchGoadsAndComesBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := auraBear(g, me.ID)
	other := auraBear(g, me.ID)
	aura := auraCast(t, g, "Ghoulish Impetus", socB6GhoulishOracle, bear)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the enchanted creature lacks deathtouch")
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Fatal("the Aura is not in the graveyard after the creature died")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if auraHostChoiceFor(g, me.ID) != nil {
		answerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: other})
		passPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not return at the beginning of the next end step")
	}
	if host := attachmentHostOf(t, g, aura); host.ID != other {
		t.Errorf("the Aura enchants %v, want the chosen creature", host.ID)
	}
}

func TestSocB6GhoulishImpetusStaysIfItLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := auraBear(g, me.ID)
	auraBear(g, me.ID)
	aura := auraCast(t, g, "Ghoulish Impetus", socB6GhoulishOracle, bear)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, aura); err != nil {
		t.Fatalf("exile the Aura: %v", err)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || auraHostChoiceFor(g, me.ID) != nil {
		t.Error("an Aura exiled from the graveyard came back")
	}
}

// --- Gorma, the Gullet ---

func TestSocB6GormaGrowsOnDeathsAndNontokenEntriesGetCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	gorma := b12Push(g, me.ID, "Gorma, the Gullet", "Legendary Creature — Pest Frog", socB6GormaOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, gorma, "lifelink") {
		t.Error("Gorma lacks lifelink")
	}
	a := b12Creature(g, me.ID, "Fodder A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Fodder B", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, a)
	killCreature(t, g, me.ID, b)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, gorma, game.CounterPlusOne); n != 2 {
		t.Errorf("Gorma has %d counters, want 2 (an opponent's creature does not count)", n)
	}
	bear := castHoldCreature(t, g, "New Bear", "Creature — Bear", "", 2, 2)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("the new creature entered with %d counters, want 2 (two of yours died)", n)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("2/2 black Zombie"), 1); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	for _, id := range socB6Named(g, "Zombie") {
		if n := countersOn(g, id, game.CounterPlusOne); n != 0 {
			t.Errorf("a token entered with %d counters, want 0", n)
		}
	}
}

func TestSocB6GormaGivesNothingWithNoDeathsThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Gorma, the Gullet", "Legendary Creature — Pest Frog", socB6GormaOracle, 1, 1)
	bear := castHoldCreature(t, g, "New Bear", "Creature — Bear", "", 2, 2)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("the creature entered with %d counters, want 0", n)
	}
}

// --- Guardian Augmenter ---

func TestSocB6GuardianAugmenterPumpsAndShieldsYourCommanders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Guardian Augmenter", "Creature — Troll Wizard", socB6AugmenterOracle, 2, 2)
	cmdr := b12Creature(g, me.ID, "My Commander", "Legendary Creature — Elf", 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Commander", "Legendary Creature — Elf", 3, 3)
	relic := gaPut(g, me.ID, "My Artifact Commander", "Legendary Artifact", "")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case cmdr, theirs, relic:
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	if p, tg := effectivePower(t, g, cmdr), effectiveToughness(t, g, cmdr); p != 5 || tg != 5 {
		t.Errorf("your commander is %d/%d, want 5/5", p, tg)
	}
	if !hasEffectiveKeyword(t, g, cmdr, "hexproof") || !hasEffectiveKeyword(t, g, relic, "hexproof") {
		t.Error("a commander you control lacks hexproof")
	}
	if effectivePower(t, g, bear) != 2 || hasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("a creature that is not a commander was affected")
	}
	if effectivePower(t, g, theirs) != 3 || hasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("an opponent's commander was affected")
	}
}

// --- Guardian of Faith ---

func TestSocB6GuardianOfFaithPhasesOutTheChosenOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	kept := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	guardian := castHoldCreature(t, g, "Guardian of Faith", "Creature — Spirit Knight", socB6FaithOracle, 3, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: guardian}}); err == nil {
		t.Error("the Guardian itself was accepted as an other creature")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(a) || !g.PhasedOut.Contains(b) {
		t.Error("a chosen creature did not phase out")
	}
	if !g.Battlefield.Contains(kept) || !g.Battlefield.Contains(guardian) {
		t.Error("a creature that was not chosen left")
	}
	if !hasEffectiveKeyword(t, g, guardian, "vigilance") || !hasEffectiveKeyword(t, g, guardian, "flash") {
		t.Error("the Guardian lacks flash or vigilance")
	}
}

// --- Guardian Scalelord ---

func TestSocB6ScalelordBackupOnAnotherCreatureGrantsFlyingAndTheReturn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	small := socB6Graveyard(me, "Small Rock", "Artifact", "{3}")
	big := socB6Graveyard(me, "Big Rock", "Artifact", "{4}")
	land := socB6Graveyard(me, "Old Land", "Land", "")
	scalelord := castHoldCreature(t, g, "Guardian Scalelord", "Creature — Dragon", socB6ScalelordOracle, 3, 4)
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bear has %d counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("backup did not give the Bear flying")
	}
	if countersOn(g, scalelord, game.CounterPlusOne) != 0 {
		t.Error("the Scalelord got the counter")
	}
	// The Bear is now 3/3: its granted attack trigger returns a card with
	// mana value 3 or less.
	declareAttack(t, g, opp.ID, bear)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the granted attack trigger asked for no target")
	}
	for _, bad := range []uuid.UUID{big, land} {
		if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bad}}); err == nil {
			t.Errorf("an illegal card %v was accepted", bad)
		}
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: small}}); err != nil {
		t.Fatalf("pick the small rock: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the card did not return to the battlefield")
	}
	advanceToEndStepOf(t, g, 0)
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the backup grant outlived the turn")
	}
}

func TestSocB6ScalelordBackupOnItselfGivesOnlyTheCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scalelord := castHoldCreature(t, g, "Guardian Scalelord", "Creature — Dragon", socB6ScalelordOracle, 3, 4)
	answerTriggerTargets(t, g, me.ID, scalelord)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, scalelord, game.CounterPlusOne); n != 1 {
		t.Errorf("the Scalelord has %d counters, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, scalelord, "flying") {
		t.Error("the Scalelord lacks its own flying")
	}
}

// --- Gyome, Master Chef ---

func TestSocB6GyomeMakesFoodForEachNontokenCreatureThatEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gyome := castHoldCreature(t, g, "Gyome, Master Chef", "Legendary Creature — Troll Warlock", socB6GyomeOracle, 5, 3)
	gone := castHoldCreature(t, g, "Short-Lived Bear", "Creature — Bear", "", 2, 2)
	killCreature(t, g, me.ID, gone)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, TokenCard("2/2 black Zombie"), 1); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	if !hasEffectiveKeyword(t, g, gyome, "trample") {
		t.Error("Gyome lacks trample")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := len(socB6Named(g, "Food")); n != 2 {
		t.Errorf("%d Food, want 2 (Gyome and the creature that died; the token does not count)", n)
	}
}

func TestSocB6GyomeSacrificesAFoodForIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gyome := b12Push(g, me.ID, "Gyome, Master Chef", "Legendary Creature — Troll Warlock", socB6GyomeOracle, 5, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, FoodToken(), 1); err != nil {
			t.Fatalf("food: %v", err)
		}
	})
	food := socB6Named(g, "Food")[0]
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, gyome, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the target did not gain indestructible")
	}
	if !isTapped(g, bear) {
		t.Error("the target was not tapped")
	}
}

// --- Hateful Eidolon ---

func TestSocB6HatefulEidolonDrawsForEachOfYourAurasOnTheDeadCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	eidolon := b12Push(g, me.ID, "Hateful Eidolon", "Enchantment Creature — Spirit", socB6EidolonOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, eidolon, "lifelink") {
		t.Error("the Eidolon lacks lifelink")
	}
	theirs := auraBear(g, opp.ID)
	auraCast(t, g, "Pacifism", pacifismAuraOracle, theirs)
	auraCast(t, g, "Pacifism", pacifismAuraOracle, theirs)
	oppAura := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Pacifism", TypeLine: auraTypeLine,
		OracleID: pacifismAuraOracle, Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(oppAura, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (your two Auras; the opponent's does not count)", got)
	}
}

func TestSocB6HatefulEidolonIgnoresAnUnenchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	eidolon := b12Push(g, me.ID, "Hateful Eidolon", "Enchantment Creature — Spirit", socB6EidolonOracle, 1, 2)
	bear := auraBear(g, me.ID)
	killCreature(t, g, me.ID, bear)
	if n := triggersOnStackFrom(g, eidolon); n != 0 {
		t.Errorf("an unenchanted creature's death triggered %d times", n)
	}
}

// --- Herald of Amity ---

// socB6HeraldLibrary seeds eight cards: an Aura on top, a creature
// under it, and six lands.
func socB6HeraldLibrary(p *game.Player) (aura, creature uuid.UUID, all []uuid.UUID) {
	all = socB3Library(p, 6, "Basic Land — Plains", "")
	creature = uuid.New()
	p.Library.PushTop(game.Card{InstanceID: creature, Name: "Library Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	aura = uuid.New()
	p.Library.PushTop(game.Card{InstanceID: aura, Name: "Pacifism", TypeLine: auraTypeLine, ManaCost: "{1}{W}",
		OracleID: pacifismAuraOracle, Owner: p.ID, Controller: p.ID})
	return aura, creature, append(all, creature, aura)
}

func TestSocB6HeraldOfAmityCastsAnAuraFreeAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura, creature, all := socB6HeraldLibrary(me)
	castHoldCreature(t, g, "Herald of Amity", "Creature — Griffin", socB6HeraldOracle, 2, 2)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, aura)
	if hasID(offer.DiscardOptions, creature) {
		t.Error("a creature card was offered as an Aura spell")
	}
	if err := castThroughResolutionCast(t, g, me.ID, aura, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast the Aura: %v", err)
	}
	castInsideTheResolution(t, g, aura)
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, aura); host.ID != victim {
		t.Errorf("the Aura enchants %v, want the target", host.ID)
	}
	for _, id := range all {
		if id == aura {
			continue
		}
		if !me.Library.Contains(id) {
			t.Error("an exiled card did not go to the bottom of the library")
		}
	}
	if me.Library.Size() != len(all)-1 {
		t.Errorf("library %d, want %d", me.Library.Size(), len(all)-1)
	}
}

func TestSocB6HeraldOfAmityDeclinedBottomsAllEight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura, _, all := socB6HeraldLibrary(me)
	castHoldCreature(t, g, "Herald of Amity", "Creature — Griffin", socB6HeraldOracle, 2, 2)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, aura)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, id := range all {
		if !me.Library.Contains(id) {
			t.Error("an exiled card was left in exile")
		}
	}
}

func TestSocB6HeraldOfAmityPumpsByYourAurasOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	herald := b12Push(g, me.ID, "Herald of Amity", "Creature — Griffin", socB6HeraldOracle, 2, 2)
	bear := auraBear(g, me.ID)
	auraCast(t, g, "Pacifism", pacifismAuraOracle, bear)
	auraCast(t, g, "Pacifism", pacifismAuraOracle, bear)
	if !hasEffectiveKeyword(t, g, herald, "flying") {
		t.Error("the Herald lacks flying")
	}
	declareAttack(t, g, opp.ID, herald)
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, herald), effectiveToughness(t, g, herald); p != 4 || tg != 4 {
		t.Errorf("the attacking Herald is %d/%d, want 4/4 (two Auras)", p, tg)
	}
}
