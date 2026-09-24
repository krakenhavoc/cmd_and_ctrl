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
	if grant.Duration.Kind != game.WhileInZone {
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

// #1304 / #1299: airbending an opposing COMMANDER. Exiling it opens
// the CR 903.9 window, so its owner is asked about the command zone
// before anything moves. A "no" must leave the commander in exile
// WITH the airbend grant — the old primitive stamped the grant on the
// line after a paused exile, found nothing in exile, and stranded the
// card — and the owner must then be able to cast it for {2}. A "yes"
// sends it home, where there is nothing to grant.
func TestAirbendingACommanderKeepsTheRebuyWhenItsOwnerDeclines(t *testing.T) {
	for _, takeCommandZone := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		vivi := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: vivi, Name: "Their Commander", TypeLine: "Legendary Creature — Human Wizard",
			ManaCost: "{2}{U}{R}", Power: 3, Toughness: 3,
			Owner: opp.ID, Controller: opp.ID, IsCommander: true,
		})

		castAndResolveCreature(t, g, "Aang, the Last Airbender",
			"Legendary Creature — Human Avatar Ally", aangTheLastAirbenderOracle)
		answerLatestTriggerPrompt(t, g, me.ID, true)
		pickCard(t, g, me.ID, vivi)
		passPriorityAroundTable(t, g)

		offer := latestChoiceOfKind(g, game.PendingChoiceOptionalReplacement)
		if offer == nil || offer.Chooser != opp.ID {
			t.Fatalf("the commander's owner was not asked about the command zone (CR 903.9)")
		}
		if err := g.ResolveOptionalReplacement(offer.ID, opp.ID, takeCommandZone); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
		passPriorityAroundTable(t, g)

		if takeCommandZone {
			if !opp.Command.Contains(vivi) {
				t.Fatalf("a commander whose owner took the offer is not in the command zone")
			}
			if perm := g.CastPermissionOnCardByIDForEffect(vivi); perm != nil {
				t.Errorf("a commander back in the command zone carries an airbend grant: %+v", perm)
			}
			continue
		}
		assertAirbent(t, g, vivi, opp.ID)

		// The rebuy itself: on its owner's turn, {2} casts it.
		advanceToPrecombatMainOf(t, g, 1)
		b06AddMana(opp, "C", "C")
		if err := g.CastSpell(opp.ID, vivi, game.CastSpellParams{FromZone: "exile"}); err != nil {
			t.Fatalf("owner casting the airbent commander for {2}: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(vivi) {
			t.Errorf("the airbent commander did not come back for {2}")
		}
	}
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

// "ANOTHER target nonland permanent": Aang is not a legal target of
// his own trigger. With another permanent around the picker offers
// that one and never Aang, and an answer naming Aang is refused; alone
// on the board he has nothing to target and the trigger asks nothing
// (CR 603.3d).
func TestAangWillNotAirbendHimself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	aang := castAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", aangTheLastAirbenderOracle)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt after saying yes")
	}
	if hasID(p.PickTargetCards, aang) {
		t.Errorf("the picker offered Aang to his own \"another target\" trigger")
	}
	if !hasID(p.PickTargetCards, victim) {
		t.Errorf("the picker did not offer the other nonland permanent")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: aang}}); err == nil {
		t.Errorf("an answer naming Aang himself was accepted")
	}

	alone := newCatalogGame(t)
	castAndResolveCreature(t, alone, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", aangTheLastAirbenderOracle)
	if p := latestTriggerPromptFor(alone, alone.Seats[0].ID); p != nil {
		t.Errorf("Aang alone on the board still asked to airbend")
	}
	if p := latestPickTarget(alone, alone.Seats[0].ID); p != nil {
		t.Errorf("Aang alone on the board opened a pick_target prompt offering %v", p.PickTargetCards)
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

// "Any number of OTHER target nonland permanents": Appa is not a
// legal target of his own trigger, so the picker never offers him and
// an answer naming him is refused. Alone on the board, the trigger has
// nothing to target at all and asks nothing.
func TestAppaCannotTargetHimself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	appa := castAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", appaSteadfastGuardianOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt for Appa")
	}
	if hasID(p.PickTargetCards, appa) {
		t.Errorf("the picker offered Appa to his own \"other target\" trigger")
	}
	if !hasID(p.PickTargetCards, bear) {
		t.Errorf("the picker did not offer another nonland permanent you control")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: appa}}); err == nil {
		t.Errorf("an answer naming Appa himself was accepted")
	}

	g2 := newCatalogGame(t)
	castAndResolveCreature(t, g2, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", appaSteadfastGuardianOracle)
	if p := latestPickTarget(g2, g2.Seats[0].ID); p != nil {
		t.Errorf("Appa alone on the board opened a pick_target prompt offering %v", p.PickTargetCards)
	}
}

// The card's headline play: Appa airbends your own commander (with
// other permanents) out of harm's way. The whole group leaves; the
// commander's owner is asked about the command zone, declines, and
// every card — the commander included — is left castable for {2}.
func TestAppaAirbendsYourOwnCommanderAndKeepsItsRebuy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	commander := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: commander, Name: "My Commander", TypeLine: "Legendary Creature — Human Avatar",
		ManaCost: "{3}{W}", Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})

	castAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", appaSteadfastGuardianOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt for Appa")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
		{Kind: game.TargetCard, ID: commander},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)

	offer := latestChoiceOfKind(g, game.PendingChoiceOptionalReplacement)
	if offer == nil || offer.Chooser != me.ID {
		t.Fatalf("the commander's owner was not asked about the command zone (CR 903.9)")
	}
	if err := g.ResolveOptionalReplacement(offer.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, bear, me.ID)
	assertAirbent(t, g, commander, me.ID)
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
		Player: me.ID, CastOnly: true, Duration: game.WhileInZoneDuration(), Cost: AirbendCost,
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
