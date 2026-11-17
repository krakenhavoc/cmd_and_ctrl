package game

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// dice_roll.go — a dice roll is a replaceable event that can pause for a
// choice (#4307, ADR 0208's 2026-10-26 amendment, CR 706).
//
// # The card that needs it
//
// Squid Fire Knight's first ability sticker: "{T}: The next time target
// player would roll one or more dice this turn, instead they roll that many
// dice plus one, then you choose one of those rolls to ignore." The roll is
// replaced before any die is rolled (CR 614.1a), and the choice is the
// sticker's controller's, made after seeing the dice. An ignored roll "is
// considered to have never happened. No abilities trigger because of the
// ignored roll, and no effects apply to that roll" (CR 706.6).
//
// # The one door
//
// Every card's roll goes through RollDiceThenForEffect: a DiceRoll names
// the roller, the source, the kind and number of dice, how many of the
// lowest the card itself ignores (Berserker's Frenzy), and the rest of the
// effect as a registered tail (KeyedTail), never a closure. The tail is
// told the kept rolls (TailOutcome.Rolls). With nothing to ask it runs at
// once, so a roll without the sticker behaves exactly as it did.
//
//  1. The window (RepEventRollDice) opens through the ordinary replacement
//     pipeline. Squid Fire Knight's shield is a ScopedEffect record
//     (ModRollExtraDieIgnoreOne, ADR 0041 P8): it adds a die, appends its
//     controller as a chooser, and is spent (a "next time" shield). The
//     window settles inline (mustSettleNow): a CR 616.1 order among these
//     replacements is only the order of their choosers, and the flow below
//     asks the roller that itself.
//  2. Order (dice/order-replacements): when the applied shields name two or
//     more different players, the roller (the affected player, CR 616.1)
//     chooses which applies first, one answer per remaining player.
//  3. Roll: every die is drawn from the roll stream at once, in order.
//  4. Ignore (dice/ignore-roll): each chooser in turn picks one of the
//     rolls still in play to ignore.
//  5. The card's own "ignore the lowest" (CR 706.6) together with every
//     static "plus one and ignore the lowest roll" that applied in the
//     window (Pixie Guide, #5681: RollOneMoreAndIgnoreLowest), then one
//     EventRollDie per kept roll in one batch, each with its natural
//     result and the instruction's modifier (CR 706.2,
//     DiceRoll.Modifier), then the tail.
//
// Each question is a keyed option pick, and the paused roll rides it as
// plain data (PendingChoice.dice, written as pendingChoiceSnapshot.DiceRoll),
// so a table waiting on either is a restore point.
//
// The starting player's roll (opening_roll.go) is not an effect's
// instruction during a turn and keeps the unreplaceable door,
// RollDiceForEffect; so do the tests that pin the roll stream.

// ModRollExtraDieIgnoreOne is "the next time <Player> would roll one or
// more dice this turn, instead they roll that many dice plus one, then you
// choose one of those rolls to ignore" (Squid Fire Knight's sticker). The
// record's controller is "you". One-use: the record is dropped as it
// applies. An on-disk identity.
const ModRollExtraDieIgnoreOne ModKind = "rollExtraDieIgnoreOne"

// The dice flow's keyed prompts. On-disk identities: never renamed, never
// reused.
const (
	diceOrderKey  = "dice/order-replacements"
	diceIgnoreKey = "dice/ignore-roll"
)

var (
	diceOrderThen  OptionPickThen
	diceIgnoreThen OptionPickThen
	// rollToVisitTail is the rest of "roll to visit your Attractions"
	// (CR 701.52a): the visits, once the die is settled.
	rollToVisitTail TailRef
)

func init() {
	diceOrderThen = RegisterOptionPickThen(diceOrderKey, diceOrderAnswered)
	diceIgnoreThen = RegisterOptionPickThen(diceIgnoreKey, diceIgnoreAnswered)
	rollToVisitTail = RegisterTail("dice/roll-to-visit", rollToVisitSettled)
}

// DiceRoll is one instruction to roll dice (CR 706.1).
type DiceRoll struct {
	// Player rolls; Source is the card whose effect it is. Together they
	// name the roll stream (RandomDraw).
	Player, Source uuid.UUID
	// Sides is the kind of die, N how many.
	Sides, N int
	// IgnoreLowest is the card's own "ignore the lowest" (Berserker's
	// Frenzy's "roll two d20 and ignore the lower roll"): at least 0 and
	// less than N.
	IgnoreLowest int
	// Then is the rest of the effect, told the kept rolls.
	Then KeyedTail
	// Planar is the planar die (CR 901.3a, ADR 0216 §4): a d6 drawn from
	// its own stream whose kept roll is announced as a face
	// (EventRollDie.PlanarFace) with no number (CR 901.9d). The tail is
	// still told the drawn 1..6.
	Planar bool
	// Modifier is what the instruction adds to (or, negative, subtracts
	// from) each die's natural result (CR 706.2): Diviner's Portent's
	// "roll a d20 and add the number of cards in your hand", The Deck of
	// Many Things' "subtract the number of cards in your hand". The card
	// reads the number as the roll is made. Every kept die is announced
	// with its natural result and this beside it (EventRollDie.Amount and
	// .Modifier), so "a natural 20" reads the face and "if any of those
	// results was 10 or higher" reads the sum, and the tail is told both
	// (TailOutcome.Rolls, .RollModifier). A planar die has no numerical
	// result (CR 901.9d) and carries none.
	Modifier int
}

// diceShield is one applied "then you choose one of those rolls to
// ignore": who chooses, and the shield's source and label for the prompts.
type diceShield struct {
	Chooser uuid.UUID `json:"chooser"`
	Source  uuid.UUID `json:"source,omitempty"`
	Label   string    `json:"label,omitempty"`
}

// diceRollPause is a roll between its window and its tail. Before the dice
// are rolled (Rolls empty), Pending are the shields still to be ordered and
// Ordered those already ordered; after, Ordered are the choosers still to
// answer, the first being the one asked now. Never mutated once queued.
type diceRollPause struct {
	Roll    DiceRoll
	Rolls   []int
	Pending []diceShield
	Ordered []diceShield
}

// RollDiceThenForEffect rolls the dice of r and runs r.Then with the kept
// rolls, now or after the questions a replacement asks (see the file
// comment). ErrInvalidParam for a roll that is not one.
//
// Caller must hold g.mu (write).
func (g *Game) RollDiceThenForEffect(r DiceRoll) error {
	if r.Sides <= 0 || r.N <= 0 || r.IgnoreLowest < 0 || r.IgnoreLowest >= r.N {
		return ErrInvalidParam
	}
	ev := &ReplacementEvent{
		Kind: RepEventRollDice, Actor: r.Player, Source: r.Source,
		RollPlayer: r.Player, RollSides: r.Sides, RollCount: r.N, RollPlanar: r.Planar,
		mustSettleNow: true,
	}
	out, err := g.applyReplacementsLocked(ev)
	g.clearReplacementEventLocked(ev.ID)
	if err != nil {
		// The iteration cap: the event as it stands.
		out = ev
	}
	if out == nil {
		// Nothing replaces a roll with nothing today; a cancelled roll
		// rolled no dice.
		return g.runKeyedRollTailLocked(r.Then, nil, 0)
	}
	r.N = out.RollCount
	// A static "plus one and ignore the lowest roll" (Pixie Guide) is the
	// instruction's own ignore, one more for each that applied.
	r.IgnoreLowest += out.rollIgnoreLowest
	return g.continueDiceRollLocked(diceRollPause{Roll: r, Pending: append([]diceShield(nil), out.rollShields...)})
}

// rollExtraDieAppliesLocked is ModRollExtraDieIgnoreOne's AppliesTo: a roll
// by the named player.
func rollExtraDieAppliesLocked(m Mod, ev *ReplacementEvent) bool {
	return ev.Kind == RepEventRollDice && ev.RollCount > 0 && m.Player != uuid.Nil && ev.RollPlayer == m.Player
}

// applyRollExtraDieLocked is its Replace: one more die, its controller a
// chooser, and the record spent ("the next time").
//
// Caller must hold g.mu (write).
func (g *Game) applyRollExtraDieLocked(e ScopedEffect, ev *ReplacementEvent) {
	ev.RollCount++
	shields := make([]diceShield, 0, len(ev.rollShields)+1)
	shields = append(shields, ev.rollShields...)
	label := e.Label
	if label == "" {
		label = e.SourceName
	}
	ev.rollShields = append(shields, diceShield{Chooser: e.Controller, Source: e.Source.ID, Label: label})
	seq := e.Seq
	g.dropScopedEffectsLocked(func(_ int, x ScopedEffect) bool { return x.Seq == seq })
}

// RollOneMoreAndIgnoreOne is a static replacement's "instead roll that
// many dice plus one and ignore one" (Ichor Elixir, #5648; CR 614.1a,
// 706.6) applied to a dice roll's window: one more die, and `chooser`
// picks one of the rolls to ignore once they are rolled, through the same
// keyed question Squid Fire Knight's shield asks. `source` and `label`
// name the replacement on that question. Unlike the shield nothing is
// spent: the replacement applies to every roll while it exists, once per
// roll (CR 614.5).
func (ev *ReplacementEvent) RollOneMoreAndIgnoreOne(chooser, source uuid.UUID, label string) {
	if ev.Kind != RepEventRollDice || ev.RollCount <= 0 || chooser == uuid.Nil {
		return
	}
	ev.RollCount++
	shields := make([]diceShield, 0, len(ev.rollShields)+1)
	shields = append(shields, ev.rollShields...)
	ev.rollShields = append(shields, diceShield{Chooser: chooser, Source: source, Label: label})
}

// RollOneMoreAndIgnoreLowest is a static replacement's "instead roll that
// many dice plus one and ignore the lowest roll" (Pixie Guide, Barbarian
// Class, #5681; CR 614.1a, 706.6) applied to a dice roll's window: one
// more die, and one more of the lowest rolls ignored once they are
// rolled. Nobody is asked: rolls tied for the lowest show the same number
// and carry the same modifier, so which of them goes is unobservable.
// Each such replacement applies once per roll (CR 614.5), so two of them
// roll two more dice and ignore the two lowest. A roll a player chose to
// ignore (RollOneMoreAndIgnoreOne) goes first, then the lowest of the
// rest; whichever order the roller would pick (CR 616.1), the same sets
// of kept rolls are reachable. The planar die has no numerical result, so
// no roll of it is the lowest (CR 706.7): a roll of planar dice is left
// alone.
func (ev *ReplacementEvent) RollOneMoreAndIgnoreLowest() {
	if ev.Kind != RepEventRollDice || ev.RollCount <= 0 || ev.RollPlanar {
		return
	}
	ev.RollCount++
	ev.rollIgnoreLowest++
}

// RollExtraDieNextTimeForEffect registers Squid Fire Knight's shield: the
// next time `player` would roll one or more dice this turn, they roll one
// more, then `controller` chooses one of the rolls to ignore. Reports
// whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) RollExtraDieNextTimeForEffect(source, controller, player uuid.UUID, label string) bool {
	if player == uuid.Nil || controller == uuid.Nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(source, ScopeGame, controller,
		[]Mod{{Kind: ModRollExtraDieIgnoreOne, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// continueDiceRollLocked takes the roll as far as it goes without an
// answer: order, roll, ignore, then finish.
//
// Caller must hold g.mu (write).
func (g *Game) continueDiceRollLocked(p diceRollPause) error {
	if len(p.Rolls) == 0 {
		if players := distinctChoosers(p.Pending); len(players) >= 2 && g.playerInGameLocked(p.Roll.Player) {
			return g.askDiceOrderLocked(p, players)
		}
		p.Ordered = append(append([]diceShield(nil), p.Ordered...), p.Pending...)
		p.Pending = nil
		kind := rngStreamRoll
		if p.Roll.Planar {
			// ADR 0216 §4: a stream of its own, so a planar roll never
			// moves a card's d20.
			kind = rngStreamPlanarDie
		}
		r := g.randForLocked(rngStream{kind: kind, player: p.Roll.Player, source: p.Roll.Source})
		p.Rolls = make([]int, p.Roll.N)
		for i := range p.Rolls {
			p.Rolls[i] = r.IntN(p.Roll.Sides) + 1
		}
	}
	for len(p.Ordered) > 0 && len(p.Rolls) > 1 {
		s := p.Ordered[0]
		if g.playerInGameLocked(s.Chooser) {
			return g.askDiceIgnoreLocked(p)
		}
		// A chooser who has left the game chooses nothing: the die the
		// replacement added, the last one rolled, is the one ignored.
		p.Rolls = p.Rolls[: len(p.Rolls)-1 : len(p.Rolls)-1]
		p.Ordered = p.Ordered[1:]
	}
	return g.finishDiceRollLocked(p.Roll, p.Rolls)
}

// finishDiceRollLocked applies the card's own "ignore the lowest", announces
// the kept rolls (one EventRollDie each, in one batch, so "whenever you roll
// one or more dice" triggers once) and runs the tail.
//
// Caller must hold g.mu (write).
func (g *Game) finishDiceRollLocked(r DiceRoll, rolls []int) error {
	kept := rolls
	if r.IgnoreLowest > 0 && r.IgnoreLowest < len(rolls) {
		kept = keepAllButLowest(rolls, r.IgnoreLowest)
	}
	batch := g.eventSeq + 1
	for _, result := range kept {
		ev := Event{Kind: EventRollDie, Actor: r.Player, Source: r.Source, Amount: result, Modifier: r.Modifier, Sides: r.Sides, BatchSeq: batch}
		if r.Planar {
			// CR 901.9d: the planar die has no numerical result.
			ev.Amount, ev.Modifier, ev.PlanarFace = 0, 0, planarFaceOf(result)
		}
		g.EmitEvent(ev)
	}
	modifier := r.Modifier
	if r.Planar {
		modifier = 0
	}
	return g.runKeyedRollTailLocked(r.Then, kept, modifier)
}

// RollResult is an EventRollDie's result: the natural result the die shows
// (Amount) plus the instruction's modifier (CR 706.2, #5681). Zero for a
// planar die, which has no numerical result (CR 901.9d).
func (ev Event) RollResult() int {
	if ev.Kind != EventRollDie || ev.PlanarFace != "" {
		return 0
	}
	return ev.Amount + ev.Modifier
}

// distinctChoosers lists the different choosers of shields, in order of
// first appearance.
func distinctChoosers(shields []diceShield) []uuid.UUID {
	var out []uuid.UUID
	for _, s := range shields {
		seen := false
		for _, id := range out {
			if id == s.Chooser {
				seen = true
				break
			}
		}
		if !seen {
			out = append(out, s.Chooser)
		}
	}
	return out
}

// askDiceOrderLocked asks the roller which of the remaining players'
// replacements applies first (CR 616.1).
//
// Caller must hold g.mu (write).
func (g *Game) askDiceOrderLocked(p diceRollPause, players []uuid.UUID) error {
	options := make([]ChoiceOption, 0, len(players))
	for _, id := range players {
		label := ""
		for _, s := range p.Pending {
			if s.Chooser == id {
				label = s.Label
				break
			}
		}
		options = append(options, ChoiceOption{Label: fmt.Sprintf("%s's %s first (they choose a roll to ignore first)", g.playerNameLocked(id), label)})
	}
	pause := p
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  p.Roll.Player,
		Source:   p.Roll.Source,
		Question: "Your next dice roll is replaced more than once. Which replacement applies first?",
		Options:  options,
		ThenKey:  diceOrderThen,
		dice:     &pause,
	})
	return nil
}

// diceOrderAnswered moves the first shield of the chosen player into the
// order and goes on. Unanswered (the roller has left), the order stands as
// the shields applied.
func diceOrderAnswered(g *Game, r OptionPicked) error {
	if r.dice == nil {
		return nil
	}
	p := *r.dice
	players := distinctChoosers(p.Pending)
	if r.Index < 0 || r.Index >= len(players) {
		p.Ordered = append(append([]diceShield(nil), p.Ordered...), p.Pending...)
		p.Pending = nil
		return g.continueDiceRollLocked(p)
	}
	pick := players[r.Index]
	pending := make([]diceShield, 0, len(p.Pending))
	ordered := append([]diceShield(nil), p.Ordered...)
	taken := false
	for _, s := range p.Pending {
		if !taken && s.Chooser == pick {
			ordered = append(ordered, s)
			taken = true
			continue
		}
		pending = append(pending, s)
	}
	p.Pending, p.Ordered = pending, ordered
	return g.continueDiceRollLocked(p)
}

// askDiceIgnoreLocked asks the first chooser which roll to ignore.
//
// Caller must hold g.mu (write).
func (g *Game) askDiceIgnoreLocked(p diceRollPause) error {
	s := p.Ordered[0]
	options := make([]ChoiceOption, 0, len(p.Rolls))
	for i, v := range p.Rolls {
		shown := strconv.Itoa(v)
		if p.Roll.Planar {
			shown = planarFaceOf(v)
		}
		options = append(options, ChoiceOption{Label: "Ignore the " + shown + " (die " + strconv.Itoa(i+1) + ")"})
	}
	dice := "d" + strconv.Itoa(p.Roll.Sides)
	if p.Roll.Planar {
		// CR 901.3a: the planar die is not named by its sides.
		dice = "planar dice"
	}
	pause := p
	source := s.Source
	if source == uuid.Nil {
		source = p.Roll.Source
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:    s.Chooser,
		FromPlayer: s.Chooser,
		Source:     source,
		Question: fmt.Sprintf("%s — %s rolled %d %s. Choose one of those rolls to ignore.",
			s.Label, g.playerNameLocked(p.Roll.Player), len(p.Rolls), dice),
		Options: options,
		ThenKey: diceIgnoreThen,
		dice:    &pause,
	})
	return nil
}

// diceIgnoreAnswered drops the chosen roll and goes on. Unanswered (the
// chooser has left), the last roll goes.
func diceIgnoreAnswered(g *Game, r OptionPicked) error {
	if r.dice == nil || len(r.dice.Rolls) == 0 || len(r.dice.Ordered) == 0 {
		return nil
	}
	p := *r.dice
	drop := r.Index
	if drop < 0 || drop >= len(p.Rolls) {
		drop = len(p.Rolls) - 1
	}
	rolls := make([]int, 0, len(p.Rolls)-1)
	rolls = append(rolls, p.Rolls[:drop]...)
	rolls = append(rolls, p.Rolls[drop+1:]...)
	p.Rolls = rolls
	p.Ordered = p.Ordered[1:]
	return g.continueDiceRollLocked(p)
}

// rollToVisitSettled is the rest of a roll to visit: every Attraction the
// player controls with the result lit up is visited (CR 701.52a).
func rollToVisitSettled(g *Game, o TailOutcome) error {
	if len(o.Rolls) == 0 {
		return nil
	}
	player, source, result := o.Tail.Player, o.Tail.Source, o.Rolls[0]
	// Collected before anything is announced: the harvester runs inside
	// EmitEvent, so the board read is the one the roll was made on.
	var visited []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && IsAttraction(c) && AttractionLit(c, result) {
			visited = append(visited, c.InstanceID)
		}
	}
	g.EmitEvent(Event{Kind: EventRollToVisit, Actor: player, Source: source, Amount: result})
	for _, id := range visited {
		g.EmitEvent(Event{Kind: EventVisitAttraction, Actor: player, CardID: id, Source: source, Amount: result})
	}
	return nil
}

// --- on disk -----------------------------------------------------------

// diceRollPauseSnapshot is a diceRollPause on disk
// (pendingChoiceSnapshot.DiceRoll).
type diceRollPauseSnapshot struct {
	Player       uuid.UUID          `json:"player"`
	Source       *uuid.UUID         `json:"source,omitempty"`
	Sides        int                `json:"sides"`
	N            int                `json:"n"`
	IgnoreLowest int                `json:"ignoreLowest,omitempty"`
	Then         *keyedTailSnapshot `json:"then,omitempty"`
	Planar       bool               `json:"planar,omitempty"`
	Modifier     int                `json:"modifier,omitempty"`
	Rolls        []int              `json:"rolls,omitempty"`
	Pending      []diceShield       `json:"pending,omitempty"`
	Ordered      []diceShield       `json:"ordered,omitempty"`
}

func snapshotDiceRollPause(p *diceRollPause) *diceRollPauseSnapshot {
	return &diceRollPauseSnapshot{
		Player: p.Roll.Player, Source: uuidPtrOrNil(p.Roll.Source), Sides: p.Roll.Sides, N: p.Roll.N,
		IgnoreLowest: p.Roll.IgnoreLowest, Then: snapshotKeyedTail(p.Roll.Then), Planar: p.Roll.Planar, Modifier: p.Roll.Modifier,
		Rolls: append([]int(nil), p.Rolls...), Pending: append([]diceShield(nil), p.Pending...),
		Ordered: append([]diceShield(nil), p.Ordered...),
	}
}

func restoreDiceRollPause(s *diceRollPauseSnapshot) *diceRollPause {
	return &diceRollPause{
		Roll: DiceRoll{
			Player: s.Player, Source: uuidOrNil(s.Source), Sides: s.Sides, N: s.N,
			IgnoreLowest: s.IgnoreLowest, Then: restoreKeyedTail(s.Then), Planar: s.Planar, Modifier: s.Modifier,
		},
		Rolls: append([]int(nil), s.Rolls...), Pending: append([]diceShield(nil), s.Pending...),
		Ordered: append([]diceShield(nil), s.Ordered...),
	}
}
