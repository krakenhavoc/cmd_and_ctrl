package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// airbend_test.go — S22: the three airbend cards the engine can now
// express. Aang, the Last Airbender is the mechanic's proof; Appa
// adds "cast a spell from exile" and Monk Gyatso adds "becomes the
// target".

const (
	aangTheLastAirbenderOracle  = "70564c3a-858f-498e-8b92-acb3ca54ae7e"
	appaSteadfastGuardianOracle = "03c141ca-11e4-4927-a6bd-980ee1203c73"
	monkGyatsoOracle            = "ff92fa60-f0fe-496e-8155-d9d6f5af651b"
)

// airbendGrantOn reads the permission covering a card in exile.
func airbendGrantOn(g *game.Game, id uuid.UUID) (*game.CastPermission, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID != id {
			continue
		}
		if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
			return perm, true
		}
		return &game.CastPermission{}, true
	}
	return nil, false
}

// assertAirbent checks the shape of a finished airbend: the card is
// in exile, and its OWNER may cast it for {2} with no expiry.
func assertAirbent(t *testing.T, g *game.Game, id, owner uuid.UUID) {
	t.Helper()
	if g.Battlefield.Contains(id) {
		t.Fatalf("the permanent is still on the battlefield")
	}
	grant, ok := airbendGrantOn(g, id)
	if !ok {
		t.Fatalf("the card never reached exile")
	}
	if grant.Player != owner {
		t.Errorf("grant holder = %v, want the card's owner %v", grant.Player, owner)
	}
	if !grant.WhileInZone {
		t.Errorf("airbend grant expires; it should last while the card is exiled")
	}
	if grant.Cost != AirbendCost {
		t.Errorf("cost override = %q, want %q", grant.Cost, AirbendCost)
	}
	if !grant.CastOnly {
		t.Errorf("airbend says CAST, so the grant should be cast-only")
	}
}

// --- Aang, the Last Airbender ------------------------------------

// The mechanic, end to end: Aang enters, airbends an opponent's
// permanent, and that opponent — the owner — gets it back for {2}.
func TestAangAirbendsAnOpponentsPermanentAndOwnerRebuysIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	aang := castAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", aangTheLastAirbenderOracle)
	if me.ID != g.Seats[g.Turn.ActiveSeat].ID {
		t.Fatalf("test assumes seat 0 is active")
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, victim, opp.ID)
	_ = aang
}

// Declining the prompt is how "up to one" chooses zero.
func TestAangDeclinedAirbendsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	castAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", aangTheLastAirbenderOracle)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(victim) {
		t.Errorf("a declined trigger airbent something anyway")
	}
	if g.Exile.Size() != 0 {
		t.Errorf("exile is not empty after a declined airbend")
	}
}

// "ANOTHER target nonland permanent": the picker can't express it,
// so the resolution has to. Aang pointed at himself does nothing.
func TestAangWillNotAirbendHimself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	aang := castAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", aangTheLastAirbenderOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, aang)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(aang) {
		t.Errorf("Aang airbent himself; the printed clause says ANOTHER")
	}
}

// A land is not a legal target, so an all-lands board drops the
// trigger before it ever prompts.
func TestAangDoesNotTargetLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := aangPushLand(g, opp.ID, "Their Island", false)

	castAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", aangTheLastAirbenderOracle)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(land) {
		t.Errorf("Aang airbent a land")
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Errorf("a pick_target prompt opened with no legal nonland permanent")
	}
}

// --- Appa, Steadfast Guardian ------------------------------------

// Appa airbends any number of your own permanents — multi-target
// from a trigger, which nothing in the catalog had done before.
func TestAppaAirbendsSeveralOfYourOwnPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine1 := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	mine2 := pushCreatureToBattlefieldForTest(g, me.ID, "My Other Bear")
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")

	castAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", appaSteadfastGuardianOracle)

	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt for Appa")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 0 {
		t.Errorf("count = %d..%d, want 0..unbounded for \"any number\"", p.PickTargetMin, p.PickTargetMax)
	}
	if hasID(p.PickTargetCards, theirs) {
		t.Errorf("the picker offered an opponent's permanent; the clause says YOU CONTROL")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine1},
		{Kind: game.TargetCard, ID: mine2},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, mine1, me.ID)
	assertAirbent(t, g, mine2, me.ID)
	if !g.Battlefield.Contains(theirs) {
		t.Errorf("an untargeted permanent was airbent")
	}
}

// "Whenever you cast a spell from exile" — the EventCast source
// zone in anger. Casting from HAND must not make an Ally.
func TestAppaMakesAnAllyOnlyForCastsFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	appa := castAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", appaSteadfastGuardianOracle)
	// Appa's own ETB trigger: nothing else is on the board, so the
	// prompt (if any) is answered with an empty pick.
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("ResolvePickTargets(none): %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(appa) {
		t.Fatalf("Appa never reached the battlefield")
	}
	allies := countAllies(g, me.ID)

	// A cast from hand: no Ally.
	fromHand := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: fromHand, Name: "Hand Instant", TypeLine: "Instant",
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, fromHand, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countAllies(g, me.ID); got != allies {
		t.Errorf("allies after a cast from hand = %d, want %d", got, allies)
	}

	// A cast from exile under an airbend grant: one Ally.
	fromExile := uuid.New()
	g.Exile.PushTop(game.Card{
		InstanceID: fromExile, Name: "Exiled Instant", TypeLine: "Instant",
		Owner: me.ID, Controller: me.ID,
	})
	g.GrantCastPermissionOverCardForEffect(fromExile, game.CastPermission{
		Player: me.ID, CastOnly: true, WhileInZone: true, Cost: AirbendCost,
	})
	if err := g.CastSpell(me.ID, fromExile, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countAllies(g, me.ID); got != allies+1 {
		t.Errorf("allies after a cast from exile = %d, want %d", got, allies+1)
	}
}

func countAllies(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Ally" {
			n++
		}
	}
	return n
}

// --- Monk Gyatso -------------------------------------------------

// The card's whole point: removal is announced at your creature,
// Gyatso's trigger resolves first and exiles it, and the removal
// fizzles for want of a legal target.
func TestMonkGyatsoAirbendsATargetedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gyatso := pushCreatureToBattlefieldForTest(g, me.ID, "Monk Gyatso")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == gyatso {
			g.Battlefield.Cards[i].OracleID = monkGyatsoOracle
		}
	}
	mine := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	// An opponent's Doom Blade points at the Bear.
	aangAdvanceToMain(t, g, 1)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: doomBladeOracle, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}

	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, mine, me.ID)
	if opp.Graveyard.Contains(mine) {
		t.Errorf("the creature died; Gyatso should have exiled it first")
	}
}

// Gyatso watches creatures YOU control, and not himself.
func TestMonkGyatsoIgnoresOpponentsAndHimself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gyatso := pushCreatureToBattlefieldForTest(g, me.ID, "Monk Gyatso")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == gyatso {
			g.Battlefield.Cards[i].OracleID = monkGyatsoOracle
		}
	}
	theirs := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")

	aangAdvanceToMain(t, g, 1)
	for _, target := range []uuid.UUID{theirs, gyatso} {
		bolt := uuid.New()
		opp.Hand.PushTop(game.Card{
			InstanceID: bolt, Name: "Doom Blade", TypeLine: "Instant",
			OracleID: doomBladeOracle, Owner: opp.ID, Controller: opp.ID,
		})
		if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		if p := latestTriggerPromptFor(g, me.ID); p != nil {
			t.Fatalf("Gyatso triggered on %v; he watches only OTHER creatures YOU control", target)
		}
		passPriorityAroundTable(t, g)
	}
}

// latestTriggerPromptFor returns the most recent trigger prompt
// addressed to chooser, or nil.
func latestTriggerPromptFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Chooser == chooser {
			return c
		}
	}
	return nil
}
