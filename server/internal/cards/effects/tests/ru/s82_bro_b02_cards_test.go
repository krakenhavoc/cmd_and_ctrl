package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	broB02Gift      = "2cb95721-56c8-4008-8055-579a5a97fc8b"
	broB02Machine   = "37069c66-c17a-40ac-80f4-14ef49d18008"
	broB02Defiance  = "257f81fb-b3c9-4ebe-9889-dd0ef960bba3"
	broB02Ray       = "b515dfc6-1a1e-46d6-a7a3-0e3af8279f49"
	broB02Union     = "4492c329-174a-44aa-be4d-9f018173b9a7"
	broB02Shell     = "9eaca93d-6c3d-4489-8652-b575a4cd931f"
	broB02Rebuff    = "db3b294a-87b1-4dda-b1f6-ea63a9722faf"
	broB02Whirling  = "6b37b0ef-7e89-447f-a189-b3ea7b42143f"
	broB02Raze      = "efe24eed-e1e2-4807-ae61-2e5be529ae3c"
	broB02ShootDown = "2a89feeb-9e66-4539-bd75-7d9d53f52d49"
	broB02LayDown   = "aa91d092-7f4a-4e4d-85db-e200692f3d9d"
	broB02Gruesome  = "c24e34e7-0ac6-451a-b73b-31c7f96d123f"
	broB02Desync    = "dae81b1b-28dd-4f7f-9bcb-65e72eaa37e2"
)

func TestBroB02EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB02Gift: "Gaea's Gift", broB02Machine: "Machine Over Matter", broB02Defiance: "Moment of Defiance",
		broB02Ray: "Scatter Ray", broB02Union: "Union of the Third Path", broB02Shell: "Unleash Shell",
		broB02Rebuff: "Urza's Rebuff", broB02Whirling: "Whirling Strike", broB02Raze: "Raze to the Ground",
		broB02ShootDown: "Shoot Down", broB02LayDown: "Lay Down Arms", broB02Gruesome: "Gruesome Realization",
		broB02Desync: "Desynchronize",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBroB02WhirlingStrikePumpsAndGrantsKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Whirling Strike", "Instant", broB02Whirling, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 2 {
		t.Errorf("Bear is %d/%d, want 4/2", p, tg)
	}
	for _, kw := range []string{"first strike", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("Bear lacks %s", kw)
		}
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Whirling Strike", "Instant", broB02Whirling,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Error("targeted a player")
	}
}

func TestBroB02MomentOfDefianceDrawsAndGrantsLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Moment of Defiance", "Instant", broB02Defiance, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 3 {
		t.Errorf("Bear is %d/%d, want 4/3", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("Bear lacks lifelink")
	}
	// The harness pushes the spell into the hand (+1), the cast takes it out (-1) and one card is drawn (+1).
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want %d", hand, got, hand+1)
	}
}

func TestBroB02GaeasGiftCountersAndProtectsOnlyYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Gaea's Gift", "Instant", broB02Gift, effectstest.B16TargetCard(theirs)); err == nil {
		t.Error("targeted an opponent's creature")
	}
	effectstest.CastCatalogSpell(t, g, "Gaea's Gift", "Instant", broB02Gift, effectstest.B16TargetCard(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.E2Card(t, g, mine).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	if p := effectstest.EffectivePowerWithCounters(t, g, mine); p != 3 {
		t.Errorf("Bear power is %d, want 3", p)
	}
	for _, kw := range []string{"reach", "trample", "hexproof", "indestructible"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("Bear lacks %s", kw)
		}
	}
}

func broB02CastFromOpponentAt(t *testing.T, g *game.Game, opp *game.Player, name, oracle string, target uuid.UUID) {
	t.Helper()
	effectstest.CastFromSeat(t, g, opp, name, "Instant", oracle, effectstest.B16TargetCard(target))
}

func TestBroB02ScatterRayCountersAnArtifactOrCreatureSpellUnlessFourArePaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		effectstest.FillPoolColored(me, "C", 4)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		broB02CastFromOpponentAt(t, g, opp, "Scatter Ray", broB02Ray, victim)
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if pay && me.Graveyard.Contains(victim) {
			t.Error("the spell was countered although {4} was paid")
		}
		if !pay && !me.Graveyard.Contains(victim) {
			t.Error("the spell was not countered")
		}
	}
}

func TestBroB02ScatterRayRefusesAnInstantSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Trick", "Instant", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	opp := g.Seats[1]
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Scatter Ray", TypeLine: "Instant", OracleID: broB02Ray, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(victim), Strict: true}); err == nil {
		t.Error("Scatter Ray targeted an instant that is neither artifact nor creature")
	}
}

func TestBroB02ShootDownExilesArtifactEnchantmentOrFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	art := effectstest.B12Permanent(g, opp.ID, "Trinket", "Artifact")
	ench := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"}})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Shoot Down", "Sorcery", broB02ShootDown, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("targeted a creature without flying")
	}
	for _, id := range []uuid.UUID{art, ench, flyer} {
		effectstest.CastCatalogSpell(t, g, "Shoot Down", "Sorcery", broB02ShootDown, effectstest.B16TargetCard(id))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(id) || !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
}

func TestBroB02UnionOfTheThirdPathGainsLifeEqualToHandAfterTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Union of the Third Path", "Instant", broB02Union, nil)
	effectstest.PassPriorityAroundTable(t, g)
	want := hand + 1 // the harness pushed the spell in, the cast took it out, one card was drawn
	if me.Hand.Size() != want {
		t.Fatalf("hand %d → %d, want %d", hand, me.Hand.Size(), want)
	}
	if me.Life != before+want {
		t.Errorf("life %d → %d, want +%d", before, me.Life, want)
	}
}

func TestBroB02UnleashShellDamagesThePermanentAndItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Unleash Shell", "Instant", broB02Shell, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("controller life %d → %d, want -2", before, opp.Life)
	}
	if !g.Battlefield.Contains(big) {
		t.Error("a 0/6 died to 5 damage")
	}
	small := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unleash Shell", "Instant", broB02Shell, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived 5 damage")
	}
	if opp.Life != before-4 {
		t.Errorf("controller life %d, want %d even after the creature died", opp.Life, before-4)
	}
}

func TestBroB02UrzasRebuffCountersOrTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Urza's Rebuff", "Instant", broB02Rebuff, []int{1}, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.E2Card(t, g, id).Tapped {
			t.Errorf("%s was not tapped", id)
		}
	}
	if effectstest.E2Card(t, g, c).Tapped {
		t.Error("an untargeted creature was tapped")
	}
}

func TestBroB02MachineOverMatterCostsOneLessWithAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, broB02Machine, "Instant", "{1}{U}"); got != 2 {
		t.Errorf("no artifact creature: %d, want 2", got)
	}
	effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Plain Rock", "Artifact")
	if got := effectstest.SelfPricedMV(t, g, me, broB02Machine, "Instant", "{1}{U}"); got != 2 {
		t.Errorf("no artifact creature among a creature and an artifact: %d, want 2", got)
	}
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Golem", "Artifact Creature — Golem", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, broB02Machine, "Instant", "{1}{U}"); got != 2 {
		t.Errorf("an opponent's artifact creature: %d, want 2", got)
	}
	effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, broB02Machine, "Instant", "{1}{U}"); got != 1 {
		t.Errorf("your artifact creature: %d, want 1", got)
	}
}

func TestBroB02MachineOverMatterBouncesANonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Machine Over Matter", "Instant", broB02Machine, effectstest.B16TargetCard(land)); err == nil {
		t.Error("targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Machine Over Matter", "Instant", broB02Machine, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the Bear was not returned to its owner's hand")
	}
}

func TestBroB02RazeToTheGroundDrawsOnlyForAnArtifactOfManaValueOneOrLess(t *testing.T) {
	for _, tc := range []struct {
		cost string
		draw bool
	}{{"{1}", true}, {"{2}", false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		art := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact",
			ManaCost: tc.cost, Owner: opp.ID, Controller: opp.ID})
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Raze to the Ground", "Sorcery", broB02Raze, effectstest.B16TargetCard(art))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(art) {
			t.Fatalf("%s: the artifact survived", tc.cost)
		}
		want := hand
		if tc.draw {
			want = hand + 1
		}
		if me.Hand.Size() != want {
			t.Errorf("%s: hand %d → %d, want %d", tc.cost, hand, me.Hand.Size(), want)
		}
	}
	spec, _ := effects.Lookup(broB02Raze)
	if !spec.CantBeCountered {
		t.Error("Raze to the Ground can be countered")
	}
}

func TestBroB02LayDownArmsNeedsEnoughPlainsAndGivesLifeToTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	three := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Three", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lay Down Arms", "Sorcery", broB02LayDown, effectstest.B16TargetCard(three)); err == nil {
		t.Error("targeted a mana value 3 creature with no Plains")
	}
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains A", "Basic Land — Plains")
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains B", "Basic Land — Plains")
	effectstest.SeedLandOnBattlefield(g, opp.ID, "Their Plains", "Basic Land — Plains")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Lay Down Arms", "Sorcery", broB02LayDown, effectstest.B16TargetCard(three)); err == nil {
		t.Error("targeted a mana value 3 creature with two Plains")
	}
	effectstest.SeedLandOnBattlefield(g, me.ID, "Plains C", "Basic Land — Plains")
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lay Down Arms", "Sorcery", broB02LayDown, effectstest.B16TargetCard(three))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) || !g.Exile.Contains(three) {
		t.Error("the creature was not exiled")
	}
	if opp.Life != before+3 {
		t.Errorf("its controller's life %d → %d, want +3", before, opp.Life)
	}
}

func TestBroB02GruesomeRealizationDrawsAndLosesLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	before, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpellWithModes(t, g, "Gruesome Realization", "Sorcery", broB02Gruesome, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-2 {
		t.Errorf("life %d → %d, want -2", before, me.Life)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2 (the harness pushes the spell in, the cast takes it out)", hand, me.Hand.Size())
	}
}

func TestBroB02GruesomeRealizationShrinksOnlyOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	x := effectstest.B12Creature(g, opp.ID, "Fragile", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpellWithModes(t, g, "Gruesome Realization", "Sorcery", broB02Gruesome, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 2 {
		t.Errorf("my creature is %d power, want 2", p)
	}
	if p, tg := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 1 || tg != 1 {
		t.Errorf("their creature is %d/%d, want 1/1", p, tg)
	}
	if g.Battlefield.Contains(x) {
		t.Error("their 1/1 survived -1/-1")
	}
}

func TestBroB02DesynchronizeLetsTheOwnerChooseThenTheCasterScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Desynchronize", "Instant", broB02Desync, effectstest.B16TargetCard(land)); err == nil {
		t.Error("targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Desynchronize", "Instant", broB02Desync, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID) != nil {
		t.Fatal("the caster was asked; the owner chooses")
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if c == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, []uuid.UUID{bear}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(bear) || opp.Library.Cards[0].InstanceID != bear {
		t.Error("the Bear should be on the bottom of its owner's library (the helper's bottom lane)")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Errorf("the caster was not asked to scry 2: %+v", g.PendingChoices)
	}
}

func TestBroB02UrzasRebuffCountersASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Urza's Rebuff", TypeLine: "Instant", OracleID: broB02Rebuff, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) {
		t.Error("the spell was not countered")
	}
}
