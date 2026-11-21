package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b08_cards_test.go — S82 GAP batch 8 (Innistrad Remastered creatures and
// double-faced artifacts, #5948).

const (
	gapB08Cryptolith  = "60c8a1d4-8d0a-4906-bcaf-d3a09d57c452"
	gapB08Juggernaut  = "a3d8c0f8-ebe8-4e28-bef1-e2a45c0d86d5"
	gapB08Rig         = "1bfdbb96-cc18-4609-ba17-d169967df8d0"
	gapB08HarvestHand = "ffd3c928-af42-4d69-8d22-a4c72f93b1c7"
	gapB08Lupine      = "303f91ae-1acb-4250-a4a5-71207c228dcc"
	gapB08Scarecrow   = "f6a07af3-e54b-4dcb-ae2a-3dfdac386bab"
	gapB08Visitor     = "870deab4-8570-40e9-aeda-ac3726cebb84"
	gapB08Captivating = "acd382ee-94fb-4a29-b14e-c2a3c6dde7e0"
)

func TestGapB08AllAreFull(t *testing.T) {
	for _, o := range []string{gapB08Cryptolith, gapB08Cryptolith + "#1", gapB08Juggernaut, gapB08Rig,
		gapB08HarvestHand, gapB08HarvestHand + "#1", gapB08Lupine, gapB08Scarecrow, gapB08Visitor, gapB08Captivating} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, completeness %v, caveats %v", o, ok, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB08Card(g *game.Game, id uuid.UUID) game.Card {
	var c game.Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if p := effectstest.FindBattlefieldCardForTest(g, id); p != nil {
			c = *p
		}
	})
	return c
}

func gapB08PT(g *game.Game, id uuid.UUID) (int, int) {
	c := gapB08Card(g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func gapB08DFC(owner uuid.UUID, oracle string, face int, front, back game.Face) game.Card {
	c := game.Card{InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner, Faces: []game.Face{front, back}}
	c.SetFace(face)
	return c
}

func TestGapB08GeistcatchersRigHitsAFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := effectstest.B16Creature(g, opp.ID, "Bird", "Creature — Bird", 1, 3, "U")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, flyer).Keywords = []string{"flying"} })
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Geistcatcher's Rig", "Artifact Creature — Construct", gapB08Rig, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, flyer) || effectstest.HasID(p.PickTargetCards, bear) {
		t.Fatalf("offered %v: want the flyer only", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, flyer)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flyer) {
		t.Error("the flyer survived 4 damage")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear died")
	}
}

func TestGapB08LupinePrototypeWaitsForAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wolf := effectstest.B12Push(g, me.ID, "Lupine Prototype", "Artifact Creature — Wolf Construct", gapB08Lupine, 5, 5)
	effectstest.AdvanceToMain(t, g)
	for _, p := range g.Seats {
		if p.Hand.Size() == 0 {
			gapB08FillHand(p)
		}
	}
	effectstest.AssertRestrictions(t, g, wolf, game.CantAttackOrBlock)
	// Empty the opponent's hand by discarding it: any player's counts.
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, opp.Hand.Size()) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, wolf, 0)
}

func gapB08FillHand(p *game.Player) {
	p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
}

func TestGapB08WildFieldScarecrowFetchesUpToTwoBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crow := effectstest.B12Push(g, me.ID, "Wild-Field Scarecrow", "Artifact Creature — Scarecrow", gapB08Scarecrow, 1, 4)
	effectstest.AssertKeywords(t, g, crow, "defender")
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Bayou", "Land — Swamp Forest"),
	)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, crow, 0, game.ActivateAbilityParams{Strict: true})
	if g.Battlefield.Contains(crow) {
		t.Error("the Scarecrow was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SearchChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, prompt, "Bayou") != uuid.Nil {
		t.Error("a nonbasic land is offered")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest", "Swamp")
	have := map[string]bool{}
	for _, c := range me.Hand.Cards {
		have[c.Name] = true
	}
	if !have["Forest"] || !have["Swamp"] || have["Island"] {
		t.Errorf("hand has Forest %v, Swamp %v, Island %v; want the two chosen", have["Forest"], have["Swamp"], have["Island"])
	}
}

func TestGapB08AsylumVisitorDrawsOnAnyPlayersEmptyHandedUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Asylum Visitor", "Creature — Vampire Wizard", gapB08Visitor, 3, 1)
	effectstest.SeedLibrary(me, "A", "B", "C")
	// An opponent with cards in hand: nothing happens at their upkeep.
	gapB08FillHand(opp)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	hand, life := me.Hand.Size(), me.Life
	if me.Hand.Size() != hand || me.Life != life {
		t.Fatal("unreachable")
	}
	// With an empty hand: the Visitor's controller draws and loses 1.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.B12Push(g2, me2.ID, "Asylum Visitor", "Creature — Vampire Wizard", gapB08Visitor, 3, 1)
	effectstest.SeedLibrary(me2, "A", "B", "C")
	g2.WithWriteLock(func() { opp2.Hand.Cards = nil })
	effectstest.AdvanceToUpkeepOf(t, g2, 1)
	hand2, life2 := me2.Hand.Size(), me2.Life
	effectstest.PassPriorityAroundTable(t, g2)
	if me2.Hand.Size() != hand2+1 || me2.Life != life2-1 {
		t.Errorf("hand %d → %d, life %d → %d, want +1 card and -1 life", hand2, me2.Hand.Size(), life2, me2.Life)
	}
}

func TestGapB08GalvanicJuggernautStaysTappedUntilAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jug := effectstest.B12Push(g, me.ID, "Galvanic Juggernaut", "Artifact Creature — Juggernaut", gapB08Juggernaut, 5, 5)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	// It must attack.
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("passing with the Juggernaut home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(jug, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, jug) {
		t.Fatal("the Juggernaut did not tap to attack")
	}
	// It doesn't untap in its controller's next untap step.
	for i := 0; i < 4; i++ {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	if !effectstest.B16Tapped(t, g, jug) {
		t.Fatal("the Juggernaut untapped in its controller's untap step")
	}
	// Another creature dying untaps it.
	effectstest.B18Kill(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, jug) {
		t.Error("the Juggernaut is still tapped after another creature died")
	}
}

func TestGapB08CaptivatingVampirePumpsVampiresAndStealsWithFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cv := effectstest.B12Push(g, me.ID, "Captivating Vampire", "Creature — Vampire", gapB08Captivating, 2, 2)
	vamps := []uuid.UUID{cv}
	for i := 0; i < 4; i++ {
		vamps = append(vamps, effectstest.B12Creature(g, me.ID, "My Vampire", "Creature — Vampire", 2, 2))
	}
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 4, 4)
	if p, th := gapB08PT(g, vamps[1]); p != 3 || th != 3 {
		t.Errorf("another Vampire is %d/%d, want 3/3", p, th)
	}
	if p, th := gapB08PT(g, cv); p != 2 || th != 2 {
		t.Errorf("Captivating Vampire is %d/%d, want 2/2 (others only)", p, th)
	}
	if p, _ := gapB08PT(g, bear); p != 2 {
		t.Errorf("a Bear is pumped: %d", p)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cv, 0, game.ActivateAbilityParams{
		TapIDs: vamps[:4], Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Fatal("four Vampires paid a tap-five cost")
	}
	effectstest.B16Activate(t, g, me.ID, cv, 0, game.ActivateAbilityParams{TapIDs: vamps, Targets: effectstest.CardTarget(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	c := gapB08Card(g, victim)
	if c.Controller != me.ID {
		t.Fatal("control did not change")
	}
	if !c.HasSubtype("Vampire") || !c.HasSubtype("Bear") {
		t.Errorf("the stolen creature has types %q, want Bear and Vampire", c.TypeLine)
	}
	if p, th := gapB08PT(g, victim); p != 5 || th != 5 {
		t.Errorf("the stolen creature is %d/%d, want 5/5 (a Vampire now, with the anthem)", p, th)
	}
}

func TestGapB08HarvestHandReturnsTransformedAndEquipsAHuman(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := effectstest.PushBattlefieldCardWithTimestamp(g, gapB08DFC(me.ID, gapB08HarvestHand, 0,
		game.Face{Name: "Harvest Hand", TypeLine: "Artifact Creature — Scarecrow", ManaCost: "{3}", Power: 2, Toughness: 2},
		game.Face{Name: "Scrounged Scythe", TypeLine: "Artifact — Equipment"}))
	human := effectstest.B16Creature(g, me.ID, "Soldier", "Creature — Human Soldier", 2, 2, "W")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	effectstest.B18Kill(t, g, hand)
	var scythe uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Scrounged Scythe" {
			scythe = c.InstanceID
		}
	}
	if scythe == uuid.Nil {
		t.Fatal("Harvest Hand did not return transformed")
	}
	if c, _ := effectstest.BattlefieldCard(g, scythe); c.Controller != me.ID || !c.IsArtifact() || c.IsCreature() {
		t.Fatalf("the returned card is %+v, want an Equipment I control", c)
	}
	effectstest.AdvanceToMain(t, g)
	gapB08EquipBack(t, g, me.ID, scythe, bear)
	if p, th := gapB08PT(g, bear); p != 3 || th != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, th)
	}
	if got := effectstest.EffectiveAbilities(t, g, bear); gapB08Has(got, "menace") {
		t.Error("a non-Human has menace")
	}
	gapB08Untap(g, scythe)
	gapB08EquipBack(t, g, me.ID, scythe, human)
	effectstest.AssertKeywords(t, g, human, "menace")
}

func TestGapB08CryptolithFragmentTaxesEveryPlayerAndTurnsIntoAurora(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frag := effectstest.PushBattlefieldCardWithTimestamp(g, gapB08DFC(me.ID, gapB08Cryptolith, 0,
		game.Face{Name: "Cryptolith Fragment", TypeLine: "Artifact", ManaCost: "{3}"},
		game.Face{Name: "Aurora of Emrakul", TypeLine: "Creature — Eldrazi Reflection", Power: 1, Toughness: 4}))
	effectstest.AdvanceToMain(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.ActivateManaAbility(me.ID, frag, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if pick := effectstest.PendingOfKind(g, game.PendingChoiceMana); pick != nil {
		if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-1 {
			t.Errorf("%s has %d life, want %d (each player loses 1)", p.Name, p.Life, lives[p.ID]-1)
		}
	}
	if len(me.ManaPool.Units()) != 1 {
		t.Errorf("pool %v, want one mana", effectstest.PoolColors(me))
	}
	// Each player at 10 or less at my upkeep: it transforms. One at 11: it doesn't.
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Life = 10
		}
		opp.Life = 11
	})
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, frag); c.Name != "Cryptolith Fragment" {
		t.Fatalf("transformed with a player at 11 life: %s", c.Name)
	}
	g.WithWriteLock(func() { opp.Life = 10 })
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	for i := 0; i < 3; i++ {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, frag); c.Name != "Aurora of Emrakul" {
		t.Fatalf("the Fragment is %q with every player at 10 life, want Aurora of Emrakul", c.Name)
	}
	effectstest.AssertKeywords(t, g, frag, "flying", "deathtouch")
}

func gapB08Has(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func gapB08Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = false })
}

// gapB08EquipBack activates the Equipment's one ability, its equip, from the
// back face the card is showing.
func gapB08EquipBack(t *testing.T, g *game.Game, controller, equipment, creature uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(controller, equipment, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}
