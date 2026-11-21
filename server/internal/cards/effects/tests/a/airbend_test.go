package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Aang, the Last Airbender ------------------------------------

// The mechanic, end to end: Aang enters, airbends an opponent's
// permanent, and that opponent — the owner — gets it back for {2}.
func TestAangAirbendsAnOpponentsPermanentAndOwnerRebuysIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	aang := effectstest.CastAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", effectstest.AangTheLastAirbenderOracle)
	if me.ID != g.Seats[g.Turn.ActiveSeat].ID {
		t.Fatalf("test assumes seat 0 is active")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertAirbent(t, g, victim, opp.ID)
	_ = aang
}

// #1304 / #1299: airbending an opposing COMMANDER. It is exiled like
// any other permanent and, since ADR 0115, the CR 903.9a state-based
// action then asks its owner about the command zone. A "no" must leave
// the commander in exile WITH the airbend grant, and the owner must then
// be able to cast it for {2}. A "yes" sends it home, a new object
// (CR 400.7), where there is nothing to grant.
func TestAirbendingACommanderKeepsTheRebuyWhenItsOwnerDeclines(t *testing.T) {
	for _, takeCommandZone := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		vivi := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: vivi, Name: "Their Commander", TypeLine: "Legendary Creature — Human Wizard",
			ManaCost: "{2}{U}{R}", Power: 3, Toughness: 3,
			Owner: opp.ID, Controller: opp.ID, IsCommander: true,
		})

		effectstest.CastAndResolveCreature(t, g, "Aang, the Last Airbender",
			"Legendary Creature — Human Avatar Ally", effectstest.AangTheLastAirbenderOracle)
		effectstest.PickCard(t, g, me.ID, vivi)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)

		if !g.Exile.Contains(vivi) {
			t.Fatalf("the airbent commander is not in exile before its owner answers")
		}
		offer := effectstest.CommanderReturnPromptFor(g, opp.ID)
		if offer == nil {
			t.Fatalf("the commander's owner was not asked about the command zone (CR 903.9a)")
		}
		if err := g.ResolveCommanderReturn(offer.ID, opp.ID, takeCommandZone); err != nil {
			t.Fatalf("ResolveCommanderReturn: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)

		if takeCommandZone {
			if !opp.Command.Contains(vivi) {
				t.Fatalf("a commander whose owner took the offer is not in the command zone")
			}
			if perm := g.CastPermissionOnCardByIDForEffect(vivi); perm != nil {
				t.Errorf("a commander back in the command zone carries an airbend grant: %+v", perm)
			}
			continue
		}
		effectstest.AssertAirbent(t, g, vivi, opp.ID)

		// The rebuy itself: on its owner's turn, {2} casts it.
		effectstest.AdvanceToPrecombatMainOf(t, g, 1)
		effectstest.B06AddMana(opp, "C", "C")
		if err := g.CastSpell(opp.ID, vivi, game.CastSpellParams{FromZone: "exile"}); err != nil {
			t.Fatalf("owner casting the airbent commander for {2}: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(vivi) {
			t.Errorf("the airbent commander did not come back for {2}")
		}
	}
}

// Declining the prompt is how "up to one" chooses zero.
func TestAangDeclinedAirbendsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	effectstest.CastAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", effectstest.AangTheLastAirbenderOracle)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	aang := effectstest.CastAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", effectstest.AangTheLastAirbenderOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt as the trigger went on the stack")
	}
	if effectstest.HasID(p.PickTargetCards, aang) {
		t.Errorf("the picker offered Aang to his own \"another target\" trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, victim) {
		t.Errorf("the picker did not offer the other nonland permanent")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: aang}}); err == nil {
		t.Errorf("an answer naming Aang himself was accepted")
	}

	alone := effectstest.NewCatalogGame(t)
	effectstest.CastAndResolveCreature(t, alone, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", effectstest.AangTheLastAirbenderOracle)
	if p := effectstest.LatestTriggerPromptFor(alone, alone.Seats[0].ID); p != nil {
		t.Errorf("Aang alone on the board still asked to airbend")
	}
	if p := effectstest.LatestPickTarget(alone, alone.Seats[0].ID); p != nil {
		t.Errorf("Aang alone on the board opened a pick_target prompt offering %v", p.PickTargetCards)
	}
}

// A land is not a legal target, so an all-lands board drops the
// trigger before it ever prompts.
func TestAangDoesNotTargetLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.AangPushLand(g, opp.ID, "Their Island", false)

	effectstest.CastAndResolveCreature(t, g, "Aang, the Last Airbender",
		"Legendary Creature — Human Avatar Ally", effectstest.AangTheLastAirbenderOracle)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(land) {
		t.Errorf("Aang airbent a land")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Errorf("a pick_target prompt opened with no legal nonland permanent")
	}
}

// --- Appa, Steadfast Guardian ------------------------------------

// Appa airbends any number of your own permanents — multi-target
// from a trigger, which nothing in the catalog had done before.
func TestAppaAirbendsSeveralOfYourOwnPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine1 := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	mine2 := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Other Bear")
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")

	effectstest.CastAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", effectstest.AppaSteadfastGuardianOracle)

	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt for Appa")
	}
	if p.PickTargetMin != 0 || p.PickTargetMax != 0 {
		t.Errorf("count = %d..%d, want 0..unbounded for \"any number\"", p.PickTargetMin, p.PickTargetMax)
	}
	if effectstest.HasID(p.PickTargetCards, theirs) {
		t.Errorf("the picker offered an opponent's permanent; the clause says YOU CONTROL")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine1},
		{Kind: game.TargetCard, ID: mine2},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertAirbent(t, g, mine1, me.ID)
	effectstest.AssertAirbent(t, g, mine2, me.ID)
	if !g.Battlefield.Contains(theirs) {
		t.Errorf("an untargeted permanent was airbent")
	}
}

// "Any number of OTHER target nonland permanents": Appa is not a
// legal target of his own trigger, so the picker never offers him and
// an answer naming him is refused. Alone on the board, the trigger has
// nothing to target at all and asks nothing.
func TestAppaCannotTargetHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	appa := effectstest.CastAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", effectstest.AppaSteadfastGuardianOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt for Appa")
	}
	if effectstest.HasID(p.PickTargetCards, appa) {
		t.Errorf("the picker offered Appa to his own \"other target\" trigger")
	}
	if !effectstest.HasID(p.PickTargetCards, bear) {
		t.Errorf("the picker did not offer another nonland permanent you control")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: appa}}); err == nil {
		t.Errorf("an answer naming Appa himself was accepted")
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.CastAndResolveCreature(t, g2, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", effectstest.AppaSteadfastGuardianOracle)
	if p := effectstest.LatestPickTarget(g2, g2.Seats[0].ID); p != nil {
		t.Errorf("Appa alone on the board opened a pick_target prompt offering %v", p.PickTargetCards)
	}
}

// The card's headline play: Appa airbends your own commander (with
// other permanents) out of harm's way. The whole group is exiled; the
// commander's owner is then asked about the command zone (CR 903.9a),
// declines, and every card — the commander included — is left castable
// for {2}.
func TestAppaAirbendsYourOwnCommanderAndKeepsItsRebuy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	commander := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: commander, Name: "My Commander", TypeLine: "Legendary Creature — Human Avatar",
		ManaCost: "{3}{W}", Power: 3, Toughness: 2,
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})

	effectstest.CastAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", effectstest.AppaSteadfastGuardianOracle)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt for Appa")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
		{Kind: game.TargetCard, ID: commander},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	offer := effectstest.CommanderReturnPromptFor(g, me.ID)
	if offer == nil {
		t.Fatalf("the commander's owner was not asked about the command zone (CR 903.9a)")
	}
	if err := g.ResolveCommanderReturn(offer.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertAirbent(t, g, bear, me.ID)
	effectstest.AssertAirbent(t, g, commander, me.ID)
}

// "Whenever you cast a spell from exile" — the EventCast source
// zone in anger. Casting from HAND must not make an Ally.
func TestAppaMakesAnAllyOnlyForCastsFromExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	appa := effectstest.CastAndResolveCreature(t, g, "Appa, Steadfast Guardian",
		"Legendary Creature — Bison Ally", effectstest.AppaSteadfastGuardianOracle)
	// Appa's own ETB trigger: nothing else is on the board, so the
	// prompt (if any) is answered with an empty pick.
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("ResolvePickTargets(none): %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.PassPriorityAroundTable(t, g)
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
		Player: me.ID, CastOnly: true, Duration: game.WhileInZoneDuration(), Cost: effects.AirbendCost,
	})
	if err := g.CastSpell(me.ID, fromExile, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gyatso := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Monk Gyatso")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == gyatso {
			g.Battlefield.Cards[i].OracleID = effectstest.MonkGyatsoOracle
		}
	}
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")

	// An opponent's Doom Blade points at the Bear.
	effectstest.AangAdvanceToMain(t, g, 1)
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: effectstest.DoomBladeOracle, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}

	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertAirbent(t, g, mine, me.ID)
	if opp.Graveyard.Contains(mine) {
		t.Errorf("the creature died; Gyatso should have exiled it first")
	}
}

// Gyatso watches creatures YOU control, and not himself.
func TestMonkGyatsoIgnoresOpponentsAndHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gyatso := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "Monk Gyatso")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == gyatso {
			g.Battlefield.Cards[i].OracleID = effectstest.MonkGyatsoOracle
		}
	}
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their Bear")

	effectstest.AangAdvanceToMain(t, g, 1)
	for _, target := range []uuid.UUID{theirs, gyatso} {
		bolt := uuid.New()
		opp.Hand.PushTop(game.Card{
			InstanceID: bolt, Name: "Doom Blade", TypeLine: "Instant",
			OracleID: effectstest.DoomBladeOracle, Owner: opp.ID, Controller: opp.ID,
		})
		if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		if p := effectstest.LatestTriggerPromptFor(g, me.ID); p != nil {
			t.Fatalf("Gyatso triggered on %v; he watches only OTHER creatures YOU control", target)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}
