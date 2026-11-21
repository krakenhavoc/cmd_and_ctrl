package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB103MolimoOracle   = "b87a2d20-4e9f-4256-9d24-80c38515119e"
	gapB103NemataOracle   = "5ce58279-7f1e-4ef7-a352-ecbaee31b6ff"
	gapB103NessianOracle  = "45703364-660a-4e73-a467-5ee47d936c1a"
	gapB103NyxbornOracle  = "58e20ca7-5fe3-49ff-bee0-0388722e3cab"
	gapB103OviyaOracle    = "41e03224-bcdd-4127-a19a-fe66196339b9"
	gapB103PollenOracle   = "828fe0c0-e280-4a8b-9e49-68cb34c2da97"
	gapB103QuickOracle    = "530a5740-041c-42df-8fe0-fcea87258038"
	gapB103SakikoOracle   = "e4021e30-16bf-4178-9a3b-565bfdce4384"
	gapB103SaprolingToken = "Token Creature — Saproling"
)

func TestGapB103CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB103MolimoOracle:  "Molimo, Maro-Sorcerer",
		gapB103NemataOracle:  "Nemata, Grove Guardian",
		gapB103NessianOracle: "Nessian Wanderer",
		gapB103NyxbornOracle: "Nyxborn Behemoth",
		gapB103OviyaOracle:   "Oviya Pashiri, Sage Lifecrafter",
		gapB103PollenOracle:  "Pollenbright Druid",
		gapB103QuickOracle:   "Quick Sliver",
		gapB103SakikoOracle:  "Sakiko, Mother of Summer",
	} {
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

func TestGapB103MolimoIsAsBigAsYourLandCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	molimo := b12Push(g, me.ID, "Molimo, Maro-Sorcerer", "Legendary Creature — Elemental Sorcerer", gapB103MolimoOracle, 0, 0)
	land := func(owner uuid.UUID, name, typeLine string) {
		apaPush(g, owner, owner, game.Card{Name: name, TypeLine: typeLine})
	}
	for i := 0; i < 4; i++ {
		land(me.ID, "Forest", "Basic Land — Forest")
	}
	land(opp.ID, "Island", "Basic Land — Island")
	c := apaLive(g, molimo)
	if c.PowerForComparison() != 4 || rprB30Toughness(t, g, molimo) != 4 {
		t.Fatalf("Molimo is %d/%d with four lands, want 4/4", c.PowerForComparison(), rprB30Toughness(t, g, molimo))
	}
	land(me.ID, "Forest", "Basic Land — Forest")
	if got := apaLive(g, molimo).PowerForComparison(); got != 5 {
		t.Fatalf("Molimo is %d/x with five lands, want 5", got)
	}
	if !hasEffectiveKeyword(t, g, molimo, "trample") {
		t.Error("Molimo has no trample")
	}
}

func TestGapB103NemataMakesSaprolingsAndPumpsEveryoneForASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	nemata := b12Push(g, me.ID, "Nemata, Grove Guardian", "Legendary Creature — Treefolk", gapB103NemataOracle, 4, 5)
	apaMana(me, "C", "C", "G")
	apaActivate(t, g, me, nemata, 0, game.ActivateAbilityParams{})
	if n := b27CountNamed(g, me.ID, "Saproling"); n != 1 {
		t.Fatalf("%d Saprolings after the token ability, want 1", n)
	}
	mine := b12Creature(g, me.ID, "Saproling", gapB103SaprolingToken, 1, 1)
	theirs := b12Creature(g, opp.ID, "Saproling", gapB103SaprolingToken, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, nemata, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a non-Saproling was accepted as the sacrifice")
	}
	apaActivate(t, g, me, nemata, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{mine}})
	if g.Battlefield.Contains(mine) {
		t.Fatal("the Saproling was not sacrificed")
	}
	if got := apaLive(g, theirs).PowerForComparison(); got != 2 {
		t.Errorf("an opponent's Saproling has power %d, want 2", got)
	}
	if got := apaLive(g, bear).PowerForComparison(); got != 2 {
		t.Errorf("a Bear has power %d, want it unchanged at 2", got)
	}
	if got := apaLive(g, nemata).PowerForComparison(); got != 4 {
		t.Errorf("Nemata has power %d, want 4: she is not a Saproling", got)
	}
}

func TestGapB103NessianWandererTakesALandFromTheTopThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Nessian Wanderer", "Creature — Satyr Scout", gapB103NessianOracle, 1, 3)
	rock := seedLibraryTop(me, "Deep Rock", "Artifact", "{1}")
	bear := seedLibraryTop(me, "Deep Bear", "Creature — Bear", "{1}{G}")
	forest := seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	top := me.Library.Size()
	castCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || !slices.Contains(pick.ChooseCards, forest) || slices.Contains(pick.ChooseCards, bear) || slices.Contains(pick.ChooseCards, rock) {
		t.Fatalf("the prompt should offer only the land: %+v", pick)
	}
	answerChooseCards(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(forest) {
		t.Fatal("the land is not in hand")
	}
	if me.Library.Size() != top-1 || me.Library.Cards[0].InstanceID != bear && me.Library.Cards[0].InstanceID != rock {
		t.Errorf("the other two did not go to the bottom (library %d, was %d)", me.Library.Size(), top)
	}
}

func TestGapB103NessianWandererIgnoresAnOpponentsEnchantmentAndItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	seedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	b12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Nessian Wanderer", "Creature — Satyr Scout", gapB103NessianOracle, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("an opponent's enchantment, or Nessian Wanderer itself, triggered it")
	}
}

// Nyxborn Behemoth costs {X} less, X the total mana value of noncreature
// enchantments I control: creature enchantments and an opponent's count for
// nothing.
func TestGapB103NyxbornBehemothCostsLessForNoncreatureEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	for _, c := range []struct {
		owner    uuid.UUID
		typeLine string
		cost     string
	}{
		{me.ID, "Enchantment", "{2}{W}"},
		{me.ID, "Enchantment — Aura", "{1}"},
		{me.ID, "Enchantment Creature — Beast", "{5}"},
		{opp.ID, "Enchantment", "{4}"},
	} {
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test " + c.cost, TypeLine: c.typeLine, ManaCost: c.cost,
			Owner: c.owner, Controller: c.owner})
	}
	behemoth := trekB39Hand(g, me, "Nyxborn Behemoth", "Enchantment Creature — Beast", "{10}{G}{G}", gapB103NyxbornOracle, 10, 10)
	// {10}{G}{G} less 4 is {6}{G}{G}: eight mana.
	floatMana(t, g, me, "{G}{G}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, behemoth, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for seven mana: the discount is more than four")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, behemoth, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {6}{G}{G}: %v", err)
	}
}

func TestGapB103NyxbornBehemothSacrificesAnotherEnchantmentForIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	behemoth := b12Push(g, me.ID, "Nyxborn Behemoth", "Enchantment Creature — Beast", gapB103NyxbornOracle, 10, 10)
	apaMana(me, "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, behemoth, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{behemoth}}); err == nil {
		t.Fatal("the Behemoth sacrificed itself")
	}
	apaMana(me, "G", "C")
	ench := b12Permanent(g, me.ID, "Test Enchantment", "Enchantment")
	me.ManaPool = game.ManaPool{}
	apaMana(me, "G", "C")
	apaActivate(t, g, me, behemoth, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{ench}})
	if g.Battlefield.Contains(ench) {
		t.Fatal("the other enchantment was not sacrificed")
	}
	if !hasEffectiveKeyword(t, g, behemoth, "indestructible") {
		t.Error("the Behemoth is not indestructible")
	}
	if !hasEffectiveKeyword(t, g, behemoth, "trample") {
		t.Error("the Behemoth has no trample")
	}
}

func TestGapB103OviyaMakesAServoAndAConstructAsBigAsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	oviya := pushCatalogPermanent(g, me.ID, "Oviya Pashiri, Sage Lifecrafter", "Legendary Creature — Human Artificer", gapB103OviyaOracle, false)
	apaMana(me, "C", "C", "G")
	apaActivate(t, g, me, oviya, 0, game.ActivateAbilityParams{})
	if n := b27CountNamed(g, me.ID, "Servo"); n != 1 {
		t.Fatalf("%d Servos, want 1", n)
	}
	untap(g, oviya)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Oviya, the Servo and the Bear: three creatures of mine.
	apaMana(me, "C", "C", "C", "C", "G")
	apaActivate(t, g, me, oviya, 1, game.ActivateAbilityParams{})
	var construct game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" && c.Controller == me.ID {
			construct = c
		}
	}
	if construct.InstanceID == uuid.Nil {
		t.Fatal("no Construct token")
	}
	live := apaLive(g, construct.InstanceID)
	if live.PowerForComparison() != 3 || rprB30Toughness(t, g, construct.InstanceID) != 3 {
		t.Errorf("the Construct is %d/%d, want 3/3", live.PowerForComparison(), rprB30Toughness(t, g, construct.InstanceID))
	}
	if !live.IsArtifact() || !live.IsCreature() {
		t.Error("the Construct is not an artifact creature")
	}
}

func TestGapB103PollenbrightDruidPutsACounterOrProliferates(t *testing.T) {
	t.Run("counter", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		castRingCreature(t, g, "Pollenbright Druid", "Creature — Elf Druid", gapB103PollenOracle, 1, 1)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
		if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
			t.Errorf("the target has %d +1/+1 counters, want 1", n)
		}
	})
	t.Run("proliferate", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		s82OneGivePoison(g, opp, 1)
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		castRingCreature(t, g, "Pollenbright Druid", "Creature — Elf Druid", gapB103PollenOracle, 1, 1)
		passPriorityAroundTable(t, g)
		pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
		if pick == nil {
			t.Fatalf("no mode prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		settleAnsweringProliferate(t, g)
		if poisonOn(opp) != 2 {
			t.Errorf("poison %d, want 2", poisonOn(opp))
		}
		if n := countersOn(g, bear, game.CounterPlusOne); n != 0 {
			t.Errorf("the Bear has %d counters: the proliferate mode put a counter on it", n)
		}
	})
}

// Quick Sliver has flash itself, and gives every player flash for Sliver
// spells and no other.
func TestGapB103QuickSliverGivesFlashToSliverSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	advanceTo(t, g, game.StepBeginCombat)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cast := func(name, typeLine string) error {
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{0}", Power: 1, Toughness: 1,
			Owner: active.ID, Controller: active.ID})
		return g.CastSpell(active.ID, id, game.CastSpellParams{})
	}
	if err := cast("Early Sliver", "Creature — Sliver"); err == nil {
		t.Fatal("a Sliver was cast at the beginning of combat with no Quick Sliver")
	}
	quick := seedEnchantment(g, gapB103QuickOracle, "Quick Sliver", "Creature — Sliver", owner.ID)
	if err := cast("Flash Sliver", "Creature — Sliver"); err != nil {
		t.Errorf("an opponent's Quick Sliver gave me no flash for a Sliver: %v", err)
	}
	if err := cast("Late Bear", "Creature — Bear"); err == nil {
		t.Error("a Bear was cast at the beginning of combat")
	}
	if !hasEffectiveKeyword(t, g, quick, "flash") {
		t.Error("Quick Sliver has no flash of its own")
	}
}

// Sakiko adds as much {G} as each creature of mine dealt to a player, and
// the mana stays through the steps until the end of the turn.
func TestGapB103SakikoAddsGreenForCombatDamageAndKeepsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, me.ID, "Sakiko, Mother of Summer", "Legendary Creature — Snake Shaman", gapB103SakikoOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 3, 3)
	declareAttack(t, g, opp.ID, bear, wolf)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("G"); got != 5 {
		t.Fatalf("%d {G} after a 2 and a 3 connected, want 5", got)
	}
	advanceTo(t, g, game.StepEndCombat)
	if got := me.ManaPool.CountColor("G"); got != 5 {
		t.Errorf("%d {G} after combat ended, want the mana kept", got)
	}
}

func TestGapB103SakikoIgnoresAnOpponentsCreatureAndANonPlayerTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	b12Push(g, opp.ID, "Sakiko, Mother of Summer", "Legendary Creature — Snake Shaman", gapB103SakikoOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := opp.ManaPool.CountColor("G"); got != 0 {
		t.Errorf("an opponent's Sakiko added %d {G} for my creature", got)
	}
}
