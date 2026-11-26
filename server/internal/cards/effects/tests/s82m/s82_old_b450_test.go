package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b450_test.go — OLD batch 450 (#5572): Innistrad cards Somberwald
// Spider, Woodland Sleuth, Caravan Vigil, Nevermore, Thraben Sentry,
// Frightful Delusion, Lost in the Mist, Night Revelers, Essence of the
// Wild, Infernal Plunge, Gutter Grime and Cloistered Youth.

const (
	oldB450Spider    = "8b3526c4-efc8-42be-a0ab-de10c4cc986d"
	oldB450Sleuth    = "8bca5627-27c2-4d10-85c5-2451daa94c5d"
	oldB450Vigil     = "925a3ded-b7bc-4384-8684-79c20c59ab12"
	oldB450Nevermore = "0a1f4892-c9fa-4c60-a41e-7986cdd59096"
	oldB450Sentry    = "9636073e-e361-4d2a-92f6-5fe049e5eb23"
	oldB450Delusion  = "13ca3ede-a4c8-463d-b160-99b4071034b0"
	oldB450Mist      = "98014ad2-7a2b-45f5-808c-edd4d0b2c597"
	oldB450Revelers  = "56697773-8a04-49b6-8327-895e972da7e7"
	oldB450Essence   = "9d105f83-583d-438b-8f26-b39e818ed295"
	oldB450Plunge    = "0b67dff0-5bb7-44fe-a72b-d3b1a6c0ee72"
	oldB450Grime     = "0b8adb9e-97a5-43d9-87d6-993f22edaf3b"
	oldB450Youth     = "7552a9b4-b82f-491f-ad52-e271cf730211"
)

func TestOldB450EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB450Spider: "Somberwald Spider", oldB450Sleuth: "Woodland Sleuth", oldB450Vigil: "Caravan Vigil",
		oldB450Nevermore: "Nevermore", oldB450Sentry: "Thraben Sentry", oldB450Delusion: "Frightful Delusion",
		oldB450Mist: "Lost in the Mist", oldB450Revelers: "Night Revelers", oldB450Essence: "Essence of the Wild",
		oldB450Plunge: "Infernal Plunge", oldB450Grime: "Gutter Grime", oldB450Youth: "Cloistered Youth",
		oldB450Sentry + "#1": "Thraben Militia", oldB450Youth + "#1": "Unholy Fiend",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB450Died kills a creature of the opponent's so morbid is on.
func oldB450Died(t *testing.T, g *game.Game, opp *game.Player) {
	t.Helper()
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
}

func TestOldB450SomberwaldSpiderEntersWithCountersOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		if died {
			oldB450Died(t, g, opp)
		}
		effectstest.CastCatalogSpell(t, g, "Somberwald Spider", "Creature — Spider", oldB450Spider, nil)
		effectstest.PassPriorityAroundTable(t, g)
		id := effectstest.BattlefieldIDNamed(g, me.ID, "Somberwald Spider")
		want := 0
		if died {
			want = 2
		}
		if got := effectstest.LayeredCard(t, g, id).Counters[game.CounterPlusOne]; got != want {
			t.Errorf("died=%v: %d +1/+1 counters, want %d", died, got, want)
		}
	}
}

func TestOldB450WoodlandSleuthReturnsACreatureCardOnlyAfterADeath(t *testing.T) {
	for _, died := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		land := effectstest.PushGraveyardCardTyped(me, "Dead Forest", "Land — Forest")
		crit := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
		if died {
			oldB450Died(t, g, opp)
		}
		effectstest.CastCatalogSpell(t, g, "Woodland Sleuth", "Creature — Human Scout", oldB450Sleuth, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(crit); got != died {
			t.Errorf("died=%v: creature card in hand = %v", died, got)
		}
		if me.Hand.Contains(land) {
			t.Errorf("died=%v: a land card was returned", died)
		}
	}
}

func TestOldB450NightRevelersHasHasteOnlyWhileAnOpponentControlsAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Night Revelers", "Creature — Vampire", oldB450Revelers, true)
	if oldB450Haste(t, g, id) {
		t.Fatal("hasty with no Human in play")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Human", "Creature — Human", 1, 1)
	if oldB450Haste(t, g, id) {
		t.Error("hasty because I control a Human, want an opponent's")
	}
	_ = mine
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if oldB450Haste(t, g, id) {
		t.Error("hasty because an opponent controls a non-Human")
	}
	effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human Soldier", 1, 1)
	if !oldB450Haste(t, g, id) {
		t.Error("no haste while an opponent controls a Human")
	}
}

func oldB450Haste(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "haste")
}

func oldB450Basic(name string) game.Card {
	return effectstest.SearchTestLand(name, "Basic Land — "+name)
}

func TestOldB450CaravanVigilFetchesToHandOrWithMorbidMayPutItOntoTheBattlefield(t *testing.T) {
	for _, tc := range []struct {
		name          string
		died, accept  bool
		onBattlefield bool
	}{
		{"no morbid", false, false, false},
		{"morbid, declined", true, false, false},
		{"morbid, accepted", true, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			effectstest.SeedSearchLibrary(me, oldB450Basic("Forest"), game.Card{Name: "Elf", TypeLine: "Creature — Elf"})
			if tc.died {
				oldB450Died(t, g, opp)
			}
			effectstest.CastCatalogSpell(t, g, "Caravan Vigil", "Sorcery", oldB450Vigil, nil)
			effectstest.PassPriorityUntilChoice(t, g)
			if tc.died {
				effectstest.AnswerMayChoice(t, g, me.ID, tc.accept)
			} else if effectstest.LatestConfirmFor(g, me.ID) != nil {
				t.Fatal("asked a morbid question without morbid")
			}
			effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
			effectstest.PassPriorityAroundTable(t, g)
			onBF := effectstest.BattlefieldIDNamed(g, me.ID, "Forest") != uuid.Nil
			inHand := false
			for _, c := range me.Hand.Cards {
				inHand = inHand || c.Name == "Forest"
			}
			if onBF != tc.onBattlefield || inHand == tc.onBattlefield {
				t.Errorf("Forest on battlefield %v, in hand %v; want battlefield=%v", onBF, inHand, tc.onBattlefield)
			}
		})
	}
}

func TestOldB450NevermoreBarsTheNamedSpellForEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Nevermore", "Enchantment", oldB450Nevermore, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatalf("no card-name prompt as it entered: %+v", g.PendingChoices)
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Lightning Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	const bolt = "4457ed35-7c10-48c8-9776-456485fdf070"
	// My own named spell is barred too: "spells", not "your opponents'".
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", bolt, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err == nil {
		t.Error("the controller cast the named spell")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Burst Lightning", "Instant", effectstest.BurstLightningOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Errorf("a differently named spell was refused: %v", err)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lightning Bolt", "Instant", bolt, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Error("an opponent cast the named spell")
	}
}

func TestOldB450InfernalPlungeSacrificesACreatureForRRR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Infernal Plunge", "Sorcery", oldB450Plunge, nil); err == nil {
		t.Fatal("cast with no creature to sacrifice")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PaCast(t, g, "Infernal Plunge", "Sorcery", "{R}", oldB450Plunge, game.CastSpellParams{SacrificeIDs: []uuid.UUID{bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, bear) {
		t.Error("the creature was not sacrificed")
	}
	if got := me.ManaPool.CountColor("R"); got != 3 {
		t.Errorf("%d {R} in the pool, want 3", got)
	}
}

func TestOldB450EssenceOfTheWildMakesYourCreaturesEnterAsAnEssence(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	essence := effectstest.PushCatalogPermanent(g, me.ID, "Essence of the Wild", "Creature — Avatar", oldB450Essence, false)
	effectstest.AssertKeywords(t, g, essence) // registered and layered
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Essence of the Wild" && c.InstanceID != essence {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d new Essence copies on the battlefield, want the Bears to have entered as 1", n)
	}
}

func TestOldB450FrightfulDelusionCountersUnlessPaidAndTheyDiscardEitherWay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		keep := effectstest.HandCardFull(me, "Spare Card", "Sorcery", "{1}", "", nil)
		victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.CastFromSeat(t, g, opp, "Frightful Delusion", "Instant", oldB450Delusion, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		if pay {
			effectstest.B06AddMana(me, "C")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.AnswerDiscard(t, g, me.ID, keep)
		effectstest.PassPriorityAroundTable(t, g)
		if _, resolved := effectstest.BattlefieldCard(g, victim); resolved != pay {
			t.Errorf("paid=%v but the spell resolved=%v", pay, resolved)
		}
		if me.Hand.Contains(keep) {
			t.Errorf("paid=%v: the spell's controller did not discard", pay)
		}
	}
}

func TestOldB450LostInTheMistCountersTheSpellAndBouncesThePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	rock := effectstest.B12Creature(g, opp.ID, "Rock", "Creature — Golem", 1, 1)
	victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, opp, "Lost in the Mist", "Instant", oldB450Mist,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}, {Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the spell was not countered")
	}
	if effectstest.BattlefieldHas(g, rock) || !opp.Hand.Contains(rock) {
		t.Errorf("the permanent was not returned to its owner's hand (battlefield %v, hand %v)", effectstest.BattlefieldHas(g, rock), opp.Hand.Contains(rock))
	}
	_ = me
}

var (
	oldB450SentryCard = effectstest.IwCard{
		Oracle: oldB450Sentry, Front: "Thraben Sentry", FrontType: "Creature — Human Soldier", Cost: "{3}{W}",
		FrontText: "Vigilance\nWhenever another creature you control dies, you may transform this creature.", Fp: "2", Ft: "2",
		Back: "Thraben Militia", BackText: "Trample", Bp: "5", Bt: "4", Colors: []string{"W"},
	}
	oldB450YouthCard = effectstest.IwCard{
		Oracle: oldB450Youth, Front: "Cloistered Youth", FrontType: "Creature — Human", Cost: "{1}{W}",
		FrontText: "At the beginning of your upkeep, you may transform this creature.", Fp: "1", Ft: "1",
		Back: "Unholy Fiend", BackText: "At the beginning of your end step, you lose 1 life.", Bp: "3", Bt: "3", Colors: []string{"B"},
	}
)

func TestOldB450ThrabenSentryMayTransformWhenAnotherCreatureYouControlDies(t *testing.T) {
	for _, yes := range []bool{false, true} {
		g, me := effectstest.IwTable(t)
		id := effectstest.IwPut(t, g, oldB450SentryCard, me)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.KillCreature(t, g, me.ID, bear)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if yes {
			want = 1
		}
		if got := effectstest.IwFace(g, id); got != want {
			t.Errorf("yes=%v: face %d, want %d", yes, got, want)
		}
	}
}

func TestOldB450ThrabenSentryIgnoresAnOpponentsCreatureDying(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.IwPut(t, g, oldB450SentryCard, me)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil || effectstest.IwFace(g, id) != 0 {
		t.Error("the Sentry reacted to an opponent's creature dying")
	}
}

func TestOldB450CloisteredYouthMayTransformAtYourUpkeep(t *testing.T) {
	for _, yes := range []bool{false, true} {
		g, me := effectstest.IwTable(t)
		id := effectstest.IwPut(t, g, oldB450YouthCard, me)
		seat := g.Turn.ActiveSeat
		effectstest.IwNextUpkeep(t, g, nil)
		for g.Turn.ActiveSeat != seat {
			effectstest.IwNextUpkeep(t, g, nil)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if yes {
			want = 1
		}
		if got := effectstest.IwFace(g, id); got != want {
			t.Errorf("yes=%v: face %d, want %d", yes, got, want)
		}
	}
}

func TestOldB450UnholyFiendCostsALifeAtYourEndStep(t *testing.T) {
	g, me := effectstest.IwTable(t)
	id := effectstest.IwPut(t, g, oldB450YouthCard, me)
	effectstest.IwTransform(t, g, id)
	if effectstest.IwFace(g, id) != 1 {
		t.Fatal("test setup: could not turn the Youth over")
	}
	life := me.Life
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d after the Fiend's end step", me.Life, life-1)
	}
}

func TestOldB450GutterGrimeMakesAnOozeSizedByItsSlimeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	grime := effectstest.PushCatalogPermanent(g, me.ID, "Gutter Grime", "Enchantment", oldB450Grime, false)
	slime := func() int { return effectstest.LayeredCard(t, g, grime).Counters["slime"] }
	oozes := func() []game.Card {
		var out []game.Card
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == me.ID && c.Name == "Ooze" {
				out = append(out, c)
			}
		}
		return out
	}
	die := func(name string) {
		id := effectstest.B12Creature(g, me.ID, name, "Creature — Bear", 2, 2)
		effectstest.KillCreature(t, g, me.ID, id)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	die("Bear A")
	if slime() != 1 || len(oozes()) != 1 {
		t.Fatalf("after one death: %d slime counters, %d Oozes", slime(), len(oozes()))
	}
	die("Bear B")
	if slime() != 2 || len(oozes()) != 2 {
		t.Fatalf("after two deaths: %d slime counters, %d Oozes", slime(), len(oozes()))
	}
	for _, o := range oozes() {
		e := effectstest.LayeredCard(t, g, o.InstanceID).Effective()
		if e.Power != 2 || e.Toughness != 2 {
			t.Errorf("Ooze is %d/%d, want 2/2 (each Ooze reads the Grime's counters)", e.Power, e.Toughness)
		}
	}
	// A token dying does not trigger it.
	tok := effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1, PrintedPTKnown: true})
	effectstest.KillCreature(t, g, me.ID, tok)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if slime() != 2 {
		t.Errorf("a token dying added a slime counter (now %d)", slime())
	}
	// Once the Grime leaves, the Oozes are 0/0 and die.
	effectstest.ExileCreature(t, g, grime)
	g.RunStateChecksForTest()
	if len(oozes()) != 0 {
		t.Errorf("%d Oozes survived the Grime leaving", len(oozes()))
	}
}
