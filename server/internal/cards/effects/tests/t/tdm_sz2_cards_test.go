package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_sz2_cards_test.go — Tarkir: Dragonstorm, the second batch of cards
// from S to Z (#4463). Each test plays the real catalog entry and asserts
// what its text says.

const (
	tdmSZ2SageOfTheFangOracle  = "d6eb2bfc-1e86-4e36-9d48-12de16955b98"
	tdmSZ2HeadlinerOracle      = "d8cfa20a-343e-4661-9f5a-a8265f6917a8"
	tdmSZ2SkyriderOracle       = "04e615e6-4088-4170-904c-f2d4bcebed59"
	tdmSZ2StormbeaconOracle    = "f4b3ee43-fc01-425f-8322-29bf6d16cb3f"
	tdmSZ2TempestHawkOracle    = "7423b3b9-56eb-4cf2-8ada-135918219c4b"
	tdmSZ2UnderdogsOracle      = "3740b86d-b0e6-4837-b408-e7f62b95c787"
	tdmSZ2UnrootedOracle       = "de5c6c9d-04a2-4287-8c6e-fdf56c411507"
	tdmSZ2BoltcasterOracle     = "0a56a33b-033f-4398-8dc1-e531be059710"
	tdmSZ2IceClimberOracle     = "ccaadf8d-22c0-4139-b734-03ac2fcb4494"
	tdmSZ2WayspeakerOracle     = "b93eaf66-91b3-44da-851c-79810901be0c"
	tdmSZ2WingspanStrideOracle = "e75cee5d-ab02-4227-b36e-61c5d2f564a8"
	tdmSZ2WorthyCostOracle     = "0f8ebf11-b5ec-4f21-8c49-90d569b5a4c5"
)

func TestTdmSZ2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmSZ2SageOfTheFangOracle:  "Sage of the Fang",
		tdmSZ2HeadlinerOracle:      "Stadium Headliner",
		tdmSZ2SkyriderOracle:       "Starry-Eyed Skyrider",
		tdmSZ2StormbeaconOracle:    "Stormbeacon Blade",
		tdmSZ2TempestHawkOracle:    "Tempest Hawk",
		tdmSZ2UnderdogsOracle:      "Underfoot Underdogs",
		tdmSZ2UnrootedOracle:       "Unrooted Ancestor",
		tdmSZ2BoltcasterOracle:     "Unsparing Boltcaster",
		tdmSZ2IceClimberOracle:     "Veteran Ice Climber",
		tdmSZ2WayspeakerOracle:     "Wayspeaker Bodyguard",
		tdmSZ2WingspanStrideOracle: "Wingspan Stride",
		tdmSZ2WorthyCostOracle:     "Worthy Cost",
	} {
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

// --- fixtures ------------------------------------------------------------------

// tdmSZ2Damage deals `n` noncombat damage to `target` from a source the
// test makes up, so the target "was dealt damage this turn".
func tdmSZ2Damage(t *testing.T, g *game.Game, by, target uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: by, Owner: by})
		if err := (effects.DealDamage{Source: uuid.Nil, Target: target, Amount: n}).Apply(ctx); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
}

// --- Sage of the Fang ----------------------------------------------------------

func TestTdmSZ2SageOfTheFangEntersWithACounterForATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Sage of the Fang", TypeLine: "Creature — Human Druid",
		OracleID: tdmSZ2SageOfTheFangOracle, Power: 2, Toughness: 2})
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("entering asked for no target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestTdmSZ2SageOfTheFangRenewsAndDoubles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID})
		if err := (effects.AddCounter{Target: bear, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
			t.Fatalf("seed counters: %v", err)
		}
	})
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Sage of the Fang", "Creature — Human Druid",
		tdmSZ2SageOfTheFangOracle, 2, 2, "{G}{G}{G}{G}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Sage")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 6 {
		t.Errorf("%d +1/+1 counters, want 6 (2, plus one, then doubled)", got)
	}
}

func TestTdmSZ2SageOfTheFangRenewIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushBattlefieldCreature(g, me.ID)
	id := effectstest.PushCatalogGraveyardCard(me, "Sage of the Fang", "Creature — Human Druid", tdmSZ2SageOfTheFangOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "G", 4)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Error("renew was activated in combat")
	}
}

// --- Stadium Headliner ---------------------------------------------------------

func TestTdmSZ2StadiumHeadlinerDealsDamageEqualToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	headliner := effectstest.B12Push(g, me.ID, "Stadium Headliner", "Creature — Goblin Warrior", tdmSZ2HeadlinerOracle, 1, 1)
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}
	// Three creatures besides the Headliner: it is sacrificed as the cost,
	// so the damage is 3, enough for a 3-toughness creature and no more.
	three := effectstest.B12Creature(g, opp.ID, "Three", "Creature — Wall", 0, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, headliner, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: three}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(headliner) {
		t.Fatal("the Headliner was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) {
		t.Error("3 damage did not kill the 0/3")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], g2.Seats[(g2.Turn.ActiveSeat+1)%len(g2.Seats)]
	headliner2 := effectstest.B12Push(g2, me2.ID, "Stadium Headliner", "Creature — Goblin Warrior", tdmSZ2HeadlinerOracle, 1, 1)
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	}
	four := effectstest.B12Creature(g2, opp2.ID, "Four", "Creature — Wall", 0, 4)
	effectstest.AdvanceToMain(t, g2)
	effectstest.FillPoolColored(me2, "R", 2)
	if err := g2.ActivateCatalogAbility(me2.ID, headliner2, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: four}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(four) {
		t.Error("the Headliner counted itself: a 0/4 died to what should be 3 damage")
	}
}

func TestTdmSZ2StadiumHeadlinerMobilizesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	headliner := effectstest.B12Push(g, me.ID, "Stadium Headliner", "Creature — Goblin Warrior", tdmSZ2HeadlinerOracle, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, headliner)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TdmPlaceWarriors(t, g, me.ID)
	if n := len(effectstest.SocB8Named(g, me.ID, "Warrior")); n != 1 {
		t.Fatalf("mobilize 1 made %d Warriors", n)
	}
}

// --- Starry-Eyed Skyrider ------------------------------------------------------

func TestTdmSZ2StarryEyedSkyriderGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skyrider := effectstest.B12Push(g, me.ID, "Starry-Eyed Skyrider", "Creature — Human Scout", tdmSZ2SkyriderOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, skyrider, "flying") {
		t.Error("Starry-Eyed Skyrider has no flying")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := effectstest.PushTokens(g, me.ID, effects.TokenCard("1/1 red Goblin"), 1)[0]
	attacker := effectstest.PushTokens(g, me.ID, effects.TokenCard("1/1 red Goblin"), 1)[0]
	effectstest.DeclareAttack(t, g, opp.ID, skyrider, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("attacking asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, skyrider) {
		t.Error("the Skyrider offered itself as \"another\" target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the targeted creature did not gain flying")
	}
	if !effectstest.HasEffectiveKeyword(t, g, attacker, "flying") {
		t.Error("an attacking token has no flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, idle, "flying") {
		t.Error("a token that is not attacking has flying")
	}
}

// --- Stormbeacon Blade ---------------------------------------------------------

func TestTdmSZ2StormbeaconBladeDrawsWithThreeAttackers(t *testing.T) {
	for _, tc := range []struct {
		name      string
		attackers int
		draws     int
	}{
		{"two attackers", 2, 0},
		{"three attackers", 3, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			bearer := effectstest.B12Creature(g, me.ID, "Bearer", "Creature — Bear", 2, 2)
			blade := effectstest.SeedEquipment(g, me.ID, "Stormbeacon Blade", tdmSZ2StormbeaconOracle)
			effectstest.BrAttach(t, g, blade, bearer)
			if got := effectstest.EffectivePower(t, g, bearer); got != 5 {
				t.Errorf("equipped power %d, want 5", got)
			}
			attackers := []uuid.UUID{bearer}
			for i := 1; i < tc.attackers; i++ {
				attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2))
			}
			hand := me.Hand.Size()
			effectstest.DeclareAttack(t, g, opp.ID, attackers...)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.draws {
				t.Errorf("drew %d, want %d", got, tc.draws)
			}
		})
	}
}

// --- Tempest Hawk --------------------------------------------------------------

func TestTdmSZ2TempestHawkFetchesAnotherHawk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := effectstest.B12Push(g, me.ID, "Tempest Hawk", "Creature — Bird", tdmSZ2TempestHawkOracle, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, hawk, "flying") {
		t.Error("Tempest Hawk has no flying")
	}
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Tempest Owl", TypeLine: "Creature — Bird"},
		game.Card{Name: "Tempest Hawk", TypeLine: "Creature — Bird", OracleID: tdmSZ2TempestHawkOracle},
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, hawk, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search after agreeing")
	}
	if effectstest.SearchOptionNamed(g, c, "Tempest Owl") != uuid.Nil {
		t.Error("a card not named Tempest Hawk was offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Tempest Hawk")
	found := false
	for _, card := range me.Hand.Cards {
		if card.Name == "Tempest Hawk" {
			found = true
		}
	}
	if !found {
		t.Error("the second Tempest Hawk is not in hand")
	}
}

func TestTdmSZ2TempestHawkDeclinedSearchesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hawk := effectstest.B12Push(g, me.ID, "Tempest Hawk", "Creature — Bird", tdmSZ2TempestHawkOracle, 2, 2)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Tempest Hawk", TypeLine: "Creature — Bird", OracleID: tdmSZ2TempestHawkOracle})
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, hawk, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("declining still searched")
	}
}

// --- Underfoot Underdogs -------------------------------------------------------

func TestTdmSZ2UnderfootUnderdogsMakeAGoblinAndSneakASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dogs := effectstest.Msc17Play(t, g, me, game.Card{Name: "Underfoot Underdogs", TypeLine: "Creature — Goblin Warrior",
		OracleID: tdmSZ2UnderdogsOracle, Power: 1, Toughness: 2})
	goblins := effectstest.SocB8Named(g, me.ID, "Goblin")
	if len(goblins) != 1 || goblins[0].Power != 1 || goblins[0].Toughness != 1 ||
		len(goblins[0].Colors) != 1 || goblins[0].Colors[0] != "R" {
		t.Fatalf("Goblins %+v, want one 1/1 red Goblin", goblins)
	}
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	// The Underdogs entered this turn and can't tap yet; a fresh copy
	// already on the battlefield can.
	_ = dogs
	ready := effectstest.B12Push(g, me.ID, "Underfoot Underdogs", "Creature — Goblin Warrior", tdmSZ2UnderdogsOracle, 1, 2)
	effectstest.FillPoolColored(me, "R", 2)
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err == nil {
		t.Fatal("a 3-power creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, ready, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, small)&game.CantBeBlocked == 0 {
		t.Error("the 2-power creature can still be blocked")
	}
}

// --- Unrooted Ancestor ---------------------------------------------------------

func TestTdmSZ2UnrootedAncestorSacrificesForIndestructibleAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ancestor := effectstest.B12Push(g, me.ID, "Unrooted Ancestor", "Creature — Spirit Cleric", tdmSZ2UnrootedOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, ancestor, "flash") {
		t.Error("Unrooted Ancestor has no flash")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 2)
	if err := g.ActivateCatalogAbility(me.ID, ancestor, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ancestor}}); err == nil {
		t.Fatal("the Ancestor sacrificed itself to its \"another creature\" cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, ancestor, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the other creature was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, ancestor, "indestructible") {
		t.Error("the Ancestor did not gain indestructible")
	}
	if !effectstest.IsTapped(g, ancestor) {
		t.Error("the Ancestor did not tap")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ancestor) })
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(ancestor) {
		t.Error("an indestructible Ancestor was destroyed")
	}
}

// --- Unsparing Boltcaster ------------------------------------------------------

func TestTdmSZ2UnsparingBoltcasterFinishesADamagedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurt := effectstest.B12Creature(g, opp.ID, "Hurt", "Creature — Giant", 5, 6)
	fresh := effectstest.B12Creature(g, opp.ID, "Fresh", "Creature — Giant", 5, 6)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Giant", 5, 6)
	effectstest.AdvanceToMain(t, g)
	tdmSZ2Damage(t, g, me.ID, hurt, 1)
	tdmSZ2Damage(t, g, me.ID, mine, 1)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Unsparing Boltcaster", TypeLine: "Creature — Ogre Wizard",
		OracleID: tdmSZ2BoltcasterOracle, Power: 3, Toughness: 3})
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("entering asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, fresh) || effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("an undamaged creature, or your own, was offered")
	}
	effectstest.PickCard(t, g, me.ID, hurt)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hurt) {
		t.Error("1 + 5 damage did not kill the 5/6")
	}
}

func TestTdmSZ2UnsparingBoltcasterWithNoDamagedCreatureAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, opp.ID, "Fresh", "Creature — Giant", 5, 6)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Unsparing Boltcaster", TypeLine: "Creature — Ogre Wizard",
		OracleID: tdmSZ2BoltcasterOracle, Power: 3, Toughness: 3})
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("asked for a target with no damaged creature")
	}
}

// --- Veteran Ice Climber -------------------------------------------------------

func TestTdmSZ2VeteranIceClimberMillsByPowerAndCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	climber := effectstest.B12Push(g, me.ID, "Veteran Ice Climber", "Creature — Human Scout", tdmSZ2IceClimberOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, climber, "vigilance") {
		t.Error("Veteran Ice Climber has no vigilance")
	}
	if effectstest.RestrictionsOf(t, g, climber)&game.CantBeBlocked == 0 {
		t.Error("Veteran Ice Climber can be blocked")
	}
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID})
		if err := (effects.AddCounter{Target: climber, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
			t.Fatalf("seed counters: %v", err)
		}
	})
	effectstest.Msc18Library(opp, 6)
	graveyard := opp.Graveyard.Size()
	effectstest.DeclareAttack(t, g, opp.ID, climber)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - graveyard; got != 3 {
		t.Errorf("milled %d, want 3 (power 1 plus two counters)", got)
	}
	if effectstest.IsTapped(g, climber) {
		t.Error("a vigilant Climber tapped to attack")
	}
}

// --- Wayspeaker Bodyguard ------------------------------------------------------

func TestTdmSZ2WayspeakerBodyguardReturnsACheapPermanentCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.PushGraveyardPermanent(me, "Cheap Charm", "Artifact", "{2}")
	dear := effectstest.PushGraveyardPermanent(me, "Dear Charm", "Artifact", "{3}")
	land := effectstest.PushGraveyardPermanent(me, "Old Forest", "Land", "")
	spell := effectstest.PushGraveyardPermanent(me, "Old Shock", "Instant", "{R}")
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Wayspeaker Bodyguard", TypeLine: "Creature — Orc Monk",
		OracleID: tdmSZ2WayspeakerOracle, Power: 3, Toughness: 4})
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("entering asked for no target")
	}
	for _, id := range []uuid.UUID{dear, land, spell} {
		if effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("%s was offered", id)
		}
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Error("the mana value 2 permanent card is not in hand")
	}
}

func TestTdmSZ2WayspeakerBodyguardFlurryTapsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Wayspeaker Bodyguard", "Creature — Orc Monk", tdmSZ2WayspeakerOracle, 3, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B24CastNoncreature(t, g, "First", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the first spell triggered flurry")
	}
	effectstest.B24CastNoncreature(t, g, "Second", "Instant", "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("the second spell did not trigger flurry")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, theirs) {
		t.Error("the opponent's creature was not tapped")
	}
}

// --- Wingspan Stride -----------------------------------------------------------

func TestTdmSZ2WingspanStridePumpsGrantsFlyingAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Wingspan Stride", "Enchantment — Aura", tdmSZ2WingspanStrideOracle, false)
	effectstest.BrAttach(t, g, aura, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("enchanted creature %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has no flying")
	}
	effectstest.FillPoolColored(me, "U", 3)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature kept flying after the Aura left")
	}
}

// --- Worthy Cost ---------------------------------------------------------------

func TestTdmSZ2WorthyCostSacrificesToExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 6, 6)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Worthy Cost", TypeLine: "Sorcery",
		OracleID: tdmSZ2WorthyCostOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err == nil {
		t.Fatal("cast without sacrificing a creature")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the creature was not sacrificed as the additional cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Error("the target creature was not exiled")
	}
}
