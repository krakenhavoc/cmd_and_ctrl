package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b160_test.go — Modern Horizons, batch 160 (#5519): each test plays
// the real catalog entry through the real activation paths.

const (
	oldB160AccompliceOracle = "01076e04-3106-4968-8ef4-b26fc3786699"
	oldB160TenderOracle     = "1141a12a-2ebd-4e7b-8c8c-c7a1f160244c"
	oldB160InfluenceOracle  = "25d955e3-2b98-4a75-bb51-6072000111be"
	oldB160WurmOracle       = "8f780497-23bb-44c3-982c-e2b8a7a29999"
	oldB160FountainOracle   = "9082d364-9fe4-47bd-9097-8888ae75b2e1"
	oldB160DecoyOracle      = "b22e55ef-fb74-4991-a43b-9d3ce06c23e5"
	oldB160BoughsOracle     = "c6836d78-ca71-4413-aced-d4de19976221"
)

func TestOldB160CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB160AccompliceOracle: "Planebound Accomplice",
		oldB160TenderOracle:     "Rime Tender",
		oldB160InfluenceOracle:  "Ayula's Influence",
		oldB160WurmOracle:       "Conifer Wurm",
		oldB160FountainOracle:   "Fountain of Ichor",
		oldB160DecoyOracle:      "Zhalfirin Decoy",
		oldB160BoughsOracle:     "Birthing Boughs",
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

func oldB160Activate(g *game.Game, who, source uuid.UUID, idx int, params game.ActivateAbilityParams) error {
	params.Strict = true
	return g.ActivateCatalogAbility(who, source, idx, params)
}

func oldB160Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func oldB160Tokens(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == owner && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func oldB160Tap(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
			}
		}
	})
}

// --- Birthing Boughs ---

func TestOldB160BirthingBoughsMakesAChangelingShapeshifter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	boughs := effectstest.PushCatalogPermanent(g, me.ID, "Birthing Boughs", "Artifact", oldB160BoughsOracle, false)
	effectstest.B06AddMana(me, "G", "G", "G", "G")
	if err := oldB160Activate(g, me.ID, boughs, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	tokens := oldB160Tokens(g, me.ID, "Shapeshifter")
	if len(tokens) != 1 {
		t.Fatalf("%d Shapeshifter tokens, want 1", len(tokens))
	}
	id := tokens[0].InstanceID
	if p, tt := effectstest.PtOf(t, g, id); p != 2 || tt != 2 {
		t.Errorf("the token is %d/%d, want 2/2", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "changeling") || len(tokens[0].Colors) != 0 {
		t.Error("the token is not a colorless changeling")
	}
	if !effectstest.IsTapped(g, boughs) {
		t.Error("the Boughs did not tap")
	}
	effectstest.B06AddMana(me, "G", "G", "G", "G")
	if err := oldB160Activate(g, me.ID, boughs, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Boughs activated while tapped")
	}
}

func TestOldB160BirthingBoughsRefusesWithTooLittleMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	boughs := effectstest.PushCatalogPermanent(g, me.ID, "Birthing Boughs", "Artifact", oldB160BoughsOracle, false)
	effectstest.B06AddMana(me, "G", "G", "G")
	if err := oldB160Activate(g, me.ID, boughs, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Boughs activated for three mana")
	}
}

// --- Ayula's Influence ---

func TestOldB160AyulasInfluenceDiscardsALandForABear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	infl := effectstest.PushCatalogPermanent(g, me.ID, "Ayula's Influence", "Enchantment", oldB160InfluenceOracle, false)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := oldB160Activate(g, me.ID, infl, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Error("a nonland card paid the discard")
	}
	if err := oldB160Activate(g, me.ID, infl, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(land) {
		t.Error("the land was not discarded")
	}
	bears := oldB160Tokens(g, me.ID, "Bear")
	if len(bears) != 1 {
		t.Fatalf("%d Bear tokens, want 1", len(bears))
	}
	if p, tt := effectstest.PtOf(t, g, bears[0].InstanceID); p != 2 || tt != 2 || len(bears[0].Colors) != 1 || bears[0].Colors[0] != "G" {
		t.Errorf("the Bear is %d/%d %v, want a green 2/2", p, tt, bears[0].Colors)
	}
}

// --- Conifer Wurm ---

func TestOldB160ConiferWurmGrowsByYourSnowPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wurm := effectstest.PushCatalogPermanent(g, me.ID, "Conifer Wurm", "Snow Creature — Wurm", oldB160WurmOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wurm {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
			}
		}
	})
	effectstest.PushCatalogPermanent(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Snow-Covered Island", "Basic Snow Land — Island", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Plain Forest", "Basic Land — Forest", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Snow Land", "Basic Snow Land — Forest", "", false)
	effectstest.B06AddMana(me, "G", "G", "G", "G")
	if err := oldB160Activate(g, me.ID, wurm, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, wurm); p != 7 || tt != 7 {
		t.Errorf("the Wurm is %d/%d, want 7/7 (+3 for itself and two snow lands)", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wurm, "trample") {
		t.Error("the Wurm lacks trample")
	}
}

// --- Rime Tender ---

func TestOldB160RimeTenderUntapsAnotherSnowPermanentOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tender := effectstest.PushCatalogPermanent(g, me.ID, "Rime Tender", "Snow Creature — Human Druid", oldB160TenderOracle, false)
	snow := effectstest.PushCatalogPermanent(g, me.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "", false)
	plain := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	oldB160Tap(g, snow, true)
	oldB160Tap(g, plain, true)
	if err := oldB160Activate(g, me.ID, tender, 0, game.ActivateAbilityParams{Targets: oldB160Target(plain)}); err == nil {
		t.Error("the Tender targeted a permanent that is not snow")
	}
	if err := oldB160Activate(g, me.ID, tender, 0, game.ActivateAbilityParams{Targets: oldB160Target(tender)}); err == nil {
		t.Error("the Tender targeted itself")
	}
	if err := oldB160Activate(g, me.ID, tender, 0, game.ActivateAbilityParams{Targets: oldB160Target(snow)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, snow) {
		t.Error("the snow land was not untapped")
	}
	if !effectstest.IsTapped(g, tender) {
		t.Error("the Tender did not tap to pay")
	}
}

// --- Fountain of Ichor ---

func TestOldB160FountainOfIchorMakesManaAndBecomesADinosaur(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	fountain := effectstest.PushCatalogPermanent(g, me.ID, "Fountain of Ichor", "Artifact", oldB160FountainOracle, false)
	if err := g.ActivateManaAbility(me.ID, fountain, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if me.ManaPool.Total() != 1 {
		t.Errorf("the pool holds %d mana, want 1", me.ManaPool.Total())
	}
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := oldB160Activate(g, me.ID, fountain, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, fountain); p != 3 || tt != 3 {
		t.Errorf("the Fountain is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveType(t, g, fountain, "Creature") || !effectstest.HasEffectiveType(t, g, fountain, "Artifact") {
		t.Errorf("types %v, want an artifact creature", effectstest.EffectiveTypes(t, g, fountain))
	}
	// It ends with the turn.
	effectstest.NextTurnOfMine(t, g)
	if effectstest.HasEffectiveType(t, g, fountain, "Creature") {
		t.Error("the Fountain is still a creature on a later turn")
	}
}

// --- Zhalfirin Decoy ---

func TestOldB160DecoyTapsACreatureOnlyAfterACreatureEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	decoy := effectstest.PushCatalogPermanent(g, me.ID, "Zhalfirin Decoy", "Creature — Human Soldier", oldB160DecoyOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := oldB160Activate(g, me.ID, decoy, 0, game.ActivateAbilityParams{Targets: oldB160Target(victim)}); err == nil {
		t.Fatal("the Decoy tapped a creature before one entered under my control this turn")
	}
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "New Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := oldB160Activate(g, me.ID, decoy, 0, game.ActivateAbilityParams{Targets: oldB160Target(victim)}); err != nil {
		t.Fatalf("activate after a creature entered: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, victim) {
		t.Error("the target creature was not tapped")
	}
}

func TestOldB160DecoyIgnoresAnOpponentsCreatureEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	decoy := effectstest.PushCatalogPermanent(g, me.ID, "Zhalfirin Decoy", "Creature — Human Soldier", oldB160DecoyOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Their New Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := oldB160Activate(g, me.ID, decoy, 0, game.ActivateAbilityParams{Targets: oldB160Target(victim)}); err == nil {
		t.Error("an opponent's creature entering opened the Decoy's window")
	}
}

// --- Planebound Accomplice ---

func oldB160Planeswalker(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nissa of Shadowed Boughs", TypeLine: "Legendary Planeswalker — Nissa", OracleID: "15e3e226-0dd1-4c9b-a7ec-af67509cbfc5", Owner: me.ID, Controller: me.ID})
	return id
}

func TestOldB160PlaneboundAccompliceCheatsInAPlaneswalkerAndSacrificesItAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	acc := effectstest.PushCatalogPermanent(g, me.ID, "Planebound Accomplice", "Creature — Human Wizard", oldB160AccompliceOracle, false)
	walker := oldB160Planeswalker(me)
	other := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: other, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R")
	if err := oldB160Activate(g, me.ID, acc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(walker) {
		t.Fatal("the planeswalker card did not enter the battlefield")
	}
	if g.Battlefield.Contains(other) {
		t.Error("a card that is not a planeswalker entered")
	}
	if got := effectstest.CounterOn(g, walker, game.CounterLoyalty); got != 4 {
		t.Errorf("the planeswalker entered with %d loyalty, want its starting 4", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker was not sacrificed at the beginning of the end step")
	}
	if !me.Graveyard.Contains(walker) {
		t.Error("the sacrificed planeswalker is not in its owner's graveyard")
	}
}

func TestOldB160PlaneboundAccompliceAsksNothingWithoutAPlaneswalkerInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	acc := effectstest.PushCatalogPermanent(g, me.ID, "Planebound Accomplice", "Creature — Human Wizard", oldB160AccompliceOracle, false)
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R")
	if err := oldB160Activate(g, me.ID, acc, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			t.Errorf("a %v prompt was opened with no planeswalker card in hand", c.Kind)
		}
	}
	if g.Battlefield.Contains(bear) {
		t.Error("a creature card was put onto the battlefield")
	}
}
