package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// threat_gate_test.go — #3555, ADR 0009 amendment (owner, 2026-10-15:
// Q1 (a), Q2 (a), Q3 (a)). An untargeted protect, survive, pump, prevent or
// sac_outlet answer sets interacts only while an opponent's item on the
// stack threatens the seat: it targets the seat, a permanent the seat
// controls or an item the seat controls, or it is an instant or sorcery
// spell with no target. Otherwise the answer counts in the combat tier
// (combat_interacts), so combat is unchanged. remove keeps interacting
// against anything. A mana sacrifice outlet reads the combat step
// itself, since a mana move carries no combat_interacts.

const (
	oracleSporeFrogTG     = "97db6c39-e690-49b6-93a6-e51b8dfad10b"
	oraclePestilenceTG    = "dafe63ef-f3d6-45e7-877a-573da92ba85e"
	oracleAshnodsAltarTG  = "4d18bcba-a346-445e-a182-6cc30b7e066d"
	oracleSelflessSpiritT = "71d785a9-ddc8-472e-a778-a551b444a4bd"
)

type threatBoard struct {
	g                                *game.Game
	seat, opp                        *game.Player
	bears, oppBears                  uuid.UUID
	seer, troll, shade, frog, spirit uuid.UUID
	pestilence, altar                uuid.UUID
}

func newThreatBoard(t *testing.T, step game.Step) *threatBoard {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, step)
	b := &threatBoard{g: g, seat: seat, opp: opponentOf(g, seat)}
	for range 4 {
		battlefieldCard(g, seat, basic("Forest", "Forest"))
	}
	for range 3 {
		battlefieldCard(g, seat, basic("Swamp", "Swamp"))
	}
	b.bears = battlefieldCard(g, seat, creature("Grizzly Bears", "{1}{G}", 2, 2))
	b.oppBears = battlefieldCard(g, b.opp, creature("Runeclaw Bear", "{1}{G}", 2, 2))
	b.seer = battlefieldCard(g, seat, game.Card{
		Name: "Viscera Seer", TypeLine: "Creature — Vampire Wizard", OracleID: oracleViscerSeerHT,
		ManaCost: "{B}", Power: 1, Toughness: 1,
	})
	b.troll = battlefieldCard(g, seat, game.Card{
		Name: "Undercity Troll", TypeLine: "Creature — Troll", OracleID: oracleUndercityTrollHT,
		ManaCost: "{1}{G}", Power: 2, Toughness: 2,
	})
	b.shade = battlefieldCard(g, seat, game.Card{
		Name: "Evernight Shade", TypeLine: "Creature — Shade", OracleID: oracleEvernightShadeHT,
		ManaCost: "{3}{B}", Power: 1, Toughness: 1,
	})
	b.frog = battlefieldCard(g, seat, game.Card{
		Name: "Spore Frog", TypeLine: "Creature — Frog", OracleID: oracleSporeFrogTG,
		ManaCost: "{G}", Power: 1, Toughness: 1,
	})
	b.spirit = battlefieldCard(g, seat, game.Card{
		Name: "Selfless Spirit", TypeLine: "Creature — Spirit", OracleID: oracleSelflessSpiritT,
		ManaCost: "{1}{W}", Power: 2, Toughness: 1,
	})
	b.pestilence = battlefieldCard(g, seat, game.Card{
		Name: "Pestilence", TypeLine: "Enchantment", OracleID: oraclePestilenceTG, ManaCost: "{2}{B}{B}",
	})
	b.altar = battlefieldCard(g, seat, game.Card{
		Name: "Ashnod's Altar", TypeLine: "Artifact", OracleID: oracleAshnodsAltarTG, ManaCost: "{3}",
	})
	return b
}

// push puts an item controlled by `by` on the stack: a spell when
// typeLine is set, else a triggered ability of a permanent of theirs.
func (b *threatBoard) push(by *game.Player, typeLine string, targets ...game.TargetRef) uuid.UUID {
	id := uuid.New()
	kind := game.StackItemSpell
	if typeLine == "" {
		kind = game.StackItemTriggered
	}
	b.g.WithWriteLock(func() {
		b.g.Stack.PushTop(game.Card{
			InstanceID: id, Name: "Stack item", TypeLine: typeLine,
			Owner: by.ID, Controller: by.ID,
		})
		if b.g.StackMeta == nil {
			b.g.StackMeta = make(map[uuid.UUID]*game.StackItem)
		}
		b.g.StackMeta[id] = &game.StackItem{
			ID: id, Kind: kind, Controller: by.ID, Owner: by.ID, SourceCardID: id, Targets: targets,
		}
	})
	return id
}

// check asserts every move of `kind` from source: interacts and
// combat_interacts as wanted.
func (b *threatBoard) check(t *testing.T, moves []legal.Move, name string, source uuid.UUID, kind legal.Kind, interacts, combat bool) {
	t.Helper()
	got := movesFrom(moves, source, kind)
	if len(got) == 0 {
		t.Fatalf("%s offered no %s move: %v", name, kind, labels(moves))
	}
	for _, m := range got {
		if m.Interacts != interacts || m.CombatInteracts != combat {
			t.Errorf("%s: interacts=%v combat_interacts=%v, want %v/%v (%s)",
				name, m.Interacts, m.CombatInteracts, interacts, combat, m.Label)
		}
		if m.CombatDefenderOnly {
			t.Errorf("%s: combat_defender_only set on a threat-gated answer (%s)", name, m.Label)
		}
	}
}

// checkGated asserts the threat-gated answers, and that Pestilence (remove)
// interacts whatever is on the stack.
func (b *threatBoard) checkGated(t *testing.T, threat bool) {
	t.Helper()
	moves := legal.EnumerateFor(b.g, b.seat.ID)
	combat := !threat
	b.check(t, moves, "Viscera Seer (sac_outlet)", b.seer, legal.KindActivate, threat, combat)
	b.check(t, moves, "Undercity Troll (survive)", b.troll, legal.KindActivate, threat, combat)
	b.check(t, moves, "Evernight Shade (pump)", b.shade, legal.KindActivate, threat, combat)
	b.check(t, moves, "Spore Frog (prevent)", b.frog, legal.KindActivate, threat, combat)
	b.check(t, moves, "Selfless Spirit (survive)", b.spirit, legal.KindActivate, threat, combat)
	b.check(t, moves, "Pestilence (remove)", b.pestilence, legal.KindActivate, true, false)
	b.check(t, moves, "Ashnod's Altar (mana sac_outlet)", b.altar, legal.KindMana, threat, false)
}

func TestThreatGatedAnswersSkipAStackThatThreatensNothing(t *testing.T) {
	t.Run("empty stack", func(t *testing.T) {
		newThreatBoard(t, game.StepPrecombatMain).checkGated(t, false)
	})
	t.Run("an opponent's removal aimed at another player's creature", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "Instant", game.TargetRef{Kind: game.TargetCard, ID: b.oppBears})
		b.checkGated(t, false)
	})
	t.Run("an opponent's creature spell", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "Creature — Bear")
		b.checkGated(t, false)
	})
	t.Run("an opponent's untargeted trigger", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "")
		b.checkGated(t, false)
	})
	t.Run("an opponent's spell that targets its own controller", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "Sorcery", game.TargetRef{Kind: game.TargetSelf})
		b.checkGated(t, false)
	})
	t.Run("the seat's own untargeted sorcery", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.seat, "Sorcery")
		b.checkGated(t, false)
	})
}

func TestThreatGatedAnswersStopForAThreat(t *testing.T) {
	t.Run("removal on the seat's creature", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "Instant", game.TargetRef{Kind: game.TargetCard, ID: b.bears})
		b.checkGated(t, true)
	})
	t.Run("a sweeper (an untargeted sorcery)", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "Sorcery")
		b.checkGated(t, true)
	})
	t.Run("an untargeted instant under another item", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "Instant")
		b.push(b.opp, "")
		b.checkGated(t, true)
	})
	t.Run("a spell aimed at the seat", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "Instant", game.TargetRef{Kind: game.TargetPlayer, ID: b.seat.ID})
		b.checkGated(t, true)
	})
	t.Run("a counterspell aimed at the seat's spell", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		mine := b.push(b.seat, "Sorcery")
		b.push(b.opp, "Instant", game.TargetRef{Kind: game.TargetCard, ID: mine})
		b.checkGated(t, true)
	})
	t.Run("a triggered ability aimed at the seat's creature", func(t *testing.T) {
		b := newThreatBoard(t, game.StepPrecombatMain)
		b.push(b.opp, "", game.TargetRef{Kind: game.TargetCard, ID: b.troll})
		b.checkGated(t, true)
	})
}

// Combat is unchanged: with nothing on the stack, an activated answer is
// a combat answer (the client counts it in its combat window), and a mana
// sacrifice outlet interacts in a combat step.
func TestThreatGatedAnswersStillCountInCombat(t *testing.T) {
	b := newThreatBoard(t, game.StepBeginCombat)
	moves := legal.EnumerateFor(b.g, b.seat.ID)
	b.check(t, moves, "Viscera Seer (sac_outlet)", b.seer, legal.KindActivate, false, true)
	b.check(t, moves, "Selfless Spirit (survive)", b.spirit, legal.KindActivate, false, true)
	b.check(t, moves, "Ashnod's Altar (mana sac_outlet)", b.altar, legal.KindMana, true, false)
}
