package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSocB6EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.SocB6FurygaleOracle:  "Furygale Flocking",
		effectstest.SocB6GalazethOracle:  "Galazeth Prismari",
		effectstest.SocB6GhoulishOracle:  "Ghoulish Impetus",
		effectstest.SocB6GormaOracle:     "Gorma, the Gullet",
		effectstest.SocB6AugmenterOracle: "Guardian Augmenter",
		effectstest.SocB6FaithOracle:     "Guardian of Faith",
		effectstest.SocB6ScalelordOracle: "Guardian Scalelord",
		effectstest.SocB6GyomeOracle:     "Gyome, Master Chef",
		effectstest.SocB6EidolonOracle:   "Hateful Eidolon",
		effectstest.SocB6HeraldOracle:    "Herald of Amity",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Furygale Flocking ---

func TestSocB6FurygaleFlockingCostsLessPerInstantAndSorceryCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	for _, tl := range []string{"Instant", "Instant", "Sorcery", "Sorcery"} {
		effectstest.SocB6Graveyard(me, "Spent "+tl, tl, "{1}")
	}
	effectstest.SocB6Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.HandCardFull(me, "Furygale Flocking", "Sorcery", "{8}{R}{R}", effectstest.SocB6FurygaleOracle, []string{"R"})
	effectstest.B06AddMana(me, "C", "C", "C", "C", "R", "R")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("four instant and sorcery cards should make it cost {4}{R}{R}: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in the pool, want 0 (the creature card does not count)", me.ManaPool.Total())
	}
}

func TestSocB6FurygaleFlockingMakesTwoBoundHastyFliersPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Furygale Flocking", "Sorcery", effectstest.SocB6FurygaleOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.SocB6Named(g, "Elemental")
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
		if !effectstest.HasEffectiveKeyword(t, g, tok, "flying") || !effectstest.HasEffectiveKeyword(t, g, tok, "haste") {
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
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.AdvanceStep(); err == nil {
		t.Error("the bound tokens were allowed not to attack")
	}
}

// --- Galazeth Prismari ---

func TestSocB6GalazethMakesATreasureAndGivesArtifactsRestrictedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastHoldCreature(t, g, "Galazeth Prismari", "Legendary Creature — Elder Dragon", effectstest.SocB6GalazethOracle, 3, 4)
	treasures := effectstest.SocB6Named(g, "Treasure")
	if len(treasures) != 1 {
		t.Fatalf("%d Treasures, want 1", len(treasures))
	}
	rock := effectstest.GaPut(g, me.ID, "Mind Stone", "Artifact", "")
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := effectstest.GaPut(g, g.Seats[1].ID, "Their Rock", "Artifact", "")
	if effectstest.GrantedManaRowsOf(g, rock) != 1 || effectstest.GrantedManaRowsOf(g, treasures[0]) != 1 {
		t.Fatal("an artifact you control lacks the granted mana ability")
	}
	if effectstest.GrantedManaRowsOf(g, bear) != 0 || effectstest.GrantedManaRowsOf(g, theirs) != 0 {
		t.Error("a nonartifact or an opponent's artifact got the ability")
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, rock, "U"); err != nil {
		t.Fatalf("tap the rock: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 1 || units[0].Color != "U" || len(units[0].Restrictions) == 0 {
		t.Fatalf("pool %+v, want one restricted {U}", units)
	}
	creature := effectstest.HandCardFull(me, "Blue Bear", "Creature — Bear", "{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err == nil && !me.Hand.Contains(creature) {
		t.Error("the restricted mana paid for a creature spell")
	}
	if me.ManaPool.Total() != 1 {
		t.Fatal("the restricted mana was spent on a creature spell")
	}
	instant := effectstest.HandCardFull(me, "Opt", "Instant", "{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the mana should pay for an instant: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the instant was not paid from the pool")
	}
}

// --- Ghoulish Impetus ---

func TestSocB6GhoulishImpetusPumpsGrantsDeathtouchGoadsAndComesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	other := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Ghoulish Impetus", effectstest.SocB6GhoulishOracle, bear)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tg != 3 {
		t.Errorf("enchanted bear is %d/%d, want 3/3", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("the enchanted creature lacks deathtouch")
	}
	if c, ok := g.LookupCardForEffect(bear); !ok || !c.Goaded() {
		t.Error("the enchanted creature is not goaded")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Fatal("the Aura is not in the graveyard after the creature died")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		effectstest.AnswerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: other})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not return at the beginning of the next end step")
	}
	if host := effectstest.AttachmentHostOf(t, g, aura); host.ID != other {
		t.Errorf("the Aura enchants %v, want the chosen creature", host.ID)
	}
}

func TestSocB6GhoulishImpetusStaysIfItLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Ghoulish Impetus", effectstest.SocB6GhoulishOracle, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, aura); err != nil {
		t.Fatalf("exile the Aura: %v", err)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || effectstest.AuraHostChoiceFor(g, me.ID) != nil {
		t.Error("an Aura exiled from the graveyard came back")
	}
}

// --- Gorma, the Gullet ---

func TestSocB6GormaGrowsOnDeathsAndNontokenEntriesGetCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	gorma := effectstest.B12Push(g, me.ID, "Gorma, the Gullet", "Legendary Creature — Pest Frog", effectstest.SocB6GormaOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, gorma, "lifelink") {
		t.Error("Gorma lacks lifelink")
	}
	a := effectstest.B12Creature(g, me.ID, "Fodder A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Fodder B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, a)
	effectstest.KillCreature(t, g, me.ID, b)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, gorma, game.CounterPlusOne); n != 2 {
		t.Errorf("Gorma has %d counters, want 2 (an opponent's creature does not count)", n)
	}
	bear := effectstest.CastHoldCreature(t, g, "New Bear", "Creature — Bear", "", 2, 2)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("the new creature entered with %d counters, want 2 (two of yours died)", n)
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 black Zombie"), 1); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	for _, id := range effectstest.SocB6Named(g, "Zombie") {
		if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 0 {
			t.Errorf("a token entered with %d counters, want 0", n)
		}
	}
}

func TestSocB6GormaGivesNothingWithNoDeathsThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Gorma, the Gullet", "Legendary Creature — Pest Frog", effectstest.SocB6GormaOracle, 1, 1)
	bear := effectstest.CastHoldCreature(t, g, "New Bear", "Creature — Bear", "", 2, 2)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("the creature entered with %d counters, want 0", n)
	}
}

// --- Guardian Augmenter ---

func TestSocB6GuardianAugmenterPumpsAndShieldsYourCommanders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Guardian Augmenter", "Creature — Troll Wizard", effectstest.SocB6AugmenterOracle, 2, 2)
	cmdr := effectstest.B12Creature(g, me.ID, "My Commander", "Legendary Creature — Elf", 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Commander", "Legendary Creature — Elf", 3, 3)
	relic := effectstest.GaPut(g, me.ID, "My Artifact Commander", "Legendary Artifact", "")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case cmdr, theirs, relic:
				g.Battlefield.Cards[i].IsCommander = true
			}
		}
	})
	if p, tg := effectstest.EffectivePower(t, g, cmdr), effectstest.EffectiveToughness(t, g, cmdr); p != 5 || tg != 5 {
		t.Errorf("your commander is %d/%d, want 5/5", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, cmdr, "hexproof") || !effectstest.HasEffectiveKeyword(t, g, relic, "hexproof") {
		t.Error("a commander you control lacks hexproof")
	}
	if effectstest.EffectivePower(t, g, bear) != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "hexproof") {
		t.Error("a creature that is not a commander was affected")
	}
	if effectstest.EffectivePower(t, g, theirs) != 3 || effectstest.HasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("an opponent's commander was affected")
	}
}

// --- Guardian of Faith ---

func TestSocB6GuardianOfFaithPhasesOutTheChosenOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	kept := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	guardian := effectstest.CastHoldCreature(t, g, "Guardian of Faith", "Creature — Spirit Knight", effectstest.SocB6FaithOracle, 3, 2)
	pick := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(a) || !g.PhasedOut.Contains(b) {
		t.Error("a chosen creature did not phase out")
	}
	if !g.Battlefield.Contains(kept) || !g.Battlefield.Contains(guardian) {
		t.Error("a creature that was not chosen left")
	}
	if !effectstest.HasEffectiveKeyword(t, g, guardian, "vigilance") || !effectstest.HasEffectiveKeyword(t, g, guardian, "flash") {
		t.Error("the Guardian lacks flash or vigilance")
	}
}

// --- Guardian Scalelord ---

func TestSocB6ScalelordBackupOnAnotherCreatureGrantsFlyingAndTheReturn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	small := effectstest.SocB6Graveyard(me, "Small Rock", "Artifact", "{3}")
	big := effectstest.SocB6Graveyard(me, "Big Rock", "Artifact", "{4}")
	land := effectstest.SocB6Graveyard(me, "Old Land", "Land", "")
	scalelord := effectstest.CastHoldCreature(t, g, "Guardian Scalelord", "Creature — Dragon", effectstest.SocB6ScalelordOracle, 3, 4)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("the Bear has %d counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("backup did not give the Bear flying")
	}
	if effectstest.CountersOn(g, scalelord, game.CounterPlusOne) != 0 {
		t.Error("the Scalelord got the counter")
	}
	// The Bear is now 3/3: its granted attack trigger returns a card with
	// mana value 3 or less.
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the card did not return to the battlefield")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the backup grant outlived the turn")
	}
}

func TestSocB6ScalelordBackupOnItselfGivesOnlyTheCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	scalelord := effectstest.CastHoldCreature(t, g, "Guardian Scalelord", "Creature — Dragon", effectstest.SocB6ScalelordOracle, 3, 4)
	effectstest.AnswerTriggerTargets(t, g, me.ID, scalelord)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, scalelord, game.CounterPlusOne); n != 1 {
		t.Errorf("the Scalelord has %d counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, scalelord, "flying") {
		t.Error("the Scalelord lacks its own flying")
	}
}

// --- Gyome, Master Chef ---

func TestSocB6GyomeMakesFoodForEachNontokenCreatureThatEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gyome := effectstest.CastHoldCreature(t, g, "Gyome, Master Chef", "Legendary Creature — Troll Warlock", effectstest.SocB6GyomeOracle, 5, 3)
	gone := effectstest.CastHoldCreature(t, g, "Short-Lived Bear", "Creature — Bear", "", 2, 2)
	effectstest.KillCreature(t, g, me.ID, gone)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 black Zombie"), 1); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	if !effectstest.HasEffectiveKeyword(t, g, gyome, "trample") {
		t.Error("Gyome lacks trample")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB6Named(g, "Food")); n != 2 {
		t.Errorf("%d Food, want 2 (Gyome and the creature that died; the token does not count)", n)
	}
}

func TestSocB6GyomeSacrificesAFoodForIndestructibleAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gyome := effectstest.B12Push(g, me.ID, "Gyome, Master Chef", "Legendary Creature — Troll Warlock", effectstest.SocB6GyomeOracle, 5, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1); err != nil {
			t.Fatalf("food: %v", err)
		}
	})
	food := effectstest.SocB6Named(g, "Food")[0]
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, gyome, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the target did not gain indestructible")
	}
	if !effectstest.IsTapped(g, bear) {
		t.Error("the target was not tapped")
	}
}

// --- Hateful Eidolon ---

func TestSocB6HatefulEidolonDrawsForEachOfYourAurasOnTheDeadCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	eidolon := effectstest.B12Push(g, me.ID, "Hateful Eidolon", "Enchantment Creature — Spirit", effectstest.SocB6EidolonOracle, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, eidolon, "lifelink") {
		t.Error("the Eidolon lacks lifelink")
	}
	theirs := effectstest.AuraBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Pacifism", effectstest.PacifismAuraOracle, theirs)
	effectstest.AuraCast(t, g, "Pacifism", effectstest.PacifismAuraOracle, theirs)
	oppAura := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Pacifism", TypeLine: effectstest.AuraTypeLine,
		OracleID: effectstest.PacifismAuraOracle, Owner: opp.ID, Controller: opp.ID})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(oppAura, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (your two Auras; the opponent's does not count)", got)
	}
}

func TestSocB6HatefulEidolonIgnoresAnUnenchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	eidolon := effectstest.B12Push(g, me.ID, "Hateful Eidolon", "Enchantment Creature — Spirit", effectstest.SocB6EidolonOracle, 1, 2)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.KillCreature(t, g, me.ID, bear)
	if n := effectstest.TriggersOnStackFrom(g, eidolon); n != 0 {
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
	p.Library.PushTop(game.Card{InstanceID: aura, Name: "Pacifism", TypeLine: effectstest.AuraTypeLine, ManaCost: "{1}{W}",
		OracleID: effectstest.PacifismAuraOracle, Owner: p.ID, Controller: p.ID})
	return aura, creature, append(all, creature, aura)
}

func TestSocB6HeraldOfAmityCastsAnAuraFreeAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura, creature, all := socB6HeraldLibrary(me)
	effectstest.CastHoldCreature(t, g, "Herald of Amity", "Creature — Griffin", effectstest.SocB6HeraldOracle, 2, 2)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, aura)
	if effectstest.HasID(offer.DiscardOptions, creature) {
		t.Error("a creature card was offered as an Aura spell")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, aura, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast the Aura: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if host := effectstest.AttachmentHostOf(t, g, aura); host.ID != victim {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aura, _, all := socB6HeraldLibrary(me)
	effectstest.CastHoldCreature(t, g, "Herald of Amity", "Creature — Griffin", effectstest.SocB6HeraldOracle, 2, 2)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, aura)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range all {
		if !me.Library.Contains(id) {
			t.Error("an exiled card was left in exile")
		}
	}
}

func TestSocB6HeraldOfAmityPumpsByYourAurasOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	herald := effectstest.B12Push(g, me.ID, "Herald of Amity", "Creature — Griffin", effectstest.SocB6HeraldOracle, 2, 2)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Pacifism", effectstest.PacifismAuraOracle, bear)
	effectstest.AuraCast(t, g, "Pacifism", effectstest.PacifismAuraOracle, bear)
	if !effectstest.HasEffectiveKeyword(t, g, herald, "flying") {
		t.Error("the Herald lacks flying")
	}
	effectstest.DeclareAttack(t, g, opp.ID, herald)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, herald), effectstest.EffectiveToughness(t, g, herald); p != 4 || tg != 4 {
		t.Errorf("the attacking Herald is %d/%d, want 4/4 (two Auras)", p, tg)
	}
}
