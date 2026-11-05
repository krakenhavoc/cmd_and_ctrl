package game

import (
	"strconv"

	"github.com/google/uuid"
)

// planechase_die.go is the planar die (CR 901.3a, 901.9, ADR 0216 §4,
// slice 2, #4870). The rules, quoted from the September 25, 2026 edition
// AGENTS.md §6 pins:
//
//	901.3a A planar die is a six-sided die. One face has the Planeswalker
//	       symbol. One face has the chaos symbol. The other faces are
//	       blank.
//	901.9. Any time the active player has priority and the stack is
//	       empty, but only during a main phase of their turn, that player
//	       may roll the planar die. Taking this action costs a player an
//	       amount of mana equal to the number of times they have
//	       previously taken this action on that turn. This is a special
//	       action and doesn't use the stack. Note that this number won't
//	       be equal to the number of times the player has rolled the
//	       planar die that turn if an effect has caused the player to
//	       roll the planar die that turn. (See rule 116.2i.)
//	901.9a If the die roll is a blank face, nothing happens. The active
//	       player gets priority.
//	901.9b If the die roll is the chaos symbol, chaos ensues (see rule
//	       311.7). The active player gets priority.
//	901.9c If the die roll is the Planeswalker symbol, the "planeswalking
//	       ability" triggers and is put on the stack. The active player
//	       gets priority. (See rule 901.8.)
//	901.9d Rolling the planar die will cause any ability that triggers
//	       whenever a player rolls one or more dice to trigger. However,
//	       any effect that refers to a numerical result of a die roll,
//	       including ones that compare the results of that roll to other
//	       rolls or to a given number, ignores the rolling of the planar
//	       die. See 706, "Rolling a Die."
//
// What the text settles that ADR 0216 §4 left to this slice: 901.9's
// note says a roll an EFFECT causes does not count toward the cost. So
// only performPlanarDieLocked counts, and RollPlanarDieForEffect does
// not.
//
// The roll goes through the one dice door (RollDiceThenForEffect) with
// DiceRoll.Planar set: its own keyed stream (rngStreamPlanarDie), and one
// EventRollDie with Amount 0 and PlanarFace naming the face, so "whenever
// you roll one or more dice" triggers and no watcher of a number sees
// one (901.9d). A natural 1 is chaos, a 6 the Planeswalker symbol, 2 to 5
// blank.

// SpecialActionPlanarDie is CR 901.9 (a CR 116.2 special action): the
// active player rolls the planar die, paying {N} generic where N is the
// number of times they have taken this action this turn. It is taken on
// no card: the wire's card_id is absent. ADR 0216 §4.
const SpecialActionPlanarDie SpecialActionKind = "planar_die"

// The planar die's faces, as EventRollDie.PlanarFace and the log carry
// them. On-disk and wire identities.
const (
	PlanarFaceBlank      = "blank"
	PlanarFaceChaos      = "chaos"
	PlanarFacePlaneswalk = "planeswalk"
)

// planarFaceOf is the face a planar die's drawn number shows: one face
// each for chaos and the Planeswalker symbol, four blanks (CR 901.3a).
func planarFaceOf(n int) string {
	switch n {
	case 1:
		return PlanarFaceChaos
	case 6:
		return PlanarFacePlaneswalk
	}
	return PlanarFaceBlank
}

// planarDieTakenThisTurnLocked is how many times the planar die special
// action has been taken this turn (CR 901.9's N).
func (g *Game) planarDieTakenThisTurnLocked() int {
	if g.Planechase == nil || g.Planechase.DieTurnSeq != g.Turn.Seq {
		return 0
	}
	return g.Planechase.DieRolls
}

// PlanarDieRollsThisTurnLocked is the wire's die_rolls: the planar die
// special actions taken this turn.
//
// Caller must hold g.mu.
func (g *Game) PlanarDieRollsThisTurnLocked() int { return g.planarDieTakenThisTurnLocked() }

// PlanarDieCostLocked is the planar die's printed cost this turn, "{N}"
// (CR 901.9), or "" when the table is not playing Planechase. The cost
// modifier pass (SpecialActionManaCostForEffect) prices it from here.
//
// Caller must hold g.mu.
func (g *Game) PlanarDieCostLocked() string {
	if g.Planechase == nil {
		return ""
	}
	return "{" + strconv.Itoa(g.planarDieTakenThisTurnLocked()) + "}"
}

// PlanarDieOfferLocked is the planar die special action as an offer,
// for the engine, the enumerator and the wire alike, or nil when the
// table is not playing Planechase.
//
// Caller must hold g.mu.
func (g *Game) PlanarDieOfferLocked() *SpecialAction {
	if g.Planechase == nil {
		return nil
	}
	cost := g.PlanarDieCostLocked()
	return &SpecialAction{Kind: SpecialActionPlanarDie, Cost: cost, Label: "Roll the planar die " + cost}
}

// PlanarDieCostForEffect is the planar die's cost for `player` after the
// CR 601.2f pass, as the engine would charge it, or ok false when the
// table is not playing Planechase.
//
// Caller must hold g.mu.
func (g *Game) PlanarDieCostForEffect(player uuid.UUID) (ParsedCost, bool) {
	sa := g.PlanarDieOfferLocked()
	if sa == nil {
		return ParsedCost{}, false
	}
	cost, err := g.SpecialActionManaCostForEffect(player, Card{}, SpecialActionPlanarDie, *sa)
	if err != nil {
		return ParsedCost{}, false
	}
	return cost, true
}

// performPlanarDieLocked takes the planar die special action (CR 901.9):
// refused unless the table plays Planechase and the window is open, then
// the {N} is paid, the count goes up, and the die is rolled. Like
// PerformSpecialAction's other kinds a refusal leaves everything as it
// was, and the player keeps priority.
//
// Caller must hold g.mu (write).
func (g *Game) performPlanarDieLocked(p *Player, cardID uuid.UUID, params SpecialActionParams) error {
	if cardID != uuid.Nil || params.SacrificeID != uuid.Nil || params.Door != DoorNone || params.Cost != "" {
		return ErrInvalidParam
	}
	sa := g.PlanarDieOfferLocked()
	if sa == nil {
		return ErrSpecialActionNotOffered
	}
	if !g.SpecialActionTimingOKLocked(p.ID, Card{}, SpecialActionPlanarDie) {
		return ErrSpecialActionTiming
	}
	g.beginEventBatchLocked()
	cost, err := g.SpecialActionManaCostForEffect(p.ID, Card{}, SpecialActionPlanarDie, *sa)
	if err != nil {
		return ErrInvalidParam
	}
	if _, err := g.payAbilityManaCostLocked(p, uuid.Nil, "the planar die", cost, ActivateAbilityParams{
		Strict:  params.Strict,
		AutoTap: params.AutoTap,
	}, specialActionSpendContext(SpecialActionPlanarDie), nil); err != nil {
		return err
	}
	// The count goes up as the action is taken, before the roll: a roll
	// a replacement pauses has still been paid for (ADR 0216 §4).
	ps := g.Planechase
	if ps.DieTurnSeq != g.Turn.Seq {
		ps.DieTurnSeq, ps.DieRolls = g.Turn.Seq, 0
	}
	ps.DieRolls++
	g.EmitEvent(Event{Kind: EventSpecialAction, Actor: p.ID, Label: sa.Label})
	if err := g.rollPlanarDieLocked(p.ID, uuid.Nil); err != nil {
		return err
	}
	// CR 117.3c: the player keeps priority, and passes made before the
	// action no longer count.
	g.noteActionTakenLocked(p.ID)
	g.runStateChecksLocked()
	return nil
}

// rollPlanarDieLocked rolls the planar die for `player`, through the one
// dice door. `source` is the object whose effect makes them roll, nil
// for the special action.
//
// Caller must hold g.mu (write).
func (g *Game) rollPlanarDieLocked(player, source uuid.UUID) error {
	return g.RollDiceThenForEffect(DiceRoll{Player: player, Source: source, Sides: 6, N: 1, Planar: true})
}

// RollPlanarDieForEffect is "roll the planar die" said by a card
// (Fractured Powerstone, later): the same roll as the special action,
// with Source set, no cost, and not counted toward the special action's
// cost (CR 901.9's note). ErrInvalidParam when the table is not playing
// Planechase.
//
// Caller must hold g.mu (write).
func (g *Game) RollPlanarDieForEffect(player, source uuid.UUID) error {
	if g.Planechase == nil || !g.playerInGameLocked(player) {
		return ErrInvalidParam
	}
	return g.rollPlanarDieLocked(player, source)
}
