package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pirates batch (#5368, cmdr2 batch 6).

const (
	cmdr2B06BeckettOracle   = "43b5e462-d860-473d-828f-6c513fc7768a"
	cmdr2B06RavagerOracle   = "b83080af-9616-4dc3-bc47-6ef84b9ccc1b"
	cmdr2B06RamirezOracle   = "a3b8690e-ac47-444f-924e-ab2a12a85431"
	cmdr2B06BrassOracle     = "4df179b4-7c6b-4e63-9d7d-62f4df377989"
	cmdr2B06FranciscoOracle = "1dbe4eb4-8313-459a-89ce-982442cd0cd0"
	cmdr2B06SaboteurOracle  = "d36f64a0-53f7-4c56-9f40-a7927a5c2608"
	cmdr2B06CaptainOracle   = "7bcf3090-95c6-4384-a857-9fe571e6da85"
	cmdr2B06WarkiteOracle   = "8b566757-e5e4-4d47-b5ef-80ed7d853679"
	cmdr2B06SkeletonOracle  = "2536750c-3c4f-4e00-b517-680ad4fd2ecc"
	cmdr2B06RaidersOracle   = "8b388331-eaa7-41aa-99de-4451e51cb221"
)

func TestCmdr2B06Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2B06BeckettOracle:   "Admiral Beckett Brass",
		cmdr2B06RavagerOracle:   "Dire Fleet Ravager",
		cmdr2B06RamirezOracle:   "Ramirez DePietro, Pillager",
		cmdr2B06BrassOracle:     "Admiral Brass, Unsinkable",
		cmdr2B06FranciscoOracle: "Francisco, Fowl Marauder",
		cmdr2B06SaboteurOracle:  "Daring Saboteur",
		cmdr2B06CaptainOracle:   "Fathom Fleet Captain",
		cmdr2B06WarkiteOracle:   "Warkite Marauder",
		cmdr2B06SkeletonOracle:  "Skeleton Crew",
		cmdr2B06RaidersOracle:   "Merchant Raiders",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || spec.Name != name {
			t.Errorf("%s: %+v %v", name, spec, ok)
		}
	}
}

// --- Daring Saboteur ------------------------------------------------

func TestCmdr2B06SaboteurPaysToBecomeUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sab := effectstest.Cmdr2B05Card(g, me.ID, "Daring Saboteur", "Creature — Human Pirate", cmdr2B06SaboteurOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.AssertRestrictions(t, g, sab, 0)
	if err := g.ActivateCatalogAbility(me.ID, sab, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatForTest(g, me, "CCU")
	effectstest.B16Activate(t, g, me.ID, sab, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, sab, game.CantBeBlocked)
}

func TestCmdr2B06SaboteurLootsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 8)
	sab := effectstest.Cmdr2B05Card(g, me.ID, "Daring Saboteur", "Creature — Human Pirate", cmdr2B06SaboteurOracle, 2, 1)
	effectstest.DealCombatDamageToPlayer(g, sab, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no loot prompt")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

// --- Dire Fleet Ravager -----------------------------------------------

func TestCmdr2B06RavagerEveryPlayerLosesAThirdRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lives := []int{20, 10, 1, 7, 40}
	for i, p := range g.Seats {
		if i < len(lives) {
			p.Life = lives[i]
		}
	}
	want := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		want[i] = p.Life - (p.Life+2)/3
	}
	effectstest.CastCatalogSpell(t, g, "Dire Fleet Ravager", "Creature — Orc Pirate Wizard", cmdr2B06RavagerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Life != want[i] {
			t.Errorf("seat %d (%s) has %d life, want %d", i, map[bool]string{true: "controller", false: "opponent"}[p.ID == me.ID], p.Life, want[i])
		}
	}
	if g.Seats[0].Life != 13 || g.Seats[1].Life != 6 || g.Seats[2].Life != 0 {
		t.Errorf("20→13, 10→6 and 1→0 expected, got %d %d %d", g.Seats[0].Life, g.Seats[1].Life, g.Seats[2].Life)
	}
}

// --- Warkite Marauder -------------------------------------------------

func TestCmdr2B06WarkiteStripsAbilitiesAndSetsBasePT(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wm := effectstest.Cmdr2B05Card(g, me.ID, "Warkite Marauder", "Creature — Human Pirate", cmdr2B06WarkiteOracle, 2, 1)
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Flyer", TypeLine: "Creature — Dragon", Keywords: []string{"flying"},
		Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(victim, game.CounterPlusOne, 1) })
	effectstest.DeclareAttack(t, g, opp.ID, wm)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.B40PickCard(t, g, me.ID, victim)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B39CurrentPower(t, g, victim); got != 1 {
		t.Errorf("victim power %d, want 1 (base 0 plus its counter)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, victim); got != 1 {
		t.Errorf("victim toughness %d, want base 1", got)
	}
	if effectstest.HasKeywordOf(g, victim, "flying") {
		t.Error("the victim kept flying")
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("an unrelated creature changed: %d", got)
	}
	// Until end of turn.
	effectstest.B39NextTurnOf(t, g, 1)
	if got := effectstest.EffectivePower(t, g, victim); got != 5 {
		t.Errorf("the effect outlived the turn: power %d", got)
	}
}

// --- Skeleton Crew ----------------------------------------------------

func TestCmdr2B06SkeletonCrewPumpsOtherSkeletonsAndPiratesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crew := effectstest.Cmdr2B05Card(g, me.ID, "Skeleton Crew", "Creature — Skeleton Pirate", cmdr2B06SkeletonOracle, 3, 3)
	pirate := effectstest.B12Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 2, 2)
	skeleton := effectstest.B12Creature(g, me.ID, "Skeleton", "Creature — Skeleton", 1, 1)
	both := effectstest.B12Creature(g, me.ID, "Both", "Creature — Skeleton Pirate", 2, 2)
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Pirate", "Creature — Human Pirate", 2, 2)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{
		"Pirate": {pirate, 3}, "Skeleton": {skeleton, 2}, "Both (once)": {both, 3},
		"Bear": {bear, 2}, "Their Pirate": {theirs, 2}, "Crew itself": {crew, 3},
	} {
		if got := effectstest.EffectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s power %d, want %d", name, got, c.want)
		}
	}
}

func TestCmdr2B06SkeletonCrewMakesOneTokenPerBatchOfCreatureCardsLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Cmdr2B05Card(g, me.ID, "Skeleton Crew", "Creature — Skeleton Pirate", cmdr2B06SkeletonOracle, 3, 3)
	var dead []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dead Thing", TypeLine: "Creature — Zombie", Owner: me.ID, Controller: me.ID})
		dead = append(dead, id)
	}
	sorcery := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Old Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(sorcery) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.BattlefieldIDsNamed(g, "Skeleton Pirate")); got != 0 {
		t.Fatalf("a sorcery card leaving made %d tokens", got)
	}
	g.WithWriteLock(func() { _ = g.ExileCardsForEffect(dead) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.BattlefieldIDsNamed(g, "Skeleton Pirate")); got != 1 {
		t.Errorf("two creature cards leaving at once made %d tokens, want 1", got)
	}
}

func TestCmdr2B06SkeletonCrewReturnsTappedFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Skeleton Crew", "Creature — Skeleton Pirate", cmdr2B06SkeletonOracle, 3, 3, "{C}{C}{C}{C}{C}{B}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatal("Skeleton Crew did not return")
	}
	if !c.Tapped {
		t.Error("it must enter tapped")
	}
}

// --- Merchant Raiders -------------------------------------------------

func TestCmdr2B06MerchantRaidersHoldsATargetWhenItAndOtherPiratesEnter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	first := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "First")
	second := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Second")
	third := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Third")

	raiders := effectstest.CastHoldCreature(t, g, "Merchant Raiders", "Creature — Human Pirate", cmdr2B06RaidersOracle, 2, 4)
	effectstest.AnswerPickTarget(t, g, first)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, first) || effectstest.HoldCount(t, g, first) != 1 {
		t.Fatalf("the Raiders' own entry: tapped %v, holds %d", effectstest.TappedForTest(t, g, first), effectstest.HoldCount(t, g, first))
	}

	// Another Pirate you control entering triggers it again.
	effectstest.CastCatalogSpell(t, g, "Deckhand", "Creature — Human Pirate", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, second)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, second) || effectstest.HoldCount(t, g, second) != 1 {
		t.Fatalf("another Pirate entering: tapped %v, holds %d", effectstest.TappedForTest(t, g, second), effectstest.HoldCount(t, g, second))
	}

	// A non-Pirate does not.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("a Bear entering triggered the Raiders")
		}
	}
	if effectstest.TappedForTest(t, g, third) {
		t.Error("an untargeted creature was tapped")
	}

	// The hold lasts only while you control the Raiders.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, raiders); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TappedForTest(t, g, first) || effectstest.TappedForTest(t, g, second) {
		t.Error("the held creatures stayed tapped after the Raiders left")
	}
}

// --- Fathom Fleet Captain ----------------------------------------------

func cmdr2B06CaptainAttack(t *testing.T, otherPirate string, token bool) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.Cmdr2B05Card(g, me.ID, "Fathom Fleet Captain", "Creature — Human Pirate", cmdr2B06CaptainOracle, 2, 1)
	if otherPirate != "" {
		if token {
			otherPirate = "Token " + otherPirate
		}
		c := game.Card{InstanceID: uuid.New(), Name: "Deckhand", TypeLine: otherPirate, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
	}
	effectstest.DeclareAttack(t, g, opp.ID, cap)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestCmdr2B06FleetCaptainPaysTwoForAMenacePirate(t *testing.T) {
	g, me := cmdr2B06CaptainAttack(t, "Creature — Human Pirate", false)
	effectstest.B06AddMana(me, "C", "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("no payment prompt with another nontoken Pirate")
	}
	if ask.PayCost != "{2}" {
		t.Errorf("optional cost %s", ask.PayCost)
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Pirate")
	if len(ids) != 1 {
		t.Fatalf("%d Pirate tokens, want 1", len(ids))
	}
	tok, _ := g.LookupCardForEffect(ids[0])
	if tok.Power != 2 || tok.Toughness != 2 || !effectstest.HasKeywordOf(g, ids[0], "menace") {
		t.Errorf("token %d/%d, menace %v", tok.Power, tok.Toughness, effectstest.HasKeywordOf(g, ids[0], "menace"))
	}
}

func TestCmdr2B06FleetCaptainNeedsAnotherNontokenPirate(t *testing.T) {
	for name, c := range map[string]struct {
		typeLine string
		token    bool
	}{"alone": {"", false}, "only a token Pirate": {"Creature — Human Pirate", true}, "only a non-Pirate": {"Creature — Bear", false}} {
		g, _ := cmdr2B06CaptainAttack(t, c.typeLine, c.token)
		if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
			t.Errorf("%s: the trigger asked for {2}", name)
		}
	}
}

func TestCmdr2B06FleetCaptainDecliningMakesNothing(t *testing.T) {
	g, me := cmdr2B06CaptainAttack(t, "Creature — Human Pirate", false)
	effectstest.B06AddMana(me, "C", "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if got := len(effectstest.BattlefieldIDsNamed(g, "Pirate")); got != 0 {
		t.Errorf("%d tokens after declining", got)
	}
}

// --- Ramirez DePietro, Pillager ----------------------------------------

func TestCmdr2B06RamirezEntersLosesTwoAndMakesTwoTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Ramirez DePietro, Pillager", "Legendary Creature — Human Pirate", cmdr2B06RamirezOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	if got := len(effectstest.BattlefieldIDsNamed(g, "Treasure")); got != 2 {
		t.Errorf("%d Treasures, want 2", got)
	}
}

func TestCmdr2B06RamirezExilesTheVictimsTopCardAndLetsYouCastItForAsLongAsItStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Cmdr2B05Card(g, me.ID, "Ramirez DePietro, Pillager", "Legendary Creature — Human Pirate", cmdr2B06RamirezOracle, 4, 3)
	pirate := effectstest.B12Creature(g, me.ID, "Deckhand", "Creature — Human Pirate", 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	top := effectstest.SeedTopOfLibrary(opp, game.Card{Name: "Stolen Spell", TypeLine: "Creature — Bear", ManaCost: "{2}{R}"})

	// Not a Pirate: nothing.
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InExile(g, top) {
		t.Fatal("a non-Pirate connecting exiled a card")
	}
	// Two Pirates at once: one trigger, one card.
	effectstest.DealCombatDamageToPlayer(g, pirate, opp.ID, 1)
	effectstest.DealCombatDamageToPlayer(g, pirate, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, top) {
		t.Fatal("the top card of the damaged player's library was not exiled")
	}
	perm := effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile)
	if perm == nil || !perm.Granted() {
		t.Fatal("no permission to cast the exiled card")
	}
}

// --- Francisco, Fowl Marauder -------------------------------------------

func TestCmdr2B06FranciscoCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.Cmdr2B05Card(g, g.Seats[0].ID, "Francisco, Fowl Marauder", "Legendary Creature — Bird Pirate", cmdr2B06FranciscoOracle, 0, 1)
	effectstest.AssertRestrictions(t, g, id, game.CantBlock)
}

func cmdr2B06Damage(g *game.Game, src, victim uuid.UUID, combat bool) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: src, Target: victim, Amount: 1, Combat: combat})
	})
}

func TestCmdr2B06FranciscoExploresOnAnyPirateDamageToAPlayerOncePerBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Cmdr2B05Card(g, me.ID, "Francisco, Fowl Marauder", "Legendary Creature — Bird Pirate", cmdr2B06FranciscoOracle, 0, 1)
	pirate := effectstest.B12Creature(g, me.ID, "Deckhand", "Creature — Human Pirate", 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})

	cmdr2B06Damage(g, bear, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(land) {
		t.Fatal("a non-Pirate's damage made Francisco explore")
	}
	// Non-combat damage from two Pirates in one batch: one explore.
	cmdr2B06Damage(g, pirate, opp.ID, false)
	cmdr2B06Damage(g, pirate, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Fatal("Pirate damage did not make Francisco explore (a land goes to hand)")
	}
	if me.Library.Size() != 0 {
		t.Errorf("more than one explore: library %d", me.Library.Size())
	}
}

// --- Admiral Brass, Unsinkable ------------------------------------------

func TestCmdr2B06AdmiralBrassMillsFourWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.Msc18Library(me, 10)
	gy := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Admiral Brass, Unsinkable", "Legendary Creature — Human Pirate", cmdr2B06BrassOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - gy; got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
}

func cmdr2B06BrassCombat(t *testing.T) (g *game.Game, me *game.Player, pirate, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.Cmdr2B05Card(g, me.ID, "Admiral Brass, Unsinkable", "Legendary Creature — Human Pirate", cmdr2B06BrassOracle, 3, 3)
	pirate = effectstest.PushCatalogGraveyardCard(me, "Dead Pirate", "Creature — Human Pirate", "", 2, 1)
	bear = effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	return
}

func TestCmdr2B06AdmiralBrassReturnsAPirateWithFinalityBase44AndHaste(t *testing.T) {
	g, me, pirate, bear := cmdr2B06BrassCombat(t)
	if opts := effectstest.PickTargetPrompt(t, g).PickTargetCards; len(opts) != 1 || opts[0] != pirate {
		t.Fatalf("legal targets %v, want only the Pirate card %v (not %v)", opts, pirate, bear)
	}
	effectstest.AnswerPickTarget(t, g, pirate)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCardByID(g, pirate); !ok {
		t.Fatal("the Pirate did not return")
	}
	if got := effectstest.CounterOn(g, pirate, game.CounterFinality); got != 1 {
		t.Errorf("finality counters %d, want 1", got)
	}
	if got := effectstest.EffectivePower(t, g, pirate); got != 4 {
		t.Errorf("power %d, want base 4", got)
	}
	if got := effectstest.EffectiveToughness(t, g, pirate); got != 4 {
		t.Errorf("toughness %d, want base 4", got)
	}
	if !effectstest.HasKeywordOf(g, pirate, "haste") {
		t.Error("no haste")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the non-Pirate card left the graveyard")
	}
}

func TestCmdr2B06AdmiralBrassMayDecline(t *testing.T) {
	g, me, pirate, _ := cmdr2B06BrassCombat(t)
	effectstest.AnswerPickTarget(t, g, pirate)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCardByID(g, pirate); ok {
		t.Error("declining still returned the Pirate")
	}
}

// --- Admiral Beckett Brass ----------------------------------------------

func TestCmdr2B06BeckettPumpsOtherPiratesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beckett := effectstest.Cmdr2B05Card(g, me.ID, "Admiral Beckett Brass", "Legendary Creature — Human Pirate", cmdr2B06BeckettOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Human Pirate", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Human Pirate", 2, 2)
	bear := effectstest.SeedBear(g, me.ID)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"other Pirate": {mine, 3}, "itself": {beckett, 3}, "their Pirate": {theirs, 2}, "Bear": {bear, 2}} {
		if got := effectstest.EffectivePower(t, g, c.id); got != c.want {
			t.Errorf("%s power %d, want %d", name, got, c.want)
		}
	}
}

func cmdr2B06BeckettEndStep(t *testing.T, hits int) (g *game.Game, me *game.Player, rock, myThing uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	opp := g.Seats[1]
	effectstest.Cmdr2B05Card(g, me.ID, "Admiral Beckett Brass", "Legendary Creature — Human Pirate", cmdr2B06BeckettOracle, 3, 3)
	rock = effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	myThing = effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	for i := 0; i < hits; i++ {
		p := effectstest.B12Creature(g, me.ID, "Deckhand", "Creature — Human Pirate", 1, 1)
		effectstest.DealCombatDamageToPlayer(g, p, opp.ID, 1)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	return
}

func TestCmdr2B06BeckettStealsFromAPlayerHitByThreePirates(t *testing.T) {
	g, me, rock, mine := cmdr2B06BeckettEndStep(t, 3)
	opts := effectstest.PickTargetPrompt(t, g).PickTargetCards
	for _, id := range opts {
		if c, ok := g.LookupCardForEffect(id); !ok || c.Controller == me.ID || c.Name == "Their Island" {
			t.Fatalf("illegal target offered: %v", c.Name)
		}
	}
	if !effectstest.SlicesContainsID(opts, rock) || effectstest.SlicesContainsID(opts, mine) {
		t.Fatalf("targets %v: want Their Rock and not My Rock", opts)
	}
	effectstest.AnswerPickTarget(t, g, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(rock); !ok || c.Controller != me.ID {
		t.Error("Their Rock is not under my control")
	}
}

func TestCmdr2B06BeckettNeedsThreePirates(t *testing.T) {
	g, me, rock, _ := cmdr2B06BeckettEndStep(t, 2)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("two Pirates offered a target")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := g.LookupCardForEffect(rock); c.Controller == me.ID {
		t.Error("the permanent was taken with only two Pirates")
	}
}
