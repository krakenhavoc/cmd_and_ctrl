package legal_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// held_answers_3376a15c_test.go — the #4734 regression (ADR 0009's
// amendment of 2026-10-30). An MCP seat played table 3376a15c on
// cmd.labxp.io on 2026-10-30 with pass_until "smart", on the mono-white
// Adeline deck, and was stopped at 23 of its first 43 decisions with
// nothing to do but moves it held all game: Flawless Maneuver, free once
// Adeline was out, and Ranger-Captain of Eos's "Sacrifice this creature:
// Your opponents can't cast noncreature spells this turn".
//
// This rebuilds the seat's standing board (Adeline, Resplendent Cathar,
// its commander; Ranger-Captain and Mother of Runes, both summoning sick;
// two Plains, one untapped; a tapped Mind Stone; Flawless Maneuver in
// hand; the game had four Plains, three of them tapped, which changes no
// move) at the four-seat table, in the game's turn order, and walks every
// window shape the session's transcript holds. Each frame asserts both
// held moves' answers_now, and that nothing else the seat holds carries
// one. The frames join the agreement fixture (smart_pass_agreement.json),
// whose readers (slices 3 and 4) assert their stop or pass on each.
//
// Flawless Maneuver declares survive (indestructible, ADR 0142) since
// slice 5 (#4773). Each frame states both verdicts: with the card
// declared survive, and as on 2026-10-30, when it declared nothing, was
// classed an instant and stopped the seat at every one of these windows.
// The test reads the catalog to pick the column, so a change to the
// card's declaration fails here by name rather than in the fixture's
// bytes alone.
//
// Two simplifications, neither of which the windows depend on. Each frame
// is its own game, at turn 1 of whichever player's turn the window was
// in, with eleven-card decks so the fixture stays small. And an attacker
// whose attack has been locked in (the declare-blockers frames) carries
// no catalog row, so its attack trigger, which had resolved by the step
// the frame copies, is not put on the stack again; in declare attackers
// the attack is staged and the trigger is put there by hand, aimed where
// the transcript says it was.

const (
	oracleAdelineHA      = "38515f89-348b-4cf3-b7bd-1f6fe4ce2fba"
	oracleMotherOfRunes  = "60433b48-d27f-413c-905c-43839b1943f1"
	oracleFlawlessManvr  = "4e183439-17d2-47ff-9d99-5e22821d91e3"
	oracleSmaugMagnif    = "c53b7602-5e36-4900-896c-844bf6eac22e"
	oracleGoreclaw       = "befb211f-37ca-4083-98d4-9ff1f28be3f2"
	oracleGhaltaUnstop   = "8e5ea773-a60b-4502-9db5-ce93fa61cc89"
	oraclePalantir       = "7798a4d9-1c1d-48e6-b9a2-ecd6aec1efa7"
	oracleGenerousGift   = "fae37e28-e137-4177-b973-fa8b4dd8f409"
	oracleArcaneSignetHA = "0bc7f093-bef0-4f1a-852c-4b75ebf54838"
)

// seats3376 is the table's turn order: Bot 1 won the opening roll and
// took the first turn, then Claude (the seat), krakenhavoc and Slysloth.
var seats3376 = []string{"Bot 1", "Claude", "krakenhavoc", "Slysloth"}

func adeline3376() game.Card {
	return game.Card{
		InstanceID: uuid.New(),
		Name:       "Adeline, Resplendent Cathar", TypeLine: "Legendary Creature — Human Knight",
		ManaCost: "{1}{W}{W}", OracleID: oracleAdelineHA, Power: 1, Toughness: 4, IsCommander: true,
	}
}

func goreclaw3376() game.Card {
	return game.Card{
		InstanceID: uuid.New(),
		Name:       "Goreclaw, Terror of Qal Sisma", TypeLine: "Legendary Creature — Bear",
		ManaCost: "{3}{G}", OracleID: oracleGoreclaw, Power: 4, Toughness: 3, IsCommander: true,
	}
}

func smaug3376() game.Card {
	return game.Card{
		Name: "Smaug the Magnificent", TypeLine: "Legendary Creature — Dragon",
		ManaCost: "{3}{R}{R}", OracleID: oracleSmaugMagnif, Power: 4, Toughness: 3,
	}
}

func ghalta3376() game.Card {
	return game.Card{
		Name: "Ghalta the Unstoppable", TypeLine: "Legendary Creature — Dinosaur",
		ManaCost: "{6}{G}{G}", OracleID: oracleGhaltaUnstop, Power: 8, Toughness: 8,
	}
}

// standIn is an attacker with no catalog row (the header says why).
func standIn(c game.Card, keywords ...string) game.Card {
	c.OracleID = ""
	c.IsCommander = false
	c.Keywords = keywords
	return c
}

// table3376 is one frame's game.
type table3376 struct {
	g      *game.Game
	by     map[string]*game.Player
	claude *game.Player
	// The seat's held moves and the creatures an item may aim at.
	fm, rc, mother, adeline uuid.UUID
}

// newTable3376 is the table at turn 1 of `active`, past the mulligans at
// its upkeep, with every hand empty. The seat's standing board is put
// down by stockClaude once the cursor is where the frame wants it.
func newTable3376(t *testing.T, active string) *table3376 {
	t.Helper()
	start := -1
	for i, n := range seats3376 {
		if n == active {
			start = i
		}
	}
	if start < 0 {
		t.Fatalf("no seat %q", active)
	}
	g := game.NewGame()
	for i := range 4 {
		name := seats3376[(start+i)%4]
		deck := make([]game.Card, 0, 11)
		switch name {
		case "Claude":
			deck = append(deck, adeline3376())
		case "Bot 1":
			deck = append(deck, goreclaw3376())
		default:
			deck = append(deck, game.NewCommander(name+"'s commander", uuid.Nil))
		}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(name, deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3376, 4734))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	tb := &table3376{g: g, by: map[string]*game.Player{}}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
		clearHand(p)
		tb.by[p.Name] = p
	}
	tb.claude = tb.by["Claude"]
	if g.Seats[g.Turn.ActiveSeat].Name != active {
		t.Fatalf("turn 1 is %s's, want %s's", g.Seats[g.Turn.ActiveSeat].Name, active)
	}
	advanceTo(t, g, game.StepUpkeep)
	return tb
}

// commanderToBattlefield moves a seat's commander from the command zone
// onto the battlefield, where the transcript has it.
func (tb *table3376) commanderToBattlefield(t *testing.T, name string) uuid.UUID {
	t.Helper()
	p := tb.by[name]
	if len(p.Command.Cards) != 1 {
		t.Fatalf("%s's command zone holds %d cards", name, len(p.Command.Cards))
	}
	c, err := p.Command.Remove(p.Command.Cards[0].InstanceID)
	if err != nil {
		t.Fatalf("%s's commander: %v", name, err)
	}
	c.Controller = p.ID
	c.KnownBy = map[uuid.UUID]bool{}
	for _, s := range tb.g.Seats {
		c.KnownBy[s.ID] = true
	}
	tb.g.Battlefield.PushTop(c)
	return c.InstanceID
}

// stockClaude puts down the seat's standing board and empties every hand
// a draw step may have filled.
func (tb *table3376) stockClaude(t *testing.T) {
	t.Helper()
	g, c := tb.g, tb.claude
	for _, p := range g.Seats {
		clearHand(p)
	}
	tb.adeline = tb.commanderToBattlefield(t, "Claude")
	tb.rc = battlefieldCard(g, c, game.Card{
		Name: "Ranger-Captain of Eos", TypeLine: "Creature — Human Soldier Ranger",
		ManaCost: "{1}{W}{W}", OracleID: oracleRangerCaptainAW, Power: 3, Toughness: 3, SummonedThisTurn: true,
	})
	tb.mother = battlefieldCard(g, c, game.Card{
		Name: "Mother of Runes", TypeLine: "Creature — Human Cleric",
		ManaCost: "{W}", OracleID: oracleMotherOfRunes, Power: 1, Toughness: 1, SummonedThisTurn: true,
	})
	for _, tapped := range []bool{true, false} {
		plains := basic("Plains", "Plains")
		plains.Tapped = tapped
		battlefieldCard(g, c, plains)
	}
	battlefieldCard(g, c, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}", OracleID: oracleMindStoneHT, Tapped: true})
	tb.fm = handCard(c, game.Card{Name: "Flawless Maneuver", TypeLine: "Instant", ManaCost: "{2}{W}", OracleID: oracleFlawlessManvr})
}

// stockBot1 puts down Bot 1's Goreclaw, Ghalta and a tapped Arcane Signet,
// and returns Goreclaw and Ghalta. standInGoreclaw puts a stand-in in
// Goreclaw's place.
func (tb *table3376) stockBot1(t *testing.T, standInGoreclaw bool) (uuid.UUID, uuid.UUID) {
	t.Helper()
	bot := tb.by["Bot 1"]
	var goreclaw uuid.UUID
	if standInGoreclaw {
		bot.Command.Cards = nil
		goreclaw = battlefieldCard(tb.g, bot, standIn(goreclaw3376(), "trample"))
	} else {
		goreclaw = tb.commanderToBattlefield(t, "Bot 1")
	}
	ghalta := ghalta3376()
	ghalta.SummonedThisTurn = true
	ghaltaID := battlefieldCard(tb.g, bot, ghalta)
	battlefieldCard(tb.g, bot, game.Card{Name: "Arcane Signet", TypeLine: "Artifact", ManaCost: "{2}", OracleID: oracleArcaneSignetHA, Tapped: true})
	return goreclaw, ghaltaID
}

// heldFrame is one window of the game, and what it should say.
type heldFrame struct {
	name, note string
	g          *game.Game
	claude     *game.Player
	fm, rc     uuid.UUID
	// rcWhy and fmWhy say why each verdict is what it is.
	rcMatches bool
	rcWhy     string
	// fmSurvive is Flawless Maneuver's verdict once it declares survive.
	fmSurvive bool
	fmWhy     string
	// stop is the smart pass verdict at the default settings: [0] with
	// Flawless Maneuver declaring nothing (2026-10-30), [1] with it
	// declaring survive (#4773).
	stop    [2]bool
	stopWhy [2]string
}

// heldFrames3376 builds every window shape of the transcript, in the
// order the issue lists them.
func heldFrames3376(t *testing.T) []heldFrame {
	t.Helper()
	const (
		fmInstant   = "Flawless Maneuver declares nothing, so it is an instant, a response here as on 2026-10-30"
		ownMain     = "the seat's own main phase with an empty stack: a cast or an activation is a play, and the main phase is ticked by default"
		rcSpell     = "an opponent's spell is on the stack (restrict)"
		rcUpkeep    = "an opponent's upkeep (restrict)"
		rcNoSpell   = "no opponent's spell on the stack and not an opponent's upkeep; nothing targets the Captain, which is not in combat"
		fmNoThreat  = "nothing threatens the seat or its permanents, and no combat involves it"
		fmAttacked  = "a creature attacks the seat (a combat that involves you)"
		fmAtMother  = "an opponent's trigger targets Mother of Runes, a permanent of the seat's"
		fmAtSeat    = "the trigger targets the seat itself, which protect, prevent and sac_outlet answer and survive does not (Q4)"
		fmUnblocked = "an unblocked attacker at the seat, which protect and prevent answer and survive does not (Q4)"
		stopRC      = "Ranger-Captain matches"
		stopFM      = "Flawless Maneuver matches"
		passNothing = "nothing the seat holds matches, and the mana is no response"
	)
	var out []heldFrame
	add := func(f heldFrame, tb *table3376) {
		f.g, f.claude, f.fm, f.rc = tb.g, tb.claude, tb.fm, tb.rc
		givePriority(tb.g, tb.claude)
		out = append(out, f)
	}
	stops := func(why1 string) ([2]bool, [2]string) {
		return [2]bool{true, true}, [2]string{fmInstant, why1}
	}
	quiets := func() ([2]bool, [2]string) {
		return [2]bool{true, false}, [2]string{fmInstant, passNothing}
	}

	// An opponent's creature spell: Smaug the Magnificent (turn 11,
	// w0.391) and Ghalta the Unstoppable (turn 13, w0.485).
	{
		tb := newTable3376(t, "krakenhavoc")
		advanceTo(t, tb.g, game.StepPrecombatMain)
		tb.stockClaude(t)
		pushSpellAW(tb.g, tb.by["krakenhavoc"], smaug3376())
		f := heldFrame{name: "3376a15c_smaug_cast", note: "krakenhavoc's precombat main: Smaug the Magnificent on the stack",
			rcMatches: true, rcWhy: rcSpell, fmWhy: "a creature spell threatens nothing; its enters trigger would be judged on its own"}
		f.stop, f.stopWhy = stops(stopRC)
		add(f, tb)
	}
	{
		tb := newTable3376(t, "Bot 1")
		advanceTo(t, tb.g, game.StepPrecombatMain)
		tb.stockClaude(t)
		tb.commanderToBattlefield(t, "Bot 1")
		pushSpellAW(tb.g, tb.by["Bot 1"], ghalta3376())
		f := heldFrame{name: "3376a15c_ghalta_cast", note: "Bot 1's precombat main: Ghalta the Unstoppable on the stack",
			rcMatches: true, rcWhy: rcSpell, fmWhy: "a creature spell threatens nothing"}
		f.stop, f.stopWhy = stops(stopRC)
		add(f, tb)
	}
	// Smaug's trigger aimed at another player during an attack on the
	// seat (turn 11, w0.405).
	{
		tb := newTable3376(t, "krakenhavoc")
		advanceTo(t, tb.g, game.StepDeclareAttackers)
		tb.stockClaude(t)
		k := tb.by["krakenhavoc"]
		smaug := battlefieldCard(tb.g, k, smaug3376())
		if err := tb.g.DeclareAttacker(smaug, tb.claude.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		ref, label := triggerRefAW(t, smaug3376(), "deals damage equal to the number of Treasures")
		pushAbilityAW(tb.g, k, smaug, game.StackItemTriggered, label, ref, game.TargetRef{Kind: game.TargetPlayer, ID: tb.by["Bot 1"].ID})
		f := heldFrame{name: "3376a15c_smaug_trigger_at_bot1_attack_on_seat",
			note:  "krakenhavoc's declare attackers: Smaug attacks the seat, its trigger aimed at Bot 1",
			rcWhy: rcNoSpell, fmSurvive: true, fmWhy: fmAttacked}
		f.stop, f.stopWhy = stops(stopFM)
		add(f, tb)
	}
	// Goreclaw's untargeted trigger during an attack on another player
	// (turn 13, w0.500).
	{
		tb := newTable3376(t, "Bot 1")
		advanceTo(t, tb.g, game.StepDeclareAttackers)
		tb.stockClaude(t)
		goreclaw, _ := tb.stockBot1(t, false)
		if err := tb.g.DeclareAttacker(goreclaw, tb.by["krakenhavoc"].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		ref, label := triggerRefAW(t, goreclaw3376(), "pump and trample")
		pushAbilityAW(tb.g, tb.by["Bot 1"], goreclaw, game.StackItemTriggered, label, ref)
		f := heldFrame{name: "3376a15c_goreclaw_trigger_attack_on_other",
			note:  "Bot 1's declare attackers: Goreclaw attacks krakenhavoc, its untargeted pump trigger on the stack",
			rcWhy: rcNoSpell, fmWhy: "an untargeted trigger that declares no sweep, in a combat between other players"}
		f.stop, f.stopWhy = quiets()
		add(f, tb)
	}
	// Smaug's trigger aimed at Mother of Runes (turn 15, w0.619).
	{
		tb := newTable3376(t, "krakenhavoc")
		advanceTo(t, tb.g, game.StepDeclareAttackers)
		tb.stockClaude(t)
		k := tb.by["krakenhavoc"]
		smaug := battlefieldCard(tb.g, k, smaug3376())
		if err := tb.g.DeclareAttacker(smaug, tb.by["Bot 1"].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		ref, label := triggerRefAW(t, smaug3376(), "deals damage equal to the number of Treasures")
		pushAbilityAW(tb.g, k, smaug, game.StackItemTriggered, label, ref, game.TargetRef{Kind: game.TargetCard, ID: tb.mother})
		f := heldFrame{name: "3376a15c_smaug_trigger_at_mother",
			note:  "krakenhavoc's declare attackers: Smaug attacks Bot 1, its trigger aimed at Mother of Runes",
			rcWhy: rcNoSpell, fmSurvive: true, fmWhy: fmAtMother}
		f.stop, f.stopWhy = stops(stopFM)
		add(f, tb)
	}
	// Palantír of Orthanc's trigger aimed at the seat. Palantír is put
	// down at the end step, so the engine does not trigger it a second
	// time.
	{
		tb := newTable3376(t, "Slysloth")
		advanceTo(t, tb.g, game.StepEnd)
		tb.stockClaude(t)
		s := tb.by["Slysloth"]
		palantir := game.Card{Name: "Palantír of Orthanc", TypeLine: "Legendary Artifact", ManaCost: "{3}", OracleID: oraclePalantir}
		src := battlefieldCard(tb.g, s, palantir)
		ref, label := triggerRefAW(t, palantir, "influence counter")
		pushAbilityAW(tb.g, s, src, game.StackItemTriggered, label, ref, game.TargetRef{Kind: game.TargetPlayer, ID: tb.claude.ID})
		f := heldFrame{name: "3376a15c_palantir_trigger_at_seat",
			note:  "Slysloth's end step: Palantír of Orthanc's trigger targets the seat",
			rcWhy: rcNoSpell, fmWhy: fmAtSeat}
		f.stop, f.stopWhy = quiets()
		add(f, tb)
	}
	// Generous Gift on another player's Ghalta.
	{
		tb := newTable3376(t, "Slysloth")
		advanceTo(t, tb.g, game.StepPrecombatMain)
		tb.stockClaude(t)
		_, ghalta := tb.stockBot1(t, false)
		pushSpellAW(tb.g, tb.by["Slysloth"], game.Card{Name: "Generous Gift", TypeLine: "Instant", ManaCost: "{2}{W}", OracleID: oracleGenerousGift},
			game.TargetRef{Kind: game.TargetCard, ID: ghalta})
		f := heldFrame{name: "3376a15c_generous_gift_on_ghalta",
			note:      "Slysloth's precombat main: Generous Gift aimed at Bot 1's Ghalta",
			rcMatches: true, rcWhy: rcSpell, fmWhy: "the spell targets another player's permanent, which threatens nothing of the seat's"}
		f.stop, f.stopWhy = stops(stopRC)
		add(f, tb)
	}
	// An empty opponent end step (turn 11, w0.429).
	{
		tb := newTable3376(t, "krakenhavoc")
		advanceTo(t, tb.g, game.StepEnd)
		tb.stockClaude(t)
		f := heldFrame{name: "3376a15c_empty_opponent_end_step", note: "krakenhavoc's end step, the stack empty",
			rcWhy: rcNoSpell, fmWhy: fmNoThreat}
		f.stop, f.stopWhy = quiets()
		add(f, tb)
	}
	// Empty declare-attackers and declare-blockers steps, in an attack on
	// another player (turn 13, w0.504 and w0.508) and in one on the seat
	// (turn 11, w0.409 and w0.413).
	{
		tb := newTable3376(t, "Bot 1")
		advanceTo(t, tb.g, game.StepDeclareAttackers)
		tb.stockClaude(t)
		goreclaw, _ := tb.stockBot1(t, false)
		if err := tb.g.DeclareAttacker(goreclaw, tb.by["krakenhavoc"].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		f := heldFrame{name: "3376a15c_empty_declare_attackers_attack_on_other",
			note:  "Bot 1's declare attackers: Goreclaw attacks krakenhavoc, the stack empty",
			rcWhy: rcNoSpell, fmWhy: "a combat between other players"}
		f.stop, f.stopWhy = quiets()
		add(f, tb)
	}
	{
		tb := newTable3376(t, "Bot 1")
		advanceTo(t, tb.g, game.StepDeclareAttackers)
		tb.stockClaude(t)
		goreclaw, _ := tb.stockBot1(t, true)
		if err := tb.g.DeclareAttacker(goreclaw, tb.by["krakenhavoc"].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, tb.g, game.StepDeclareBlockers)
		f := heldFrame{name: "3376a15c_empty_declare_blockers_attack_on_other",
			note:  "Bot 1's declare blockers: Goreclaw attacks krakenhavoc unblocked, the stack empty",
			rcWhy: rcNoSpell, fmWhy: "a combat between other players"}
		f.stop, f.stopWhy = quiets()
		add(f, tb)
	}
	{
		tb := newTable3376(t, "krakenhavoc")
		advanceTo(t, tb.g, game.StepDeclareAttackers)
		tb.stockClaude(t)
		smaug := battlefieldCard(tb.g, tb.by["krakenhavoc"], smaug3376())
		if err := tb.g.DeclareAttacker(smaug, tb.claude.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		f := heldFrame{name: "3376a15c_empty_declare_attackers_attack_on_seat",
			note:  "krakenhavoc's declare attackers: Smaug attacks the seat, the stack empty",
			rcWhy: rcNoSpell, fmSurvive: true, fmWhy: fmAttacked}
		f.stop, f.stopWhy = stops(stopFM)
		add(f, tb)
	}
	{
		tb := newTable3376(t, "krakenhavoc")
		advanceTo(t, tb.g, game.StepDeclareAttackers)
		tb.stockClaude(t)
		smaug := battlefieldCard(tb.g, tb.by["krakenhavoc"], standIn(smaug3376(), "flying", "haste"))
		if err := tb.g.DeclareAttacker(smaug, tb.claude.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, tb.g, game.StepDeclareBlockers)
		f := heldFrame{name: "3376a15c_empty_declare_blockers_attack_on_seat",
			note:  "krakenhavoc's declare blockers: Smaug, flying, attacks the seat unblocked, the stack empty",
			rcWhy: rcNoSpell, fmWhy: fmUnblocked}
		f.stop, f.stopWhy = quiets()
		add(f, tb)
	}
	// An opponent's upkeep, empty and with Smaug's Treasure trigger on the
	// stack (turn 15, w0.596).
	{
		tb := newTable3376(t, "krakenhavoc")
		tb.stockClaude(t)
		f := heldFrame{name: "3376a15c_opponent_upkeep", note: "krakenhavoc's upkeep, the stack empty",
			rcMatches: true, rcWhy: rcUpkeep, fmWhy: fmNoThreat}
		f.stop, f.stopWhy = stops(stopRC)
		add(f, tb)
	}
	{
		tb := newTable3376(t, "krakenhavoc")
		tb.stockClaude(t)
		k := tb.by["krakenhavoc"]
		smaug := battlefieldCard(tb.g, k, smaug3376())
		ref, label := triggerRefAW(t, smaug3376(), "create a Treasure")
		pushAbilityAW(tb.g, k, smaug, game.StackItemTriggered, label, ref)
		f := heldFrame{name: "3376a15c_opponent_upkeep_treasure_trigger",
			note:      "krakenhavoc's upkeep: Smaug's Treasure trigger on the stack",
			rcMatches: true, rcWhy: rcUpkeep, fmWhy: "an untargeted trigger that declares no sweep"}
		f.stop, f.stopWhy = stops(stopRC)
		add(f, tb)
	}
	// The seat's own main phase.
	{
		tb := newTable3376(t, "Claude")
		advanceTo(t, tb.g, game.StepPrecombatMain)
		tb.stockClaude(t)
		f := heldFrame{name: "3376a15c_own_main_phase", note: "the seat's precombat main, the stack empty",
			rcWhy: "the seat's own turn, nothing on the stack", fmWhy: fmNoThreat}
		f.stop, f.stopWhy = [2]bool{true, true}, [2]string{ownMain, ownMain}
		add(f, tb)
	}
	return out
}

// fmColumn is the column of heldFrame.stop Flawless Maneuver's catalog
// declaration picks: 0 for nothing, 1 for survive (#4773).
func fmColumn(t *testing.T) int {
	t.Helper()
	switch a := game.CardPurposeOf(game.Card{OracleID: oracleFlawlessManvr}).Answers; a {
	case 0:
		return 0
	case game.AnswerSurvive:
		return 1
	default:
		t.Fatalf("Flawless Maneuver declares %s; held_answers_3376a15c_test.go expects nothing or survive, so add a column for it", a)
		return 0
	}
}

// fmWant is Flawless Maneuver's expected answers_now for a frame.
func (f heldFrame) fmWant(col int) *bool {
	if col == 0 {
		return nil
	}
	return verdict(f.fmSurvive)
}

func TestHeldAnswersTable3376a15c(t *testing.T) {
	col := fmColumn(t)
	frames := heldFrames3376(t)
	rcStops, fmStops, stops := 0, 0, 0
	for _, f := range frames {
		t.Run(f.name, func(t *testing.T) {
			moves := legal.EnumerateFor(f.g, f.claude.ID)
			wantVerdict(t, moves, "Flawless Maneuver", f.fm, legal.KindCast, f.fmWant(col))
			wantVerdict(t, moves, "Ranger-Captain of Eos", f.rc, legal.KindActivate, verdict(f.rcMatches))
			for _, m := range moves {
				if m.Source == f.fm || m.Source == f.rc || m.AnswersNow == nil {
					continue
				}
				t.Errorf("%q carries answers_now %s; the seat holds nothing else that declares", m.Label, verdictString(m.AnswersNow))
			}
			// The view the readers get carries the same verdicts.
			view := protocol.ViewOfGameFor(f.g, f.claude.ID.String())
			wantVerdict(t, view.LegalMoves, "Flawless Maneuver (frame)", f.fm, legal.KindCast, f.fmWant(col))
			wantVerdict(t, view.LegalMoves, "Ranger-Captain of Eos (frame)", f.rc, legal.KindActivate, verdict(f.rcMatches))
		})
		if f.rcMatches {
			rcStops++
		}
		if f.fmSurvive {
			fmStops++
		}
		if f.stop[col] {
			stops++
		}
	}
	// The amendment's count for these window shapes: Ranger-Captain
	// matches with an opponent's spell on the stack and in an opponent's
	// upkeep, five of the fifteen; Flawless Maneuver, as survive, in the
	// three attacks on the seat or aimed at its permanent.
	if len(frames) != 15 || rcStops != 5 || fmStops != 3 {
		t.Errorf("%d frames, Ranger-Captain matches %d and Flawless Maneuver (survive) %d; want 15, 5 and 3", len(frames), rcStops, fmStops)
	}
	t.Logf("of %d windows, %d stop with Flawless Maneuver as the catalog declares it (column %d)", len(frames), stops, col)
}
