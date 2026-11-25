package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b311_test.go — Battle for Zendikar lands (#5547). Every test
// plays the real catalog entry through the engine.

const (
	oldB311SkylineCascade = "79301ae1-8c9c-4723-be21-dc27e1646f35"
	oldB311FertileThicket = "7ba580c9-f933-43d9-b03d-a349faa6c641"
	oldB311LoomingSpires  = "7d09b136-525f-49dd-a3a2-dfaca4e8e9a8"
	oldB311AllyEncampment = "9d293b69-12b7-4b50-a0a7-c4f493dee30b"
	oldB311BlightedGorge  = "c2cb0afd-781f-4cfa-b680-ed1edfa81868"
	oldB311BlightedSteppe = "db16a2fb-dc42-4086-9928-52076043097f"
)

// oldB311PlayLand puts the catalog land into the active seat's hand and
// plays it as the land drop.
func oldB311PlayLand(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Land", OracleID: oracle, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	return id
}

func TestOldB311EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB311SkylineCascade, oldB311FertileThicket, oldB311LoomingSpires, oldB311AllyEncampment, oldB311BlightedGorge, oldB311BlightedSteppe} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB311SkylineCascadeEntersTappedAndFreezesOnlyTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	frozen := effectstest.B12Creature(g, opp.ID, "Frozen", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free", "Creature — Bear", 2, 2)
	effectstest.B02bSetTapped(g, frozen, true)
	effectstest.B02bSetTapped(g, free, true)
	land := oldB311PlayLand(t, g, "Skyline Cascade", oldB311SkylineCascade)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the land did not enter tapped")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, frozen)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if !effectstest.B16Tapped(t, g, frozen) {
		t.Error("the target untapped in its controller's next untap step")
	}
	if effectstest.B16Tapped(t, g, free) {
		t.Error("a creature that was not targeted stayed tapped")
	}
}

func TestOldB311LoomingSpiresPumpsAndGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := oldB311PlayLand(t, g, "Looming Spires", oldB311LoomingSpires)
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the land did not enter tapped")
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the creature is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the creature has no first strike")
	}
}

// oldB311ThicketTable seeds a library whose top five are
// [a, b, Forest, d, e] over one deeper card, and plays Fertile Thicket.
func oldB311ThicketTable(t *testing.T, land bool) (g *game.Game, me *game.Player, ids []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	deep := effectstest.SeedLibraryTop(me, "Deep", "Sorcery", "")
	e := effectstest.SeedLibraryTop(me, "E", "Sorcery", "")
	d := effectstest.SeedLibraryTop(me, "D", "Sorcery", "")
	var third uuid.UUID
	if land {
		third = effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	} else {
		third = effectstest.SeedLibraryTop(me, "Third", "Sorcery", "")
	}
	b := effectstest.SeedLibraryTop(me, "B", "Sorcery", "")
	a := effectstest.SeedLibraryTop(me, "A", "Sorcery", "")
	oldB311PlayLand(t, g, "Fertile Thicket", oldB311FertileThicket)
	return g, me, []uuid.UUID{a, b, third, d, e, deep}
}

// oldB311AcceptTheLook answers Fertile Thicket's "you may look" prompt and
// lets the trigger resolve.
func oldB311AcceptTheLook(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	prompt := effectstest.ConfirmForTriggerPrompt(g, me.ID)
	if prompt == nil {
		effectstest.PassPriorityAroundTable(t, g)
		prompt = effectstest.ConfirmForTriggerPrompt(g, me.ID)
	}
	if prompt == nil {
		t.Fatal("no \"you may look\" prompt")
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, yes); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB311FertileThicketKeepsTheBasicLandOnTopAndSendsTheRestToTheBottom(t *testing.T) {
	g, me, ids := oldB311ThicketTable(t, true)
	oldB311AcceptTheLook(t, g, me, true)
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards)
	if pick == nil || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[2] {
		t.Fatalf("pick = %+v, want one candidate: the Forest", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("the four other cards were not offered for the bottom")
	}
	rest := []uuid.UUID{ids[4], ids[0], ids[3], ids[1]}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, rest, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[2] {
		t.Errorf("the top card is %q, want the Forest", top.Name)
	}
	cards := me.Library.Cards
	if second := cards[len(cards)-2]; second.InstanceID != ids[5] {
		t.Errorf("the second card is %q, want the card that was sixth from the top", second.Name)
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range cards[:4] {
		bottom[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{ids[0], ids[1], ids[3], ids[4]} {
		if !bottom[id] {
			t.Errorf("card %s is not among the bottom four", id)
		}
	}
}

func TestOldB311FertileThicketWithoutABasicLandSendsAllFiveToTheBottom(t *testing.T) {
	g, me, ids := oldB311ThicketTable(t, false)
	oldB311AcceptTheLook(t, g, me, true)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards) != nil {
		t.Fatal("asked to pick a basic land from five nonlands")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("the five cards were not offered for the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, ids[:5], nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if top, _ := me.Library.Top(); top.InstanceID != ids[5] {
		t.Errorf("the top card is %q, want the deep card", top.Name)
	}
}

func TestOldB311FertileThicketMayDeclineToLook(t *testing.T) {
	g, me, ids := oldB311ThicketTable(t, true)
	oldB311AcceptTheLook(t, g, me, false)
	if top, _ := me.Library.Top(); top.InstanceID != ids[0] {
		t.Errorf("declining changed the top card to %q", top.Name)
	}
}

func TestOldB311AllyEncampmentTapsForColourlessAndAnyColourOnlyForAllies(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{{"ally", "Creature — Kor Ally", true}, {"bear", "Creature — Bear", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.SpendTable(t)
			camp := effectstest.PushCatalogPermanent(g, me.ID, "Ally Encampment", "Land", oldB311AllyEncampment, false)
			effectstest.ActivateManaFor(t, g, me.ID, camp, 1, game.ManaAbilityParams{})
			if effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana) != nil {
				effectstest.ResolveManaPickAnyColor(t, g, me.ID)
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test", TypeLine: tc.typeLine, ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if tc.ok && err != nil {
				t.Errorf("the mana could not pay for an Ally spell: %v", err)
			}
			if !tc.ok {
				effectstest.RefusedForMana(t, err, "Ally Encampment's mana for a non-Ally")
			}
		})
	}
}

func TestOldB311AllyEncampmentSacrificeReturnsAnAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	camp := effectstest.PushCatalogPermanent(g, me.ID, "Ally Encampment", "Land", oldB311AllyEncampment, false)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Kor Ally", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ally", "Creature — Kor Ally", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "C")
	for _, bad := range []uuid.UUID{bear, theirs} {
		err := g.ActivateCatalogAbility(me.ID, camp, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}}})
		if err == nil {
			t.Fatalf("activation targeting a non-Ally or an opponent's Ally was accepted")
		}
	}
	effectstest.B16Activate(t, g, me.ID, camp, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ally}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ally) {
		t.Error("the Ally did not return to hand")
	}
	if _, ok := effectstest.BattlefieldCard(g, camp); ok {
		t.Error("the land was not sacrificed")
	}
}

func TestOldB311BlightedGorgeDealsTwoDamageToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gorge := effectstest.PushCatalogPermanent(g, me.ID, "Blighted Gorge", "Land", oldB311BlightedGorge, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RCCCC")
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, gorge, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("dealt %d, want 2", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, gorge); ok {
		t.Error("the land was not sacrificed")
	}
}

func TestOldB311BlightedSteppeGainsTwoPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	steppe := effectstest.PushCatalogPermanent(g, me.ID, "Blighted Steppe", "Land", oldB311BlightedSteppe, false)
	effectstest.B12Creature(g, me.ID, "One", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WCCC")
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, steppe, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 4 {
		t.Errorf("gained %d, want 4 (two creatures)", got)
	}
	if _, ok := effectstest.BattlefieldCard(g, steppe); ok {
		t.Error("the land was not sacrificed")
	}
}
