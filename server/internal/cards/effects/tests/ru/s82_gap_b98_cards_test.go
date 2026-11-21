package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b98_cards_test.go — Commander Masters artifacts and artifact
// creatures, GAP batch 98 (#5948). Each test plays the real catalog entry
// through the engine.

const (
	gapAbstruseArchaicOracle   = "0727dc95-03df-4a2c-aed5-73fb840f3db7"
	gapBaneOfBalaGedOracle     = "38839894-1706-4e11-8310-5ea8dd8866d9"
	gapDwarvenHammerOracle     = "a6365f98-f652-492b-81de-5f8cd6fe0a1c"
	gapMatterReshaperOracle    = "38538aff-b6f9-4c30-992e-7fd78da3c44a"
	gapMyrSireOracle           = "25794033-bd17-4056-b116-68d3acf52569"
	gapMyriadConstructOracle   = "f54111c8-9c20-4a73-a7c4-24d216e7de07"
	gapTransmogrifyingWandOrcl = "921345ce-2eaf-4b95-b65d-1c27204665ed"
	gapVialOfDragonfireOracle  = "17be934e-1e33-4b95-a984-242fa018e280"
)

func TestGapB98EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapAbstruseArchaicOracle:   "Abstruse Archaic",
		gapBaneOfBalaGedOracle:     "Bane of Bala Ged",
		gapDwarvenHammerOracle:     "Dwarven Hammer",
		gapMatterReshaperOracle:    "Matter Reshaper",
		gapMyrSireOracle:           "Myr Sire",
		gapMyriadConstructOracle:   "Myriad Construct",
		gapTransmogrifyingWandOrcl: "Transmogrifying Wand",
		gapVialOfDragonfireOracle:  "Vial of Dragonfire",
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

// --- Vial of Dragonfire ---

func TestGapB98VialOfDragonfireSacrificesItselfForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	vial := effectstest.PushCatalogPermanent(g, me.ID, "Vial of Dragonfire", "Artifact", gapVialOfDragonfireOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vial) {
		t.Error("the Vial was not sacrificed")
	}
	if c := effectstest.FindBattlefieldCardForTest(g, bear); c == nil || c.DamageMarked != 2 {
		t.Errorf("the target: %+v, want 2 damage", c)
	}
}

// --- Myr Sire ---

func TestGapB98MyrSireLeavesAPhyrexianMyrWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sire := effectstest.PushCatalogPermanent(g, me.ID, "Myr Sire", "Artifact Creature — Phyrexian Myr", gapMyrSireOracle, false)
	_ = g.DestroyPermanentForEffect(sire)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Phyrexian Myr" {
			found = true
			if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || !c.IsColorless() || c.Power != 1 || c.Toughness != 1 {
				t.Errorf("the token is %+v, want a 1/1 colorless artifact creature token", c)
			}
		}
	}
	if !found {
		t.Error("no Phyrexian Myr token")
	}
}

// --- Transmogrifying Wand ---

func TestGapB98TransmogrifyingWandDestroysAndGivesAnOxAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wand := effectstest.CastCatalogSpell(t, g, "Transmogrifying Wand", "Artifact", gapTransmogrifyingWandOrcl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, wand, game.CounterCharge); n != 3 {
		t.Fatalf("charge counters %d, want 3", n)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, wand, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), Strict: true}); err == nil {
		t.Fatal("the Wand was activated outside a main phase")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, wand, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	if n := effectstest.CountersOn(g, wand, game.CounterCharge); n != 2 {
		t.Errorf("charge counters %d, want 2", n)
	}
	ox := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ox" {
			ox++
			if c.Controller != opp.ID || c.Power != 2 || c.Toughness != 4 || len(c.Colors) != 1 || c.Colors[0] != "W" {
				t.Errorf("the Ox is %+v, want a white 2/4 under the destroyed creature's controller", c)
			}
		}
	}
	if ox != 1 {
		t.Errorf("%d Ox tokens, want 1", ox)
	}
}

// --- Dwarven Hammer ---

func TestGapB98DwarvenHammerMayPayTwoForADwarfToWearIt(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		hammer := effectstest.SeedEquipment(g, me.ID, "Dwarven Hammer", gapDwarvenHammerOracle)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: hammer})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("no {2} offer")
		}
		if pay {
			effectstest.FloatMana(t, g, me, "{C}{C}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		host := effectstest.AttachmentHostOf(t, g, hammer)
		if !pay {
			if host.Kind == game.TargetCard {
				t.Errorf("declined, but the Hammer is attached to %s", host.ID)
			}
			continue
		}
		if host.Kind != game.TargetCard {
			t.Fatal("paid, but the Hammer is attached to nothing")
		}
		if p, tough := effectstest.EffectivePTOf(t, g, host.ID); p != 5 || tough != 1 {
			t.Errorf("the Dwarf is %d/%d, want 5/1", p, tough)
		}
		effectstest.AssertKeywords(t, g, host.ID, "trample")
		if c := effectstest.FindBattlefieldCardForTest(g, host.ID); c == nil || c.Name != "Dwarf Berserker" {
			t.Errorf("the token is %+v, want a Dwarf Berserker", c)
		}
	}
}

// --- Bane of Bala Ged ---

func TestGapB98BaneOfBalaGedMakesTheDefenderExileTwoOfTheirPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	bane := effectstest.PushCatalogPermanent(g, me.ID, "Bane of Bala Ged", "Creature — Eldrazi", gapBaneOfBalaGedOracle, false)
	a := effectstest.B12Creature(g, opp.ID, "Their A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "Their C", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, third.ID, "Third's", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bane)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the two chosen permanents were not exiled")
	}
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the chosen permanents are not in exile")
	}
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(other) || !g.Battlefield.Contains(bane) {
		t.Error("an unchosen permanent, or another player's, left")
	}
}

// --- Matter Reshaper ---

func gapB98ReshaperDies(t *testing.T, top game.Card) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top.InstanceID = uuid.New()
	top.Owner, top.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Library.PushTop(top) })
	reshaper := effectstest.PushCatalogPermanent(g, me.ID, "Matter Reshaper", "Creature — Eldrazi", gapMatterReshaperOracle, false)
	_ = g.DestroyPermanentForEffect(reshaper)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestGapB98MatterReshaperOffersACheapPermanentOntoTheBattlefield(t *testing.T) {
	for _, yes := range []bool{true, false} {
		card := game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
		g, me := gapB98ReshaperDies(t, card)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		onField, inHand := false, false
		for _, c := range g.Battlefield.Cards {
			onField = onField || c.Name == "Cheap Bear"
		}
		for _, c := range me.Hand.Cards {
			inHand = inHand || c.Name == "Cheap Bear"
		}
		if onField != yes || inHand == yes {
			t.Errorf("answer %v: on the battlefield %v, in hand %v", yes, onField, inHand)
		}
	}
}

func TestGapB98MatterReshaperPutsAnythingElseIntoYourHand(t *testing.T) {
	for _, card := range []game.Card{
		{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 4, Toughness: 4},
		{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{1}"},
	} {
		g, me := gapB98ReshaperDies(t, card)
		if effectstest.LatestConfirmFor(g, me.ID) != nil {
			t.Errorf("%s: asked about putting a card onto the battlefield", card.Name)
		}
		inHand := false
		for _, c := range me.Hand.Cards {
			inHand = inHand || c.Name == card.Name
		}
		if !inHand {
			t.Errorf("%s did not go into the hand", card.Name)
		}
		for _, c := range g.Battlefield.Cards {
			if c.Name == card.Name {
				t.Errorf("%s entered the battlefield", card.Name)
			}
		}
	}
}

// --- Myriad Construct ---

func TestGapB98MyriadConstructCountsOpponentsNonbasicLandsOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, opp.ID, "Nonbasic One", "Land", "", false)
			effectstest.PushCatalogPermanent(g, opp.ID, "Nonbasic Two", "Land — Forest", "", false)
			effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
			effectstest.PushCatalogPermanent(g, me.ID, "My Nonbasic", "Land", "", false)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Myriad Construct", "Artifact Creature — Construct", gapMyriadConstructOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.PlusOneCounters(g, id); n != tc.want {
				t.Errorf("+1/+1 counters %d, want %d", n, tc.want)
			}
		})
	}
}

func TestGapB98MyriadConstructBecomesTokensWhenTargetedByASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	construct := effectstest.B12Push(g, me.ID, "Myriad Construct", "Artifact Creature — Construct", gapMyriadConstructOracle, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(construct))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(construct) {
		t.Error("the Construct was not sacrificed")
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Construct" && c.IsToken() && c.Controller == me.ID {
			n++
		}
	}
	if n != 4 {
		t.Errorf("%d Construct tokens, want 4 (its power)", n)
	}
}

// --- Abstruse Archaic ---

func TestGapB98AbstruseArchaicCopiesOnlyYourAbilitiesFromColorlessSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	archaic := effectstest.PushCatalogPermanent(g, me.ID, "Abstruse Archaic", "Creature — Avatar", gapAbstruseArchaicOracle, false)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Colorless Rock", "Artifact", "", false)
	red := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Thing", TypeLine: "Creature — Goblin",
		Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	for _, a := range []struct {
		who  *game.Player
		src  uuid.UUID
		text string
	}{{me, rock, "rock trigger"}, {me, red, "red trigger"}, {opp, theirs, "their trigger"}} {
		if err := g.AnnounceTrigger(a.who.ID, a.src, game.AbilityParams{Label: a.text}); err != nil {
			t.Fatalf("AnnounceTrigger %s: %v", a.text, err)
		}
	}
	mine, notColorless, notMine := effectstest.AcItemLabelled(g, "rock trigger"), effectstest.AcItemLabelled(g, "red trigger"), effectstest.AcItemLabelled(g, "their trigger")
	offered := effectstest.AcLegalAbilityTargets(g, me.ID, gapAbstruseArchaicOracle, 0)
	if !offered[mine] {
		t.Error("your ability from a colorless source was not offered")
	}
	if offered[notColorless] {
		t.Error("an ability from a red source was offered")
	}
	if offered[notMine] {
		t.Error("an opponent's ability was offered")
	}
	effectstest.B06AddMana(me, "G")
	if err := g.ActivateCatalogAbility(me.ID, archaic, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AcPassUntilCopy(t, g)
	if n := effectstest.AcCopiesOnStack(g); n != 1 {
		t.Errorf("%d copies on the stack, want 1", n)
	}
}

func TestGapB98AbstruseArchaicStillSeesTheAbilityOfASacrificedColorlessArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Abstruse Archaic", "Creature — Avatar", gapAbstruseArchaicOracle, false)
	vial := effectstest.PushCatalogPermanent(g, me.ID, "Vial of Dragonfire", "Artifact", gapVialOfDragonfireOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate the Vial: %v", err)
	}
	var item uuid.UUID
	g.WithWriteLock(func() {
		for id, it := range g.StackMeta {
			if it != nil && it.SourceCardID == vial {
				item = id
			}
		}
	})
	if item == uuid.Nil {
		t.Fatal("the Vial's ability is not on the stack")
	}
	if !effectstest.AcLegalAbilityTargets(g, me.ID, gapAbstruseArchaicOracle, 0)[item] {
		t.Error("the sacrificed Vial's ability was not offered")
	}
}
