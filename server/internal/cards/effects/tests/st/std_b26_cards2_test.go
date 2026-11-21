package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b26_cards2_test.go — Standard-set batch 26, the second half: Niko,
// Ghost Vacuum, Reluctant Role Model, Come Back Wrong, Chainsaw and Cursed
// Recording.

// stdB26Shard puts a Shard token on the battlefield for p.
func stdB26Shard(g *game.Game, p *game.Player) uuid.UUID {
	c := effects.ShardToken()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	niko := effectstest.TrekB13Card(g, me.Hand, me, "Niko, Light of Hope", "Legendary Creature — Human Wizard", "{2}{W}{U}", stdB26NikoOracle, 3, 4)
	if err := effectstest.TrekB13Cast(t, g, me, niko, "{W}{U}{C}{C}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TrekB13Tokens(t, g, me.ID, "Shard")); n != 2 {
		t.Fatalf("Shard tokens %d, want 2", n)
	}
	shard := effectstest.TrekB13Tokens(t, g, me.ID, "Shard")[0]
	if !shard.HasSubtype("Shard") || !shard.HasCardType("enchantment") {
		t.Errorf("the token is %q, want an enchantment Shard", shard.TypeLine)
	}
}

// stdB26NikoTable is Niko with two Shards and a Hill Giant, the active seat's.
func stdB26NikoTable(t *testing.T) (g *game.Game, me *game.Player, niko, giant, s1, s2 uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	niko = effectstest.B27Push(g, me.ID, "Niko, Light of Hope", "Legendary Creature — Human Wizard", stdB26NikoOracle, "{2}{W}{U}", 3, 4, "W", "U")
	giant = effectstest.B27Push(g, me.ID, "Hill Giant", "Creature — Giant", "", "{3}{R}", 3, 3, "R")
	s1, s2 = stdB26Shard(g, me), stdB26Shard(g, me)
	return
}

func stdB26NikoActivate(t *testing.T, g *game.Game, me *game.Player, niko, target uuid.UUID) error {
	t.Helper()
	effectstest.FloatMana(t, g, me, "{C}{C}")
	return g.ActivateCatalogAbility(me.ID, niko, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target), Strict: true})
}

func TestStdB26NikoExilesTheCreatureAndTheShardsCopyItUntilTheEndStep(t *testing.T) {
	g, me, niko, giant, s1, s2 := stdB26NikoTable(t)
	legend := effectstest.B27Push(g, me.ID, "Hero", stdB26LegendaryCreat, "", "{2}", 2, 2)
	if err := stdB26NikoActivate(t, g, me, niko, legend); err == nil {
		t.Fatal("targeted a legendary creature")
	}
	if err := stdB26NikoActivate(t, g, me, niko, giant); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, giant); got != game.ZoneExile {
		t.Fatalf("the Hill Giant is in %q, want exile", got)
	}
	for _, s := range []uuid.UUID{s1, s2} {
		v := effectstest.LayeredCard(t, g, s)
		if v.Name != "Hill Giant" || effectstest.EffectivePower(t, g, s) != 3 || !v.IsCreature() {
			t.Fatalf("a Shard is %q (%s) %d power, want a copy of the Hill Giant", v.Name, v.TypeLine, effectstest.EffectivePower(t, g, s))
		}
		if !v.IsToken() {
			t.Error("the copy lost its token status")
		}
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	for _, s := range []uuid.UUID{s1, s2} {
		if v := effectstest.LayeredCard(t, g, s); v.Name != "Shard" {
			t.Errorf("a Shard is still %q as the end step begins, want the copy over", v.Name)
		}
	}
	if got := effectstest.B12ZoneOf(g, giant); got != game.ZoneExile {
		t.Errorf("the Hill Giant is in %q before the delayed trigger resolves, want exile", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(stdB26Named(g, me.ID, "Hill Giant")) != 1 {
		t.Errorf("the Hill Giant did not return at the end step: %d on the battlefield", len(stdB26Named(g, me.ID, "Hill Giant")))
	}
	if len(stdB26Named(g, me.ID, "Shard")) != 2 {
		t.Errorf("Shards %d after the end step, want 2", len(stdB26Named(g, me.ID, "Shard")))
	}
}

func TestStdB26NikoActivatedInTheEndStepLastsUntilTheNextEndStep(t *testing.T) {
	g, me, niko, giant, s1, _ := stdB26NikoTable(t)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if err := stdB26NikoActivate(t, g, me, niko, giant); err != nil {
		t.Fatalf("activate in the end step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.LayeredCard(t, g, s1); v.Name != "Hill Giant" {
		t.Fatalf("a Shard is %q right after the activation, want a Hill Giant", v.Name)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if v := effectstest.LayeredCard(t, g, s1); v.Name != "Hill Giant" {
		t.Fatalf("a Shard is %q in the next turn, want it still a Hill Giant until that end step", v.Name)
	}
	if got := effectstest.B12ZoneOf(g, giant); got != game.ZoneExile {
		t.Fatalf("the Hill Giant is in %q during the next turn, want exile", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if v := effectstest.LayeredCard(t, g, s1); v.Name != "Shard" {
		t.Errorf("a Shard is %q as the next end step begins, want the copy over", v.Name)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(stdB26Named(g, me.ID, "Hill Giant")) != 1 {
		t.Errorf("the Hill Giant did not return at the next end step")
	}
}

func TestStdB26NikoCopiesWhatACloneWasCopying(t *testing.T) {
	g, me, niko, _, s1, _ := stdB26NikoTable(t)
	clone := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clone", TypeLine: "Creature — Shapeshifter", Power: 0, Toughness: 0,
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		v := game.PrintedValues{Name: "Sengir Vampire", TypeLine: "Creature — Vampire", Power: 4, Toughness: 4}
		g.BecomeCopyForEffect(uuid.Nil, uuid.Nil, []uuid.UUID{clone}, v, game.IndefiniteDuration(), "test clone")
	})
	if v := effectstest.LayeredCard(t, g, clone); v.Name != "Sengir Vampire" {
		t.Fatalf("setup: the Clone is %q", v.Name)
	}
	if err := stdB26NikoActivate(t, g, me, niko, clone); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if v := effectstest.LayeredCard(t, g, s1); v.Name != "Sengir Vampire" || effectstest.EffectivePower(t, g, s1) != 4 {
		t.Errorf("a Shard is %q %d power, want a copy of what the Clone was copying", v.Name, effectstest.EffectivePower(t, g, s1))
	}
}

// --- Ghost Vacuum -------------------------------------------------------------

func TestStdB26VacuumExilesAndLaterReturnsTheCreatureCardsAsFlyingSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vac := effectstest.B27Push(g, me.ID, "Ghost Vacuum", "Artifact", stdB26VacuumOracle, "{1}", 0, 0)
	mine := stdB26Gy(me, "Grizzly Bears", stdB26BearTypeLine)
	theirs := stdB26Gy(opp, "Hill Giant", "Creature — Giant")
	bolt := stdB26Gy(opp, "Lightning Bolt", "Instant")

	for i, id := range []uuid.UUID{mine, theirs, bolt} {
		if i > 0 {
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vac) })
		}
		if err := g.ActivateCatalogAbility(me.ID, vac, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(id), Strict: true}); err != nil {
			t.Fatalf("exile %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.B12ZoneOf(g, id); got != game.ZoneExile {
			t.Fatalf("card %d is in %q, want exile", i, got)
		}
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vac) })

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vac, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated the second ability outside a main phase")
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vac, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.B12ZoneOf(g, vac); got != game.ZoneGraveyard {
		t.Errorf("the Vacuum is in %q, want the graveyard (sacrificed)", got)
	}
	for _, name := range []string{"Grizzly Bears", "Hill Giant"} {
		cs := stdB26Named(g, me.ID, name)
		if len(cs) != 1 {
			t.Fatalf("%s: %d under my control, want 1", name, len(cs))
		}
		id := cs[0].InstanceID
		v := effectstest.LayeredCard(t, g, id)
		if effectstest.EffectivePower(t, g, id) != 1 || effectstest.EffectiveToughness(t, g, id) != 1 {
			t.Errorf("%s is %d/%d, want 1/1", name, effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id))
		}
		if !v.HasSubtype("Spirit") || !v.HasCardType("creature") {
			t.Errorf("%s is %q, want a Spirit in addition to its other types", name, v.TypeLine)
		}
		if stdB26Counters(t, g, id, game.CounterFlying) != 1 || !effectstest.HasString(effectstest.EffectiveAbilities(t, g, id), "flying") {
			t.Errorf("%s has no flying counter or flying", name)
		}
	}
	if !effectstest.LayeredCard(t, g, stdB26Named(g, me.ID, "Hill Giant")[0].InstanceID).HasSubtype("Giant") {
		t.Error("the Hill Giant lost its Giant type")
	}
	if got := effectstest.B12ZoneOf(g, bolt); got != game.ZoneExile {
		t.Errorf("the Lightning Bolt is in %q, want it left in exile", got)
	}
}

func TestStdB26VacuumReturnsNothingWhenItExiledNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vac := effectstest.B27Push(g, me.ID, "Ghost Vacuum", "Artifact", stdB26VacuumOracle, "{1}", 0, 0)
	bolt := stdB26Gy(me, "Lightning Bolt", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, vac, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bolt), Strict: true}); err != nil {
		t.Fatalf("exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vac) })
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vac, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, bolt); got != game.ZoneExile {
		t.Errorf("the Bolt is in %q, want exile", got)
	}
}

// --- Reluctant Role Model -----------------------------------------------------

func TestStdB26RoleModelSurvivalPutsTheChosenCounterOnATappedOne(t *testing.T) {
	for i, kind := range []string{game.CounterFlying, game.CounterLifelink, game.CounterPlusOne} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		model := effectstest.B27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(model) })
		effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, me.ID, i)
		effectstest.PassPriorityAroundTable(t, g)
		if n := stdB26Counters(t, g, model, kind); n != 1 {
			t.Errorf("%s counters %d, want 1", kind, n)
		}
	}
}

func TestStdB26RoleModelSurvivalDoesNothingWhenUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	model := effectstest.B27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if survivorTriggered(g, model) {
		t.Fatal("an untapped Role Model triggered")
	}
}

// stdB26PutCounters puts counters on a permanent through the effect API.
func stdB26PutCounters(g *game.Game, id uuid.UUID, kind string, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, kind, n) })
}

func TestStdB26RoleModelMovesTheDeadCreaturesCountersToATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	dying := effectstest.B12Creature(g, me.ID, "Grizzly Bears", stdB26BearTypeLine, 2, 2)
	heir := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	stdB26PutCounters(g, dying, game.CounterPlusOne, 2)
	stdB26PutCounters(g, dying, game.CounterFlying, 1)

	effectstest.B27Kill(g, dying)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTarget(t, g, heir)
	effectstest.PassPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, heir, game.CounterPlusOne); n != 2 {
		t.Errorf("+1/+1 counters on the target %d, want 2", n)
	}
	if n := stdB26Counters(t, g, heir, game.CounterFlying); n != 1 {
		t.Errorf("flying counters on the target %d, want 1", n)
	}
}

func TestStdB26RoleModelTriggersForItselfToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	model := effectstest.B27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	heir := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	stdB26PutCounters(g, model, game.CounterLifelink, 1)

	effectstest.B27Kill(g, model)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTarget(t, g, heir)
	effectstest.PassPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, heir, game.CounterLifelink); n != 1 {
		t.Errorf("lifelink counters on the target %d, want 1", n)
	}
}

func TestStdB26RoleModelIgnoresCreaturesWithoutCountersAndTheOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	model := effectstest.B27Push(g, me.ID, "Reluctant Role Model", "Creature — Human Survivor", stdB26RoleModelOracle, "{1}{W}", 2, 2, "W")
	plain := effectstest.B12Creature(g, me.ID, "Grizzly Bears", stdB26BearTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	stdB26PutCounters(g, theirs, game.CounterPlusOne, 3)

	effectstest.B27Kill(g, plain)
	effectstest.B27Kill(g, theirs)
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, model) != nil || len(g.PendingTriggers) > 0 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Come Back Wrong", "Sorcery", stdB26ComeBackOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, victim); got != game.ZoneBattlefield {
		t.Fatalf("the creature is in %q, want the battlefield", got)
	}
	if c := effectstest.B12Card(t, g, victim); c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatalf("controller %v owner %v, want me controlling the opponent's card", c.Controller, c.Owner)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, victim); got != game.ZoneGraveyard {
		t.Errorf("the creature is in %q after my end step, want the graveyard", got)
	}
	if !opp.Graveyard.Contains(victim) {
		t.Error("the sacrificed card is not in its owner's graveyard")
	}
}

func TestStdB26ComeBackWrongDoesNotReturnAnIndestructibleCreatureOrAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Darksteel Colossus", TypeLine: "Artifact Creature — Golem", Power: 11, Toughness: 11,
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Come Back Wrong", "Sorcery", stdB26ComeBackOracle, effectstest.CardRefs(tough))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.B12Card(t, g, tough); c.Controller != opp.ID || effectstest.B12ZoneOf(g, tough) != game.ZoneBattlefield {
		t.Errorf("the indestructible creature changed hands or left: zone %q", effectstest.B12ZoneOf(g, tough))
	}

	goblin := effectstest.B12Creature(g, opp.ID, "Goblin", "Token Creature — Goblin", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Come Back Wrong", "Sorcery", stdB26ComeBackOracle, effectstest.CardRefs(goblin))
	effectstest.PassPriorityAroundTable(t, g)
	if len(stdB26Named(g, me.ID, "Goblin")) != 0 || len(stdB26Named(g, opp.ID, "Goblin")) != 0 {
		t.Error("a destroyed token came back")
	}
}

// --- Chainsaw -----------------------------------------------------------------

func TestStdB26ChainsawEntersDealingThreeToUpToOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Chainsaw", "Artifact — Equipment", stdB26ChainsawOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, victim); got != game.ZoneGraveyard {
		t.Errorf("the Hill Giant is in %q after 3 damage, want the graveyard", got)
	}
	_ = me

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	effectstest.CastCatalogSpell(t, g2, "Chainsaw", "Artifact — Equipment", stdB26ChainsawOracle, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if len(stdB26Named(g2, g2.Seats[0].ID, "Chainsaw")) != 1 {
		t.Error("Chainsaw did not enter with no creature to target")
	}
}

func TestStdB26ChainsawCountsOneRevCounterPerBatchOfDeathsAndPumpsTheEquipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	saw := effectstest.B27Push(g, me.ID, "Chainsaw", "Artifact — Equipment", stdB26ChainsawOracle, "{1}{R}", 0, 0, "R")
	bearer := effectstest.B12Creature(g, me.ID, "Grizzly Bears", stdB26BearTypeLine, 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	c := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)

	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{a, b}) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, saw, "rev"); n != 1 {
		t.Fatalf("rev counters %d after two creatures died together, want 1", n)
	}
	effectstest.B27Kill(g, c)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, saw, "rev"); n != 2 {
		t.Fatalf("rev counters %d after a second death, want 2", n)
	}

	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, saw, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bearer), Strict: true}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bearer), effectstest.EffectiveToughness(t, g, bearer); p != 4 || tough != 2 {
		t.Errorf("equipped Bears are %d/%d, want 4/2", p, tough)
	}
	effectstest.B27Kill(g, bearer)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := stdB26Counters(t, g, saw, "rev"); n != 3 {
		t.Errorf("rev counters %d after the bearer died, want 3", n)
	}
}

// --- Cursed Recording ---------------------------------------------------------

func stdB26BoltTheOpponent(t *testing.T, g *game.Game, victim uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestStdB26CursedRecordingCountsInstantsAndSorceriesAndBurnsYouAtSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rec := effectstest.B27Push(g, me.ID, "Cursed Recording", "Artifact", stdB26RecordingOracle, "{2}{R}{R}", 0, 0, "R")

	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rec := effectstest.B27Push(g, me.ID, "Cursed Recording", "Artifact", stdB26RecordingOracle, "{2}{R}{R}", 0, 0, "R")
	before := effectstest.LifeOf(g, opp.ID)

	if err := g.ActivateCatalogAbility(me.ID, rec, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if got := effectstest.SettleKeepingTarget(t, g, me.ID, opp.ID); got != 1 {
		t.Errorf("%d re-target prompts, want 1 (one copy)", got)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != before-6 {
		t.Fatalf("opponent life %d → %d, want -6 (the Bolt and its copy)", before, got)
	}
	stdB26BoltTheOpponent(t, g, opp.ID)
	if got := effectstest.LifeOf(g, opp.ID); got != before-9 {
		t.Errorf("opponent life %d → %d, want -9: the second Bolt is not copied", before, got)
	}
	if err := g.ActivateCatalogAbility(me.ID, rec, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
}
