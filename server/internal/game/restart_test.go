package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// restart_test.go — CR 727, restarting the game (#3856, ADR 0201). The
// engine verb is driven directly here; Karn Liberated's −14 through the
// catalog is effects/karn_liberated_test.go, and a table of bots
// playing through a restart is aiseat/restart_bot_test.go.

// ownedCardCount counts the cards `owner` has in the game: every zone
// the restart sweeps, tokens excluded.
func ownedCardCount(g *Game, owner uuid.UUID) int {
	n := 0
	count := func(z *Zone) {
		if z == nil {
			return
		}
		for _, c := range z.Cards {
			if c.Owner == owner && !c.IsToken() {
				n++
			}
		}
	}
	for _, p := range g.Seats {
		count(p.Library)
		count(p.Hand)
		count(p.Graveyard)
		count(p.Command)
	}
	count(g.Battlefield)
	count(g.Stack)
	count(g.Exile)
	count(g.PhasedOut)
	return n
}

// allInstanceIDs lists every card instance in the game's zones.
func allInstanceIDs(g *Game) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	add := func(z *Zone) {
		if z == nil {
			return
		}
		for _, c := range z.Cards {
			out[c.InstanceID] = true
		}
	}
	for _, p := range g.Seats {
		add(p.Library)
		add(p.Hand)
		add(p.Graveyard)
		add(p.Command)
	}
	add(g.Battlefield)
	add(g.Stack)
	add(g.Exile)
	add(g.PhasedOut)
	return out
}

// restartTable is a four-player game in seat 0's main phase with an
// old game's worth of state on it: life lost, poison, commander damage
// and tax, a monarch, a token, a tapped permanent with counters, cards
// in a graveyard and in exile, and seat 3 out of the game. It returns
// the game and the exiled creature card the restart will exempt.
func restartTable(t *testing.T) (*Game, uuid.UUID) {
	t.Helper()
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p0, p1 := g.Seats[0], g.Seats[1]
	if err := g.Concede(g.Seats[3].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	exiled := NewCard("Exiled Bear", p1.ID)
	exiled.TypeLine = "Creature — Bear"
	exiled.Power, exiled.Toughness = 2, 2
	g.WithWriteLock(func() {
		p0.Life = 12
		p0.Poison = 3
		p1.CommanderDamage[p0.Command.Cards[0].InstanceID] = 9
		p0.CommanderCasts[p0.Command.Cards[0].InstanceID] = 2
		g.Monarch = p1.ID
		tok := NewCard("Goblin", p0.ID)
		tok.TypeLine = "Token Creature — Goblin"
		g.Battlefield.PushTop(tok)
		perm := NewCard("Tapped Rock", p0.ID)
		perm.TypeLine = "Artifact"
		perm.Tapped = true
		perm.Counters = map[string]int{"charge": 4}
		g.Battlefield.PushTop(perm)
		dead, _ := p1.Library.PopTop()
		p1.Graveyard.PushTop(dead)
		gone, _ := p0.Library.PopTop()
		g.Exile.PushTop(gone)
		g.Exile.PushTop(exiled)
	})
	return g, exiled.InstanceID
}

// The whole of CR 727 from the engine's side: the old game ends, the
// players still in it begin a new one with every card they own and
// nothing else, the restarter goes first, the exempted card waits in
// exile through the mulligans and then enters under the restarter's
// control before the first turn's untap step.
func TestRestartBeginsANewGameWithTheRestarterFirst(t *testing.T) {
	g, exempt := restartTable(t)
	restarter := g.Seats[2]
	owned := map[uuid.UUID]int{}
	for _, p := range g.Seats[:3] {
		owned[p.ID] = ownedCardCount(g, p.ID)
	}
	owned[g.Seats[1].ID]-- // the exempted card is not in the deck
	before := allInstanceIDs(g)
	seq := lastSeq(g)

	g.WithWriteLock(func() {
		g.RestartTheGameForEffect(restarter.ID, uuid.Nil, []uuid.UUID{exempt})
	})
	if g.RestartPending == nil || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("the request did not park the game: pending %v, priority %d", g.RestartPending, g.Turn.PriorityHolder)
	}
	g.SettleResolution()

	if g.RestartPending != nil {
		t.Fatal("RestartPending was left set")
	}
	if g.State != StateActive || !g.MulligansOpen || g.Outcome != nil {
		t.Fatalf("state %s mulligans %v outcome %v, want a new active game in its mulligan window", g.State, g.MulligansOpen, g.Outcome)
	}
	if g.StartingSeat != 2 || g.Turn.ActiveSeat != 2 || g.Turn.Seq != 1 || g.Turn.Step != StepUntap {
		t.Errorf("new game starts at seat %d (active %d) turn %d %s, want seat 2's turn 1 untap (CR 727.1a)",
			g.StartingSeat, g.Turn.ActiveSeat, g.Turn.Seq, g.Turn.Step)
	}
	if g.MulliganDecider() != 2 {
		t.Errorf("first mulligan decision is seat %d's, want the starting player's", g.MulliganDecider())
	}
	if g.Monarch != uuid.Nil {
		t.Error("the monarch survived the restart")
	}
	for id := range allInstanceIDs(g) {
		if before[id] {
			t.Errorf("card %s kept its old-game instance ID", id)
		}
	}
	for _, p := range g.Seats[:3] {
		if p.Life != g.Settings.StartingLife || p.Poison != 0 || len(p.CommanderDamage) != 0 || len(p.CommanderCasts) != 0 {
			t.Errorf("%s carried old-game state: life %d poison %d cmdr dmg %v tax %v",
				p.Name, p.Life, p.Poison, p.CommanderDamage, p.CommanderCasts)
		}
		if p.Hand.Size() != OpeningHandSize || p.Graveyard.Size() != 0 {
			t.Errorf("%s: hand %d graveyard %d, want a fresh seven and no graveyard", p.Name, p.Hand.Size(), p.Graveyard.Size())
		}
		if p.Command.Size() != 1 || !p.Command.Cards[0].IsCommander {
			t.Errorf("%s: command zone %d cards, want their commander (CR 103.2c)", p.Name, p.Command.Size())
		}
		if got := p.Library.Size() + p.Hand.Size() + p.Command.Size(); got != owned[p.ID] {
			t.Errorf("%s has %d cards, want every card they owned: %d (CR 727.2)", p.Name, got, owned[p.ID])
		}
		for _, c := range p.Library.Cards {
			if c.Tapped || len(c.Counters) != 0 || c.Controller != p.ID {
				t.Errorf("%s came back with old-game state: %+v", c.Name, c)
			}
		}
	}
	if out := g.Seats[3]; !out.Eliminated || out.Hand.Size() != 0 {
		t.Errorf("the player who had left was dealt back in (eliminated %v, hand %d)", out.Eliminated, out.Hand.Size())
	}
	if g.Battlefield.Size() != 0 {
		t.Errorf("battlefield holds %d objects before the mulligans end, want none", g.Battlefield.Size())
	}
	if g.Exile.Size() != 1 {
		t.Fatalf("exile holds %d cards, want only the exempted one (CR 727.5)", g.Exile.Size())
	}
	waiting := g.Exile.Cards[0]
	if waiting.Name != "Exiled Bear" || g.RestartTail == nil || len(g.RestartTail.Exempt) != 1 || g.RestartTail.Exempt[0] != waiting.InstanceID {
		t.Fatalf("the exempted card is not the tail's: exile %s, tail %+v", waiting.Name, g.RestartTail)
	}
	if ev := eventsSince(g, EventGameRestarted, seq); len(ev) != 1 || ev[0].Actor != restarter.ID || ev[0].Amount != 1 {
		t.Errorf("EventGameRestarted = %+v, want one naming the restarter and one card", ev)
	}

	keepEveryHand(t, g)
	if g.RestartTail != nil {
		t.Error("RestartTail was left set after the mulligans")
	}
	bear := restartCardNamed(g.Battlefield, "Exiled Bear")
	if bear == nil {
		t.Fatal("the exempted card did not enter the battlefield (CR 727.4)")
	}
	if bear.Controller != restarter.ID || bear.Owner != g.Seats[1].ID {
		t.Errorf("the bear entered under %s owned by %s, want the restarter's control and its owner unchanged", bear.Controller, bear.Owner)
	}
	if bear.SummonedThisTurn {
		t.Error("the bear is summoning sick; it entered before the restarter's first turn began")
	}
	if g.Turn.ActiveSeat != 2 || g.Turn.Step != StepUpkeep || g.Turn.PriorityHolder != 2 {
		t.Errorf("first turn at seat %d %s priority %d, want seat 2's upkeep with priority", g.Turn.ActiveSeat, g.Turn.Step, g.Turn.PriorityHolder)
	}
}

// keepEveryHand keeps each opening hand in the order the mulligan
// window asks for them, starting with the starting player.
func keepEveryHand(t *testing.T, g *Game) {
	t.Helper()
	for range len(g.Seats) {
		seat := g.MulliganDecider()
		if seat < 0 {
			return
		}
		if err := g.KeepHand(g.Seats[seat].ID); err != nil {
			t.Fatalf("KeepHand seat %d: %v", seat, err)
		}
	}
}

func restartCardNamed(z *Zone, name string) *Card {
	for i := range z.Cards {
		if z.Cards[i].Name == name {
			return &z.Cards[i]
		}
	}
	return nil
}

const restartStubOracle = "stub-restart-the-game"

// The restart from inside a real resolution, with a spell beneath the
// restarting one: the resolution finishes, the boundary begins the new
// game instead of handing priority back, and the spell beneath is a
// card in its owner's library rather than anything still on a stack.
// An undo taken before the cast restores the old game whole (the room's
// undo is exactly this RestoreFrom).
func TestRestartFromAResolutionAndUndoAcrossIt(t *testing.T) {
	withEffectHooks(t,
		func(g *Game, item *StackItem, oracleID string) error {
			if oracleID == restartStubOracle {
				g.RestartTheGameForEffect(item.Controller, item.SourceCardID, nil)
			}
			return nil
		},
		nil,
		func(oracleID string) bool { return oracleID == restartStubOracle },
	)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	active := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { active.Life = 7 })
	undo := g.Clone()
	turnBefore := g.Turn

	under := pushTypedCardToHand(active, "Lesser Instant", "Instant")
	if err := g.CastSpell(active.ID, under, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	stub := pushTypedCardToHand(active, "Restart Stub", "Instant")
	g.WithWriteLock(func() {
		for i := range active.Hand.Cards {
			if active.Hand.Cards[i].InstanceID == stub {
				active.Hand.Cards[i].OracleID = restartStubOracle
			}
		}
	})
	if err := g.CastSpell(active.ID, stub, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for range 8 {
		if g.MulligansOpen {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.MulligansOpen || g.Turn.Seq != 1 || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("no new game: mulligans %v turn %d priority %d", g.MulligansOpen, g.Turn.Seq, g.Turn.PriorityHolder)
	}
	if len(g.Stack.Cards) != 0 || len(g.StackMeta) != 0 {
		t.Errorf("the old game's stack survived: %d cards, %d items", len(g.Stack.Cards), len(g.StackMeta))
	}
	// The seat is a new Player in the new game, as it is after an undo:
	// read it off the game again.
	now := g.PlayerByID(active.ID)
	if now == nil || now == active {
		t.Fatalf("the restarted seat is the old game's Player")
	}
	if restartCardNamed(now.Library, "Lesser Instant") == nil && restartCardNamed(now.Hand, "Lesser Instant") == nil {
		t.Error("the spell beneath is not back in its owner's deck")
	}
	if now.Life != g.Settings.StartingLife {
		t.Errorf("life %d, want the starting %d", now.Life, g.Settings.StartingLife)
	}

	g.WithWriteLock(func() { g.RestoreFrom(undo) })
	if back := g.PlayerByID(active.ID); g.MulligansOpen || g.Turn != turnBefore || back.Life != 7 {
		t.Errorf("the undo did not restore the old game: mulligans %v turn %+v life %d", g.MulligansOpen, g.Turn, back.Life)
	}
}

// A restore point written while the new game waits on its mulligans
// carries the tail, so the exempted card still enters after a server
// restart.
func TestRestartTailSurvivesASnapshot(t *testing.T) {
	g, exempt := restartTable(t)
	restarter := g.Seats[2]
	g.WithWriteLock(func() {
		g.RestartTheGameForEffect(restarter.ID, uuid.Nil, []uuid.UUID{exempt})
	})
	g.SettleResolution()
	if g.RestartTail == nil {
		t.Fatal("no tail to persist")
	}
	_, restored := roundTrip(t, g)
	if restored.RestartTail == nil || len(restored.RestartTail.Exempt) != 1 {
		t.Fatalf("the snapshot dropped the tail: %+v", restored.RestartTail)
	}
	keepEveryHand(t, restored)
	if bear := restartCardNamed(restored.Battlefield, "Exiled Bear"); bear == nil || bear.Controller != restarter.ID {
		t.Errorf("the restored game did not finish the restart: %+v", bear)
	}

	// And the pending half, written between the resolution and its
	// boundary, round-trips too.
	g2, exempt2 := restartTable(t)
	g2.WithWriteLock(func() {
		g2.RestartTheGameForEffect(g2.Seats[1].ID, uuid.Nil, []uuid.UUID{exempt2})
	})
	raw, err := json.Marshal(g2.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	if snap.RestartPending == nil || snap.RestartPending.Restarter != g2.Seats[1].ID {
		t.Errorf("the snapshot dropped the pending restart: %+v", snap.RestartPending)
	}
	if g2.Clone().RestartPending == nil {
		t.Error("Clone dropped RestartPending")
	}
}

// CR 727.3: a player whose deck cannot supply seven cards loses at the
// first state-based check of the new game, however they mulligan.
func TestRestartWithAShortDeckLosesAtTheFirstCheck(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	short := g.Seats[1]
	var exempt []uuid.UUID
	g.WithWriteLock(func() {
		// Every card of seat 1's library but three is exiled and
		// exempted, as if one Karn had eaten their deck.
		for short.Library.Size() > 3 {
			c, _ := short.Library.PopTop()
			c.TypeLine = "Artifact"
			g.Exile.PushTop(c)
			exempt = append(exempt, c.InstanceID)
		}
		short.Hand.Cards = nil
		g.RestartTheGameForEffect(g.Seats[0].ID, uuid.Nil, exempt)
	})
	g.SettleResolution()
	if !g.Seats[1].AttemptedEmptyDraw {
		t.Fatal("the short deck was not marked for CR 704.5b")
	}
	keepEveryHand(t, g)
	if g.Seats[1].Eliminated {
		t.Fatal("the player lost before the first turn began")
	}
	// The new game's first state-based check, in its first upkeep.
	if g.Turn.Step != StepUpkeep {
		t.Fatalf("the first turn is at %s, want its upkeep", g.Turn.Step)
	}
	runChecks(g)
	if !g.Seats[1].Eliminated {
		t.Error("the player who could not draw seven is still in the game (CR 727.3)")
	}
}

// Every Player field is either carried into the new game (the seat's
// identity and its own settings) or left at its newPlayer value (a fact
// about the old game). A field added to Player later must be put on one
// side, here and in restartedPlayer, or this fails.
func TestRestartClassifiesEveryPlayerField(t *testing.T) {
	carried := map[string]bool{
		"ID": true, "Name": true, "Seat": true, "DeckImported": true,
		"DiscordID": true, "DiscordAvatarHash": true, "DisplayName": true,
		"IsBot": true, "BotTier": true, "BotDeck": true, "Agent": true, "AgentClient": true,
		"TriggerOrder": true, "AutoAnswers": true,
		"Companion": true, "Sideboard": true, "HasSideboard": true,
	}
	reset := map[string]bool{
		"Life": true, "Poison": true, "Energy": true, "Speed": true,
		"Library": true, "Hand": true, "Graveyard": true, "Command": true, "Emblems": true,
		"CommanderDamage": true, "LifeHistory": true, "TurnsBegun": true, "UpkeepsBegun": true,
		"EndStepTurn": true, "LastTurnAttacks": true, "Eliminated": true, "Removed": true,
		"HandKept": true, "MulliganDecided": true, "TriggerOrderMemory": true, "MulligansTaken": true,
		"mulliganBottoming": true, "UndosRemaining": true, "AttemptedEmptyDraw": true,
		"CommanderCasts": true, "Counters": true, "MaxHandSize": true, "MaxHandSizeAt": true,
		"LandDropsPerTurn": true, "ManaPool": true, "CastPermissions": true, "Statics": true,
		"CitysBlessing": true, "Dungeon": true, "CompletedDungeons": true,
	}
	typ := reflect.TypeOf(Player{})
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if carried[name] == reset[name] {
			t.Errorf("Player.%s is not classified for a restart (carried %v, reset %v)", name, carried[name], reset[name])
		}
	}

	// And restartedPlayer does what the table says.
	p := newPlayer("Ann", 2)
	p.IsBot, p.BotTier, p.DisplayName, p.DeckImported = true, "strong", "Ann B", true
	p.TriggerOrder = TriggerOrderMode("auto")
	p.Life, p.Poison, p.TurnsBegun = 3, 9, 7
	np := restartedPlayer(p)
	if np.ID != p.ID || np.Seat != 2 || !np.IsBot || np.BotTier != "strong" || np.DisplayName != "Ann B" || !np.DeckImported || np.TriggerOrder != p.TriggerOrder {
		t.Errorf("restartedPlayer dropped the seat's identity: %+v", np)
	}
	if np.Poison != 0 || np.TurnsBegun != 0 || np.Library == p.Library {
		t.Errorf("restartedPlayer kept old-game state: %+v", np)
	}
}
