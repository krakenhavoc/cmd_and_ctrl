package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB01JeweledAmuletOracle = "e2ad4b4d-d5fe-4a64-9b40-296c8fc8bcce"
	gapB01DragonsSmileOracle  = "e7611ef9-e5fc-45eb-bc3f-4358a27107fd"
	gapB01EverforgerOracle    = "b8354ac6-5285-43dd-9ba5-93064def2e96"
	gapB01UramonOracle        = "92eeedc6-a331-445f-af56-c91dac69412b"
	gapB01IstvanOracle        = "5e277de5-793a-4364-923a-a55b53c0fa85"
	gapB01KurokiOracle        = "ce70f16c-ea45-4f1d-aafb-34a5d05d31e7"
	gapB01OlindaOracle        = "6ec8a9b8-233f-4de5-8f4c-3d677fbf2f30"
	gapB01SahirOracle         = "e1a8f5cf-9b59-46c6-8e41-6e44225ec0f4"
	gapB01ValkoOracle         = "8023d419-8fd2-4b92-b6ea-f25fd67e6410"
	gapB01DavvolOracle        = "e4aca279-b144-4517-acef-d925f4f552e2"
)

func TestGapB01EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB01JeweledAmuletOracle: "Jeweled Amulet",
		gapB01DragonsSmileOracle:  "Dragon's Smile",
		gapB01EverforgerOracle:    "The Everforger",
		gapB01UramonOracle:        "Boss Uramon, Shadow's Reach",
		gapB01IstvanOracle:        "Istvan, Butcher of Eln",
		gapB01KurokiOracle:        "Kuroki, Thief of Talents",
		gapB01OlindaOracle:        "Olinda the Oblivious",
		gapB01SahirOracle:         "Sahir, Visitor in Darkness",
		gapB01ValkoOracle:         "Valko Indorian",
		gapB01DavvolOracle:        "Davvol, Evincar of Rath",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB01Enter(t *testing.T, g *game.Game, seat int, name, typeLine string) uuid.UUID {
	t.Helper()
	id := effectstest.B13PlayAs(t, g, seat, name, typeLine, "")
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- Valko Indorian ---------------------------------------------------

func TestGapB01ValkoGivesYourPhyrexiansMenaceAndLifelinkOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	valko := effectstest.B12Push(g, me.ID, "Valko Indorian", "Legendary Creature — Human Wizard", gapB01ValkoOracle, 2, 4)
	mine := effectstest.B12Creature(g, me.ID, "Rager", "Creature — Phyrexian Horror", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Phyrexian", "Creature — Phyrexian Horror", 2, 2)
	for _, kw := range []string{"menace", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my Phyrexian lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, human, kw) || effectstest.HasEffectiveKeyword(t, g, theirs, kw) || effectstest.HasEffectiveKeyword(t, g, valko, kw) {
			t.Errorf("a non-Phyrexian or an opponent's creature has %s", kw)
		}
	}
}

// --- Davvol, Evincar of Rath ----------------------------------------

func TestGapB01DavvolLosesALifeAndAddsBlackForAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Davvol, Evincar of Rath", "Legendary Creature — Phyrexian Noble", gapB01DavvolOracle, 2, 2)
	before := me.Life
	gapB01Enter(t, g, 0, "Bear", "Creature — Bear")
	if me.Life != before-1 {
		t.Errorf("life %d, want %d", me.Life, before-1)
	}
	if effectstest.PoolCount(me, "B") != 1 {
		t.Errorf("pool holds %d {B}, want 1", effectstest.PoolCount(me, "B"))
	}
}

func TestGapB01DavvolIgnoresItselfAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Davvol, Evincar of Rath", "Legendary Creature — Phyrexian Noble", gapB01DavvolOracle, 2, 2)
	before := me.Life
	gapB01Enter(t, g, 1, "Their Bear", "Creature — Bear")
	if me.Life != before || effectstest.PoolCount(me, "B") != 0 {
		t.Errorf("an opponent's creature triggered Davvol: life %d, {B} %d", me.Life, effectstest.PoolCount(me, "B"))
	}
}

// --- Istvan, Butcher of Eln -------------------------------------------

func TestGapB01IstvanGrowsByTwoWhenAnOpponentsCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	istvan := effectstest.B12Push(g, me.ID, "Istvan, Butcher of Eln", "Legendary Creature — Human", gapB01IstvanOracle, 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, istvan, "+1/+1"); n != 0 {
		t.Fatalf("my own creature dying put %d counters on Istvan", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, istvan, "+1/+1"); n != 2 {
		t.Errorf("Istvan has %d +1/+1 counters, want 2", n)
	}
}

func TestGapB01IstvanPreventsDamageFromCreaturesButNotOtherSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	istvan := effectstest.B12Push(g, me.ID, "Istvan, Butcher of Eln", "Legendary Creature — Human", gapB01IstvanOracle, 1, 3)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	bolt := effectstest.B12Permanent(g, opp.ID, "Pinger", "Artifact")
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(ogre, istvan, 5); err != nil {
			t.Fatalf("damage from a creature: %v", err)
		}
	})
	if got := gapB01Damage(g, istvan); got != 0 {
		t.Errorf("a creature dealt %d damage to Istvan, want it all prevented", got)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(bolt, istvan, 2); err != nil {
			t.Fatalf("damage from an artifact: %v", err)
		}
	})
	if got := gapB01Damage(g, istvan); got != 2 {
		t.Errorf("an artifact dealt %d damage to Istvan, want 2 (only creatures are prevented)", got)
	}
}

func gapB01Damage(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

func gapB01CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				n++
			}
		}
	})
	return n
}

// --- Boss Uramon, Shadow's Reach --------------------------------------

func TestGapB01UramonGivesAttackingNinjasAndRoguesDeathtouchAndMakesRatsOnDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	uramon := effectstest.B12Push(g, me.ID, "Boss Uramon, Shadow's Reach", "Legendary Creature — Human Rogue", gapB01UramonOracle, 2, 2)
	ninja := effectstest.B12Creature(g, me.ID, "Ninja", "Creature — Human Ninja", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, ninja, "deathtouch") {
		t.Fatal("a Ninja that is not attacking has deathtouch")
	}
	effectstest.AttackWith(t, g, opp.ID, uramon, ninja, bear)
	for _, id := range []uuid.UUID{uramon, ninja} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Errorf("an attacking Rogue or Ninja lacks deathtouch")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("an attacking Bear has deathtouch")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Rat Rogue"); n != 1 {
		t.Errorf("%d Rat Rogue tokens, want 1: two Ninjas and Rogues dealt damage to one player together", n)
	}
}

func TestGapB01UramonMakesNothingWhenOnlyAnOtherTypeDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Boss Uramon, Shadow's Reach", "Legendary Creature — Human Rogue", gapB01UramonOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Rat Rogue"); n != 0 {
		t.Errorf("a Bear's damage made %d Rat Rogues", n)
	}
}

func TestGapB01UramonMakesOneRatPerDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	uramon := effectstest.B12Push(g, me.ID, "Boss Uramon, Shadow's Reach", "Legendary Creature — Human Rogue", gapB01UramonOracle, 2, 2)
	rogue := effectstest.B12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(uramon, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(rogue, third.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Rat Rogue"); n != 2 {
		t.Errorf("%d Rat Rogues, want 2 (one per damaged player)", n)
	}
}

// --- Olinda the Oblivious -------------------------------------------------

func TestGapB01OlindaMarksOneCreaturePerOpponentAndTheyLoseLifeAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	ca := effectstest.B12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	ca2 := effectstest.B12Creature(g, a.ID, "A's Other Bear", "Creature — Bear", 2, 2)
	cb := effectstest.B12Creature(g, b.ID, "B's Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Olinda the Oblivious", "Legendary Creature — Skunk Rogue", gapB01OlindaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, ca)
	effectstest.AnswerPickTarget(t, g, cb)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ca, cb} {
		if n := effectstest.CounterCount(g, id, "odor"); n != 1 {
			t.Errorf("%s has %d odor counters, want 1", id, n)
		}
	}
	if n := effectstest.CounterCount(g, ca2, "odor"); n != 0 {
		t.Errorf("the second creature of the same opponent got %d odor counters", n)
	}
	// The creature's controller loses 2 life at their own upkeep.
	before := a.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if a.Life != before-2 {
		t.Errorf("A's life %d, want %d", a.Life, before-2)
	}
}

func TestGapB01OlindaOnlyOffersCreaturesWithoutAnOdorCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := g.Seats[1]
	ca := effectstest.B12Creature(g, a.ID, "A's Smelly Bear", "Creature — Bear", 2, 2)
	fresh := effectstest.B12Creature(g, a.ID, "A's Fresh Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ca, "odor", 1) })
	effectstest.CastCatalogSpell(t, g, "Olinda the Oblivious", "Legendary Creature — Skunk Rogue", gapB01OlindaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, ca) || !slices.Contains(ch.PickTargetCards, fresh) {
		t.Fatalf("offered %v: only the creature without an odor counter should be", ch.PickTargetCards)
	}
}

// --- Sahir, Visitor in Darkness ---------------------------------------------

func gapB01SahirCard(p *game.Player) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Sahir, Visitor in Darkness", TypeLine: "Legendary Creature — Human Wizard",
		ManaCost: "{1}{B}{B}", OracleID: gapB01SahirOracle, Power: 3, Toughness: 3,
		Owner: p.ID, Controller: p.ID, IsCommander: true,
	}
}

func TestGapB01SahirEminenceDamagesYouAtUpkeepFromTheCommandZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Command.PushTop(gapB01SahirCard(me))
	before := me.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("life %d, want %d: Sahir deals 1 damage to you from the command zone", me.Life, before-1)
	}
	// The counter ability is the battlefield's: nothing grew, nothing drew.
	if len(me.Command.Cards) != 1 {
		t.Fatal("Sahir left the command zone")
	}
}

func TestGapB01SahirEminenceDoesNotWorkFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := gapB01SahirCard(me)
	me.Graveyard.PushTop(c)
	before := me.Life
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d, want %d: Sahir in a graveyard does nothing", me.Life, before)
	}
}

func TestGapB01SahirGrowsOnExactlyOneDamageOrOneLifeLostByYourSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sahir := effectstest.B12Push(g, me.ID, "Sahir, Visitor in Darkness", "Legendary Creature — Human Wizard", gapB01SahirOracle, 3, 3)
	mine := effectstest.B12Permanent(g, me.ID, "My Pinger", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Pinger", "Artifact")

	// Exactly one damage from a source you control: a counter.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sahir, "+1/+1"); n != 1 {
		t.Fatalf("%d counters after 1 damage from my source, want 1", n)
	}
	// Two damage, or damage from an opponent's source: none.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(mine, me.ID, 2) })
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sahir, "+1/+1"); n != 1 {
		t.Fatalf("%d counters after 2 damage and an opponent's 1 damage, want still 1", n)
	}
	// Exactly one life lost to a spell or ability you control: a counter, and
	// as the second resolution this turn it draws a card.
	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mine, me.ID, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sahir, "+1/+1"); n != 2 {
		t.Fatalf("%d counters after 1 life lost to my ability, want 2", n)
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand grew by %d, want 1 card for the second resolution this turn", len(me.Hand.Cards)-hand)
	}
	// A third resolution this turn draws nothing.
	hand = len(me.Hand.Cards)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mine, me.ID, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sahir, "+1/+1"); n != 3 {
		t.Fatalf("%d counters after the third trigger, want 3", n)
	}
	if len(me.Hand.Cards) != hand {
		t.Errorf("the third resolution drew %d cards, want 0", len(me.Hand.Cards)-hand)
	}
}

func TestGapB01SahirIgnoresLifePaidAsACostAndLifeLostToAnOpponentsSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sahir := effectstest.B12Push(g, me.ID, "Sahir, Visitor in Darkness", "Legendary Creature — Human Wizard", gapB01SahirOracle, 3, 3)
	mine := effectstest.B12Permanent(g, me.ID, "My Altar", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Drain", "Artifact")
	g.WithWriteLock(func() { _ = g.PayLifeForEffect(mine, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(theirs, me.ID, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(mine, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sahir, "+1/+1"); n != 0 {
		t.Errorf("%d counters, want 0: a paid cost, an opponent's drain and a gain are not it", n)
	}
}

// --- Kuroki, Thief of Talents -------------------------------------------------

func gapB01KurokiEndStep(t *testing.T, g *game.Game) (me, opp *game.Player, kuroki uuid.UUID) {
	t.Helper()
	me, opp = g.Seats[0], g.Seats[1]
	kuroki = effectstest.B12Push(g, me.ID, "Kuroki, Thief of Talents", "Legendary Creature — Spirit", gapB01KurokiOracle, 4, 4)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	return me, opp, kuroki
}

func TestGapB01KurokiOpponentDrawsFourThenYouCastASpellFromTheirHandForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	sorcery := effectstest.PushHandCardWithManaCost(opp, "Their Big Sorcery", "Sorcery", "{5}{U}{U}")
	hand := len(opp.Hand.Cards)
	me, _, kuroki := gapB01KurokiEndStep(t, g)
	third := g.Seats[2]
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID) == nil {
		t.Fatal("the targeted opponent was not asked")
	}
	answerKurokiConfirm(t, g, opp.ID, true)
	if got := len(opp.Hand.Cards) - hand; got != 4 {
		t.Fatalf("the opponent drew %d cards, want 4", got)
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	_ = offer
	if !effectstest.Knows(g, opp, sorcery, me.ID) || effectstest.Knows(g, opp, sorcery, third.ID) {
		t.Error("only Kuroki's controller should have looked at the hand")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast out of their hand: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, sorcery)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(sorcery) {
		t.Error("the resolved sorcery should be in its owner's graveyard")
	}
	if n := effectstest.CounterCount(g, kuroki, "+1/+1"); n != 0 {
		t.Errorf("Kuroki got %d counters although the opponent drew", n)
	}
}

func TestGapB01KurokiGetsTwoCountersWhenTheOpponentDeclines(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.PushHandCardWithManaCost(opp, "Their Big Sorcery", "Sorcery", "{5}{U}{U}")
	hand := len(opp.Hand.Cards)
	_, _, kuroki := gapB01KurokiEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	answerKurokiConfirm(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, kuroki, "+1/+1"); n != 2 {
		t.Errorf("Kuroki has %d counters, want 2", n)
	}
	if len(opp.Hand.Cards) != hand {
		t.Errorf("the opponent's hand changed from %d to %d", hand, len(opp.Hand.Cards))
	}
}

func answerKurokiConfirm(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, who)
	if c == nil {
		t.Fatalf("%s has no confirm open: %+v", who, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, who, yes); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
}

// --- The Everforger ----------------------------------------------------------

// gapB01PassUntilConfirm passes priority until `who` has a confirm open, and
// returns whether one opened before the stack emptied.
func gapB01PassUntilConfirm(t *testing.T, g *game.Game, who uuid.UUID) bool {
	t.Helper()
	for i := 0; i < 12; i++ {
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, who) != nil {
			return true
		}
		if effectstest.StackFullyEmpty(g) {
			return false
		}
		if err := g.PassPriority(); err != nil && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, who) == nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	return effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, who) != nil
}

func gapB01Everforger(g *game.Game, p *game.Player) uuid.UUID {
	return effectstest.B12Push(g, p.ID, "The Everforger", "Legendary Artifact Creature — Construct", gapB01EverforgerOracle, 4, 5)
}

func TestGapB01EverforgerCopiesAnArtifactCreatureSpellOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapB01Everforger(g, me)
	effectstest.CastCatalogSpell(t, g, "Steel Golem", "Artifact Creature — Golem", "", nil)
	if !gapB01PassUntilConfirm(t, g, me.ID) {
		t.Fatal("the copy was not offered for an artifact creature spell")
	}
	answerKurokiConfirm(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Steel Golem"); n != 2 {
		t.Fatalf("%d Steel Golems, want the spell and its token copy", n)
	}
	// Once each turn: a second artifact creature spell this turn is not offered.
	effectstest.CastCatalogSpell(t, g, "Iron Golem", "Artifact Creature — Golem", "", nil)
	if gapB01PassUntilConfirm(t, g, me.ID) {
		t.Fatal("a second copy was offered in the same turn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Iron Golem"); n != 1 {
		t.Errorf("%d Iron Golems, want 1 (no copy the second time this turn)", n)
	}
}

func TestGapB01EverforgerDecliningLeavesTheCopyForALaterSpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapB01Everforger(g, me)
	effectstest.CastCatalogSpell(t, g, "Sword of Testing", "Artifact — Equipment", "", nil)
	if !gapB01PassUntilConfirm(t, g, me.ID) {
		t.Fatal("the copy was not offered for an Equipment spell")
	}
	answerKurokiConfirm(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB01CountNamed(g, me.ID, "Sword of Testing"); n != 1 {
		t.Fatalf("%d Swords, want 1 after declining", n)
	}
	effectstest.CastCatalogSpell(t, g, "Steel Golem", "Artifact Creature — Golem", "", nil)
	if !gapB01PassUntilConfirm(t, g, me.ID) {
		t.Fatal("declining the first copy used up the once-each-turn allowance")
	}
}

func TestGapB01EverforgerIgnoresOtherSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gapB01Everforger(g, me)
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	if gapB01PassUntilConfirm(t, g, me.ID) {
		t.Fatal("a nonartifact creature spell was offered a copy")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Rock", "Artifact", "", nil)
	if gapB01PassUntilConfirm(t, g, me.ID) {
		t.Fatal("a noncreature, non-Equipment artifact spell was offered a copy")
	}
}

// --- Dragon's Smile -----------------------------------------------------------

func TestGapB01DragonsSmileOffersAFreeCastFromExileOnceOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, first := effectstest.ImpulseExileCost(t, g, "Test Bolt", "{2}{R}")
	effectstest.AdvanceToMainForTest(t, g)
	effectstest.B12Push(g, me.ID, "Dragon's Smile", "Legendary Artifact — Vehicle", gapB01DragonsSmileOracle, 4, 5)
	key := effects.GrantedAltCostZeroFromExileOnceYourTurn
	offer := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, first, game.ZoneExile), key)
	if offer == nil {
		t.Fatal("Dragon's Smile offered no {0} cast for a spell in exile")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err != nil {
		t.Fatalf("free cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// Once during each of your turns.
	_, second := effectstest.ImpulseExileCost(t, g, "Test Shock", "{R}")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, second, game.ZoneExile), key) != nil {
		t.Fatal("a second free cast was offered in one turn")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, FromZone: "exile", AlternativeCost: key}); err == nil {
		t.Fatal("the cast gate accepted a second claim in one turn")
	}
	// Not during an opponent's turn, though the claim is fresh there.
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, second, game.ZoneExile), key) != nil {
		t.Fatal("a free cast was offered during an opponent's turn")
	}
	// Back on my turn, once again.
	effectstest.AdvanceToMainOf(t, g, seat)
	_, next := effectstest.ImpulseExileCost(t, g, "Test Zap", "{R}")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, next, game.ZoneExile), key) == nil {
		t.Fatal("the once-each-turn claim did not come back on my next turn")
	}
}

func TestGapB01DragonsSmileOffersNoFreeCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Dragon's Smile", "Legendary Artifact — Vehicle", gapB01DragonsSmileOracle, 4, 5)
	spell := effectstest.HandSorcery(me, "Hand Sorcery", "{1}{B}{B}")
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, spell, game.ZoneHand), effects.GrantedAltCostZeroFromExileOnceYourTurn) != nil {
		t.Fatal("a spell in hand was offered the exile-only {0} cast")
	}
}

func TestGapB01DragonsSmileCrewedHitExilesThatManyCardsYouMayPlay(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	smile := s.PushVehicle("Dragon's Smile", gapB01DragonsSmileOracle, 4, 5)
	crew := s.PushCreature("Crew", "Creature — Human", 3, 3)
	s.CrewAndResolve(t, smile, 0, crew)
	library := len(s.Opp.Library.Cards)
	effectstest.AttackWith(t, s.G, s.Opp.ID, smile)
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := library - len(s.Opp.Library.Cards); got != 4 {
		t.Fatalf("%d cards left the damaged player's library, want 4 (the damage dealt)", got)
	}
	exiled := 0
	for _, c := range s.G.Exile.Cards {
		if c.Owner == s.Opp.ID {
			exiled++
			if s.G.CastPermissionOnCardByIDForEffect(c.InstanceID) == nil {
				t.Errorf("the exiled %s carries no permission for Dragon's Smile's controller", c.Name)
			}
		}
	}
	if exiled != 4 {
		t.Errorf("%d of the damaged player's cards are in exile, want 4", exiled)
	}
}

func TestGapB01DragonsSmileCrew3NeedsPowerThree(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	smile := s.PushVehicle("Dragon's Smile", gapB01DragonsSmileOracle, 4, 5)
	weak := s.PushCreature("Weak Crew", "Creature — Human", 2, 2)
	effectstest.ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, smile, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{weak}}); err == nil {
		t.Fatal("Crew 3 was paid with power 2")
	}
}

// --- Jeweled Amulet ------------------------------------------------------------

func gapB01Amulet(t *testing.T, g *game.Game) (me *game.Player, amulet uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	amulet = effectstest.B12Push(g, me.ID, "Jeweled Amulet", "Artifact", gapB01JeweledAmuletOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	return me, amulet
}

func gapB01Noted(g *game.Game, id uuid.UUID) string {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.ChosenColor
}

func TestGapB01JeweledAmuletNotesTheManaSpentAndPaysItBack(t *testing.T) {
	for _, color := range []string{"G", "C", "R"} {
		t.Run(color, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, amulet := gapB01Amulet(t, g)
			effectstest.FillPoolColored(me, color, 1)
			if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("charge: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CounterCount(g, amulet, game.CounterCharge); n != 1 {
				t.Fatalf("%d charge counters, want 1", n)
			}
			if got := gapB01Noted(g, amulet); got != color {
				t.Fatalf("noted %q, want %q", got, color)
			}
			// Tapped from the first ability: untap, then take the mana.
			effectstest.UntapForTest(g, amulet)
			effectstest.ActivateManaFor(t, g, me.ID, amulet, 0, game.ManaAbilityParams{})
			if got := effectstest.PoolCount(me, color); got != 1 || me.ManaPool.Total() != 1 {
				t.Errorf("pool holds %d {%s} of %d, want exactly one {%s}", got, color, me.ManaPool.Total(), color)
			}
			if n := effectstest.CounterCount(g, amulet, game.CounterCharge); n != 0 {
				t.Errorf("%d charge counters left, want 0: the mana ability removes one", n)
			}
			// The last noted type stays noted.
			if got := gapB01Noted(g, amulet); got != color {
				t.Errorf("the note changed to %q after the mana was taken", got)
			}
		})
	}
}

func TestGapB01JeweledAmuletCantChargeWhileItHasACounterAndNeedsACounterForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, amulet := gapB01Amulet(t, g)
	// No counter, no mana: the cost can't be paid.
	if err := g.ActivateManaAbility(me.ID, amulet, 0, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the mana ability was activated with no charge counter")
	}
	effectstest.AddCountersForTest(g, amulet, game.CounterCharge, 1)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the charge ability was activated while the Amulet had a charge counter")
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("the refused activation spent mana: pool %d", me.ManaPool.Total())
	}
}

func TestGapB01JeweledAmuletLastNotedTypeReplacesTheEarlierOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, amulet := gapB01Amulet(t, g)
	for _, color := range []string{"W", "B"} {
		effectstest.FillPoolColored(me, color, 1)
		if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("charge with %s: %v", color, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.UntapForTest(g, amulet)
		effectstest.ActivateManaFor(t, g, me.ID, amulet, 0, game.ManaAbilityParams{})
		if got := effectstest.PoolCount(me, color); got != 1 || me.ManaPool.Total() != 1 {
			t.Errorf("pool %v, want exactly one {%s}", effectstest.PoolColours(me), color)
		}
		me.ManaPool.EmptyPool()
		effectstest.UntapForTest(g, amulet)
	}
	if got := gapB01Noted(g, amulet); got != "B" {
		t.Errorf("last noted %q, want B", got)
	}
}
