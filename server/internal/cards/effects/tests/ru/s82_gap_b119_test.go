package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b119_test.go — S82 set GAP, batch 119 (Tales of Middle-earth
// Commander enchantments and instants), through the real catalog.

const (
	gap119Entmoot   = "1f1ad44c-1811-40e3-976b-03f19866abd5"
	gap119Song      = "440366dd-7654-4ede-aa47-7e8a7916e31e"
	gap119Oath      = "3f8f72d8-5c3a-4f31-8741-3c1a55d9b739"
	gap119Fealty    = "c0128941-240a-4ca1-9816-b32f47523fdb"
	gap119Unity     = "288d5415-ef44-400f-ab88-7c8a4deaaf4b"
	gap119Herbs     = "2efb908f-d069-4041-930b-5c1ccec00db2"
	gap119Worship   = "99f9a4c8-b59f-4a41-8577-6c1e4684b240"
	gap119Legolas   = "59154c04-6bc5-4b95-af1f-6de5fed05649"
	gap119Windswift = "39ab72e2-04f2-41cc-ae33-4cc564b1542b"
	gap119Olorin    = "e6f2ba06-2c92-48b7-a95f-ea6bf35aa0fe"
	gap119LearnPast = "f325fd7a-dcae-4892-adf5-de9893a23ac5"
)

func gap119Named(g *game.Game, owner uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == name {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func gap119Counters(g *game.Game, id uuid.UUID, kind string) int {
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[kind]
			}
		}
	})
	return n
}

func gap119ControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	var who uuid.UUID
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			who = c.Controller
		}
	})
	return who
}

func TestGap119Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		gap119Entmoot: "Assemble the Entmoot", gap119Song: "Song of Eärendil",
		gap119Oath: "Oath of Eorl", gap119Fealty: "Fealty to the Realm",
		gap119Unity: "Call for Unity", gap119Herbs: "Of Herbs and Stewed Rabbit",
		gap119Worship: "Worship", gap119Legolas: "Legolas's Quick Reflexes",
		gap119Windswift: "Windswift Slice", gap119Olorin: "Olórin's Searing Light",
		gap119LearnPast: "Learn from the Past",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestGap119EntmootGivesReachAndSacrificesForTappedTreefolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	ent := effectstest.PushCatalogPermanent(g, me.ID, "Assemble the Entmoot", "Enchantment", gap119Entmoot, false)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if !effectstest.P1g2HasKeyword(g, bear, "reach") {
		t.Fatal("a creature you control lacks reach")
	}
	effectstest.GainLifeNow(t, g, ent, me, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ent, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ent) {
		t.Error("the Entmoot was not sacrificed")
	}
	trees := gap119Named(g, me.ID, "Treefolk")
	if len(trees) != 3 {
		t.Fatalf("%d Treefolk, want 3", len(trees))
	}
	for _, id := range trees {
		var c game.Card
		g.ReadSnapshot(func() { c, _ = g.LookupCardForEffect(id) })
		if c.PowerForComparison() != 4 || c.CurrentToughness() != 4 || !c.Tapped {
			t.Errorf("Treefolk %d/%d tapped=%v, want a tapped 4/4", c.PowerForComparison(), c.CurrentToughness(), c.Tapped)
		}
		if gap119Counters(g, id, game.CounterReach) != 1 {
			t.Error("no reach counter")
		}
	}
	// The reach static ended with the enchantment, so the Bear lost it.
	if effectstest.P1g2HasKeyword(g, bear, "reach") {
		t.Error("the Bear kept reach after the Entmoot left")
	}
}

func TestGap119SongOfEarendilThroughItsThreeChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	seat := g.Turn.ActiveSeat
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Song of Eärendil", "Enchantment — Saga", gap119Song, nil)
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d after chapter I, want +2 (the card cast came from the hand pushed for it)", hand, got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Treasure"); n != 1 {
		t.Errorf("%d Treasures after chapter II, want 1", n)
	}
	if n := len(gap119Named(g, me.ID, "Bird")); n != 1 {
		t.Errorf("%d Birds after chapter II, want 1", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap119Counters(g, bear, game.CounterFlying); got != 1 {
		t.Errorf("Bear has %d flying counters, want 1", got)
	}
	for _, id := range gap119Named(g, me.ID, "Bird") {
		if gap119Counters(g, id, game.CounterFlying) != 0 {
			t.Error("a creature that already had flying got a counter")
		}
	}
}

func TestGap119OathOfEorlThroughItsThreeChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	seat := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Oath of Eorl", "Enchantment — Saga", gap119Oath, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gap119Named(g, me.ID, "Human Soldier")); n != 2 {
		t.Fatalf("%d Human Soldiers after chapter I, want 2", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	knights := gap119Named(g, me.ID, "Human Knight")
	if len(knights) != 2 {
		t.Fatalf("%d Human Knights after chapter II, want 2", len(knights))
	}
	for _, kw := range []string{"trample", "haste"} {
		if !effectstest.P1g2HasKeyword(g, knights[0], kw) {
			t.Errorf("a Knight lacks %s", kw)
		}
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: knights[0]}, true)
	if got := gap119Counters(g, knights[0], game.CounterIndestructible); got != 1 {
		t.Errorf("target Human has %d indestructible counters, want 1", got)
	}
	if g.Monarch != me.ID {
		t.Error("chapter III did not make you the monarch")
	}
}

func TestGap119OfHerbsAndStewedRabbitThroughItsThreeChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	seat := g.Turn.ActiveSeat
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Of Herbs and Stewed Rabbit", "Enchantment — Saga", gap119Herbs, nil)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if got := gap119Counters(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("Bear has %d +1/+1 counters, want 1", got)
	}
	if n := effectstest.CountBattlefieldByName(g, "Food"); n != 1 {
		t.Fatalf("%d Foods after chapter I, want 1", n)
	}
	hand := me.Hand.Size()
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldByName(g, "Food"); n != 2 {
		t.Fatalf("%d Foods after chapter II, want 2", n)
	}
	if me.Hand.Size() < hand+1 {
		t.Errorf("hand %d -> %d, want the chapter's draw", hand, me.Hand.Size())
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(gap119Named(g, me.ID, "Halfling")); n != 2 {
		t.Errorf("%d Halflings after chapter III, want one per Food (2)", n)
	}
}

func TestGap119FealtyGivesTheCreatureToTheMonarch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fealty to the Realm", "Enchantment — Aura", gap119Fealty, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Fatalf("monarch = %v, want the Aura's controller", g.Monarch)
	}
	if got := gap119ControllerOf(g, theirs); got != me.ID {
		t.Fatalf("the monarch (you) does not control the enchanted creature: %v", got)
	}
	// The crown moves: so does the creature.
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(opp.ID) })
	if got := gap119ControllerOf(g, theirs); got != opp.ID {
		t.Fatalf("the creature did not follow the crown: %v", got)
	}
	// Now the opponent controls it and it can't attack the Aura's controller.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(theirs, me.ID); err == nil {
		t.Error("the enchanted creature attacked the Aura's controller")
	}
}

func TestGap119CallForUnityGrowsOnRevoltOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	cfu := effectstest.BlbB1Push(g, me.ID, "Call for Unity", "Enchantment", gap119Unity, 0, 0)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	fodder := effectstest.BlbB1Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	// No permanent left: no counter at the end step.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap119Counters(g, cfu, "unity"); got != 0 {
		t.Fatalf("%d unity counters without revolt, want 0", got)
	}
	// A permanent leaves under your control; next end step (this one has begun,
	// so make a fresh turn state).
	g2 := effectstest.NewCatalogGame(t)
	me2, _, _ := effectstest.BlbB1Seats(g2)
	cfu2 := effectstest.BlbB1Push(g2, me2.ID, "Call for Unity", "Enchantment", gap119Unity, 0, 0)
	bear2 := effectstest.BlbB1Push(g2, me2.ID, "Bear", "Creature — Bear", "", 2, 2)
	fodder2 := effectstest.BlbB1Push(g2, me2.ID, "Fodder", "Creature — Bear", "", 1, 1)
	g2.WithWriteLock(func() { _ = g2.DestroyPermanentForEffect(fodder2) })
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := gap119Counters(g2, cfu2, "unity"); got != 1 {
		t.Fatalf("%d unity counters with revolt, want 1", got)
	}
	var power int
	g2.ReadSnapshot(func() { c, _ := g2.LookupCardForEffect(bear2); power = c.PowerForComparison() })
	if power != 3 {
		t.Errorf("Bear power %d, want 2+1", power)
	}
	_, _ = bear, fodder
}

func TestGap119WorshipFloorsDamageOnlyWhileYouControlACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Worship", "Enchantment", gap119Worship, false)
	src := effectstest.BlbB1Push(g, opp.ID, "Attacker", "Creature — Bear", "", 2, 2)
	me.Life = 5
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 9) })
	if me.Life != -4 {
		t.Fatalf("life = %d with no creature, want -4", me.Life)
	}
	effectstest.BlbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	me.Life = 5
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 9) })
	if me.Life != 1 {
		t.Errorf("life = %d with a creature, want 1", me.Life)
	}
	// Life loss is not damage.
	me.Life = 5
	effectstest.GainLifeNow(t, g, src, me, -9)
	if me.Life != -4 {
		t.Errorf("life = %d after losing 9 life, want -4 (only damage is floored)", me.Life)
	}
}

func TestGap119LegolasUntapsGrantsAndDealsDamageWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	elf := effectstest.BlbB1Push(g, me.ID, "Elf", "Creature — Elf Archer", "", 3, 3)
	victim := effectstest.BlbB1Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elf {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Legolas's Quick Reflexes", TypeLine: "Instant", OracleID: gap119Legolas, ManaCost: "{G}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elf}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// Split second: nothing else can be cast while it is on the stack.
	other := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: other, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, other, game.CastSpellParams{}); err == nil {
		t.Error("a spell was cast in response to a split second spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	var tapped bool
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(elf); tapped = c.Tapped })
	if tapped {
		t.Error("the creature was not untapped")
	}
	for _, kw := range []string{"reach", "hexproof"} {
		if !effectstest.P1g2HasKeyword(g, elf, kw) {
			t.Errorf("lacks %s", kw)
		}
	}
	// Tap it: the granted trigger deals its power to a target creature.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(elf) })
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 3 damage from the tapped creature")
	}
}

func TestGap119WindswiftSliceMakesElvesForExcessDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	biter := effectstest.BlbB1Push(g, me.ID, "Big Elf", "Creature — Elf", "", 5, 5)
	victim := effectstest.BlbB1Push(g, opp.ID, "Small Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Windswift Slice", "Instant", gap119Windswift, []game.TargetRef{
		{Kind: game.TargetCard, ID: biter}, {Kind: game.TargetCard, ID: victim, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 5 damage")
	}
	if n := len(gap119Named(g, me.ID, "Elf Warrior")); n != 3 {
		t.Errorf("%d Elf Warriors, want the 3 excess damage", n)
	}
	// Exactly lethal: no excess, no tokens.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, _ := effectstest.BlbB1Seats(g2)
	biter2 := effectstest.BlbB1Push(g2, me2.ID, "Elf", "Creature — Elf", "", 2, 2)
	victim2 := effectstest.BlbB1Push(g2, opp2.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g2, "Windswift Slice", "Instant", gap119Windswift, []game.TargetRef{
		{Kind: game.TargetCard, ID: biter2}, {Kind: game.TargetCard, ID: victim2, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g2)
	if n := len(gap119Named(g2, me2.ID, "Elf Warrior")); n != 0 {
		t.Errorf("%d Elf Warriors with no excess damage", n)
	}
}

func TestGap119OlorinsSearingLightExilesEachOpponentsBiggestAndBurnsWithMastery(t *testing.T) {
	for _, mastery := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, other := effectstest.BlbB1Seats(g)
		bigOpp := effectstest.BlbB1Push(g, opp.ID, "Big", "Creature — Giant", "", 5, 5)
		smallOpp := effectstest.BlbB1Push(g, opp.ID, "Small", "Creature — Bear", "", 2, 2)
		bigOther := effectstest.BlbB1Push(g, other.ID, "Mid", "Creature — Ogre", "", 3, 3)
		if mastery {
			effectstest.PushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
			effectstest.PushGraveyardPermanent(me, "Old Sorcery", "Sorcery", "{1}{R}")
		}
		oppLife, otherLife := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, other.ID)
		effectstest.CastCatalogSpell(t, g, "Olórin's Searing Light", "Instant", gap119Olorin, nil)
		for i := 0; i < 8; i++ {
			if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID); c != nil {
				if err := g.ResolveOwnPermanents(c.ID, opp.ID, []uuid.UUID{bigOpp}); err != nil {
					t.Fatalf("opponent pick: %v", err)
				}
			}
			if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, other.ID); c != nil {
				if err := g.ResolveOwnPermanents(c.ID, other.ID, []uuid.UUID{bigOther}); err != nil {
					t.Fatalf("opponent pick: %v", err)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		if g.Battlefield.Contains(bigOpp) || g.Battlefield.Contains(bigOther) {
			t.Fatalf("mastery=%v: a creature with the greatest power was not exiled", mastery)
		}
		if !g.Battlefield.Contains(smallOpp) {
			t.Errorf("mastery=%v: the smaller creature was exiled", mastery)
		}
		dOpp, dOther := oppLife-effectstest.LifeOf(g, opp.ID), otherLife-effectstest.LifeOf(g, other.ID)
		if mastery && (dOpp != 5 || dOther != 3) {
			t.Errorf("spell mastery damage %d and %d, want 5 and 3", dOpp, dOther)
		}
		if !mastery && (dOpp != 0 || dOther != 0) {
			t.Errorf("damage %d and %d without spell mastery", dOpp, dOther)
		}
	}
}

func TestGap119LearnFromThePastShufflesAGraveyardInAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.PushGraveyardPermanent(opp, "Dead One", "Creature — Bear", "{1}")
	effectstest.PushGraveyardPermanent(opp, "Dead Two", "Creature — Bear", "{1}")
	lib, hand := opp.Library.Size(), me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Learn from the Past", "Instant", gap119LearnPast, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 {
		t.Errorf("graveyard still has %d cards", opp.Graveyard.Size())
	}
	if got := opp.Library.Size(); got != lib+2 {
		t.Errorf("library %d -> %d, want +2", lib, got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want cast one and drew one", hand, got)
	}
}
