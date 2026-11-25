package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b198_cards_test.go — Core Set 2019, batch 198 (S82 older sets,
// #5527). Every test plays the real catalog entry through the engine.

const (
	oldB198GraveyardMarshalOracle   = "216c5a72-ef87-4ada-bf52-437e50874be5"
	oldB198ViviensJaguarOracle      = "32412a41-2929-4417-90bc-21c5659eeb72"
	oldB198BoneDragonOracle         = "51926430-b98a-424d-9347-c36938132825"
	oldB198SarkhansDragonfireOracle = "2a08d878-7816-4d0a-9532-4f2d3c8eee90"
	oldB198ViviensInvocationOracle  = "5e2b30d2-81e9-40cd-86dd-3ec5dd06ced0"
	oldB198RadiatingLightningOracle = "9794f397-dd86-4cfe-a431-5f8d8c6b68ea"
	oldB198FrayingOmnipotenceOracle = "3b00068d-ad07-417c-bbb3-5f980615fb06"
	oldB198NightmaresThirstOracle   = "9ecf56b9-1f36-4030-8333-1675ccc663d9"
	oldB198CourtClericOracle        = "dec9edbb-677b-4773-8bc4-b388d7592276"
	oldB198SuncleanserOracle        = "437661f7-a32f-45c9-a006-05a1e38ee8d9"
	oldB198GatebreakerOracle        = "eb35d7e5-eddd-4e02-99ac-1a42cd575041"
	oldB198AbnormalEnduranceOracle  = "edbf1b87-2d1e-47e6-a04e-a2b1646af7d9"

	oldB198BearType = "Creature — Bear"
)

func TestOldB198EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB198GraveyardMarshalOracle, oldB198ViviensJaguarOracle, oldB198BoneDragonOracle, oldB198SarkhansDragonfireOracle,
		oldB198ViviensInvocationOracle, oldB198RadiatingLightningOracle, oldB198FrayingOmnipotenceOracle,
		oldB198NightmaresThirstOracle, oldB198CourtClericOracle, oldB198SuncleanserOracle, oldB198GatebreakerOracle,
		oldB198AbnormalEnduranceOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB198Graveyard puts a catalog card in p's graveyard.
func oldB198Graveyard(p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, power, toughness
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// oldB198LibraryTop stacks cards on top of p's library, the first on top.
func oldB198LibraryTop(p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

func oldB198Planeswalker(g *game.Game, owner uuid.UUID, subtype string) uuid.UUID {
	id := effectstest.B12Permanent(g, owner, subtype+" Test", "Legendary Planeswalker — "+subtype)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, 4) })
	return id
}

func TestOldB198GraveyardMarshalExilesACreatureCardForATappedZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := effectstest.SeatsForCombat(g)
	marshal := effectstest.B12Push(g, me.ID, "Graveyard Marshal", "Creature — Zombie Soldier", oldB198GraveyardMarshalOracle, 3, 2)
	spell := oldB198Graveyard(me, "Old Spell", "Sorcery", "", 0, 0)
	dead := oldB198Graveyard(me, "Dead Bear", oldB198BearType, "", 2, 2)
	effectstest.FloatForTest(g, me, "BCC")
	if err := g.ActivateCatalogAbility(me.ID, marshal, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("a sorcery card paid for a creature-card exile cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, marshal, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{dead}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	var zombie *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == me.ID && c.IsToken() {
			zombie = c
		}
	}
	if zombie == nil {
		t.Fatal("no token was created")
	}
	if zombie.Name != "Zombie" || zombie.Power != 2 || zombie.Toughness != 2 || !zombie.Tapped {
		t.Errorf("token = %s %d/%d tapped=%v, want a tapped 2/2 Zombie", zombie.Name, zombie.Power, zombie.Toughness, zombie.Tapped)
	}
}

func TestOldB198GraveyardMarshalNeedsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := effectstest.SeatsForCombat(g)
	marshal := effectstest.B12Push(g, me.ID, "Graveyard Marshal", "Creature — Zombie Soldier", oldB198GraveyardMarshalOracle, 3, 2)
	oldB198Graveyard(me, "Old Spell", "Sorcery", "", 0, 0)
	effectstest.FloatForTest(g, me, "BCC")
	if err := g.ActivateCatalogAbility(me.ID, marshal, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no creature card in the graveyard")
	}
}

func TestOldB198ViviensJaguarReturnsOnlyWithAVivien(t *testing.T) {
	for _, tc := range []struct {
		subtype string
		want    bool
	}{{"", false}, {"Ajani", false}, {"Vivien", true}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, _ := effectstest.SeatsForCombat(g)
		jag := oldB198Graveyard(me, "Vivien's Jaguar", "Creature — Cat Spirit", oldB198ViviensJaguarOracle, 3, 2)
		if tc.subtype != "" {
			oldB198Planeswalker(g, me.ID, tc.subtype)
		}
		effectstest.FloatForTest(g, me, "GCC")
		err := g.ActivateCatalogAbility(me.ID, jag, 0, game.ActivateAbilityParams{Strict: true})
		if (err == nil) != tc.want {
			t.Fatalf("planeswalker %q: activate err = %v, want success %v", tc.subtype, err, tc.want)
		}
		if !tc.want {
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(jag) || me.Graveyard.Contains(jag) {
			t.Error("the Jaguar did not return to hand")
		}
	}
}

func TestOldB198ViviensJaguarHasReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Vivien's Jaguar", "Creature — Cat Spirit", oldB198ViviensJaguarOracle, 3, 2)
	if !effectstest.HasEffectiveKeyword(t, g, id, "reach") {
		t.Error("no reach")
	}
}

func TestOldB198BoneDragonNeedsSevenOtherCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := effectstest.SeatsForCombat(g)
	dragon := oldB198Graveyard(me, "Bone Dragon", "Creature — Dragon Skeleton", oldB198BoneDragonOracle, 5, 4)
	var fodder []uuid.UUID
	for i := 0; i < 6; i++ {
		fodder = append(fodder, oldB198Graveyard(me, "Filler", "Sorcery", "", 0, 0))
	}
	effectstest.FloatForTest(g, me, "BBCCC")
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: fodder}); err == nil {
		t.Fatal("six cards paid for seven")
	}
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: append(append([]uuid.UUID{}, fodder...), dragon)}); err == nil {
		t.Fatal("the Dragon itself paid for its own cost")
	}
	fodder = append(fodder, oldB198Graveyard(me, "Filler", "Land", "", 0, 0))
	if err := g.ActivateCatalogAbility(me.ID, dragon, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: fodder}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range fodder {
		if !g.Exile.Contains(id) {
			t.Fatal("a paid card is not in exile")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dragon) {
		t.Fatal("the Dragon did not return to the battlefield")
	}
	c, _ := g.LookupCardForEffect(dragon)
	if !c.Tapped {
		t.Error("the Dragon returned untapped")
	}
}

func TestOldB198SarkhansDragonfireDamagesAndTakesARedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	ids := oldB198LibraryTop(me,
		game.Card{Name: "Green Card", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}},
		game.Card{Name: "Red Card", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}},
		game.Card{Name: "Land", TypeLine: "Land"}, game.Card{Name: "Land", TypeLine: "Land"}, game.Card{Name: "Land", TypeLine: "Land"},
		game.Card{Name: "Deep Red", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}})
	green, red, deep := ids[0], ids[1], ids[5]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Sarkhan's Dragonfire", "Sorcery", oldB198SarkhansDragonfireOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to take a red card")
	}
	for _, id := range pick.ChooseCards {
		if id == green || id == deep {
			t.Errorf("offered a card it may not take: %s", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, red)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(red) {
		t.Error("the red card is not in hand")
	}
}

func TestOldB198ViviensInvocationPutsACreatureAndItBitesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 5)
	ids := oldB198LibraryTop(me,
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Big Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4})
	spell, beast := ids[0], ids[1]
	effectstest.CastCatalogSpell(t, g, "Vivien's Invocation", "Sorcery", oldB198ViviensInvocationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to put a creature onto the battlefield")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(beast) {
		t.Fatal("the creature did not enter")
	}
	if me.Library.Contains(spell) && me.Library.Cards[0].InstanceID == spell {
		t.Error("the rest stayed on top of the library")
	}
	effectstest.PickCard(t, g, me.ID, victim) // the reflexive trigger's target
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.DamageOn(t, g, victim); d != 4 {
		t.Errorf("their creature has %d damage, want 4 (the Beast's power)", d)
	}
}

func TestOldB198RadiatingLightningHitsThePlayerAndTheirCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB198BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB198BearType, 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Radiating Lightning", "Instant", oldB198RadiatingLightningOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, opp.Life)
	}
	if d := effectstest.DamageOn(t, g, theirs); d != 1 {
		t.Errorf("their creature has %d damage, want 1", d)
	}
	if d := effectstest.DamageOn(t, g, mine); d != 0 {
		t.Errorf("my creature has %d damage, want 0", d)
	}
}

func TestOldB198FrayingOmnipotenceHalvesEverythingRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	me.Life, opp.Life = 11, 20
	var myBears, theirBears []uuid.UUID
	for i := 0; i < 3; i++ {
		myBears = append(myBears, effectstest.B12Creature(g, me.ID, "My Bear", oldB198BearType, 2, 2))
	}
	theirBears = append(theirBears, effectstest.B12Creature(g, opp.ID, "Their Bear", oldB198BearType, 2, 2))
	effectstest.CastCatalogSpell(t, g, "Fraying Omnipotence", "Sorcery", oldB198FrayingOmnipotenceOracle, nil)
	hands := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		hands[p.ID] = p.Hand.Size()
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 5 || opp.Life != 10 {
		t.Errorf("life = %d/%d, want 5/10 (half of 11 and 20, rounded up)", me.Life, opp.Life)
	}
	for _, p := range g.Seats {
		hand := hands[p.ID]
		c := effectstest.ChooseCardsChoiceFor(g, p.ID)
		if c == nil {
			t.Fatalf("no discard prompt for a player with %d cards", hand)
		}
		want := (hand + 1) / 2
		if c.Count != want {
			t.Errorf("a hand of %d owes %d discards, want %d (rounded up)", hand, c.Count, want)
		}
		var ids []uuid.UUID
		for _, card := range p.Hand.Cards[:want] {
			ids = append(ids, card.InstanceID)
		}
		effectstest.AnswerChooseCards(t, g, p.ID, ids...)
		if got := p.Hand.Size(); got != hand-want {
			t.Errorf("a hand of %d has %d cards left, want %d", hand, got, hand-want)
		}
	}
	// 3 creatures -> 2 sacrificed (round up); 1 -> 1.
	effectstest.AnswerSacrifice(t, g, me.ID, myBears[0], myBears[1])
	effectstest.AnswerSacrifice(t, g, opp.ID, theirBears[0])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myBears[0]) || g.Battlefield.Contains(myBears[1]) || !g.Battlefield.Contains(myBears[2]) {
		t.Error("I should have sacrificed exactly the two chosen of my three creatures")
	}
	if g.Battlefield.Contains(theirBears[0]) {
		t.Error("their lone creature should have been sacrificed (half of one, rounded up)")
	}
}

func TestOldB198NightmaresThirstShrinksByTheLifeGainedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB198BearType, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Nightmare's Thirst", "Instant", oldB198NightmaresThirstOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
	// X = 2 (earlier gain) + 1 = 3.
	if p, tough := effectstest.PtOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("target is %d/%d, want 1/1 (-3/-3)", p, tough)
	}
}

func TestOldB198NightmaresThirstWithNoOtherGain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB198BearType, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Nightmare's Thirst", "Instant", oldB198NightmaresThirstOracle, effectstest.CardTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, big); p != 3 || tough != 3 {
		t.Errorf("target is %d/%d, want 3/3 (-1/-1)", p, tough)
	}
}

func TestOldB198CourtClericGrowsWithAnAjani(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	cleric := effectstest.B12Push(g, me.ID, "Court Cleric", "Creature — Human Cleric", oldB198CourtClericOracle, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, cleric, "lifelink") {
		t.Error("no lifelink")
	}
	oldB198Planeswalker(g, me.ID, "Vivien")
	if p, tough := effectstest.PtOf(t, g, cleric); p != 1 || tough != 1 {
		t.Errorf("with a Vivien it is %d/%d, want 1/1", p, tough)
	}
	oldB198Planeswalker(g, me.ID, "Ajani")
	if p, tough := effectstest.PtOf(t, g, cleric); p != 2 || tough != 2 {
		t.Errorf("with an Ajani it is %d/%d, want 2/2", p, tough)
	}
}

// oldB198Suncleanser casts Suncleanser, picks `mode` for its enters trigger
// and answers the target prompt with `pick`, then lets the trigger resolve.
func oldB198Suncleanser(t *testing.T, g *game.Game, me *game.Player, mode int, pick func()) uuid.UUID {
	t.Helper()
	id := effectstest.CastAndResolveCreature(t, g, "Suncleanser", "Creature — Human Cleric", oldB198SuncleanserOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	pick()
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB198SuncleanserFirstModeStripsAndHoldsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB198BearType, 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2) })
	oldB198Suncleanser(t, g, me, 0, func() { effectstest.PickCard(t, g, me.ID, bear) })
	if n := effectstest.PlusOneCounters(g, bear); n != 0 {
		t.Errorf("the creature still has %d +1/+1 counters", n)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if n := effectstest.PlusOneCounters(g, bear); n != 0 {
		t.Errorf("a counter was put on the creature while Suncleanser remains (%d)", n)
	}
}

func TestOldB198SuncleanserSecondModeStripsAndHoldsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	g.WithWriteLock(func() {
		_ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 3)
		_ = g.AddPlayerCounterForEffect(opp.ID, game.CounterEnergy, 2)
	})
	oldB198Suncleanser(t, g, me, 1, func() { effectstest.PickPlayer(t, g, me.ID, opp.ID) })
	if opp.Counters[game.CounterPoison] != 0 || opp.Counters[game.CounterEnergy] != 0 {
		t.Errorf("the opponent still has counters: %v", opp.Counters)
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 1) })
	if opp.Counters[game.CounterPoison] != 0 {
		t.Error("the opponent got a poison counter while Suncleanser remains")
	}
}

func TestOldB198SuncleanserHoldEndsWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB198BearType, 2, 2)
	cleanser := oldB198Suncleanser(t, g, me, 0, func() { effectstest.PickCard(t, g, me.ID, bear) })
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if n := effectstest.PlusOneCounters(g, bear); n != 0 {
		t.Fatalf("a counter landed while Suncleanser remains (%d)", n)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(cleanser) })
	g.RunStateChecksForTest()
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if n := effectstest.PlusOneCounters(g, bear); n != 1 {
		t.Errorf("the creature still can't have counters after Suncleanser left (%d)", n)
	}
}

func TestOldB198GatebreakerLooksForABlueOrArtifactCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	ids := oldB198LibraryTop(me,
		game.Card{Name: "Red", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}},
		game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Blue", TypeLine: "Sorcery", ManaCost: "{U}", Colors: []string{"U"}},
		game.Card{Name: "Land", TypeLine: "Land"}, game.Card{Name: "Land", TypeLine: "Land"},
		game.Card{Name: "Deep Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	red, rock, blue, deep := ids[0], ids[1], ids[2], ids[5]
	effectstest.CastHoldCreature(t, g, "Tezzeret's Gatebreaker", "Artifact", oldB198GatebreakerOracle, 0, 0)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to take a card")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if offered[red] || offered[deep] || !offered[rock] || !offered[blue] {
		t.Errorf("offered %v: want the blue and artifact cards among the top five only", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the artifact is not in hand")
	}
}

func TestOldB198GatebreakerMakesYourCreaturesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.SeatsForCombat(g)
	gate := effectstest.B12Push(g, me.ID, "Tezzeret's Gatebreaker", "Artifact", oldB198GatebreakerOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB198BearType, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB198BearType, 2, 2)
	effectstest.FloatForTest(g, me, "UCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, gate, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(gate) {
		t.Error("the Gatebreaker was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.PorBlock(t, g, mine, opp.ID, blocker); err == nil {
		t.Error("an unblockable creature was blocked")
	}
}

func TestOldB198AbnormalEnduranceReturnsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", oldB198BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Abnormal Endurance", "Instant", oldB198AbnormalEnduranceOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Fatalf("target is %d/%d, want 4/2", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "My Bear" && c.Controller == me.ID {
			back = c
		}
	}
	if back == nil {
		t.Fatal("the creature did not return")
	}
	if !back.Tapped {
		t.Error("the creature returned untapped")
	}
}
