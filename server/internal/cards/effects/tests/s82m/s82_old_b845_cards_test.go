package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b845_cards_test.go — Ice Age, batch 845 (#5636): each test plays the
// real catalog entry through the real engine paths.

const (
	b845MountainTitan = "160c8e94-ccf9-4058-b7fd-04529456f0fa"
	b845TimeBomb      = "29062094-594d-4557-9496-11a15d7e3bb0"
	b845Hydra         = "29a12f75-d66d-457f-85c4-6b9450adc0d8"
	b845IceCauldron   = "2b7e4e70-53b1-4d4f-bdcc-56c2e15f367a"
	b845Wiitigo       = "7f956a21-854e-41cc-96f9-ad774e69fa8e"
	b845Armory        = "aba53a54-8666-430d-899e-45ad9f783a4a"
	b845Iceberg       = "cff3fba2-9c5d-4e98-9ff0-0cc0b3bc8dbe"
	b845Winds         = "ffeb792a-6df0-46c1-badc-4deb857f7d48"
)

func TestS82OldB845CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b845MountainTitan: "Mountain Titan", b845TimeBomb: "Time Bomb", b845Hydra: "Balduvian Hydra",
		b845IceCauldron: "Ice Cauldron", b845Wiitigo: "Wiitigo", b845Armory: "Dwarven Armory",
		b845Iceberg: "Iceberg", b845Winds: "Freyalise's Winds",
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

func b845Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[kind]
}

// b845Enter puts the card onto the battlefield from the hand, so its entry
// replacements apply.
func b845Enter(t *testing.T, g *game.Game, owner *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	c := game.NewCard(name, owner.ID)
	c.TypeLine, c.OracleID, c.Power, c.Toughness = typeLine, oracle, power, toughness
	owner.Hand.PushTop(c)
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, id).SummonedThisTurn = false })
	return id
}

func TestS82OldB845DwarvenArmoryWorksOnlyInAnUpkeepAndSacrificesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	armory := effectstest.B12Push(g, me.ID, "Dwarven Armory", "Enchantment", b845Armory, 0, 0)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		lands = append(lands, effectstest.BlbB1Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0))
	}
	effectstest.AdvanceToMain(t, g)
	params := game.ActivateAbilityParams{Strict: true, AutoTap: true, Targets: effectstest.CardRefs(bear), SacrificeIDs: []uuid.UUID{lands[2]}}
	if err := g.ActivateCatalogAbility(me.ID, armory, 0, params); err == nil {
		t.Fatal("activated in the main phase")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if err := g.ActivateCatalogAbility(me.ID, armory, 0, params); err != nil {
		t.Fatalf("activation during an opponent's upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, bear, "+2/+2"); got != 1 {
		t.Fatalf("+2/+2 counters = %d, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tough)
	}
	left := 0
	for _, id := range lands {
		if g.Battlefield.Contains(id) {
			left++
		}
	}
	if left != 2 {
		t.Errorf("%d of three lands left, want 2 (one sacrificed)", left)
	}
}

func TestS82OldB845TimeBombCountsUpkeepsThenBlowsUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.B12Push(g, me.ID, "Time Bomb", "Artifact", b845TimeBomb, 0, 0)
	mine := effectstest.B16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 5, 5, "G")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, bomb, "time"); got != 0 {
		t.Fatalf("an opponent's upkeep put %d time counters on it", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, bomb, "time"); got != 2 {
		t.Fatalf("time counters = %d, want 2", got)
	}
	effectstest.AdvanceToMain(t, g)
	life, oppLife := me.Life, opp.Life
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bomb) {
		t.Error("the Bomb was not sacrificed")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the 5/5 died to 2 damage")
	}
	if me.Life != life-2 || opp.Life != oppLife-2 {
		t.Errorf("life %d and %d, want %d and %d", me.Life, opp.Life, life-2, oppLife-2)
	}
}

func TestS82OldB845IcebergEntersWithXAndIsAManaRock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U", "U", "U")
	effectstest.Pr8CastX(t, g, "Iceberg", "Enchantment", b845Iceberg, 2, nil)
	var iceberg uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == b845Iceberg {
			iceberg = c.InstanceID
		}
	}
	if iceberg == uuid.Nil {
		t.Fatal("the Iceberg did not resolve")
	}
	if got := b845Counters(t, g, iceberg, "ice"); got != 2 {
		t.Fatalf("entered with %d ice counters, want 2", got)
	}
	before := me.ManaPool.Total()
	if err := g.ActivateManaAbility(me.ID, iceberg, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("remove an ice counter: %v", err)
	}
	if got := b845Counters(t, g, iceberg, "ice"); got != 1 {
		t.Errorf("after the mana ability %d ice counters, want 1", got)
	}
	if n := me.ManaPool.Total() - before; n != 1 {
		t.Errorf("the mana ability added %d mana, want 1", n)
	}
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, iceberg, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{3}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, iceberg, "ice"); got != 2 {
		t.Errorf("after {3} %d ice counters, want 2", got)
	}
}

func TestS82OldB845WiitigoGrowsOnlyAfterABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wiitigo := b845Enter(t, g, me, "Wiitigo", "Creature — Yeti", b845Wiitigo, 0, 0)
	if got := b845Counters(t, g, wiitigo, game.CounterPlusOne); got != 6 {
		t.Fatalf("entered with %d counters, want 6", got)
	}
	// No block yet: the next upkeep takes one off.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, wiitigo, game.CounterPlusOne); got != 5 {
		t.Fatalf("after an idle turn %d counters, want 5", got)
	}
	// Wiitigo blocks during the opponent's turn: the next upkeep adds one.
	attacker := effectstest.B16Creature(g, opp.ID, "Attacker", "Creature — Bear", 1, 1, "G")
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, attacker).SummonedThisTurn = false })
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wiitigo, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, wiitigo, game.CounterPlusOne); got != 6 {
		t.Fatalf("after blocking %d counters, want 6", got)
	}
	// The block was before that upkeep, so the next idle turn shrinks it again.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, wiitigo, game.CounterPlusOne); got != 5 {
		t.Errorf("after another idle turn %d counters, want 5", got)
	}
}

func TestS82OldB845WiitigoGrowsWhenItIsBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wiitigo := b845Enter(t, g, me, "Wiitigo", "Creature — Yeti", b845Wiitigo, 0, 0)
	blocker := effectstest.B16Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 1, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, wiitigo)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, wiitigo); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, wiitigo, game.CounterPlusOne); got != 7 {
		t.Errorf("a blocked Wiitigo has %d counters, want 7", got)
	}
}

func TestS82OldB845BalduvianHydraCountersShieldAndGrow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R")
	c := game.NewCard("Balduvian Hydra", me.ID)
	c.TypeLine, c.OracleID, c.ManaCost, c.Power, c.Toughness, c.Colors = "Creature — Hydra", b845Hydra, "{X}{R}{R}", 0, 1, []string{"R"}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hydra := c.InstanceID
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("the Hydra did not resolve")
	}
	if p, tough := effectstest.PtOf(t, g, hydra); p != 3 || tough != 1 {
		t.Fatalf("X=3 Hydra is %d/%d, want 3/1", p, tough)
	}
	// Shield: remove a counter, prevent 1 damage, so 1 damage leaves it alive.
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("remove a counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), hydra, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("1 damage with one prevented killed the Hydra")
	}
	if p, _ := effectstest.PtOf(t, g, hydra); p != 2 {
		t.Errorf("power %d after paying a counter, want 2", p)
	}
	// {R}{R}{R}: only during your upkeep.
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, hydra, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("grew in the main phase")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, hydra, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("grow in my upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, hydra, "+1/+0"); got != 3 {
		t.Errorf("+1/+0 counters = %d, want 3", got)
	}
}

func TestS82OldB845MountainTitanGrowsForEachBlackSpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	titan := effectstest.B12Push(g, me.ID, "Mountain Titan", "Creature — Giant", b845MountainTitan, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, titan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	cast := func(name, color string) {
		c := game.NewCard(name, me.ID)
		c.TypeLine, c.ManaCost, c.Colors = "Sorcery", "{"+color+"}", []string{color}
		me.Hand.PushTop(c)
		effectstest.B06AddMana(me, color)
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast("Red Spell", "R")
	if got := b845Counters(t, g, titan, game.CounterPlusOne); got != 0 {
		t.Fatalf("a red spell gave %d counters", got)
	}
	cast("Black Spell", "B")
	cast("Another Black Spell", "B")
	if got := b845Counters(t, g, titan, game.CounterPlusOne); got != 2 {
		t.Errorf("two black spells gave %d counters, want 2", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if got := b845Counters(t, g, titan, game.CounterPlusOne); got != 2 {
		t.Errorf("counters = %d at the next turn, want 2", got)
	}
}

func TestS82OldB845FreyalisesWindsKeepAPermanentTappedOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Freyalise's Winds", "Enchantment", b845Winds, 0, 0)
	mine := effectstest.B16Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2, "G")
	theirs := effectstest.B16Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		_ = g.TapTargetForEffect(mine)
		_ = g.TapTargetForEffect(theirs)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := b845Counters(t, g, mine, "wind"); got != 1 {
		t.Fatalf("my tapped creature has %d wind counters, want 1", got)
	}
	if got := b845Counters(t, g, theirs, "wind"); got != 1 {
		t.Fatalf("their tapped creature has %d wind counters, want 1", got)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || !c.Tapped {
		t.Fatal("mine untapped in an opponent's untap step")
	}
	if got := b845Counters(t, g, theirs, "wind"); got != 0 {
		t.Errorf("their creature kept %d wind counters through its untap step", got)
	}
	if c := effectstest.FindBattlefieldCardByID(g, theirs); c == nil || !c.Tapped {
		t.Error("their creature untapped though it had a wind counter")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if got := b845Counters(t, g, mine, "wind"); got != 0 {
		t.Errorf("my creature kept %d wind counters", got)
	}
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || !c.Tapped {
		t.Error("my creature untapped through its first untap step")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if c := effectstest.FindBattlefieldCardByID(g, mine); c == nil || c.Tapped {
		t.Error("my creature stayed tapped through its second untap step")
	}
}

func TestS82OldB845IceCauldronExilesNotesAndPaysOnlyForThatCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cauldron := effectstest.B12Push(g, me.ID, "Ice Cauldron", "Artifact", b845IceCauldron, 0, 0)
	spell := game.Card{InstanceID: uuid.New(), Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{2}{R}{R}", Owner: me.ID, Controller: me.ID}
	land := game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(spell)
	me.Hand.PushTop(land)
	effectstest.AdvanceToMain(t, g)
	// X = 3, paid with R, R, G.
	effectstest.B06AddMana(me, "R", "R", "G")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no choice of a card from hand")
	}
	if effectstest.ContainsID(ch.ChooseCards, land.InstanceID) {
		t.Error("a land was offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, spell.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell.InstanceID) {
		t.Fatal("the card was not exiled")
	}
	if _, ok := effectstest.AirbendGrantOn(g, spell.InstanceID); !ok {
		t.Fatal("no permission to cast the exiled card")
	}
	if got := b845Counters(t, g, cauldron, game.CounterCharge); got != 1 {
		t.Fatalf("charge counters = %d, want 1", got)
	}
	// A second activation is refused while a charge counter is on it.
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, cauldron).Tapped = false })
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{Strict: true, XValue: 1}); err == nil {
		t.Error("activated with a charge counter on it")
	}
	// The mana ability adds the noted mana, spendable only on that card.
	before := me.ManaPool.Total()
	if err := g.ActivateManaAbility(me.ID, cauldron, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := me.ManaPool.Total() - before; got != 3 {
		t.Fatalf("added %d mana, want 3", got)
	}
	if got := b845Counters(t, g, cauldron, game.CounterCharge); got != 0 {
		t.Errorf("charge counters = %d after the mana ability, want 0", got)
	}
	restricted := 0
	for _, tok := range me.ManaPool {
		for _, r := range tok.Restrictions {
			if r == game.ManaRestrictSpellCard(spell.InstanceID) {
				n := tok.Count
				if n == 0 {
					n = 1
				}
				restricted += n
			}
		}
	}
	if restricted != 3 {
		t.Errorf("%d of the mana is tied to the exiled card, want 3", restricted)
	}
	// The exiled card can be cast with it; another card cannot be paid for with it.
	other := game.Card{InstanceID: uuid.New(), Name: "Other Spell", TypeLine: "Sorcery", ManaCost: "{R}", Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(other)
	g.WithWriteLock(func() {
		me.ManaPool = append(game.ManaPool(nil), game.ManaToken{Color: "R", Count: 1, Restrictions: []string{game.ManaRestrictSpellCard(spell.InstanceID)}})
	})
	if err := g.CastSpell(me.ID, other.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the Cauldron's mana paid for a different card")
	}
}
