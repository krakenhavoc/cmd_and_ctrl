package legal_test

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // catalog hooks
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// Catalog oracle IDs used by the scenarios below. Each is registered
// by the effects package and has the structured spec the enumerator
// needs (target clause, additional cost, activated / mana ability).
const (
	oracleLightningBolt     = "4457ed35-7c10-48c8-9776-456485fdf070"
	oracleVillageRites      = "365548fb-5acc-4a8a-b20b-26d28b7d029f"
	oracleGoblinBombardment = "edad60c6-80de-4033-af1b-a703ac332983"
	oracleLlanowarElves     = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	oraclePreordain         = "ac641490-ca14-48d7-8cc4-b69ce984befa"
	oracleCounterspell      = "cc187110-1148-4090-bbb8-e205694a39f5"
)

// --- wire compatibility ------------------------------------------

func TestWireTypesMatchActions(t *testing.T) {
	pairs := map[string]actions.Type{
		legal.TypePassPriority:        actions.TypePassPriority,
		legal.TypeCastSpell:           actions.TypeCastSpell,
		legal.TypeActivateAbility:     actions.TypeActivateAbility,
		legal.TypeActivateManaAbility: actions.TypeActivateManaAbility,
		legal.TypeDeclareAttacker:     actions.TypeDeclareAttacker,
		legal.TypeDeclareBlocker:      actions.TypeDeclareBlocker,
		legal.TypeResolveChoice:       actions.TypeResolveChoice,
		legal.TypeKeepHand:            actions.TypeKeepHand,
		legal.TypeMulligan:            actions.TypeMulligan,
		legal.TypeDiscardSelection:    actions.TypeDiscardSelection,
	}
	for got, want := range pairs {
		if got != string(want) {
			t.Errorf("legal type %q != actions type %q", got, want)
		}
	}
}

// --- scenario helpers --------------------------------------------

// newTable builds a four-seat game with filler decks, starts it, and
// closes the mulligan window. Seat 0 is the active player on turn 1.
func newTable(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 0, 21)
		deck = append(deck, game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil))
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 9))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// newTableMulligans is newTable with the mulligan window still open.
func newTableMulligans(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 0, 21)
		deck = append(deck, game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil))
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

// clearHand empties a seat's hand so scenarios control exactly what
// is castable.
func clearHand(p *game.Player) { p.Hand.Cards = nil }

func handCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner = p.ID
	c.Controller = p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

func battlefieldCard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner = p.ID
	c.Controller = p.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func basic(name, sub string) game.Card {
	return game.Card{Name: name, TypeLine: "Basic Land — " + sub}
}

func creature(name string, cost string, p, tgh int) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Beast", ManaCost: cost, Power: p, Toughness: tgh}
}

func advanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 40; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to %s: %v", step, err)
		}
	}
	t.Fatalf("never reached step %s (at %s)", step, g.Turn.Step)
}

// dispatchAll is the soundness check: every enumerated move for seat
// must be accepted by the dispatcher when sent by that seat, each
// against a fresh clone so moves do not interfere.
func dispatchAll(t *testing.T, g *game.Game, seat uuid.UUID, moves []legal.Move) {
	t.Helper()
	for _, m := range moves {
		clone := g.Clone()
		a := actions.Action{
			Type:   actions.Type(m.Type),
			Player: m.Player,
			Caller: seat,
			Params: m.Params,
		}
		if err := actions.Dispatch(clone, a); err != nil {
			t.Errorf("move %q (%s %s) rejected: %v", m.Label, m.Type, string(m.Params), err)
		}
	}
}

func labels(moves []legal.Move) []string {
	out := make([]string, 0, len(moves))
	for _, m := range moves {
		out = append(out, m.Label)
	}
	return out
}

func hasLabel(moves []legal.Move, prefix string) bool {
	for _, m := range moves {
		if strings.HasPrefix(m.Label, prefix) {
			return true
		}
	}
	return false
}

func countKind(moves []legal.Move, k legal.Kind) int {
	n := 0
	for _, m := range moves {
		if m.Kind == k {
			n++
		}
	}
	return n
}

// --- mulligan window ---------------------------------------------

func TestMulliganWindow(t *testing.T) {
	g := newTableMulligans(t)
	for _, p := range g.Seats {
		moves := legal.EnumerateFor(g, p.ID)
		if countKind(moves, legal.KindMulligan) != 2 {
			t.Fatalf("%s: want keep + mulligan, got %v", p.Name, labels(moves))
		}
		if !hasLabel(moves, "Mulligan to 7") {
			t.Errorf("%s: first mulligan should be free (to 7): %v", p.Name, labels(moves))
		}
		dispatchAll(t, g, p.ID, moves)
	}
	// After one mulligan the next offer is a card fewer.
	p := g.Seats[0]
	if err := g.Mulligan(p.ID, 7); err != nil {
		t.Fatal(err)
	}
	moves := legal.EnumerateFor(g, p.ID)
	if !hasLabel(moves, "Mulligan to 6") {
		t.Errorf("second mulligan should go to 6: %v", labels(moves))
	}
	// A seat that has kept has nothing left to do in the window.
	if err := g.KeepHand(p.ID); err != nil {
		t.Fatal(err)
	}
	if moves := legal.EnumerateFor(g, p.ID); len(moves) != 0 {
		t.Errorf("kept seat should have no moves while mulligans open: %v", labels(moves))
	}
}

// --- main phase: lands, casts, affordability ---------------------

func TestMainPhaseLandAndCasts(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	handCard(active, basic("Mountain", "Mountain"))
	handCard(active, basic("Forest", "Forest"))
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	bear := handCard(active, creature("Grizzly Bears", "{1}{G}", 2, 2))
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	battlefieldCard(g, opp, creature("Opposing Bear", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	if !hasLabel(moves, "Pass priority") {
		t.Errorf("active seat in main phase must be able to pass: %v", labels(moves))
	}
	if countKind(moves, legal.KindLand) != 2 {
		t.Errorf("want two land drops offered, got %v", labels(moves))
	}
	// Bolt: one Mountain on the battlefield pays {R}; targets are the
	// four players plus the opposing bear.
	bolts := 0
	for _, m := range moves {
		if m.Source == bolt {
			bolts++
		}
	}
	if bolts != 5 {
		t.Errorf("want Bolt expanded to 5 targets (4 players + 1 creature), got %d: %v", bolts, labels(moves))
	}
	// Bears: {1}{G} with only one Mountain is unaffordable.
	for _, m := range moves {
		if m.Source == bear {
			t.Errorf("Grizzly Bears should be unaffordable with one Mountain: %q", m.Label)
		}
	}

	// Play the Forest; now Bears is affordable and no second land is
	// offered.
	var forestMove *legal.Move
	for i := range moves {
		if moves[i].Label == "Play Forest" {
			forestMove = &moves[i]
		}
	}
	if forestMove == nil {
		t.Fatal("no Play Forest move")
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(forestMove.Type), Player: active.ID, Caller: active.ID, Params: forestMove.Params}); err != nil {
		t.Fatalf("play forest: %v", err)
	}
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	if countKind(moves, legal.KindLand) != 0 {
		t.Errorf("second land drop must not be offered: %v", labels(moves))
	}
	found := false
	for _, m := range moves {
		if m.Source == bear {
			found = true
		}
	}
	if !found {
		t.Errorf("Grizzly Bears should be castable with Mountain + Forest: %v", labels(moves))
	}
}

// TestUnparseableCostIsNotEnumerated covers the enumerator half of
// #289. A split card imports a joined cost ("{1}{R} // {1}{U}") that
// ParseCost rejects. The enumerator used to mirror the engine's old
// silent downgrade and offer the cast as FREE; the engine now
// rejects it with ErrUnparseableCost, so offering the move would
// hand the client an action guaranteed to fail.
func TestUnparseableCostIsNotEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	split := handCard(active, game.Card{
		Name:     "Fire // Ice",
		TypeLine: "Instant",
		ManaCost: "{1}{R} // {1}{U}",
	})
	bolt := handCard(active, game.Card{
		Name:     "Lightning Bolt",
		TypeLine: "Instant",
		ManaCost: "{R}",
		OracleID: oracleLightningBolt,
	})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	for _, m := range moves {
		if m.Source == split {
			t.Errorf("split card with an unparseable cost must not be enumerated: %q", m.Label)
		}
	}
	// The sibling card with a readable cost is unaffected — this is
	// a targeted refusal, not a blanket one.
	seenBolt := false
	for _, m := range moves {
		if m.Source == bolt {
			seenBolt = true
		}
	}
	if !seenBolt {
		t.Errorf("Lightning Bolt should still be enumerated: %v", labels(moves))
	}
}

func TestOpponentPriorityOnlyInstants(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(opp)
	handCard(opp, basic("Mountain", "Mountain"))
	handCard(opp, creature("Grizzly Bears", "{1}{G}", 2, 2))
	bolt := handCard(opp, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, opp, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	// Opponent does not hold priority yet: nothing at all.
	if moves := legal.EnumerateFor(g, opp.ID); len(moves) != 0 {
		t.Fatalf("non-holder should have no moves in main phase: %v", labels(moves))
	}
	// Active passes → opponent holds priority.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	_ = active
	moves := legal.EnumerateFor(g, opp.ID)
	dispatchAll(t, g, opp.ID, moves)
	if countKind(moves, legal.KindLand) != 0 {
		t.Errorf("opponent may not play a land on someone else's turn: %v", labels(moves))
	}
	if hasLabel(moves, "Cast Grizzly Bears") {
		t.Errorf("sorcery-speed cast offered off-turn: %v", labels(moves))
	}
	bolts := 0
	for _, m := range moves {
		if m.Source == bolt {
			bolts++
		}
	}
	if bolts == 0 {
		t.Errorf("instant should be castable when holding priority off-turn: %v", labels(moves))
	}
}

func TestNoPriorityStepsHaveNoMoves(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	// Turn 1 starts at upkeep after mulligans close (untap auto-skips).
	// Force the cursor onto cleanup via a fresh, unwound cursor: the
	// easiest no-priority state to reach is the untap step of the next
	// turn, which the engine passes through immediately; instead check
	// the invariant directly on the enumerator's view of a cursor with
	// no holder.
	g.Turn.PriorityHolder = game.NoPriority
	if moves := legal.EnumerateFor(g, active.ID); len(moves) != 0 {
		t.Errorf("no holder → no moves, got %v", labels(moves))
	}
}

// --- split second -------------------------------------------------

func TestSplitSecondBlocksCastsNotPass(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)
	g.SplitSecondActive = true
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	if countKind(moves, legal.KindCast) != 0 {
		t.Errorf("split second must suppress casts: %v", labels(moves))
	}
	if !hasLabel(moves, "Pass priority") {
		t.Errorf("pass must survive split second: %v", labels(moves))
	}
}

// --- additional costs ---------------------------------------------

func TestVillageRitesNeedsACreatureToSacrifice(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	rites := handCard(active, game.Card{Name: "Village Rites", TypeLine: "Instant", ManaCost: "{B}", OracleID: oracleVillageRites})
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	for _, m := range moves {
		if m.Source == rites {
			t.Errorf("Village Rites with no creature must not be offered: %q", m.Label)
		}
	}
	a := battlefieldCard(g, active, creature("Fodder A", "{G}", 1, 1))
	b := battlefieldCard(g, active, creature("Fodder B", "{G}", 1, 1))
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	sacs := map[string]bool{}
	for _, m := range moves {
		if m.Source != rites {
			continue
		}
		var p struct {
			SacrificeIDs []string `json:"sacrifice_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || len(p.SacrificeIDs) != 1 {
			t.Fatalf("bad Village Rites params: %s", string(m.Params))
		}
		sacs[p.SacrificeIDs[0]] = true
	}
	if !sacs[a.String()] || !sacs[b.String()] || len(sacs) != 2 {
		t.Errorf("want one Village Rites cast per creature, got %v", sacs)
	}
}

// --- activated + mana abilities -----------------------------------

func TestGoblinBombardmentSacrificeOutlet(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	bomb := battlefieldCard(g, active, game.Card{Name: "Goblin Bombardment", TypeLine: "Enchantment", OracleID: oracleGoblinBombardment})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	for _, m := range moves {
		if m.Source == bomb {
			t.Errorf("no creature to sacrifice → no activation: %q", m.Label)
		}
	}
	battlefieldCard(g, active, creature("Goblin", "{R}", 1, 1))
	battlefieldCard(g, opp, creature("Target Bear", "{1}{G}", 2, 2))
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	n := 0
	for _, m := range moves {
		if m.Source == bomb && m.Kind == legal.KindActivate {
			n++
		}
	}
	// One fodder × (4 players + 2 creatures) targets.
	if n != 6 {
		t.Errorf("want 6 Bombardment activations, got %d: %v", n, labels(moves))
	}
}

func TestLlanowarElvesSummoningSick(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	elves := battlefieldCard(g, active, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Power: 1, Toughness: 1, OracleID: oracleLlanowarElves, SummonedThisTurn: true})
	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, active.ID)
	for _, m := range moves {
		if m.Source == elves {
			t.Errorf("summoning-sick dork must not offer its mana ability: %q", m.Label)
		}
	}
	// Not sick → offered, and dispatchable.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == elves {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	}
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	if countKind(moves, legal.KindMana) != 1 {
		t.Errorf("want the Elves mana ability once, got %v", labels(moves))
	}
}

// --- combat -------------------------------------------------------

func TestDeclareAttackersAndBlockers(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%4]
	elim := g.Seats[(g.Turn.ActiveSeat+2)%4]
	clearHand(active)
	ready := battlefieldCard(g, active, creature("Ready", "{G}", 2, 2))
	battlefieldCard(g, active, game.Card{Name: "Tapped", TypeLine: "Creature — Beast", Power: 2, Toughness: 2, Tapped: true})
	battlefieldCard(g, active, game.Card{Name: "Sick", TypeLine: "Creature — Beast", Power: 2, Toughness: 2, SummonedThisTurn: true})
	battlefieldCard(g, active, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	blocker := battlefieldCard(g, def, creature("Blocker", "{W}", 1, 3))
	battlefieldCard(g, def, game.Card{Name: "Tapped Blocker", TypeLine: "Creature — Beast", Power: 1, Toughness: 1, Tapped: true})
	if err := g.Concede(elim.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	attacks := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack {
			continue
		}
		attacks++
		if m.Source != ready {
			t.Errorf("only the untapped, unsick, non-defender creature may attack: %q", m.Label)
		}
		if strings.Contains(m.Label, elim.Name) {
			t.Errorf("eliminated seat offered as an attack target: %q", m.Label)
		}
	}
	// Two live opponents (one conceded, and never yourself).
	if attacks != 2 {
		t.Errorf("want 2 attack declarations, got %d: %v", attacks, labels(moves))
	}
	// A non-active seat has nothing to declare here.
	if moves := legal.EnumerateFor(g, def.ID); len(moves) != 0 {
		t.Errorf("defender should have no moves during declare attackers: %v", labels(moves))
	}

	// Attack the defender, move to blockers.
	if err := g.DeclareAttacker(ready, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	moves = legal.EnumerateFor(g, def.ID)
	dispatchAll(t, g, def.ID, moves)
	blocks := 0
	for _, m := range moves {
		if m.Kind == legal.KindBlock {
			blocks++
			if m.Source != blocker {
				t.Errorf("tapped creature offered as blocker: %q", m.Label)
			}
		}
	}
	if blocks != 1 {
		t.Errorf("want exactly one block, got %d: %v", blocks, labels(moves))
	}
	// A seat that is not being attacked has no blocks to declare.
	other := g.Seats[(g.Turn.ActiveSeat+3)%4]
	battlefieldCard(g, other, creature("Bystander", "{W}", 1, 1))
	for _, m := range legal.EnumerateFor(g, other.ID) {
		if m.Kind == legal.KindBlock {
			t.Errorf("seat not under attack offered a block: %q", m.Label)
		}
	}
}

// --- pending choices ----------------------------------------------

func TestPreordainScryChoices(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	pre := handCard(active, game.Card{Name: "Preordain", TypeLine: "Sorcery", ManaCost: "{U}", OracleID: oraclePreordain})
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	var cast *legal.Move
	for i := range moves {
		if moves[i].Source == pre {
			cast = &moves[i]
		}
	}
	if cast == nil {
		t.Fatalf("Preordain not offered: %v", labels(moves))
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(cast.Type), Player: active.ID, Caller: active.ID, Params: cast.Params}); err != nil {
		t.Fatal(err)
	}
	// Resolve: pass around the table until the scry prompt appears.
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.PendingChoices) == 0 {
		t.Fatal("expected a scry prompt")
	}
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	// Keep all, bottom all, bottom each of the two: four answers, and
	// nothing else while a choice is owed.
	if countKind(moves, legal.KindChoice) != 4 || len(moves) != 4 {
		t.Errorf("want exactly 4 scry answers, got %v", labels(moves))
	}
	// Other seats are blocked entirely while a choice is open.
	for _, p := range g.Seats {
		if p.ID == active.ID {
			continue
		}
		if m := legal.EnumerateFor(g, p.ID); len(m) != 0 {
			t.Errorf("%s should have no moves while %s owes a choice: %v", p.Name, active.Name, labels(m))
		}
	}
}

func TestCounterspellOnlyWithASpellOnTheStack(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	clearHand(opp)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	cs := handCard(opp, game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", OracleID: oracleCounterspell})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	battlefieldCard(g, opp, basic("Island", "Island"))
	battlefieldCard(g, opp, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	// Empty stack, opponent holds priority: Counterspell has no target.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	for _, m := range legal.EnumerateFor(g, opp.ID) {
		if m.Source == cs {
			t.Errorf("Counterspell offered with an empty stack: %q", m.Label)
		}
	}
	// Back to the active player, who bolts a face; now it is a target.
	for i := 0; i < 4 && !holdsPriority(g, active.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(active.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true, AutoTap: true}); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	moves := legal.EnumerateFor(g, opp.ID)
	dispatchAll(t, g, opp.ID, moves)
	n := 0
	for _, m := range moves {
		if m.Source == cs {
			n++
		}
	}
	if n != 1 {
		t.Errorf("want Counterspell offered once against the Bolt, got %d: %v", n, labels(moves))
	}
}

func holdsPriority(g *game.Game, id uuid.UUID) bool {
	ph := g.Turn.PriorityHolder
	return ph >= 0 && ph < len(g.Seats) && g.Seats[ph].ID == id
}

// --- lobby / ended / unknown seat ---------------------------------

func TestInactiveGameAndUnknownSeat(t *testing.T) {
	g := game.NewGame()
	if m := legal.EnumerateFor(g, uuid.New()); m != nil {
		t.Errorf("lobby game should enumerate nil, got %v", labels(m))
	}
	g = newTable(t)
	if m := legal.EnumerateFor(g, uuid.New()); m != nil {
		t.Errorf("unknown seat should enumerate nil, got %v", labels(m))
	}
	if m := legal.EnumerateFor(nil, uuid.New()); m != nil {
		t.Errorf("nil game should enumerate nil")
	}
}

// --- modal spells -------------------------------------------------

const (
	oracleIzzetCharm     = "a07698f6-5ad5-49a3-9da2-f82d407f5cd7"
	oracleAustereCommand = "09cc8709-fe10-472a-b05c-e89f3523018d"
)

func TestModalSpellsExpandPerModeAndTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	charm := handCard(active, game.Card{Name: "Izzet Charm", TypeLine: "Instant", ManaCost: "{U}{R}", OracleID: oracleIzzetCharm})
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	battlefieldCard(g, opp, creature("Bear One", "{1}{G}", 2, 2))
	battlefieldCard(g, opp, creature("Bear Two", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	// Empty stack: mode 0 (counter target spell) has no target, so it
	// is not offered; mode 1 expands to the two bears; mode 2 is
	// untargeted. Three casts.
	modes := map[string]int{}
	for _, m := range moves {
		if m.Source != charm {
			continue
		}
		var p struct {
			Modes   []int         `json:"modes"`
			Targets []interface{} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		modes[fmt.Sprintf("%v/%d", p.Modes, len(p.Targets))]++
	}
	want := map[string]int{"[1]/1": 2, "[2]/0": 1}
	if len(modes) != len(want) {
		t.Fatalf("want %v, got %v", want, modes)
	}
	for k, v := range want {
		if modes[k] != v {
			t.Errorf("mode/targets %s: want %d, got %d (all: %v)", k, v, modes[k], modes)
		}
	}
}

func TestChooseTwoEnumeratesPairs(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	cmd := handCard(active, game.Card{Name: "Austere Command", TypeLine: "Sorcery", ManaCost: "{4}{W}{W}", OracleID: oracleAustereCommand})
	for i := 0; i < 6; i++ {
		battlefieldCard(g, active, basic("Plains", "Plains"))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	n := 0
	for _, m := range moves {
		if m.Source == cmd {
			n++
		}
	}
	// Four untargeted options, choose exactly two: C(4,2) = 6.
	if n != 6 {
		t.Errorf("want 6 mode pairs, got %d: %v", n, labels(moves))
	}
}

func TestExpansionCapIsHonoured(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	bolt := handCard(active, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	for i := 0; i < 20; i++ {
		battlefieldCard(g, opp, creature(fmt.Sprintf("Bear %d", i), "{1}{G}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateForWithOptions(g, active.ID, legal.Options{MaxExpansionPerSource: 5})
	n := 0
	for _, m := range moves {
		if m.Source == bolt {
			n++
		}
	}
	if n != 5 {
		t.Errorf("cap of 5 should yield 5 Bolt moves, got %d", n)
	}
}

// --- cleanup discard ----------------------------------------------

func TestCleanupDiscardToHandSize(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	for i := 0; i < 9; i++ {
		handCard(active, creature(fmt.Sprintf("Card %d", i), "{1}", 1, 1))
	}
	// Walk to end of turn; the cleanup hook parks the cursor.
	advanceTo(t, g, game.StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step != game.StepCleanup || len(g.DiscardPending) == 0 {
		t.Fatalf("expected a cleanup discard pause, at %s pending=%v", g.Turn.Step, g.DiscardPending)
	}
	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	// 9 cards, max 7 → discard 2: C(9,2)=36 capped at the default 12.
	if len(moves) != 12 {
		t.Fatalf("want 12 capped discard moves, got %d: %v", len(moves), labels(moves))
	}
	for _, m := range moves {
		if m.Type != legal.TypeDiscardSelection {
			t.Errorf("only discard_selection is legal at a cleanup pause: %q", m.Label)
		}
	}
	// Nobody else has anything to do.
	for _, p := range g.Seats {
		if p.ID != active.ID {
			if m := legal.EnumerateFor(g, p.ID); len(m) != 0 {
				t.Errorf("%s should have no moves during another seat's discard: %v", p.Name, labels(m))
			}
		}
	}
	// Discarding clears the pause and the turn advances.
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeDiscardSelection, Player: active.ID, Caller: active.ID, Params: moves[0].Params}); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Step == game.StepCleanup {
		t.Errorf("turn should have advanced after the discard; still at cleanup")
	}
}

// --- S22 choice kinds: search_library, entry_pay_life ---------------

const (
	oraclePollutedDelta    = "ef86989d-ce80-4e55-aece-7d11710eeffa"
	oracleHallowedFountain = "f1750962-a87c-49f6-b731-02ae971ac6ea"
)

func TestSearchLibraryOffersFailToFindAndEachCandidate(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	delta := battlefieldCard(g, active, game.Card{Name: "Polluted Delta", TypeLine: "Land", OracleID: oraclePollutedDelta})
	// Two matching lands in the library → a real choice (limit 1).
	for _, n := range []string{"Island A", "Island B"} {
		c := game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Basic Land — Island", Owner: active.ID, Controller: active.ID}
		active.Library.PushTop(c)
	}
	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, active.ID)
	var crack *legal.Move
	for i := range moves {
		if moves[i].Source == delta && moves[i].Kind == legal.KindActivate {
			crack = &moves[i]
		}
	}
	if crack == nil {
		t.Fatalf("Polluted Delta's ability not offered: %v", labels(moves))
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(crack.Type), Player: active.ID, Caller: active.ID, Params: crack.Params}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceSearchLibrary {
		t.Fatalf("expected a search prompt, got %+v", g.PendingChoices)
	}
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	fail, takes := 0, 0
	for _, m := range moves {
		switch {
		case strings.HasSuffix(m.Label, ": fail to find"):
			fail++
		case strings.Contains(m.Label, ": take Island"):
			takes++
		}
	}
	if len(moves) != 3 || fail != 1 || takes != 2 {
		t.Errorf("want fail-to-find + one take per candidate (3), got %v", labels(moves))
	}
}

func TestShocklandEntryOffersPayOrTapped(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	fountain := handCard(active, game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island", OracleID: oracleHallowedFountain})
	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, active.ID)
	var play *legal.Move
	for i := range moves {
		if moves[i].Source == fountain {
			play = &moves[i]
		}
	}
	if play == nil {
		t.Fatalf("land drop not offered: %v", labels(moves))
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(play.Type), Player: active.ID, Caller: active.ID, Params: play.Params}); err != nil {
		t.Fatal(err)
	}
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceEntryPayLife {
		t.Fatalf("expected the pay-life prompt, got %+v", g.PendingChoices)
	}
	moves = legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)
	pay, tapped := 0, 0
	for _, m := range moves {
		switch {
		case strings.HasSuffix(m.Label, ": pay 2 life"):
			pay++
		case strings.HasSuffix(m.Label, ": enter tapped"):
			tapped++
		}
	}
	if len(moves) != 2 || pay != 1 || tapped != 1 {
		t.Errorf("want pay / enter tapped, got %v", labels(moves))
	}
}
