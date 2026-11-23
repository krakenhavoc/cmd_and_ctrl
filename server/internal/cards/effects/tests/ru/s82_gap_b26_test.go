package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	scalelordReckonerOracle  = "9da65d66-71f0-4e17-a1f9-6da6cd1cfa8a"
	transcendentMasterOracle = "11332981-f06a-48c7-a7b4-fce68e629e82"
	soulOfTherosOracle       = "d820b920-191c-43e3-ba6a-4cdceab1c06d"
	halvingSeasonOracle      = "4d4ad5b2-b22d-497f-bd9a-282bf5c33947"
	sliverOfHopeOracle       = "4d57d6a0-2c4d-466c-8545-f97058a8f1c5"
	manabondOracle           = "482006a0-bb2a-4533-891b-2195da5e0f01"
)

func TestScalelordReckonerPunishesTheOpponentWhoTargetsAnyOfYourDragons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Scalelord Reckoner", "Creature — Dragon", scalelordReckonerOracle, 4, 4)
	whelp := effectstest.B12Push(g, me.ID, "Hellkite Whelp", "Creature — Dragon", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	theirRelic := effectstest.PushPermanentForTest(g, them.ID, "Their Relic", "", "Artifact")
	theirLand := effectstest.PushPermanentForTest(g, them.ID, "Their Forest", "", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)

	cast := func(who *game.Player, target uuid.UUID) {
		spell := game.NewCard("Test Removal", who.ID)
		spell.TypeLine = "Instant"
		who.Hand.PushTop(spell)
		if err := g.CastSpell(who.ID, spell.InstanceID, game.CastSpellParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("cast: %v", err)
		}
	}
	// A non-Dragon of mine being targeted does nothing.
	cast(them, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBoard(g, theirRelic) == false || effectstest.PendingOfKind(g, game.PendingChoicePickTarget) != nil {
		t.Fatal("targeting a non-Dragon triggered Scalelord Reckoner")
	}
	// Targeting the other Dragon does.
	cast(them, whelp)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != me.ID {
		t.Errorf("trigger target chosen by %s, want Scalelord's controller", ch.Chooser)
	}
	for _, id := range ch.PickTargetCards {
		if id == theirLand {
			t.Error("a land was offered as a target")
		}
	}
	effectstest.AnswerPickTarget(t, g, theirRelic)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBoard(g, theirRelic) {
		t.Error("the opponent's nonland permanent was not destroyed")
	}
	if !effectstest.OnBoard(g, theirLand) {
		t.Error("a land was destroyed")
	}
}

func TestScalelordReckonerIgnoresItsOwnControllerTargeting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Scalelord Reckoner", "Creature — Dragon", scalelordReckonerOracle, 4, 4)
	whelp := effectstest.B12Push(g, me.ID, "Hellkite Whelp", "Creature — Dragon", "", 2, 2)
	relic := effectstest.PushPermanentForTest(g, them.ID, "Their Relic", "", "Artifact")
	effectstest.AdvanceToMain(t, g)
	spell := game.NewCard("Test Pump", me.ID)
	spell.TypeLine = "Instant"
	me.Hand.PushTop(spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: whelp}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoicePickTarget) != nil || !effectstest.OnBoard(g, relic) {
		t.Error("Scalelord Reckoner triggered off its controller's own spell")
	}
}

func TestTranscendentMasterLevelsUp(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	master := effectstest.B12Push(g, me.ID, "Transcendent Master", "Creature — Human Cleric Avatar", transcendentMasterOracle, 3, 3)
	have := 0
	setLevel := func(n int) {
		t.Helper()
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(master, game.CounterLevel, n-have); err != nil {
				t.Fatalf("level counters: %v", err)
			}
		})
		have = n
	}
	want := func(level, pt int, lifelink, indestructible bool) {
		t.Helper()
		setLevel(level)
		if p, tg := effectstest.EffectivePower(t, g, master), effectstest.EffectiveToughness(t, g, master); p != pt || tg != pt {
			t.Errorf("level %d: %d/%d, want %d/%d", level, p, tg, pt, pt)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, master, "lifelink"); got != lifelink {
			t.Errorf("level %d: lifelink %v, want %v", level, got, lifelink)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, master, "indestructible"); got != indestructible {
			t.Errorf("level %d: indestructible %v, want %v", level, got, indestructible)
		}
	}
	want(0, 3, false, false)
	want(5, 3, false, false)
	want(6, 6, true, false)
	want(11, 6, true, false)
	want(12, 9, true, true)
	want(20, 9, true, true)

	// Level up is a sorcery-speed {1} that adds one level counter.
	setLevel(5)
	effectstest.ApaMana(me, "C")
	effectstest.P7Activate(t, g, me, master, 0, game.ActivateAbilityParams{Strict: true})
	want(6, 6, true, false)
}

func TestSoulOfTherosPumpsYourCreaturesFromTheBattlefieldAndTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	soul := effectstest.PushCatalogPermanent(g, me.ID, "Soul of Theros", "Creature — Avatar", soulOfTherosOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	foe := effectstest.PushVanillaCreature(g, them.ID, "Foe", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{W}{W}{W}{W}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, soul, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("my Bear has power %d, want 4", p)
	}
	for _, kw := range []string{"first strike", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("my Bear lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, foe, kw) {
			t.Errorf("their creature has %s", kw)
		}
	}
	if p := effectstest.EffectivePower(t, g, foe); p != 2 {
		t.Errorf("their creature has power %d, want 2", p)
	}
	// A creature that arrives afterwards gets nothing (CR 611.2c).
	late := effectstest.PushVanillaCreature(g, me.ID, "Late", 1, 1)
	if p := effectstest.EffectivePower(t, g, late); p != 1 {
		t.Errorf("a creature that entered later has power %d, want 1", p)
	}
}

func TestSoulOfTherosFromTheGraveyardExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	soul := effectstest.PushCatalogGraveyardCard(me, "Soul of Theros", "Creature — Avatar", soulOfTherosOracle, 6, 6)
	effectstest.DcMana(t, g, me.ID, "{W}{W}{W}{W}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, soul, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("Bear has power %d, want 4", p)
	}
	if !g.Exile.Contains(soul) {
		t.Error("Soul of Theros was not exiled as a cost")
	}
	// The battlefield row is not available from the graveyard.
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: soulOfTherosOracle})
	if len(abilities) != 2 {
		t.Fatalf("activated abilities = %d, want 2", len(abilities))
	}
}

func TestHalvingSeasonHalvesOnlyOpponentsTokensAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Halving Season", "Enchantment", halvingSeasonOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	soldiers := func(owner uuid.UUID) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Soldier" && c.Controller == owner {
				n++
			}
		}
		return n
	}
	mk := func(who uuid.UUID, n int) {
		g.WithWriteLock(func() {
			if err := (effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: n}).Apply(effects.NewContext(g, &game.StackItem{Controller: who})); err != nil {
				t.Fatalf("token: %v", err)
			}
		})
	}
	mk(them.ID, 5)
	if got := soldiers(them.ID); got != 2 {
		t.Errorf("an opponent making 5 tokens got %d, want 2 (half, rounded down)", got)
	}
	mk(them.ID, 1)
	if got := soldiers(them.ID); got != 2 {
		t.Errorf("an opponent making 1 token got %d more, want none", got-2)
	}
	mk(me.ID, 3)
	if got := soldiers(me.ID); got != 3 {
		t.Errorf("my own 3 tokens became %d, want 3", got)
	}

	put := func(who uuid.UUID, n int) {
		g.WithWriteLock(func() {
			if err := g.AddCounterByForEffect(who, bear, game.CounterPlusOne, n); err != nil {
				t.Fatalf("counter: %v", err)
			}
		})
	}
	put(them.ID, 5)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("an opponent putting 5 counters put %d, want 2", got)
	}
	put(me.ID, 3)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 5 {
		t.Errorf("after my own 3 counters: %d, want 5", got)
	}
}

func TestSliverOfHopePreventsDamageToAttackingSlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	hope := effectstest.PushCatalogPermanent(g, me.ID, "Sliver of Hope", "Creature — Sliver", sliverOfHopeOracle, false)
	other := effectstest.B12Push(g, me.ID, "Plated Sliver", "Creature — Sliver", "", 3, 3)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 3, 3)
	foe := effectstest.PushVanillaCreature(g, them.ID, "Foe", 2, 2)
	for _, id := range []uuid.UUID{hope, other} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "hope") {
			t.Fatalf("a Sliver you control lacks hope")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "hope") || effectstest.HasEffectiveKeyword(t, g, foe, "hope") {
		t.Error("a non-Sliver has hope")
	}
	dmg := func(target uuid.UUID, n int) {
		g.WithWriteLock(func() {
			if err := g.DealDamageToCreatureForEffect(foe, target, n); err != nil {
				t.Fatalf("damage: %v", err)
			}
		})
	}
	marked := func(id uuid.UUID) int { return effectstest.E2Card(t, g, id).DamageMarked }

	// Not attacking: damaged normally.
	dmg(other, 1)
	if marked(other) != 1 {
		t.Errorf("a non-attacking Sliver took %d damage, want 1", marked(other))
	}
	effectstest.DeclareAttack(t, g, them.ID, hope, other, bear)
	dmg(hope, 1)
	dmg(other, 1)
	dmg(bear, 1)
	if marked(hope) != 0 {
		t.Errorf("an attacking Sliver of Hope took %d damage, want 0", marked(hope))
	}
	if marked(other) != 1 {
		t.Errorf("an attacking Sliver took more damage (%d), want it still at 1", marked(other))
	}
	if marked(bear) != 1 {
		t.Errorf("an attacking Bear took %d damage, want 1", marked(bear))
	}
}

func TestManabondPutsEveryLandThenDiscardsTheHand(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Manabond", "Enchantment", manabondOracle, false)
		var lands []uuid.UUID
		var spell uuid.UUID
		for _, n := range []string{"Forest", "Island"} {
			c := game.NewCard(n, me.ID)
			c.TypeLine = "Basic Land — " + n
			me.Hand.PushTop(c)
			lands = append(lands, c.InstanceID)
		}
		sp := game.NewCard("Giant Growth", me.ID)
		sp.TypeLine = "Instant"
		me.Hand.PushTop(sp)
		spell = sp.InstanceID
		effectstest.WalkToEndStepOfSeat0(t, g)
		before := len(me.Hand.Cards)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if yes {
			for _, l := range lands {
				if !g.Battlefield.Contains(l) {
					t.Errorf("a land card was not put onto the battlefield")
				}
			}
			if len(me.Hand.Cards) != 0 {
				t.Errorf("hand has %d cards, want 0 after discarding it", len(me.Hand.Cards))
			}
			if !me.Graveyard.Contains(spell) {
				t.Error("the nonland card was not discarded")
			}
		} else {
			if len(me.Hand.Cards) != before || g.Battlefield.Contains(lands[0]) {
				t.Errorf("declined Manabond changed the hand: %d cards", len(me.Hand.Cards))
			}
		}
	}
}
