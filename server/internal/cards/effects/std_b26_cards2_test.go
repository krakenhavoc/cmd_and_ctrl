package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b26_cards2_test.go — Standard-set batch 26, the second half: Niko,
// Ghost Vacuum, Reluctant Role Model, Come Back Wrong, Chainsaw and Cursed
// Recording.

// stdB26Shard puts a Shard token on the battlefield for p.
func stdB26Shard(g *game.Game, p *game.Player) uuid.UUID {
	c := ShardToken()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	return pushBattlefieldCardWithTimestamp(g, c)
}

// stdB26Named is the battlefield permanents named `name` under controller.
func stdB26Named(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Niko, Light of Hope ------------------------------------------------------

func TestStdB26NikoEntersWithTwoShards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	niko := trekB13Card(g, me.Hand, me, "Niko, Light of Hope", "Legendary Creature — Human Wizard", "{2}{W}{U}", stdB26NikoOracle, 3, 4)
	if err := trekB13Cast(t, g, me, niko, "{W}{U}{C}{C}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(trekB13Tokens(t, g, me.ID, "Shard")); n != 2 {
		t.Fatalf("Shard tokens %d, want 2", n)
	}
	shard := trekB13Tokens(t, g, me.ID, "Shard")[0]
	if !shard.HasSubtype("Shard") || !shard.HasCardType("enchantment") {
		t.Errorf("the token is %q, want an enchantment Shard", shard.TypeLine)
	}
}

// stdB26NikoTable is Niko with two Shards and a Hill Giant, the active seat's.
func stdB26NikoTable(t *testing.T) (g *game.Game, me *game.Player, niko, giant, s1, s2 uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	advanceToMain(t, g)
	niko = b27Push(g, me.ID, "Niko, Light of Hope", "Legendary Creature — Human Wizard", stdB26NikoOracle, "{2}{W}{U}", 3, 4, "W", "U")
	giant = b27Push(g, me.ID, "Hill Giant", "Creature — Giant", "", "{3}{R}", 3, 3, "R")
	s1, s2 = stdB26Shard(g, me), stdB26Shard(g, me)
	return
}

func stdB26NikoActivate(t *testing.T, g *game.Game, me *game.Player, niko, target uuid.UUID) error {
	t.Helper()
	floatMana(t, g, me, "{C}{C}")
	return g.ActivateCatalogAbility(me.ID, niko, 0, game.ActivateAbilityParams{Targets: cardRefs(target), Strict: true})
}

func TestStdB26NikoExilesTheCreatureAndTheShardsCopyItUntilTheEndStep(t *testing.T) {
	g, me, niko, giant, s1, s2 := stdB26NikoTable(t)
	legend := b27Push(g, me.ID, "Hero", stdB26LegendaryCreat, "", "{2}", 2, 2)
	if err := stdB26NikoActivate(t, g, me, niko, legend); err == nil {
		t.Fatal("targeted a legendary creature")
	}
	if err := stdB26NikoActivate(t, g, me, niko, giant); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, giant); got != game.ZoneExile {
		t.Fatalf("the Hill Giant is in %q, want exile", got)
	}
	for _, s := range []uuid.UUID{s1, s2} {
		v := layeredCard(t, g, s)
		if v.Name != "Hill Giant" || effectivePower(t, g, s) != 3 || !v.IsCreature() {
			t.Fatalf("a Shard is %q (%s) %d power, want a copy of the Hill Giant", v.Name, v.TypeLine, effectivePower(t, g, s))
		}
		if !v.IsToken() {
			t.Error("the copy lost its token status")
		}
	}

	advanceTo(t, g, game.StepEnd)
	for _, s := range []uuid.UUID{s1, s2} {
		if v := layeredCard(t, g, s); v.Name != "Shard" {
			t.Errorf("a Shard is still %q as the end step begins, want the copy over", v.Name)
		}
	}
	if got := b12ZoneOf(g, giant); got != game.ZoneExile {
		t.Errorf("the Hill Giant is in %q before the delayed trigger resolves, want exile", got)
	}
	passPriorityAroundTable(t, g)
	if len(stdB26Named(g, me.ID, "Hill Giant")) != 1 {
		t.Errorf("the Hill Giant did not return at the end step: %d on the battlefield", len(stdB26Named(g, me.ID, "Hill Giant")))
	}
	if len(stdB26Named(g, me.ID, "Shard")) != 2 {
		t.Errorf("Shards %d after the end step, want 2", len(stdB26Named(g, me.ID, "Shard")))
	}
}

func TestStdB26NikoActivatedInTheEndStepLastsUntilTheNextEndStep(t *testing.T) {
	g, me, niko, giant, s1, _ := stdB26NikoTable(t)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if err := stdB26NikoActivate(t, g, me, niko, giant); err != nil {
		t.Fatalf("activate in the end step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if v := layeredCard(t, g, s1); v.Name != "Hill Giant" {
		t.Fatalf("a Shard is %q right after the activation, want a Hill Giant", v.Name)
	}
	advanceToNextSeatsTurn(t, g)
	if v := layeredCard(t, g, s1); v.Name != "Hill Giant" {
		t.Fatalf("a Shard is %q in the next turn, want it still a Hill Giant until that end step", v.Name)
	}
	if got := b12ZoneOf(g, giant); got != game.ZoneExile {
		t.Fatalf("the Hill Giant is in %q during the next turn, want exile", got)
	}
	advanceTo(t, g, game.StepEnd)
	if v := layeredCard(t, g, s1); v.Name != "Shard" {
		t.Errorf("a Shard is %q as the next end step begins, want the copy over", v.Name)
	}
	passPriorityAroundTable(t, g)
	if len(stdB26Named(g, me.ID, "Hill Giant")) != 1 {
		t.Errorf("the Hill Giant did not return at the next end step")
	}
}

func TestStdB26NikoCopiesWhatACloneWasCopying(t *testing.T) {
	g, me, niko, _, s1, _ := stdB26NikoTable(t)
	clone := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", Power: 0, Toughness: 0,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		v := game.PrintedValues{Name: "Sengir Vampire", TypeLine: "Creature — Vampire", Power: 4, Toughness: 4}
		g.BecomeCopyForEffect(uuid.Nil, uuid.Nil, []uuid.UUID{clone}, v, game.IndefiniteDuration(), "test clone")
	})
	if v := layeredCard(t, g, clone); v.Name != "Sengir Vampire" {
		t.Fatalf("setup: the Clone is %q", v.Name)
	}
	if err := stdB26NikoActivate(t, g, me, niko, clone); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if v := layeredCard(t, g, s1); v.Name != "Sengir Vampire" || effectivePower(t, g, s1) != 4 {
		t.Errorf("a Shard is %q %d power, want a copy of what the Clone was copying", v.Name, effectivePower(t, g, s1))
	}
}

// --- Ghost Vacuum -------------------------------------------------------------

func TestStdB26VacuumExilesAndLaterReturnsTheCreatureCardsAsFlyingSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vac := b27Push(g, me.ID, "Ghost Vacuum", "Artifact", stdB26VacuumOracle, "{1}", 0, 0)
	mine := stdB26Gy(me, "Grizzly Bears", stdB26BearTypeLine)
	theirs := stdB26Gy(opp, "Hill Giant", "Creature — Giant")
	bolt := stdB26Gy(opp, "Lightning Bolt", "Instant")

	for i, id := range []uuid.UUID{mine, theirs, bolt} {
		if i > 0 {
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vac) })
		}
		if err := g.ActivateCatalogAbility(me.ID, vac, 0, game.ActivateAbilityParams{Targets: cardRefs(id), Strict: true}); err != nil {
			t.Fatalf("exile %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if got := b12ZoneOf(g, id); got != game.ZoneExile {
			t.Fatalf("card %d is in %q, want exile", i, got)
		}
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vac) })

	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vac, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated the second ability outside a main phase")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vac, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := b12ZoneOf(g, vac); got != game.ZoneGraveyard {
		t.Errorf("the Vacuum is in %q, want the graveyard (sacrificed)", got)
	}
	for _, name := range []string{"Grizzly Bears", "Hill Giant"} {
		cs := stdB26Named(g, me.ID, name)
		if len(cs) != 1 {
			t.Fatalf("%s: %d under my control, want 1", name, len(cs))
		}
		id := cs[0].InstanceID
		v := layeredCard(t, g, id)
		if effectivePower(t, g, id) != 1 || effectiveToughness(t, g, id) != 1 {
			t.Errorf("%s is %d/%d, want 1/1", name, effectivePower(t, g, id), effectiveToughness(t, g, id))
		}
		if !v.HasSubtype("Spirit") || !v.HasCardType("creature") {
			t.Errorf("%s is %q, want a Spirit in addition to its other types", name, v.TypeLine)
		}
		if stdB26Counters(t, g, id, game.CounterFlying) != 1 || !hasString(effectiveAbilities(t, g, id), "flying") {
			t.Errorf("%s has no flying counter or flying", name)
		}
	}
	if !layeredCard(t, g, stdB26Named(g, me.ID, "Hill Giant")[0].InstanceID).HasSubtype("Giant") {
		t.Error("the Hill Giant lost its Giant type")
	}
	if got := b12ZoneOf(g, bolt); got != game.ZoneExile {
		t.Errorf("the Lightning Bolt is in %q, want it left in exile", got)
	}
}

func TestStdB26VacuumReturnsNothingWhenItExiledNoCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vac := b27Push(g, me.ID, "Ghost Vacuum", "Artifact", stdB26VacuumOracle, "{1}", 0, 0)
	bolt := stdB26Gy(me, "Lightning Bolt", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, vac, 0, game.ActivateAbilityParams{Targets: cardRefs(bolt), Strict: true}); err != nil {
		t.Fatalf("exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vac) })
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vac, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, bolt); got != game.ZoneExile {
		t.Errorf("the Bolt is in %q, want exile", got)
	}
}

// --- Reluctant Role Model -----------------------------------------------------

func TestStdB26RoleModelSurvivalPutsTheChosenCounterOnATappedOne(t *testing.T) {
	for i, kind := range []string{game.CounterFlying, game.CounterLifelink, game.CounterPlusOne} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		model := b27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(model) })
		advanceTo(t, g, game.StepPostcombatMain)
		passPriorityAroundTable(t, g)
		answerOptionPick(t, g, me.ID, i)
		passPriorityAroundTable(t, g)
		if n := stdB26Counters(t, g, model, kind); n != 1 {
			t.Errorf("%s counters %d, want 1", kind, n)
		}
	}
}

func TestStdB26RoleModelSurvivalDoesNothingWhenUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	model := b27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	advanceTo(t, g, game.StepPostcombatMain)
	if survivorTriggered(g, model) {
		t.Fatal("an untapped Role Model triggered")
	}
}

// stdB26PutCounters puts counters on a permanent through the effect API.
func stdB26PutCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
}

func TestStdB26RoleModelMovesTheDeadCreaturesCountersToATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	dying := b12Creature(g, me.ID, "Grizzly Bears", stdB26BearTypeLine, 2, 2)
	heir := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	stdB26PutCounters(g, dying, game.CounterPlusOne, 2)
	stdB26PutCounters(g, dying, game.CounterFlying, 1)

	b27Kill(g, dying)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, heir)
	passPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, heir, game.CounterPlusOne); n != 2 {
		t.Errorf("+1/+1 counters on the target %d, want 2", n)
	}
	if n := stdB26Counters(t, g, heir, game.CounterFlying); n != 1 {
		t.Errorf("flying counters on the target %d, want 1", n)
	}
}

func TestStdB26RoleModelTriggersForItselfToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	model := b27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	heir := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	stdB26PutCounters(g, model, game.CounterLifelink, 1)

	b27Kill(g, model)
	g.RunStateChecksForTest()
	answerPickTarget(t, g, heir)
	passPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, heir, game.CounterLifelink); n != 1 {
		t.Errorf("lifelink counters on the target %d, want 1", n)
	}
}

func TestStdB26RoleModelIgnoresCreaturesWithoutCountersAndTheOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	model := b27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	plain := b12Creature(g, me.ID, "Grizzly Bears", stdB26BearTypeLine, 2, 2)
	theirs := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	stdB26PutCounters(g, theirs, game.CounterPlusOne, 3)

	b27Kill(g, plain)
	b27Kill(g, theirs)
	g.RunStateChecksForTest()
	if triggerOnStack(g, model) != nil || len(g.PendingTriggers) > 0 {
		t.Fatal("the Role Model triggered for a creature without counters, or an opponent's")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("a target was asked for")
		}
	}
}

// --- Come Back Wrong ----------------------------------------------------------

func TestStdB26ComeBackWrongReturnsTheCreatureUnderYourControlAndSacrificesItAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	castCatalogSpell(t, g, "Come Back Wrong", "Sorcery", stdB26ComeBackOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, victim); got != game.ZoneBattlefield {
		t.Fatalf("the creature is in %q, want the battlefield", got)
	}
	if c := b12Card(t, g, victim); c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatalf("controller %v owner %v, want me controlling the opponent's card", c.Controller, c.Owner)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, victim); got != game.ZoneGraveyard {
		t.Errorf("the creature is in %q after my end step, want the graveyard", got)
	}
	if !opp.Graveyard.Contains(victim) {
		t.Error("the sacrificed card is not in its owner's graveyard")
	}
}

func TestStdB26ComeBackWrongDoesNotReturnAnIndestructibleCreatureOrAToken(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Colossus", TypeLine: "Artifact Creature — Golem", Power: 11, Toughness: 11,
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Come Back Wrong", "Sorcery", stdB26ComeBackOracle, cardRefs(tough))
	passPriorityAroundTable(t, g)
	if c := b12Card(t, g, tough); c.Controller != opp.ID || b12ZoneOf(g, tough) != game.ZoneBattlefield {
		t.Errorf("the indestructible creature changed hands or left: zone %q", b12ZoneOf(g, tough))
	}

	goblin := b12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	castCatalogSpell(t, g, "Come Back Wrong", "Sorcery", stdB26ComeBackOracle, cardRefs(goblin))
	passPriorityAroundTable(t, g)
	if len(stdB26Named(g, me.ID, "Goblin")) != 0 || len(stdB26Named(g, opp.ID, "Goblin")) != 0 {
		t.Error("a destroyed token came back")
	}
}

// --- Chainsaw -----------------------------------------------------------------

func TestStdB26ChainsawEntersDealingThreeToUpToOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	castCatalogSpell(t, g, "Chainsaw", "Artifact — Equipment", stdB26ChainsawOracle, nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, victim); got != game.ZoneGraveyard {
		t.Errorf("the Hill Giant is in %q after 3 damage, want the graveyard", got)
	}
	_ = me

	g2 := newCatalogGame(t)
	advanceToMain(t, g2)
	castCatalogSpell(t, g2, "Chainsaw", "Artifact — Equipment", stdB26ChainsawOracle, nil)
	passPriorityAroundTable(t, g2)
	if len(stdB26Named(g2, g2.Seats[0].ID, "Chainsaw")) != 1 {
		t.Error("Chainsaw did not enter with no creature to target")
	}
}

func TestStdB26ChainsawCountsOneRevCounterPerBatchOfDeathsAndPumpsTheEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	saw := b27Push(g, me.ID, "Chainsaw", "Artifact — Equipment", stdB26ChainsawOracle, "{1}{R}", 0, 0, "R")
	bearer := b12Creature(g, me.ID, "Grizzly Bears", stdB26BearTypeLine, 2, 2)
	a := b12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	b := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	c := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)

	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{a, b}) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, saw, "rev"); n != 1 {
		t.Fatalf("rev counters %d after two creatures died together, want 1", n)
	}
	b27Kill(g, c)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, saw, "rev"); n != 2 {
		t.Fatalf("rev counters %d after a second death, want 2", n)
	}

	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, saw, 0, game.ActivateAbilityParams{Targets: cardRefs(bearer), Strict: true}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bearer), effectiveToughness(t, g, bearer); p != 4 || tough != 2 {
		t.Errorf("equipped Bears are %d/%d, want 4/2", p, tough)
	}
	b27Kill(g, bearer)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, saw, "rev"); n != 3 {
		t.Errorf("rev counters %d after the bearer died, want 3", n)
	}
}

// --- Cursed Recording ---------------------------------------------------------

func stdB26BoltTheOpponent(t *testing.T, g *game.Game, victim uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	passPriorityAroundTable(t, g)
}

func TestStdB26CursedRecordingCountsInstantsAndSorceriesAndBurnsYouAtSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rec := b27Push(g, me.ID, "Cursed Recording", "Artifact", stdB26RecordingOracle, "{2}{R}{R}", 0, 0, "R")

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, rec, "time"); n != 0 {
		t.Fatalf("a creature spell put %d time counters on it", n)
	}
	for i := 1; i <= 6; i++ {
		stdB26BoltTheOpponent(t, g, opp.ID)
		if n := stdB26Counters(t, g, rec, "time"); n != i {
			t.Fatalf("after %d spells: %d time counters", i, n)
		}
	}
	if me.Life != 40 {
		t.Fatalf("life %d before the seventh spell, want 40", me.Life)
	}
	stdB26BoltTheOpponent(t, g, opp.ID)
	if n := stdB26Counters(t, g, rec, "time"); n != 0 {
		t.Errorf("time counters %d after the seventh, want them all removed", n)
	}
	if me.Life != 20 {
		t.Errorf("life %d after the seventh spell, want 20 (20 damage to me)", me.Life)
	}
}

func TestStdB26CursedRecordingTapCopiesTheNextInstantOrSorceryThisTurnOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rec := b27Push(g, me.ID, "Cursed Recording", "Artifact", stdB26RecordingOracle, "{2}{R}{R}", 0, 0, "R")
	before := lifeOf(g, opp.ID)

	if err := g.ActivateCatalogAbility(me.ID, rec, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if got := settleKeepingTarget(t, g, me.ID, opp.ID); got != 1 {
		t.Errorf("%d re-target prompts, want 1 (one copy)", got)
	}
	if got := lifeOf(g, opp.ID); got != before-6 {
		t.Fatalf("opponent life %d → %d, want -6 (the Bolt and its copy)", before, got)
	}
	stdB26BoltTheOpponent(t, g, opp.ID)
	if got := lifeOf(g, opp.ID); got != before-9 {
		t.Errorf("opponent life %d → %d, want -9: the second Bolt is not copied", before, got)
	}
	if err := g.ActivateCatalogAbility(me.ID, rec, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
}
