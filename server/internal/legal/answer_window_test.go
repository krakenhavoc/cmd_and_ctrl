package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// answer_window_test.go — ADR 0009's amendment of 2026-10-30 (#4734),
// accepted 2026-10-31 with the owner's answers Q1 to Q5 (a). A move whose
// declared answers are something other than `value` alone carries
// answers_now: whether one of them matches the window the seat holds
// priority in. The tables here are the amendment's, kind against window,
// on real four-seat games.
//
// The per-kind rows are free, untargeted rows carried on artifacts of the
// seat's, one declared kind each, so a case asserts exactly which kinds
// match and nothing else on the board can answer for them. Stack items
// are put on the stack the way the engine keeps them: a spell is a card
// on the stack zone with its metadata, an ability is metadata alone
// (game/stack.go).

const (
	oracleWrathOfGodAW    = "34515b16-c9a4-4f98-8c77-416a7a523407"
	oracleSwordsAW        = "b1544f21-7e98-461b-aed5-e748b0168c52"
	oracleRangerCaptainAW = "cada3481-cc2b-4412-b9b5-0436af53aad2"
)

// answerKinds is every kind the match judges, in the vocabulary's order.
// remove and value carry no verdict and are TestAnswersNowAbsentKeepsTodaysClass's.
var answerKinds = []game.Answers{
	game.AnswerProtect, game.AnswerSurvive, game.AnswerPump, game.AnswerPrevent,
	game.AnswerSacOutlet, game.AnswerRestrict, game.AnswerGraveyard,
	game.AnswerCombatGrant, game.AnswerAnimate, game.AnswerMakesBlocker,
}

// saving is the five kinds that keep or trade away something of yours.
const saving = game.AnswerProtect | game.AnswerSurvive | game.AnswerPump | game.AnswerPrevent | game.AnswerSacOutlet

// onSpell is what an opponent's spell on the stack matches whatever it
// targets: restrict, and graveyard, which Q5 split out of it.
const onSpell = game.AnswerRestrict | game.AnswerGraveyard

// declaredRelic is an artifact carrying one free, untargeted row that
// declares `a`.
func declaredRelic(a game.Answers) game.Card {
	return game.Card{
		Name: "Relic (" + a.String() + ")", TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:   "Answer: " + a.String(),
			Purpose: game.Purpose{Answers: a},
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

// pushSpellAW puts a spell controlled by `by` on the stack: the card on
// the stack zone and its metadata, as a cast leaves them.
func pushSpellAW(g *game.Game, by *game.Player, c game.Card, targets ...game.TargetRef) uuid.UUID {
	id := uuid.New()
	c.InstanceID, c.Owner, c.Controller = id, by.ID, by.ID
	c.KnownBy = map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		c.KnownBy[s.ID] = true
	}
	g.WithWriteLock(func() {
		g.Stack.PushTop(c)
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[id] = &game.StackItem{
			ID: id, Kind: game.StackItemSpell, Controller: by.ID, Owner: by.ID,
			SourceCardID: id, Label: c.Name, Targets: targets, Seq: nextSeqAW(g),
		}
	})
	return id
}

// pushAbilityAW puts an activated or triggered ability of `source` on the
// stack under `by`: metadata alone, with no card on the stack zone, as
// the engine keeps an ability. A nil ref names no catalog row.
func pushAbilityAW(g *game.Game, by *game.Player, source uuid.UUID, kind game.StackItemKind, label string, ref *game.AbilityRef, targets ...game.TargetRef) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		item := &game.StackItem{
			ID: id, Kind: kind, Controller: by.ID, Owner: by.ID,
			SourceCardID: source, Label: label, Targets: targets, Seq: nextSeqAW(g),
		}
		if ref != nil {
			item.Params = game.EffectParams{Ability: ref}
		}
		g.StackMeta[id] = item
	})
	return id
}

// nextSeqAW is the next stack Seq, so pushed items keep their order.
// Caller holds the write lock.
func nextSeqAW(g *game.Game) uint64 {
	var top uint64
	for _, it := range g.StackMeta {
		if it != nil && it.Seq > top {
			top = it.Seq
		}
	}
	return top + 1
}

// triggerRefAW is the ref the harvest gives a catalog card's triggered
// row whose key contains `label`.
func triggerRefAW(t *testing.T, c game.Card, label string) (*game.AbilityRef, string) {
	t.Helper()
	for _, row := range game.CatalogTriggers(game.CatalogKey(c)) {
		if strings.Contains(row.Key, label) {
			ref := game.TriggeredAbilityRef(row)
			if ref == nil {
				t.Fatalf("%s: trigger %q cannot be named by a ref", c.Name, row.Key)
			}
			return ref, row.Key
		}
	}
	t.Fatalf("%s has no triggered row containing %q", c.Name, label)
	return nil, ""
}

// activatedRefAW is the ref an activation gives a catalog card's
// activated row whose label contains `label`.
func activatedRefAW(t *testing.T, c game.Card, label string) (*game.AbilityRef, string) {
	t.Helper()
	key := game.CatalogKey(c)
	for i, ab := range game.CatalogActivatedAbilities(key) {
		if strings.Contains(ab.Label, label) {
			return &game.AbilityRef{Key: key, Slot: game.AbilitySlotActivated, Ref: game.OwnAbilityRef(i), Name: ab.Label}, ab.Label
		}
	}
	t.Fatalf("%s has no activated row containing %q", c.Name, label)
	return nil, ""
}

// givePriority hands `p` priority, as the players before it in turn
// order passing would.
func givePriority(g *game.Game, p *game.Player) {
	g.WithWriteLock(func() { g.Turn.PriorityHolder = p.Seat })
}

// verdictString renders an answers_now for a message.
func verdictString(v *bool) string {
	switch {
	case v == nil:
		return "absent"
	case *v:
		return "true"
	}
	return "false"
}

// wantVerdict asserts answers_now on every move of `kind` from source.
func wantVerdict(t *testing.T, moves []legal.Move, name string, source uuid.UUID, kind legal.Kind, want *bool) {
	t.Helper()
	got := movesFrom(moves, source, kind)
	if len(got) == 0 {
		t.Fatalf("%s offered no %s move: %v", name, kind, labels(moves))
	}
	for _, m := range got {
		if verdictString(m.AnswersNow) != verdictString(want) {
			t.Errorf("%s: answers_now %s, want %s (%s)", name, verdictString(m.AnswersNow), verdictString(want), m.Label)
		}
	}
}

func verdict(b bool) *bool { return &b }

// kindBoard is a four-seat table with the seat holding one declared row
// per kind, a creature, and an opponent who is not the active player
// holding a creature of their own.
type kindBoard struct {
	g                   *game.Game
	seat, active, other *game.Player
	rows                map[game.Answers]uuid.UUID
	bear, otherBear     uuid.UUID
}

// newKindBoard builds the table. The seat is `offset` seats after the
// player whose turn 1 it is (0 is that player), and `other` the seat two
// after it, so it is an opponent of both. noCreature leaves the seat
// without its bear.
func newKindBoard(t *testing.T, offset int, noCreature bool) *kindBoard {
	t.Helper()
	g := newTable(t)
	for _, p := range g.Seats {
		clearHand(p)
	}
	a := g.Turn.ActiveSeat
	b := &kindBoard{
		g: g, active: g.Seats[a], seat: g.Seats[(a+offset)%4],
		rows: map[game.Answers]uuid.UUID{},
	}
	for i := 1; i < 4; i++ {
		if s := g.Seats[(a+i)%4]; s != b.seat {
			b.other = s
			break
		}
	}
	for _, k := range answerKinds {
		b.rows[k] = battlefieldCard(g, b.seat, declaredRelic(k))
	}
	if !noCreature {
		b.bear = battlefieldCard(g, b.seat, creature("Seat Bear", "{1}{G}", 2, 2))
	}
	b.otherBear = battlefieldCard(g, b.other, creature("Other Bear", "{1}{G}", 2, 2))
	return b
}

// check asserts every kind row's verdict: true exactly for the kinds in
// want.
func (b *kindBoard) check(t *testing.T, want game.Answers) {
	t.Helper()
	moves := legal.EnumerateFor(b.g, b.seat.ID)
	for _, k := range answerKinds {
		wantVerdict(t, moves, k.String(), b.rows[k], legal.KindActivate, verdict(want.HasAny(k)))
	}
}

// attack declares `attacker` (a creature of the active player's) at
// `target` in declare attackers.
func (b *kindBoard) attack(t *testing.T, attacker, target uuid.UUID) {
	t.Helper()
	advanceTo(t, b.g, game.StepDeclareAttackers)
	if err := b.g.DeclareAttacker(attacker, target); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

// toBlockers moves to declare blockers, has each defender in `blocks`
// block (blocker → attacker) and every defender finish, so the active
// player has priority again (CR 509.2).
func (b *kindBoard) toBlockers(t *testing.T, blocks map[uuid.UUID]uuid.UUID, defenders ...*game.Player) {
	t.Helper()
	advanceTo(t, b.g, game.StepDeclareBlockers)
	for blocker, attacker := range blocks {
		if err := b.g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	for _, d := range defenders {
		if err := b.g.FinishBlocks(d.ID); err != nil {
			t.Fatalf("FinishBlocks(%s): %v", d.Name, err)
		}
	}
}

func TestAnswersNowByKind(t *testing.T) {
	murder := game.Card{Name: "Murder", TypeLine: "Instant", ManaCost: "{1}{B}{B}", OracleID: oracleMurder}
	bolt := game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt}
	counter := game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", OracleID: oracleCounterspell}
	wrath := game.Card{Name: "Wrath of God", TypeLine: "Sorcery", ManaCost: "{2}{W}{W}", OracleID: oracleWrathOfGodAW}
	opt := game.Card{Name: "Opt", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleOptVO}
	mystery := game.Card{Name: "Mystery Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{B}"}
	bears := game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	marauder := game.Card{Name: "Fleshbag Marauder", TypeLine: "Creature — Zombie Warrior", OracleID: oracleFleshbagMarauder, ManaCost: "{2}{B}", Power: 3, Toughness: 1}
	pestilence := game.Card{Name: "Pestilence", TypeLine: "Enchantment", OracleID: oraclePestilenceTG, ManaCost: "{2}{B}{B}"}

	// onTheirMain is an opponent's precombat main, the seat holding
	// priority with `push` on the stack.
	onTheirMain := func(noCreature bool, push func(t *testing.T, b *kindBoard)) func(t *testing.T) *kindBoard {
		return func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, noCreature)
			advanceTo(t, b.g, game.StepPrecombatMain)
			push(t, b)
			givePriority(b.g, b.seat)
			return b
		}
	}
	at := func(offset int, step game.Step) func(t *testing.T) *kindBoard {
		return func(t *testing.T) *kindBoard {
			b := newKindBoard(t, offset, false)
			advanceTo(t, b.g, step)
			givePriority(b.g, b.seat)
			return b
		}
	}
	attacker := func(b *kindBoard) uuid.UUID {
		return battlefieldCard(b.g, b.active, creature("Attacker", "{2}{G}", 3, 3))
	}

	cases := []struct {
		name  string
		build func(t *testing.T) *kindBoard
		want  game.Answers
	}{
		// The stack (Q4 (a)).
		{"an opponent's Murder on your creature", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, murder, game.TargetRef{Kind: game.TargetCard, ID: b.bear})
		}), saving | onSpell},
		{"an opponent's Murder on another player's creature", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, murder, game.TargetRef{Kind: game.TargetCard, ID: b.otherBear})
		}), onSpell},
		{"an opponent's Lightning Bolt at you", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, bolt, game.TargetRef{Kind: game.TargetPlayer, ID: b.seat.ID})
		}), game.AnswerProtect | game.AnswerPrevent | game.AnswerSacOutlet | onSpell},
		{"an opponent's Counterspell at your spell", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			mine := pushSpellAW(b.g, b.seat, bolt, game.TargetRef{Kind: game.TargetCard, ID: b.otherBear})
			pushSpellAW(b.g, b.active, counter, game.TargetRef{Kind: game.TargetCard, ID: mine})
		}), game.AnswerProtect | onSpell},
		{"an opponent's declared creature sweep, with a creature of yours", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, wrath)
		}), saving | onSpell},
		{"an opponent's declared creature sweep, with no creature of yours", onTheirMain(true, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, wrath)
		}), onSpell},
		{"an opponent's declared-value instant (Opt)", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, opt)
		}), onSpell},
		{"an opponent's undeclared untargeted sorcery", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, mystery)
		}), saving | onSpell},
		{"an opponent's creature spell", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.active, bears)
		}), onSpell},
		{"Fleshbag Marauder's edict, a trigger", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, marauder)
			ref, label := triggerRefAW(t, marauder, "each player sacrifices")
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, label, ref)
		}), saving},
		{"Pestilence's sweep, an activation", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, pestilence)
			ref, label := activatedRefAW(t, pestilence, "damage to each creature")
			pushAbilityAW(b.g, b.active, src, game.StackItemActivated, label, ref)
		}), saving},
		{"an opponent's untargeted ability that declares nothing", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{G}", 1, 1))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — something", nil)
		}), 0},
		{"an opponent's ability aimed at your creature", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{G}", 1, 1))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — damage", nil,
				game.TargetRef{Kind: game.TargetCard, ID: b.bear})
		}), saving},
		// The graveyard kind (Q5): an opponent's item that targets a card
		// in a graveyard, which exiling the graveyard answers (CR 608.2b).
		{"an opponent's reanimation trigger on a card in their graveyard (Eternal Witness)", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{1}{G}{G}", 2, 1))
			card := graveyardCard(b.active, creature("Fallen Bear", "{1}{G}", 2, 2))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — return target card from your graveyard to your hand", nil,
				game.TargetRef{Kind: game.TargetCard, ID: card})
		}), game.AnswerGraveyard},
		{"an opponent's reanimation trigger on a card in your graveyard", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{1}{B}{B}", 2, 1))
			card := graveyardCard(b.seat, creature("Fallen Bear", "{1}{G}", 2, 2))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — return target creature card from a graveyard to the battlefield", nil,
				game.TargetRef{Kind: game.TargetCard, ID: card})
		}), game.AnswerGraveyard},
		{"an opponent's reanimation spell on a card in a graveyard", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			card := graveyardCard(b.other, creature("Fallen Bear", "{1}{G}", 2, 2))
			pushSpellAW(b.g, b.active, game.Card{Name: "Reanimate", TypeLine: "Sorcery", ManaCost: "{B}"},
				game.TargetRef{Kind: game.TargetCard, ID: card})
		}), onSpell},
		{"an opponent's trigger on a graveyard card that has left it", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{1}{G}{G}", 2, 1))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — return target card from your graveyard to your hand", nil,
				game.TargetRef{Kind: game.TargetCard, ID: uuid.New()})
		}), 0},
		{"an opponent's trigger on a creature in play, with a graveyard beside it", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			graveyardCard(b.active, creature("Fallen Bear", "{1}{G}", 2, 2))
			src := battlefieldCard(b.g, b.active, creature("Trigger Source", "{G}", 1, 1))
			pushAbilityAW(b.g, b.active, src, game.StackItemTriggered, "Trigger Source — damage", nil,
				game.TargetRef{Kind: game.TargetCard, ID: b.otherBear})
		}), 0},
		{"your own reanimation trigger", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			src := battlefieldCard(b.g, b.seat, creature("Own Trigger Source", "{1}{G}{G}", 2, 1))
			card := graveyardCard(b.seat, creature("Fallen Bear", "{1}{G}", 2, 2))
			pushAbilityAW(b.g, b.seat, src, game.StackItemTriggered, "Own Trigger Source — return target card from your graveyard to your hand", nil,
				game.TargetRef{Kind: game.TargetCard, ID: card})
		}), 0},
		{"your own Murder", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.seat, murder, game.TargetRef{Kind: game.TargetCard, ID: b.otherBear})
		}), 0},
		{"your own sweep", onTheirMain(false, func(t *testing.T, b *kindBoard) {
			pushSpellAW(b.g, b.seat, wrath)
		}), 0},

		// Combat (Q1, Q4 (a)).
		{"an attack on you", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			b.attack(t, attacker(b), b.seat.ID)
			givePriority(b.g, b.seat)
			return b
		}, saving | game.AnswerCombatGrant | game.AnswerAnimate | game.AnswerMakesBlocker},
		{"an attack on your planeswalker", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			walker := battlefieldCard(b.g, b.seat, game.Card{
				Name: "Seat Walker", TypeLine: "Legendary Planeswalker — Test",
				Counters: map[string]int{game.CounterLoyalty: 4},
			})
			b.attack(t, attacker(b), walker)
			givePriority(b.g, b.seat)
			return b
		}, saving | game.AnswerCombatGrant | game.AnswerAnimate | game.AnswerMakesBlocker},
		{"an attack on a battle you protect", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			battle := battlefieldCard(b.g, b.other, game.Card{
				Name: "Test Siege", TypeLine: "Battle — Siege",
				ProtectorPlayerID: b.seat.ID, Counters: map[string]int{game.CounterDefense: 5},
			})
			b.attack(t, attacker(b), battle)
			givePriority(b.g, b.seat)
			return b
		}, saving | game.AnswerCombatGrant | game.AnswerAnimate | game.AnswerMakesBlocker},
		{"an attack on another player", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			b.attack(t, attacker(b), b.other.ID)
			givePriority(b.g, b.seat)
			return b
		}, 0},
		{"your attack", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 0, false)
			b.attack(t, b.bear, b.other.ID)
			return b
		}, saving | game.AnswerCombatGrant},
		{"declare blockers: your blocker", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			a := attacker(b)
			b.attack(t, a, b.seat.ID)
			b.toBlockers(t, map[uuid.UUID]uuid.UUID{b.bear: a}, b.seat)
			givePriority(b.g, b.seat)
			return b
		}, saving | game.AnswerCombatGrant},
		{"declare blockers: an unblocked attacker at you", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			b.attack(t, attacker(b), b.seat.ID)
			b.toBlockers(t, nil, b.seat)
			givePriority(b.g, b.seat)
			return b
		}, game.AnswerProtect | game.AnswerPrevent},
		{"declare blockers: your attacker, blocked", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 0, false)
			b.attack(t, b.bear, b.other.ID)
			b.toBlockers(t, map[uuid.UUID]uuid.UUID{b.otherBear: b.bear}, b.other)
			return b
		}, saving | game.AnswerCombatGrant},
		{"declare blockers: your attacker, unblocked", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 0, false)
			b.attack(t, b.bear, b.other.ID)
			b.toBlockers(t, nil, b.other)
			return b
		}, game.AnswerPump | game.AnswerPrevent | game.AnswerCombatGrant},
		{"declare blockers: an attack on another player", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			b.attack(t, attacker(b), b.other.ID)
			b.toBlockers(t, nil, b.other)
			givePriority(b.g, b.seat)
			return b
		}, 0},
		// Beginning of combat matches nothing on its own, even on your
		// turn (the owner's decision of 2026-10-31, #4734): a manland, a
		// Crew row or a keyword grant does not stop smart auto-pass there.
		// With nothing on the stack, an animate or combat_grant row there
		// carries no verdict instead of false, which
		// TestAnswersNowOwnBeginningOfCombat pins. An opponent's item on
		// the stack still matches, as in any step, and the two rows are
		// false beside it.
		{"beginning of combat, your turn, with an opponent's Murder on your creature", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 0, false)
			advanceTo(t, b.g, game.StepBeginCombat)
			pushSpellAW(b.g, b.other, murder, game.TargetRef{Kind: game.TargetCard, ID: b.bear})
			givePriority(b.g, b.seat)
			return b
		}, saving | onSpell},
		{"beginning of combat, an opponent's turn", at(1, game.StepBeginCombat), 0},
		{"the first-strike damage step: your blocker", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			a := battlefieldCard(b.g, b.active, game.Card{
				Name: "First Striker", TypeLine: "Creature — Knight", ManaCost: "{W}",
				Power: 1, Toughness: 1, Keywords: []string{"first strike"},
			})
			b.attack(t, a, b.seat.ID)
			b.toBlockers(t, map[uuid.UUID]uuid.UUID{b.bear: a}, b.seat)
			advanceTo(t, b.g, game.StepFirstStrikeDamage)
			givePriority(b.g, b.seat)
			return b
		}, saving | game.AnswerCombatGrant},
		{"the regular combat damage step after it", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 1, false)
			a := battlefieldCard(b.g, b.active, game.Card{
				Name: "First Striker", TypeLine: "Creature — Knight", ManaCost: "{W}",
				Power: 1, Toughness: 1, Keywords: []string{"first strike"},
			})
			b.attack(t, a, b.seat.ID)
			b.toBlockers(t, map[uuid.UUID]uuid.UUID{b.bear: a}, b.seat)
			advanceTo(t, b.g, game.StepCombatDamage)
			givePriority(b.g, b.seat)
			return b
		}, 0},

		// The turn's other steps.
		{"an opponent's upkeep", at(3, game.StepUpkeep), game.AnswerRestrict},
		{"an opponent's upkeep, with their trigger on the stack", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 3, false)
			advanceTo(t, b.g, game.StepUpkeep)
			upkeeper := b.g.Seats[b.g.Turn.ActiveSeat]
			src := battlefieldCard(b.g, upkeeper, creature("Upkeep Source", "{G}", 1, 1))
			pushAbilityAW(b.g, upkeeper, src, game.StackItemTriggered, "Upkeep Source — at the beginning of your upkeep", nil)
			givePriority(b.g, b.seat)
			return b
		}, game.AnswerRestrict},
		{"an opponent's upkeep, with their reanimation trigger on the stack", func(t *testing.T) *kindBoard {
			b := newKindBoard(t, 3, false)
			advanceTo(t, b.g, game.StepUpkeep)
			upkeeper := b.g.Seats[b.g.Turn.ActiveSeat]
			src := battlefieldCard(b.g, upkeeper, creature("Upkeep Source", "{G}", 1, 1))
			card := graveyardCard(upkeeper, creature("Fallen Bear", "{1}{G}", 2, 2))
			pushAbilityAW(b.g, upkeeper, src, game.StackItemTriggered, "Upkeep Source — return target card from your graveyard to your hand", nil,
				game.TargetRef{Kind: game.TargetCard, ID: card})
			givePriority(b.g, b.seat)
			return b
		}, game.AnswerRestrict | game.AnswerGraveyard},
		{"your own upkeep", at(1, game.StepUpkeep), 0},
		{"an opponent's draw step", at(1, game.StepDraw), 0},
		{"an opponent's end step", at(1, game.StepEnd), 0},
		{"your own main phase", at(0, game.StepPrecombatMain), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b := tc.build(t)
			if g := b.g; g.Seats[g.Turn.PriorityHolder].ID != b.seat.ID {
				t.Fatalf("the seat does not hold priority (holder %d, step %s)", g.Turn.PriorityHolder, g.Turn.Step)
			}
			b.check(t, tc.want)
		})
	}
}

// TestAnswersNowOwnBeginningOfCombat is the owner's decision of
// 2026-10-31 (#4734): animate and combat_grant no longer match the seat's
// own beginning of combat, so a manland, a Crew row or a keyword grant is
// no stop there every turn. With nothing of an opponent's on the stack the
// two kinds carry no verdict rather than false, so the move keeps the
// class it had before the amendment: a player who ticks the step still
// stops there through the stops grid, even with "only when I can do
// something" on, which a false verdict would end. Every other kind is
// false. At an opponent's beginning of combat, with an opponent's item on
// the stack (where only what that item threatens is true) and in the
// declare attackers step, the two kinds are false too.
func TestAnswersNowOwnBeginningOfCombat(t *testing.T) {
	both := game.AnswerCombatGrant | game.AnswerAnimate
	// check asserts each kind row: true for the kinds in wantTrue, absent
	// for the kinds in wantAbsent, false for the rest.
	check := func(t *testing.T, b *kindBoard, wantTrue, wantAbsent game.Answers) {
		t.Helper()
		moves := legal.EnumerateFor(b.g, b.seat.ID)
		for _, k := range answerKinds {
			want := verdict(wantTrue.HasAny(k))
			if wantAbsent.HasAny(k) {
				want = nil
			}
			wantVerdict(t, moves, k.String(), b.rows[k], legal.KindActivate, want)
		}
	}
	t.Run("your own beginning of combat", func(t *testing.T) {
		b := newKindBoard(t, 0, false)
		advanceTo(t, b.g, game.StepBeginCombat)
		givePriority(b.g, b.seat)
		check(t, b, 0, both)
	})
	t.Run("a row that declares animate and a kind that does not match", func(t *testing.T) {
		b := newKindBoard(t, 0, false)
		row := battlefieldCard(b.g, b.seat, declaredRelic(game.AnswerAnimate|game.AnswerRestrict))
		advanceTo(t, b.g, game.StepBeginCombat)
		givePriority(b.g, b.seat)
		wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "animate and restrict", row, legal.KindActivate, nil)
	})
	t.Run("an opponent's beginning of combat", func(t *testing.T) {
		b := newKindBoard(t, 1, false)
		advanceTo(t, b.g, game.StepBeginCombat)
		givePriority(b.g, b.seat)
		check(t, b, 0, 0)
	})
	t.Run("your own beginning of combat, an opponent's Murder on the stack", func(t *testing.T) {
		b := newKindBoard(t, 0, false)
		advanceTo(t, b.g, game.StepBeginCombat)
		pushSpellAW(b.g, b.other, game.Card{Name: "Murder", TypeLine: "Instant", ManaCost: "{1}{B}{B}", OracleID: oracleMurder},
			game.TargetRef{Kind: game.TargetCard, ID: b.bear})
		givePriority(b.g, b.seat)
		check(t, b, saving|onSpell, 0)
	})
	t.Run("your own beginning of combat, your own spell on the stack", func(t *testing.T) {
		b := newKindBoard(t, 0, false)
		advanceTo(t, b.g, game.StepBeginCombat)
		pushSpellAW(b.g, b.seat, game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
		givePriority(b.g, b.seat)
		check(t, b, 0, both)
	})
	t.Run("your own declare attackers step with no attack", func(t *testing.T) {
		b := newKindBoard(t, 0, false)
		advanceTo(t, b.g, game.StepDeclareAttackers)
		givePriority(b.g, b.seat)
		check(t, b, 0, 0)
	})
}

// TestAnswersNowCounter: a move that targets an item on the stack is a
// counter, which matches an opponent's item there whatever its row
// declares, and not a stack that is all the seat's own (#323's). The
// row here declares combat_grant only, so nothing else matches outside
// combat. Counterspell, which declares remove (#4773), keeps today's
// class.
func TestAnswersNowCounter(t *testing.T) {
	build := func(t *testing.T, theirs bool) (*kindBoard, uuid.UUID, uuid.UUID) {
		b := newKindBoard(t, 1, false)
		advanceTo(t, b.g, game.StepPrecombatMain)
		row := battlefieldCard(b.g, b.seat, game.Card{
			Name: "Counter Relic", TypeLine: "Artifact",
			ActivatedAbilities: []game.ActivatedAbilityShape{{
				Label:   "Counter target spell.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Targets: effects.TargetSpell("target spell"),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		})
		for range 2 {
			battlefieldCard(b.g, b.seat, basic("Island", "Island"))
		}
		cs := handCard(b.seat, game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", OracleID: oracleCounterspell})
		caster := b.seat
		if theirs {
			caster = b.active
		}
		pushSpellAW(b.g, caster, game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
		givePriority(b.g, b.seat)
		return b, row, cs
	}
	t.Run("an opponent's spell", func(t *testing.T) {
		b, row, cs := build(t, true)
		moves := legal.EnumerateFor(b.g, b.seat.ID)
		wantVerdict(t, moves, "Counter Relic", row, legal.KindActivate, verdict(true))
		wantVerdict(t, moves, "Counterspell (remove)", cs, legal.KindCast, nil)
	})
	t.Run("your own spell", func(t *testing.T) {
		b, row, cs := build(t, false)
		moves := legal.EnumerateFor(b.g, b.seat.ID)
		wantVerdict(t, moves, "Counter Relic", row, legal.KindActivate, verdict(false))
		wantVerdict(t, moves, "Counterspell (remove)", cs, legal.KindCast, nil)
	})
}

// TestAnswersNowUnion: a move that declares several kinds matches when
// one of them does (ADR 0142 sweep ruling 4). A row declaring survive
// and restrict matches in either one's window, and in neither's it does
// not.
func TestAnswersNowUnion(t *testing.T) {
	both := game.AnswerSurvive | game.AnswerRestrict
	build := func(t *testing.T, offset int, step game.Step) (*kindBoard, uuid.UUID) {
		b := newKindBoard(t, offset, false)
		row := battlefieldCard(b.g, b.seat, declaredRelic(both))
		advanceTo(t, b.g, step)
		return b, row
	}
	cases := []struct {
		name   string
		offset int
		step   game.Step
		push   func(b *kindBoard)
		want   bool
	}{
		{"an opponent's upkeep (restrict)", 3, game.StepUpkeep, nil, true},
		{"an opponent's ability aimed at your creature (survive)", 1, game.StepPrecombatMain, func(b *kindBoard) {
			pushAbilityAW(b.g, b.active, b.otherBear, game.StackItemTriggered, "Other Bear — damage", nil,
				game.TargetRef{Kind: game.TargetCard, ID: b.bear})
		}, true},
		{"an opponent's Murder on your creature (both)", 1, game.StepPrecombatMain, func(b *kindBoard) {
			pushSpellAW(b.g, b.active, game.Card{Name: "Murder", TypeLine: "Instant", OracleID: oracleMurder},
				game.TargetRef{Kind: game.TargetCard, ID: b.bear})
		}, true},
		{"an opponent's untargeted ability (neither)", 1, game.StepPrecombatMain, func(b *kindBoard) {
			pushAbilityAW(b.g, b.active, b.otherBear, game.StackItemTriggered, "Other Bear — something", nil)
		}, false},
		{"an opponent's end step (neither)", 1, game.StepEnd, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, row := build(t, tc.offset, tc.step)
			if tc.push != nil {
				tc.push(b)
			}
			givePriority(b.g, b.seat)
			wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "survive|restrict", row, legal.KindActivate, verdict(tc.want))
		})
	}
}

// TestAnswersNowDeclaredTargetedRow is Q2 (a), ahead of slice 7: a
// targeted row that declares is judged by its kind, and still carries
// has_targets. One that declares nothing keeps today's class.
func TestAnswersNowDeclaredTargetedRow(t *testing.T) {
	targeted := func(name string, a game.Answers) game.Card {
		return game.Card{
			Name: name, TypeLine: "Artifact",
			ActivatedAbilities: []game.ActivatedAbilityShape{{
				Label:   name + ": target creature you control gains protection.",
				Purpose: game.Purpose{Answers: a},
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect:  func(*game.Game, *game.StackItem) error { return nil },
			}},
		}
	}
	for _, tc := range []struct {
		name   string
		threat bool
		want   bool
	}{
		{"an opponent's Murder on your creature", true, true},
		{"an opponent's Murder on another player's creature", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := newKindBoard(t, 1, false)
			declared := battlefieldCard(b.g, b.seat, targeted("Declared Shield", game.AnswerProtect))
			undeclared := battlefieldCard(b.g, b.seat, targeted("Undeclared Shield", 0))
			advanceTo(t, b.g, game.StepPrecombatMain)
			victim := b.otherBear
			if tc.threat {
				victim = b.bear
			}
			pushSpellAW(b.g, b.active, game.Card{Name: "Murder", TypeLine: "Instant", OracleID: oracleMurder},
				game.TargetRef{Kind: game.TargetCard, ID: victim})
			givePriority(b.g, b.seat)
			moves := legal.EnumerateFor(b.g, b.seat.ID)
			wantVerdict(t, moves, "Declared Shield", declared, legal.KindActivate, verdict(tc.want))
			wantVerdict(t, moves, "Undeclared Shield", undeclared, legal.KindActivate, nil)
			for _, m := range movesFrom(moves, declared, legal.KindActivate) {
				if !m.HasTargets || m.Interacts {
					t.Errorf("Declared Shield: has_targets=%v interacts=%v, want true/false (%s)", m.HasTargets, m.Interacts, m.Label)
				}
			}
		})
	}
}

// TestAnswersNowSelfSacrifice is ADR 0142 owner answer 4 as a match:
// Ranger-Captain of Eos, which declares restrict, matches at any step
// while an opponent's item targets it, and while it attacks or blocks
// where combat counts. An opponent's ability is used where Swords to
// Plowshares, a spell, would match restrict anyway.
func TestAnswersNowSelfSacrifice(t *testing.T) {
	rc := game.Card{Name: "Ranger-Captain of Eos", TypeLine: "Creature — Human Soldier Ranger", OracleID: oracleRangerCaptainAW, ManaCost: "{1}{W}{W}", Power: 3, Toughness: 3}
	swords := game.Card{Name: "Swords to Plowshares", TypeLine: "Instant", ManaCost: "{W}", OracleID: oracleSwordsAW}
	for _, step := range []game.Step{game.StepDraw, game.StepPrecombatMain, game.StepBeginCombat, game.StepEnd} {
		t.Run("Swords to Plowshares on it, "+string(step), func(t *testing.T) {
			b := newKindBoard(t, 1, false)
			captain := battlefieldCard(b.g, b.seat, rc)
			advanceTo(t, b.g, step)
			pushSpellAW(b.g, b.active, swords, game.TargetRef{Kind: game.TargetCard, ID: captain})
			givePriority(b.g, b.seat)
			wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "Ranger-Captain", captain, legal.KindActivate, verdict(true))
		})
		t.Run("an opponent's ability on it, "+string(step), func(t *testing.T) {
			b := newKindBoard(t, 1, false)
			captain := battlefieldCard(b.g, b.seat, rc)
			advanceTo(t, b.g, step)
			pushAbilityAW(b.g, b.active, b.otherBear, game.StackItemTriggered, "Other Bear — damage", nil,
				game.TargetRef{Kind: game.TargetCard, ID: captain})
			givePriority(b.g, b.seat)
			wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "Ranger-Captain", captain, legal.KindActivate, verdict(true))
		})
		t.Run("an opponent's ability on another creature, "+string(step), func(t *testing.T) {
			b := newKindBoard(t, 1, false)
			captain := battlefieldCard(b.g, b.seat, rc)
			advanceTo(t, b.g, step)
			pushAbilityAW(b.g, b.active, b.otherBear, game.StackItemTriggered, "Other Bear — damage", nil,
				game.TargetRef{Kind: game.TargetCard, ID: b.bear})
			givePriority(b.g, b.seat)
			wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "Ranger-Captain", captain, legal.KindActivate, verdict(false))
		})
	}
	t.Run("attacking", func(t *testing.T) {
		b := newKindBoard(t, 0, false)
		captain := battlefieldCard(b.g, b.seat, rc)
		b.attack(t, captain, b.other.ID)
		wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "Ranger-Captain", captain, legal.KindActivate, verdict(true))
	})
	t.Run("blocking", func(t *testing.T) {
		b := newKindBoard(t, 1, false)
		captain := battlefieldCard(b.g, b.seat, rc)
		a := battlefieldCard(b.g, b.active, creature("Attacker", "{2}{G}", 3, 3))
		b.attack(t, a, b.seat.ID)
		b.toBlockers(t, map[uuid.UUID]uuid.UUID{captain: a}, b.seat)
		givePriority(b.g, b.seat)
		wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "Ranger-Captain", captain, legal.KindActivate, verdict(true))
	})
	t.Run("not in combat, an attack on another player", func(t *testing.T) {
		b := newKindBoard(t, 1, false)
		captain := battlefieldCard(b.g, b.seat, rc)
		b.attack(t, battlefieldCard(b.g, b.active, creature("Attacker", "{2}{G}", 3, 3)), b.other.ID)
		givePriority(b.g, b.seat)
		wantVerdict(t, legal.EnumerateFor(b.g, b.seat.ID), "Ranger-Captain", captain, legal.KindActivate, verdict(false))
	})
}

// TestAnswersNowCasts: a cast reads the announcement's declaration
// (castAnswersOf). Altar's Reap declares sac_outlet, and matches like
// one; a cast that declares nothing (Fog), exactly value (Opt) or remove
// (Lightning Bolt) carries no verdict.
func TestAnswersNowCasts(t *testing.T) {
	for _, tc := range []struct {
		name   string
		threat bool
	}{
		{"an opponent's Murder on your creature", true},
		{"an opponent's creature spell", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := newKindBoard(t, 1, false)
			for range 2 {
				battlefieldCard(b.g, b.seat, basic("Swamp", "Swamp"))
			}
			battlefieldCard(b.g, b.seat, basic("Island", "Island"))
			battlefieldCard(b.g, b.seat, basic("Mountain", "Mountain"))
			battlefieldCard(b.g, b.seat, basic("Forest", "Forest"))
			reap := handCard(b.seat, game.Card{Name: "Altar's Reap", TypeLine: "Instant", ManaCost: "{1}{B}", OracleID: oracleAltarsReapVO})
			opt := handCard(b.seat, game.Card{Name: "Opt", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleOptVO})
			bolt := handCard(b.seat, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracleLightningBolt})
			fog := handCard(b.seat, game.Card{Name: "Fog", TypeLine: "Instant", ManaCost: "{G}", OracleID: oracleFogVO})
			advanceTo(t, b.g, game.StepPrecombatMain)
			if tc.threat {
				pushSpellAW(b.g, b.active, game.Card{Name: "Murder", TypeLine: "Instant", OracleID: oracleMurder},
					game.TargetRef{Kind: game.TargetCard, ID: b.bear})
			} else {
				pushSpellAW(b.g, b.active, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
			}
			givePriority(b.g, b.seat)
			moves := legal.EnumerateFor(b.g, b.seat.ID)
			wantVerdict(t, moves, "Altar's Reap (sac_outlet)", reap, legal.KindCast, verdict(tc.threat))
			wantVerdict(t, moves, "Opt (value)", opt, legal.KindCast, nil)
			wantVerdict(t, moves, "Lightning Bolt (remove)", bolt, legal.KindCast, nil)
			wantVerdict(t, moves, "Fog (undeclared)", fog, legal.KindCast, nil)
		})
	}
}

// TestAnswersNowAbsentKeepsTodaysClass: the moves the match does not
// judge carry no answers_now, and their old flags say what they said
// before it (the flags' own tests, threat_gate_test.go,
// combat_interacts_test.go, value_only_test.go and
// interacts_internal_test.go, pin the rest). An opponent's creature
// spell is on the stack, which no kind on this board answers.
func TestAnswersNowAbsentKeepsTodaysClass(t *testing.T) {
	b := newKindBoard(t, 1, false)
	for range 3 {
		battlefieldCard(b.g, b.seat, basic("Island", "Island"))
	}
	for range 4 {
		battlefieldCard(b.g, b.seat, basic("Swamp", "Swamp"))
	}
	battlefieldCard(b.g, b.seat, basic("Forest", "Forest"))
	undeclaredRow := battlefieldCard(b.g, b.seat, game.Card{
		Name: "Instance Relic", TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Do something.",
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	undeclaredTargeted := battlefieldCard(b.g, b.seat, game.Card{
		Name: "Instance Wand", TypeLine: "Artifact",
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:   "Target creature: do something.",
			Targets: effects.TargetCreature("target creature"),
			Effect:  func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	pestilence := battlefieldCard(b.g, b.seat, game.Card{Name: "Pestilence", TypeLine: "Enchantment", OracleID: oraclePestilenceTG, ManaCost: "{2}{B}{B}"})
	stone := battlefieldCard(b.g, b.seat, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}", OracleID: oracleMindStoneHT})
	opt := handCard(b.seat, game.Card{Name: "Opt", TypeLine: "Instant", ManaCost: "{U}", OracleID: oracleOptVO})
	cs := handCard(b.seat, game.Card{Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", OracleID: oracleCounterspell})
	fog := handCard(b.seat, game.Card{Name: "Fog", TypeLine: "Instant", ManaCost: "{G}", OracleID: oracleFogVO})
	advanceTo(t, b.g, game.StepPrecombatMain)
	pushSpellAW(b.g, b.active, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	givePriority(b.g, b.seat)
	moves := legal.EnumerateFor(b.g, b.seat.ID)

	wantVerdict(t, moves, "an undeclared cast (Fog)", fog, legal.KindCast, nil)
	wantVerdict(t, moves, "a remove cast (Counterspell)", cs, legal.KindCast, nil)
	wantVerdict(t, moves, "a value cast (Opt)", opt, legal.KindCast, nil)
	wantVerdict(t, moves, "a run-time row that declares nothing", undeclaredRow, legal.KindActivate, nil)
	wantVerdict(t, moves, "a targeted row that declares nothing", undeclaredTargeted, legal.KindActivate, nil)
	wantVerdict(t, moves, "a remove row (Pestilence)", pestilence, legal.KindActivate, nil)
	wantVerdict(t, moves, "a value row (Mind Stone's draw)", stone, legal.KindActivate, nil)
	wantVerdict(t, moves, "a value mana ability (Mind Stone)", stone, legal.KindMana, nil)
	for _, m := range moves {
		switch {
		case m.Source == cs && !m.TargetsStack:
			t.Errorf("Counterspell: targets_stack unset (%s)", m.Label)
		case m.Source == opt && !m.ValueOnly:
			t.Errorf("Opt: value_only unset (%s)", m.Label)
		case m.Source == undeclaredRow && (!m.Interacts || m.CombatInteracts):
			t.Errorf("the undeclared row: interacts=%v combat_interacts=%v, want true/false", m.Interacts, m.CombatInteracts)
		case m.Source == undeclaredTargeted && (!m.HasTargets || m.Interacts):
			t.Errorf("the undeclared targeted row: has_targets=%v interacts=%v, want true/false", m.HasTargets, m.Interacts)
		case m.Source == pestilence && m.Kind == legal.KindActivate && !m.Interacts:
			t.Errorf("Pestilence: interacts unset (%s)", m.Label)
		case m.Source == stone && (m.Interacts || m.CombatInteracts):
			t.Errorf("Mind Stone: interacts=%v combat_interacts=%v, want neither (%s)", m.Interacts, m.CombatInteracts, m.Label)
		}
		if m.Kind == legal.KindPass && m.AnswersNow != nil {
			t.Errorf("pass carries answers_now")
		}
	}
	// The per-kind rows keep their #3555 flags: no threat on the stack,
	// so the saving kinds are combat answers and restrict interacts.
	for _, k := range answerKinds {
		for _, m := range movesFrom(moves, b.rows[k], legal.KindActivate) {
			wantInteracts := k == game.AnswerRestrict || k == game.AnswerGraveyard
			if m.Interacts != wantInteracts || m.CombatInteracts == wantInteracts {
				t.Errorf("%s: interacts=%v combat_interacts=%v, want %v/%v", k, m.Interacts, m.CombatInteracts, wantInteracts, !wantInteracts)
			}
		}
	}
}
