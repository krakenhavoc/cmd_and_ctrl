package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// inapp_playtest_reports_test.go — the three cards named by the
// 2026-09-20 playtest reports (#1156, #1157, #1158), each played the
// way the reporter played it.
//
// The mechanics are fixed at their choke points and pinned in the
// packages that own them (game/replacement_resume_test.go for the
// CR 614 resume boundary, game/loyalty_test.go for CR 606.3,
// client/src/lib/loyalty.test.ts for the greyed row). These are the
// cards as proof: the sequence of a real table, with the printed card.

// --- #1156: "Curiosity should go to the yard" --------------------

// TestIssue1156CuriosityFallsOffACommanderTakingTheCommandZone is the
// reporter's board: Curiosity on a commander creature, the commander
// leaves, and its owner takes CR 903.9's offer.
//
// Every other way that host can leave was already covered (#1046's
// theme deck kills it with the toughness SBA), and the difference is
// not the exit — it is that this one PAUSES on a prompt. The move
// returns with nothing moved, the resume lands the commander, and
// before #1156 nobody ran the state-based actions afterwards, so the
// Aura sat on the battlefield attached to a card that had left. The
// client draws an unattached enchantment in the enchantments row,
// which is exactly what the report says it saw.
func TestIssue1156CuriosityFallsOffACommanderTakingTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmd := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:  cmd,
		Name:        "Malcolm, Alluring Scoundrel",
		TypeLine:    "Legendary Creature — Siren Pirate",
		Power:       2,
		Toughness:   1,
		Owner:       me.ID,
		Controller:  me.ID,
		IsCommander: true,
	})
	aura := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: aura,
		Name:       "Curiosity",
		TypeLine:   "Enchantment — Aura",
		OracleID:   curiosityOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: aura, OldZone: game.ZoneStack, NewZone: game.ZoneBattlefield})
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: cmd}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, cmd); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected the CR 903.9 prompt, got %d pending choices", len(g.PendingChoices))
	}
	p := g.PendingChoices[0]
	if err := g.ResolveOptionalReplacement(p.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if !me.Command.Contains(cmd) {
		t.Fatal("the commander did not reach the command zone")
	}
	if g.Battlefield.Contains(aura) {
		t.Error("CR 704.5m: Curiosity stayed on the battlefield with nothing to enchant")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("CR 704.5m: Curiosity is not in its owner's graveyard")
	}
}

// --- #1157: "Aetherspark abilities are not functioning" ----------

// TestIssue1157AethersparkPlusOneIsLiveWithNoCreatureToAttachTo pins
// the three server-side facts the client's greyed row contradicted.
//
// The Aetherspark's +1 is "Attach The Aetherspark to up to one target
// creature you control. Put a +1/+1 counter on that creature." — min
// 0, so declining is a complete activation that still ticks the
// loyalty. On a board with nothing to attach to, the engine accepts
// the activation, internal/legal offers it to a bot, and the view
// publishes the clause WITH its min. The client's ability-row
// predicate read only "is the candidate list empty" and greyed it,
// which — with the −5 and −10 correctly greyed at four loyalty —
// is every row on the card.
func TestIssue1157AethersparkPlusOneIsLiveWithNoCreatureToAttachTo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "The Aetherspark",
		// What the deck importer stamps from Scryfall, verbatim.
		TypeLine:        "Legendary Artifact Planeswalker — Equipment",
		ManaCost:        "{4}",
		OracleID:        theAethersparkOracle,
		StartingLoyalty: 4,
		Owner:           me.ID,
		Controller:      me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, id); got != 4 {
		t.Fatalf("printed loyalty on entry: got %d, want 4", got)
	}

	// 1. The view publishes the clause, empty, with min 0 — which is
	//    everything the client needs to know it is still activatable.
	var plusOne *protocol.ActivatedAbilityView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, cv := range v.Battlefield.Cards {
		if cv.InstanceID != id.String() {
			continue
		}
		for i := range cv.ActivatedAbilities {
			if cv.ActivatedAbilities[i].Index == 0 {
				plusOne = &cv.ActivatedAbilities[i]
			}
		}
	}
	if plusOne == nil {
		t.Fatal("the view published no +1 row")
	}
	if plusOne.LegalTargets == nil {
		t.Fatal("the +1 row carries no legal_targets clause")
	}
	if n := len(plusOne.LegalTargets.Cards) + len(plusOne.LegalTargets.Players); n != 0 {
		t.Fatalf("legal targets on an empty board: got %d, want 0", n)
	}
	if plusOne.LegalTargets.Min != 0 {
		t.Errorf(`"up to one target" published min %d, want 0`, plusOne.LegalTargets.Min)
	}

	// 2. internal/legal offers it — a bot may take this activation.
	offered := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "activate_ability" && m.Source == id {
			offered = true
		}
	}
	if !offered {
		t.Error("the enumerator withheld the +1 on a board with no creature")
	}

	// 3. And the engine accepts it, with no target and no creature.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("the +1 with no target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyCount(g, id); got != 5 {
		t.Errorf("loyalty after a targetless +1: got %d, want 5", got)
	}
}

// --- #1158: "Improv capstone … off of malcolm" -------------------

// seedCapstoneTargets is the top of the library the Capstone's exile
// run walks.
func seedCapstoneTargets(me *game.Player) []uuid.UUID {
	return seedSearchLibrary(me,
		improvisationCapstoneCard("Bolt", "Instant", "{R}"),
		improvisationCapstoneCard("Wurm", "Creature — Wurm", "{5}{G}"),
		improvisationCapstoneCard("Deep", "Instant", "{R}"),
	)
}

// pitchCapstoneToMalcolm plays the reported line: Malcolm connects at
// four chorus counters, the Capstone is the card pitched to the loot,
// and the offer is taken. Returns the Capstone's instance ID.
func pitchCapstoneToMalcolm(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	mal := pushMalcolm(t, g, me.ID, 3)
	capstone := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: capstone, Name: "Improvisation Capstone",
		TypeLine: "Sorcery — Lesson", OracleID: improvisationCapstoneOracle,
		ManaCost: "{5}{R}{R}", Owner: me.ID, Controller: me.ID,
	})
	dealCombatDamageToPlayer(g, mal, opp.ID, 2)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, capstone)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(capstone) {
		t.Fatal("the Capstone is not in the graveyard")
	}
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatalf("no free cast offered at four chorus counters: %+v", g.PendingChoices)
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	return capstone
}

// TestIssue1158CapstoneResolvesOffMalcolmsFreeCast is the reported
// line end to end, and it PASSES on develop — the report's "having
// issues resolving off of malcolm" half does not reproduce. Left in
// place as the pin: the Capstone cast out of the graveyard for {0}
// exiles off the top until the running total reaches four and grants
// one cast per card that really arrived.
func TestIssue1158CapstoneResolvesOffMalcolmsFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ids := seedCapstoneTargets(me)
	capstone := pitchCapstoneToMalcolm(t, g, me, opp)

	if err := g.CastSpell(me.ID, capstone, game.CastSpellParams{
		Strict: true, FromZone: "graveyard",
	}); err != nil {
		t.Fatalf("the granted free cast: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool after a free cast: %d tokens, want 0", len(me.ManaPool))
	}
	passPriorityAroundTable(t, g)

	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventEffectError {
				t.Errorf("EventEffectError while the Capstone resolved: %s", ev.ErrorMsg)
			}
		}
	})
	// Malcolm's loot drew ids[0], so the Wurm is on top when the
	// Capstone resolves: mana value 6 clears the printed 4 on the
	// first card and the run stops there.
	if !g.Exile.Contains(ids[1]) {
		t.Error("the Capstone exiled nothing off the top of the library")
	}
	if g.Exile.Contains(ids[2]) {
		t.Error("the run did not stop on the card that carried the total over four")
	}
	perm := exiledPermission(g, ids[1])
	if perm == nil || perm.Cost != "{0}" || !perm.CastOnly {
		t.Errorf("the exiled card carries no free-cast grant: %+v", perm)
	}
	if !me.Graveyard.Contains(capstone) {
		t.Error("the Capstone did not reach the graveyard — Paradigm's self-exile is a declared gap, so the graveyard is right")
	}
	// CR 400.7: what came back to the graveyard is a new object, so
	// Malcolm's grant must not pay for it twice.
	if err := g.CastSpell(me.ID, capstone, game.CastSpellParams{
		Strict: true, FromZone: "graveyard",
	}); err == nil {
		t.Error("CR 400.7: Malcolm's grant paid for the Capstone a second time")
	}
}

// The same line where the table actually plays it. Malcolm's trigger
// resolves in the combat damage step and its grant is flash-timed
// (CR 608.2g, the card's declared simplification), so the Capstone —
// a sorcery — is cast right there rather than in a main phase.
func TestIssue1158CapstoneResolvesOffMalcolmAtFlashSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepCombatDamage {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	seedCapstoneTargets(me)
	capstone := pitchCapstoneToMalcolm(t, g, me, opp)

	if err := g.CastSpell(me.ID, capstone, game.CastSpellParams{
		Strict: true, FromZone: "graveyard",
	}); err != nil {
		t.Errorf("Malcolm's flash grant did not pay for the sorcery: %v", err)
	}
}

// TestIssue1158CapstoneCountsACardTheWindowDivertedIsTheDeclaredCAVEAT
// pins the one real defect the #1158 reproduction turned up, which is
// NOT what the report describes and is not new: MillToZone's `until`
// predicate is answered in millPlanLocked against the cards that came
// OFF the library, before the CR 614 window has said where any of
// them went. A commander whose owner takes the command zone never
// reached exile, so CR 400.7 says it is not one of the cards "you
// exile" and the run should carry on — and it stops instead, one card
// short and granting nothing.
//
// Helm of Obedience declares the same limitation in its own words and
// millPlanLocked's doc comment names it. The Capstone did not declare
// it; it does now. Pinned here so the day the mill reads the landed
// truth, this test fails and both caveats come off together.
func TestIssue1158CapstoneCountsACardTheWindowDivertedIsTheDeclaredCaveat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmd := improvisationCapstoneCard("Pirate Commander", "Legendary Creature — Pirate", "{2}{U}{R}")
	cmd.IsCommander = true
	cmd.Owner = me.ID
	ids := seedSearchLibrary(me, cmd,
		improvisationCapstoneCard("Wurm", "Creature — Wurm", "{5}{G}"))

	castCatalogSpell(t, g, "Improvisation Capstone", "Sorcery — Lesson", improvisationCapstoneOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.PendingChoices {
		if p.Kind == game.PendingChoiceOptionalReplacement {
			if err := g.ResolveOptionalReplacement(p.ID, p.Chooser, true); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
		}
	}
	passPriorityAroundTable(t, g)

	if !me.Command.Contains(ids[0]) {
		t.Fatal("the commander did not take the command zone")
	}
	// THE CAVEAT: the commander's mana value 4 satisfied the run even
	// though nothing was exiled, so the second card stays in the
	// library and no card is granted. The rules answer is the
	// opposite on both lines.
	if g.Exile.Contains(ids[1]) {
		t.Error("the run continued past a diverted card — the caveat on Improvisation Capstone is stale, remove it")
	}
	g.ReadSnapshot(func() {
		if len(me.CastPermissions) != 0 {
			t.Errorf("grants held: %d, want 0 while the caveat stands", len(me.CastPermissions))
		}
	})
}
