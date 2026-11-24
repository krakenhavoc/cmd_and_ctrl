package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB09Miles      = "98dd8231-5679-404a-b976-0bbfa6a8493c"
	oldB09Grenzo     = "a374f99e-6443-486c-8026-8eba4c7e6ad7"
	oldB09Betrayal   = "aa5e9269-b1db-4bf0-a548-24f999f60e44"
	oldB09Vikya      = "ad43ccce-3052-4817-8c46-f5c23cb49300"
	oldB09Sage       = "bdbf70da-af38-4827-9b8b-39f553ebd34a"
	oldB09Paul       = "c23ecf92-eda4-43b3-a5e7-7235b936b02b"
	oldB09Zethi      = "ac136d46-32b7-4222-a840-b8be011c1872"
	oldB09Goo        = "d1005625-0a78-4f8e-bf51-80458fc0e9d6"
	oldB09ParallaxTd = "e36a92be-537d-4a63-858a-afaab064fbf7"
)

func TestOldB09CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB09Miles: "Miles \"Tails\" Prower", oldB09Grenzo: "Grenzo, Dungeon Warden",
		oldB09Betrayal: "Inevitable Betrayal", oldB09Vikya: "Vikya, Scorching Stalwart",
		oldB09Sage: "Sage of Hours", oldB09Paul: "Paul Atreides, Ducal Heir",
		oldB09Zethi: "Zethi, Arcane Blademaster", oldB09Goo: "Chaotic Goo", oldB09ParallaxTd: "Parallax Tide",
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

// castVehicle puts a Vehicle (flying or not) in hand and casts it.
func oldB09CastVehicle(t *testing.T, g *game.Game, flying bool) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: "Skiff", TypeLine: "Artifact — Vehicle", Owner: me.ID, Controller: me.ID, Power: 3, Toughness: 3}
	if flying {
		c.Keywords = []string{"flying"}
	}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast vehicle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB09MilesPutsAFlyingCounterOnAVehicleWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Miles \"Tails\" Prower", "Legendary Creature — Fox Artificer", oldB09Miles, 3, 1)
	hand := me.Hand.Size()
	id := oldB09CastVehicle(t, g, false)
	if got := effectstest.CounterCount(g, id, game.CounterFlying); got != 1 {
		t.Errorf("flying counters: %d, want 1", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, a non-flying Vehicle draws nothing", hand, me.Hand.Size())
	}
}

func TestOldB09MilesDrawsForAFlyingVehicleInsteadOfACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Miles \"Tails\" Prower", "Legendary Creature — Fox Artificer", oldB09Miles, 3, 1)
	hand := me.Hand.Size()
	id := oldB09CastVehicle(t, g, true)
	if got := effectstest.CounterCount(g, id, game.CounterFlying); got != 0 {
		t.Errorf("flying counters: %d, want 0", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestOldB09MilesIgnoresANonVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Miles \"Tails\" Prower", "Legendary Creature — Fox Artificer", oldB09Miles, 3, 1)
	hand := me.Hand.Size()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterFlying); got != 0 || me.Hand.Size() != hand-0 && me.Hand.Size() != hand {
		t.Errorf("non-Vehicle: flying %d, hand %d (was %d)", got, me.Hand.Size(), hand)
	}
}

func TestOldB09InevitableBetrayalSuspendsAndStealsACreature(t *testing.T) {
	effectstest.S40SuspendAction(t, oldB09Betrayal, 3, "{1}{U}{U}")
	g := effectstest.NewCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	creature := effectstest.PushLibraryCardForTest(opponent, game.Card{Name: "Opponent's Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Inevitable Betrayal", TypeLine: "Sorcery", OracleID: oldB09Betrayal, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opponent.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	perm, ok := g.LookupCardForEffect(creature)
	if !ok || !g.Battlefield.Contains(creature) {
		t.Fatal("the creature should be on the battlefield")
	}
	if perm.Controller != caster.ID || perm.Owner != opponent.ID {
		t.Errorf("controller %v owner %v, want the caster controlling the opponent's creature", perm.Controller, perm.Owner)
	}
}

func oldB09BottomOfLibrary(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() { id = effectstest.PushLibraryCardForTest(p, c) })
	return id
}

func TestOldB09GrenzoPutsTheBottomCreatureOntoTheBattlefieldIfItsPowerIsLowEnough(t *testing.T) {
	for _, tc := range []struct {
		name     string
		card     game.Card
		wantZone game.ZoneKind
	}{
		{"small creature", game.Card{Name: "Imp", TypeLine: "Creature — Imp", Power: 2, Toughness: 1}, game.ZoneBattlefield},
		{"too big", game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3}, game.ZoneGraveyard},
		{"noncreature", game.Card{Name: "Rock", TypeLine: "Artifact"}, game.ZoneGraveyard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.ToMain(t, g)
			grenzo := effectstest.B12Push(g, me.ID, "Grenzo, Dungeon Warden", "Legendary Creature — Goblin Rogue", oldB09Grenzo, 2, 2)
			id := oldB09BottomOfLibrary(g, me, tc.card)
			if err := effectstest.ConB1Activate(t, g, me, grenzo, 0, "RR", game.ActivateAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.B12ZoneOf(g, id); got != tc.wantZone {
				t.Errorf("%s ended in %s, want %s", tc.name, got, tc.wantZone)
			}
		})
	}
}

func TestOldB09GrenzoEntersWithXCounters(t *testing.T) {
	spec, _ := effects.Lookup(oldB09Grenzo)
	if !spec.XMatters || len(spec.EntersWithCountersFromCast) != 1 {
		t.Fatalf("Grenzo's X counters are not declared: %+v", spec.EntersWithCountersFromCast)
	}
}

func oldB09PushVikya(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Vikya, Scorching Stalwart", "Legendary Creature — Human Warrior", oldB09Vikya, 3, 4)
	effectstest.B16Tap(g, id)
	return id
}

func TestOldB09VikyaDrawsOnlyForExcessDamageToACreature(t *testing.T) {
	for _, tc := range []struct {
		name      string
		toughness int
		wantDraw  bool
	}{{"excess", 2, true}, {"exactly lethal", 3, false}, {"not lethal", 5, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.ToMain(t, g)
			vikya := oldB09PushVikya(t, g)
			victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, tc.toughness)
			fodder := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: fodder, Name: "Fodder", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
			hand := me.Hand.Size()
			effectstest.FloatForTest(g, me, "RRRRR")
			err := g.ActivateCatalogAbility(me.ID, vikya, 0, game.ActivateAbilityParams{
				Strict: true, DiscardIDs: []uuid.UUID{fodder},
				Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
			})
			if err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := hand - 1
			if tc.wantDraw {
				want++
			}
			if me.Hand.Size() != want {
				t.Errorf("hand %d, want %d", me.Hand.Size(), want)
			}
			if c, _ := effectstest.BattlefieldCard(g, vikya); c.Tapped {
				t.Error("Vikya's {Q} should have untapped her")
			}
		})
	}
}

func TestOldB09VikyaDamageToAPlayerDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	vikya := oldB09PushVikya(t, g)
	fodder := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: fodder, Name: "Fodder", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	life, hand := opp.Life, me.Hand.Size()
	effectstest.FloatForTest(g, me, "RRRRR")
	if err := g.ActivateCatalogAbility(me.ID, vikya, 0, game.ActivateAbilityParams{
		Strict: true, DiscardIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 || me.Hand.Size() != hand-1 {
		t.Errorf("life %d -> %d, hand %d -> %d; want 3 damage and no draw", life, opp.Life, hand, me.Hand.Size())
	}
}

func TestOldB09VikyaAndPaulTrain(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Vikya, Scorching Stalwart", oldB09Vikya}, {"Paul Atreides, Ducal Heir", oldB09Paul},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := effectstest.B12Push(g, me.ID, tc.name, "Legendary Creature — Human", tc.oracle, 1, 1)
			big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 5, 5)
			effectstest.AttackWith(t, g, opp.ID, id, big)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
				t.Errorf("+1/+1 counters: %d, want 1", got)
			}
		})
	}
}

func TestOldB09PaulNeedsPowerFourToActivate(t *testing.T) {
	for _, tc := range []struct {
		power  int
		wantOK bool
	}{{3, false}, {4, true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.ToMain(t, g)
		paul := effectstest.B12Push(g, me.ID, "Paul Atreides, Ducal Heir", "Legendary Creature — Human Noble", oldB09Paul, tc.power, 2)
		other := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
		err := g.ActivateCatalogAbility(me.ID, paul, 0, game.ActivateAbilityParams{Strict: true})
		if (err == nil) != tc.wantOK {
			t.Fatalf("power %d: activation err = %v, want ok=%v", tc.power, err, tc.wantOK)
		}
		if !tc.wantOK {
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterCount(g, other, game.CounterPlusOne); got != 1 {
			t.Errorf("ally counters: %d, want 1", got)
		}
		if got := effectstest.CounterCount(g, paul, game.CounterPlusOne); got != 0 {
			t.Errorf("Paul's own counters: %d, want 0 (each OTHER creature)", got)
		}
	}
}

func TestOldB09ChaoticGooEntersWithThreeCountersAndFlipsEachUpkeep(t *testing.T) {
	spec, _ := effects.Lookup(oldB09Goo)
	if len(spec.Replacements) != 1 || len(spec.Triggered) != 1 {
		t.Fatalf("replacements %d triggered %d", len(spec.Replacements), len(spec.Triggered))
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	id := effectstest.B12PlayFromHand(t, g, "Chaotic Goo", "Creature — Ooze", oldB09Goo, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("entered with %d counters, want 3", got)
	}
	_ = me
}

func TestOldB09SageOfHoursTakesAnExtraTurnPerFiveCounters(t *testing.T) {
	for _, tc := range []struct{ counters, turns int }{{4, 0}, {5, 1}, {11, 2}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.ToMain(t, g)
		id := effectstest.B12Push(g, me.ID, "Sage of Hours", "Creature — Human Wizard", oldB09Sage, 1, 1)
		_ = g.AddCounter(id, game.CounterPlusOne, tc.counters)
		before := len(g.ExtraTurnsQueued())
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := len(g.ExtraTurnsQueued()) - before; got != tc.turns {
			t.Errorf("%d counters: %d extra turns, want %d", tc.counters, got, tc.turns)
		}
		if c := effectstest.CounterCount(g, id, game.CounterPlusOne); c != 0 {
			t.Errorf("counters left: %d", c)
		}
	}
}

func TestOldB09ParallaxTideExilesLandsAndReturnsThemWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	tide := effectstest.B12Push(g, me.ID, "Parallax Tide", "Enchantment", oldB09ParallaxTd, 0, 0)
	_ = g.AddCounter(tide, game.CounterFade, 5)
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, tide, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, land); got != game.ZoneExile {
		t.Fatalf("land is in %s, want exile", got)
	}
	if got := effectstest.CounterCount(g, tide, game.CounterFade); got != 4 {
		t.Errorf("fade counters %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(tide) })
	effectstest.PassPriorityAroundTable(t, g)
	// It returns as a new object (CR 400.7) under its owner's control.
	if got := effectstest.CountOnBattlefield(g, "Island", opp.ID); got != 1 {
		t.Errorf("Islands on the battlefield under their owner: %d, want 1", got)
	}
}

func TestOldB09ChaoticGooFlipsAtUpkeepAndGrowsOrShrinks(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 60 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.ToMain(t, g)
		goo := effectstest.B12Push(g, me.ID, "Chaotic Goo", "Creature — Ooze", oldB09Goo, 0, 0)
		_ = g.AddCounter(goo, game.CounterPlusOne, 3)
		effectstest.B12ToMyNextUpkeep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.LatestTriggerPrompt(g, me.ID)
		if prompt == nil {
			t.Fatal("no optional-trigger prompt")
		}
		if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins flipped, want 1", len(evs)-flips)
		}
		got := effectstest.CounterCount(g, goo, game.CounterPlusOne)
		if evs[len(evs)-1].Won {
			won++
			if got != 4 {
				t.Errorf("won the flip: %d counters, want 4", got)
			}
		} else {
			lost++
			if got != 2 {
				t.Errorf("lost the flip: %d counters, want 2", got)
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("never saw both outcomes (won %d, lost %d)", won, lost)
	}
}

func TestOldB09ZethiExilesKickedInstantsAndCopiesThemWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.GyCard(me, "Shock", "Instant", "{R}", "")
	bolt2 := effectstest.GyCard(me, "Opt", "Instant", "{U}", "")
	sorc := effectstest.GyCard(me, "Rock", "Sorcery", "{1}", "")
	zethi, err := effectstest.Msc17Cast(t, g, me, game.Card{
		Name: "Zethi, Arcane Blademaster", TypeLine: "Legendary Creature — Human Soldier",
		OracleID: oldB09Zethi, ManaCost: "{1}{W}{U}", Power: 3, Toughness: 3, Keywords: []string{"haste"},
	}, game.CastSpellParams{OptionalCosts: []int{0}})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: bolt}, {Kind: game.TargetCard, ID: bolt2},
	}); err == nil {
		t.Fatal("'up to X' with X=1 accepted two targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCountInExile(g, bolt, "kick"); got != 1 {
		t.Fatalf("the exiled instant has %d kick counters, want 1", got)
	}
	if effectstest.CounterCountInExile(g, bolt2, "kick") != -1 || effectstest.B12ZoneOf(g, sorc) != game.ZoneGraveyard {
		t.Error("only the chosen instant should have been exiled")
	}
	// Attack: the exiled kicked card is copied and offered.
	effectstest.AttackWithTrigger(t, g, zethi, opp.ID, uuid.Nil)
	offer := effectstest.ResolutionCastOfferFor(g, me.ID)
	if offer == nil {
		t.Fatalf("no cast offer for the copies: %+v", g.PendingChoices)
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
}
