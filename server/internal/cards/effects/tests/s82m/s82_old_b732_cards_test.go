package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b732_cards_test.go — Urza's Saga, batch 732 (#5624). Each test plays
// the real catalog entry and asserts what its text says.

const (
	b732AttendantOracle  = "4fa30e29-94c1-4537-bebd-dc7124e4ef71"
	b732EdictOracle      = "544eca0d-63cd-4cc5-9318-c2006cdec170"
	b732HealerOracle     = "c657e110-c204-48df-a396-9822ac48798d"
	b732SporoOracle      = "33005c63-fe9f-4051-b781-eacf1ca5eec6"
	b732PasturesOracle   = "f9e1346a-af72-4644-b384-8f7b0a35ec31"
	b732PowerSinkOracle  = "39412e6d-2837-4729-abf9-e64a5ba87e40"
	b732ProctorOracle    = "5120bfdc-7201-420d-85a1-eeb4aba6632b"
	b732WitchOracle      = "48adccf6-adb4-49b2-b013-4dc5e0fc7bcb"
	b732TurbineOracle    = "6c31f06d-bf02-48a7-9c17-be2a226b8eae"
	b732WirecatOracle    = "57828973-df3d-4288-988d-147a3016120e"
	b732ScalesOracle     = "aebddadd-f05b-4ad5-bda4-603a134f68dc"
	b732CreatureTypeLine = "Creature — Test"
)

func TestOldB732EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b732AttendantOracle: "Silent Attendant", b732EdictOracle: "Yawgmoth's Edict", b732HealerOracle: "Faith Healer",
		b732SporoOracle: "Sporogenesis", b732PasturesOracle: "Greener Pastures", b732PowerSinkOracle: "Power Sink",
		b732ProctorOracle: "Stern Proctor", b732WitchOracle: "Witch Engine", b732TurbineOracle: "Thran Turbine",
		b732WirecatOracle: "Wirecat", b732ScalesOracle: "Noetic Scales",
	}
	for oracle, name := range want {
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

func b732OnBattlefield(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

func b732CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

func TestOldB732SilentAttendantGainsOneLifeByTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	att := effectstest.PushCatalogPermanent(g, me.ID, "Silent Attendant", "Creature — Human Cleric", b732AttendantOracle, false)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.ApaActivate(t, g, me, att, 0, game.ActivateAbilityParams{})
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	if !effectstest.TappedForTest(t, g, att) {
		t.Error("the Attendant did not tap")
	}
}

// b732Cast puts a spell with the given colours in `p`'s hand and casts it where
// the table stands.
func b732Cast(t *testing.T, g *game.Game, p *game.Player, name string, colors []string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Colors: colors, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func TestOldB732YawgmothsEdictDrainsAnOpponentWhoCastsWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Yawgmoth's Edict", "Enchantment", b732EdictOracle, false)
	effectstest.AdvanceToMain(t, g)
	myLife, theirLife := me.Life, opp.Life
	b732Cast(t, g, opp, "White Spell", []string{"W"})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("life: opponent %d (want %d), you %d (want %d)", opp.Life, theirLife-1, me.Life, myLife+1)
	}
	b732Cast(t, g, opp, "Blue Spell", []string{"U"})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-1 || me.Life != myLife+1 {
		t.Error("a nonwhite spell drained")
	}
	b732Cast(t, g, me, "My White Spell", []string{"W"})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 {
		t.Error("your own white spell drained")
	}
}

func TestOldB732FaithHealerGainsLifeEqualToTheSacrificedEnchantmentsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	healer := effectstest.PushCatalogPermanent(g, me.ID, "Faith Healer", "Creature — Human Cleric", b732HealerOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no enchantment to sacrifice")
	}
	ench := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Aura", TypeLine: "Enchantment — Aura", ManaCost: "{2}{W}"})
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{ench}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if b732OnBattlefield(g, ench) {
		t.Error("the enchantment was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (the enchantment's mana value is 3)", me.Life, life+3)
	}
}

func TestOldB732SternProctorReturnsATargetArtifactOrEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.SeedPermanent(g, opp.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b732CreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Stern Proctor", "Creature — Human Wizard", b732ProctorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	_ = bear
	effectstest.PickTriggerTarget(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if b732OnBattlefield(g, rock) {
		t.Error("the artifact stayed on the battlefield")
	}
	if countHandNamedB732(opp, "Rock") != 1 {
		t.Error("the artifact is not in its owner's hand")
	}
}

func countHandNamedB732(p *game.Player, name string) int {
	n := 0
	for _, c := range p.Hand.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

func TestOldB732WirecatCantAttackOrBlockWhileAnEnchantmentIsOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	free := effectstest.B12Push(g, me.ID, "Wirecat", "Artifact Creature — Cat", b732WirecatOracle, 4, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(free, opp.ID); err != nil {
		t.Fatalf("a Wirecat could not attack with no enchantment on the battlefield: %v", err)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	cat2 := effectstest.B12Push(g2, me2.ID, "Wirecat", "Artifact Creature — Cat", b732WirecatOracle, 4, 3)
	ench := effectstest.SeedPermanent(g2, opp2.ID, "Their Enchantment", "Enchantment")
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(cat2, opp2.ID); err == nil {
		t.Error("a Wirecat attacked while an enchantment was on the battlefield")
	}
	// Block: the opponent's Wirecat can't block while an enchantment is on the table.
	g3 := effectstest.NewCatalogGame(t)
	me3, opp3 := g3.Seats[0], g3.Seats[1]
	atk := effectstest.ReqCreature(g3, me3.ID, "Attacker", "", 2, 2)
	blocker := effectstest.B12Push(g3, opp3.ID, "Wirecat", "Artifact Creature — Cat", b732WirecatOracle, 4, 3)
	effectstest.SeedPermanent(g3, me3.ID, "Enchantment", "Enchantment")
	effectstest.BrAttack(t, g3, atk)
	if err := effectstest.ReqBlock(t, g3, atk, blocker); err == nil {
		t.Error("a Wirecat blocked while an enchantment was on the battlefield")
	}
	// Removing the last enchantment lifts the restriction.
	effectstest.B14Kill(g2, ench)
	effectstest.PassPriorityAroundTable(t, g2)
	if err := g2.DeclareAttacker(cat2, opp2.ID); err != nil {
		t.Errorf("the Wirecat still can't attack with no enchantment on the battlefield: %v", err)
	}
}

func TestOldB732ThranTurbineAddsRestrictedColorlessAtUpkeep(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Thran Turbine", "Artifact", b732TurbineOracle, false)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityUntilChoice(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		if !yes {
			if me.ManaPool.Total() != 0 {
				t.Errorf("declined, but the pool holds %d", me.ManaPool.Total())
			}
			continue
		}
		if got := me.ManaPool.Total(); got != 2 {
			t.Fatalf("the pool holds %d mana, want {C}{C}", got)
		}
		for _, tok := range me.ManaPool {
			if tok.Color != "C" || len(tok.Restrictions) != 1 || tok.Restrictions[0] != game.ManaRestrictNotCast {
				t.Errorf("token %+v: want colorless carrying the can't-cast restriction", tok)
			}
		}
		cost, _ := game.ParseCost("{C}")
		if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendContext{Purpose: game.SpendPurposeCast, Types: []string{"Artifact"}}) {
			t.Error("the Turbine's mana paid for a spell")
		}
		if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendContext{Purpose: game.SpendPurposeActivate}) {
			t.Error("the Turbine's mana could not pay for an ability")
		}
	}
}

func TestOldB732NoeticScalesReturnsCreaturesWithPowerAboveThePlayersHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Noetic Scales", "Artifact", b732ScalesOracle, false)
	big := effectstest.B12Creature(g, me.ID, "Big", b732CreatureTypeLine, 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", b732CreatureTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b732CreatureTypeLine, 6, 6)
	me.Hand.Cards = nil
	for i := 0; i < 2; i++ {
		effectstest.HandCardForTest(me, "Card", "Basic Land — Forest", "")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if b732OnBattlefield(g, big) {
		t.Error("a 3-power creature stayed with two cards in hand")
	}
	if !b732OnBattlefield(g, small) {
		t.Error("a 2-power creature returned with two cards in hand")
	}
	if !b732OnBattlefield(g, theirs) {
		t.Error("an opponent's creature returned at your upkeep")
	}
	if countHandNamedB732(me, "Big") != 1 {
		t.Error("the creature is not in its owner's hand")
	}
}

func TestOldB732NoeticScalesAlsoWorksOnAnOpponentsUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Noetic Scales", "Artifact", b732ScalesOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", b732CreatureTypeLine, 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b732CreatureTypeLine, 1, 1)
	opp.Hand.Cards = nil
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if b732OnBattlefield(g, theirs) {
		t.Error("their 1-power creature stayed with no cards in hand")
	}
	if !b732OnBattlefield(g, mine) {
		t.Error("your creature returned at their upkeep")
	}
}

func TestOldB732GreenerPasturesMakesASaprolingForTheLandLeader(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		upkeepSeat   int
		wantFor      int // seat that gets the token, -1 for none
	}{
		{"you lead at your upkeep", 2, 1, 0, 0},
		{"tied at your upkeep", 2, 2, 0, -1},
		{"they lead at their upkeep", 1, 3, 1, 1},
		{"you lead at their upkeep", 3, 1, 1, -1},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Greener Pastures", "Enchantment", b732PasturesOracle, false)
		for i := 0; i < tc.mine; i++ {
			effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		}
		for i := 0; i < tc.theirs; i++ {
			effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
		}
		effectstest.AdvanceToUpkeepOfSeat(t, g, tc.upkeepSeat)
		effectstest.PassPriorityAroundTable(t, g)
		mineN, theirN := b732CountNamed(g, me.ID, "Saproling"), b732CountNamed(g, opp.ID, "Saproling")
		wantMine, wantTheirs := 0, 0
		switch tc.wantFor {
		case 0:
			wantMine = 1
		case 1:
			wantTheirs = 1
		}
		if mineN != wantMine || theirN != wantTheirs {
			t.Errorf("%s: Saprolings you %d (want %d), they %d (want %d)", tc.name, mineN, wantMine, theirN, wantTheirs)
		}
	}
}

func TestOldB732SporogenesisPutsAFungusCounterOnATargetNontokenCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Sporogenesis", "Enchantment", b732SporoOracle, false)
		bear := effectstest.B12Creature(g, me.ID, "Bear", b732CreatureTypeLine, 2, 2)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityUntilChoice(t, g)
		if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
			effectstest.AnswerFirstPickTarget(t, g)
			effectstest.PassPriorityUntilChoice(t, g)
		}
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if yes {
			want = 1
		}
		if got := effectstest.CountersOn(g, bear, "fungus"); got != want {
			t.Errorf("yes=%v: %d fungus counters, want %d", yes, got, want)
		}
	}
}

func TestOldB732SporogenesisMakesASaprolingPerFungusCounterWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sporogenesis", "Enchantment", b732SporoOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b732CreatureTypeLine, 2, 2)
	plain := effectstest.B12Creature(g, opp.ID, "Plain", b732CreatureTypeLine, 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, "fungus", 2) })
	effectstest.B14Kill(g, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b732CountNamed(g, me.ID, "Saproling"); n != 0 {
		t.Fatalf("a creature without a fungus counter made %d Saprolings", n)
	}
	effectstest.B14Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b732CountNamed(g, me.ID, "Saproling"); n != 2 {
		t.Errorf("%d Saprolings under your control, want 2 (one per counter)", n)
	}
	if n := b732CountNamed(g, opp.ID, "Saproling"); n != 0 {
		t.Errorf("the dead creature's controller got %d Saprolings", n)
	}
}

func TestOldB732SporogenesisLeavingRemovesEveryFungusCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spore := effectstest.PushCatalogPermanent(g, me.ID, "Sporogenesis", "Enchantment", b732SporoOracle, false)
	a := effectstest.B12Creature(g, me.ID, "A", b732CreatureTypeLine, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", b732CreatureTypeLine, 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(a, "fungus", 1)
		_ = g.AddCounterForEffect(b, "fungus", 3)
	})
	effectstest.B14Kill(g, spore)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, a, "fungus") != 0 || effectstest.CountersOn(g, b, "fungus") != 0 {
		t.Error("a fungus counter survived Sporogenesis leaving the battlefield")
	}
	if n := b732CountNamed(g, me.ID, "Saproling") + b732CountNamed(g, opp.ID, "Saproling"); n != 0 {
		t.Errorf("%d Saprolings made though nothing died", n)
	}
}

func TestOldB732WitchEngineAddsBlackManaThenHandsItToAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.PushCatalogPermanent(g, me.ID, "Witch Engine", "Creature — Horror", b732WitchOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("you were accepted as the target opponent")
	}
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the mana arrived before the ability resolved; it is not a mana ability")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 4 {
		t.Errorf("the pool holds %d, want {B}{B}{B}{B}", got)
	}
	if effectstest.ControllerOf(t, g, engine) != opp.ID {
		t.Error("the chosen opponent did not gain control of the Engine")
	}
}

// b732PowerSinkScene has the opponent cast a spell, then the caster answer it
// with Power Sink for X = 2, and returns the opponent's spell.
func b732PowerSinkScene(t *testing.T, g *game.Game, me, opp *game.Player, oppMana ...string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(opp, oppMana...)
	theirs := effectstest.CastFromHandForTest(t, g, opp, "Their Spell", "Instant", "", "", game.CastSpellParams{})
	effectstest.ApaMana(me, "U", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Power Sink", "Instant", "{X}{U}", b732PowerSinkOracle, game.CastSpellParams{
		Targets: effectstest.CardTarget(theirs), XValue: 2,
	})
	effectstest.PassPriorityAroundTable(t, g)
	return theirs
}

func b732PayPrompt(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless {
			return c
		}
	}
	return nil
}

func TestOldB732PowerSinkDeclinedCountersTapsTheirLandsAndEmptiesTheirPool(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	idle := effectstest.B12Permanent(g, opp.ID, "Oddity", "Land")
	mine := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	theirs := b732PowerSinkScene(t, g, me, opp, "G")
	pay := b732PayPrompt(g)
	if pay == nil {
		t.Fatal("the spell's controller was not asked to pay")
	}
	if pay.Chooser != opp.ID {
		t.Errorf("asked %s, want the spell's controller %s", pay.Chooser, opp.ID)
	}
	if err := g.ResolvePayUnless(pay.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P7Countered(g, theirs) {
		t.Error("the spell was not countered")
	}
	if !effectstest.TappedForTest(t, g, forest) {
		t.Error("their land with a mana ability was not tapped")
	}
	if effectstest.TappedForTest(t, g, idle) {
		t.Error("a land with no mana ability was tapped")
	}
	if effectstest.TappedForTest(t, g, mine) {
		t.Error("the caster's land was tapped")
	}
	if opp.ManaPool.Total() != 0 {
		t.Errorf("their pool holds %d, want it emptied", opp.ManaPool.Total())
	}
}

func TestOldB732PowerSinkPaidLetsTheSpellResolveAndTapsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := b732PowerSinkScene(t, g, me, opp, "C", "C")
	pay := b732PayPrompt(g)
	if pay == nil {
		t.Fatal("the spell's controller was not asked to pay")
	}
	if err := g.ResolvePayUnless(pay.ID, opp.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.P7Countered(g, theirs) {
		t.Error("the spell was countered though {2} was paid")
	}
	if effectstest.TappedForTest(t, g, forest) {
		t.Error("their land was tapped though they paid")
	}
}
