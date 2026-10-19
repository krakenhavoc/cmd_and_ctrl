package game

import (
	"errors"
	"slices"

	"github.com/google/uuid"
)

// declaration_control.go — "you choose which creatures attack this turn"
// and "you choose which creatures block this turn and how those
// creatures block" (#4154, ADR 0167's 2026-10-19 amendment, the second):
// Master Warcraft, Odric, Master Tactician, Melee.
//
// This is not player control (CR 723). Nobody controls the active
// player or the defenders: they keep every other decision, their mana,
// their hands and what they can see. One player makes one turn-based
// choice for them, the CR 508.1a choice of which creatures attack or the
// CR 509.1a choice of which creatures block and what they block. So the
// effect is its own record (DeclarationControl), and only while one of
// those choices is being made does it show up as a control: a
// projection onto the player-control seam with stage ControlDeclaring
// (declarationControlOverLocked). Routing, the enumerator's controlled
// moves, the bots and the client's "Playing as" board then let the
// chooser make that one choice on the other player's board, and nothing
// else: the verb row for ControlDeclaring is the declaration verbs.
//
// # Who chooses what
//
//   - Attacks. The record's declarer chooses the attacking creatures
//     (CR 508.1a). The active player still announces what each one
//     attacks (CR 508.1b) and decides the costs: the ruling, "the person
//     who cast Master Warcraft first chooses the complete group of
//     creatures that are going to attack. Then, for each of those
//     creatures, the active player chooses who or what it's going to
//     attack." So the declaration has two halves, kept in
//     Game.AttackHandoff:
//
//     1. The declarer's half (AttackHandoff nil, or Stage ""). The
//     active player holds priority, but their decisions are the
//     declarer's (ControlDeclaring over them). The declarer stages
//     attackers with the ordinary verbs, naming a target for each as a
//     proposal; every restriction, limit and requirement applies as it
//     always does (the 2005 ruling, "Your choices must be legal within
//     the normal rules"). Nothing is paid: no attack tax and no exert,
//     because those are costs, and costs are the active player's
//     (CR 508.1g-j). The declarer's pass, sent as the active player,
//     finishes this half (finishDeclarerAttacksLocked), after the
//     CR 508.1d checkpoint.
//     2. The active player's half (Stage "targets"). The active player
//     decides for themselves again. They may point each chosen attacker
//     at another target (the per-creature verb, which re-points and
//     pays nothing), and they finish it either with declare_attackers
//     naming every chosen attacker (the targets, the exerts and the
//     whole attack tax, paid at once: confirmAttackTargetsLocked) or
//     with a pass, which accepts the targets as they stand. A creature
//     whose attack costs something at every target it could attack may
//     be left out (or, on a pass, is left out): the active player
//     declined that cost (CR 508.1d: "that player is not required to
//     pay that cost"), the creature is taken out, it can't be chosen
//     again this combat, and the declaration goes back to the declarer
//     to choose again (the Odric ruling: "you must propose a new set").
//     No cost is ever paid before the set is final, so a declaration
//     that goes back owes nobody a refund.
//
//     Once the active player's half is done (Stage "chosen") the
//     declaration is locked in as usual (commitAttackDeclarationLocked),
//     which the open halves hold back.
//
//   - Blocks. The declarer chooses, for every defending player other
//     than themselves, which creatures block and what each blocks
//     (CR 509.1a). Each defender's declaration is already a
//     turn-based action with priority parked (block_completion.go), so
//     the projection is simply live over each defender while their
//     declaration is pending, and the declarer declares and finishes
//     it on their board with the ordinary block verbs. The engine has
//     no cost to block, so the Odric ruling about one has nothing to
//     ask yet.
//
// # Precedence
//
// The projection is asked before any player-control record
// (ControllerOfPlayerLocked): ADR 0167's ruling, "If an effect causes
// another player to make decisions that the affected player would
// normally make (such as Master Warcraft does), that effect takes
// precedence." The choice is the declarer's, so whoever makes the
// declarer's decisions makes it (a player controlling the declarer's
// turn, CR 723.5).
//
// The newest record wins when two apply (CR 613.7's timestamp order,
// as CR 723.1a has it for player control).
//
// # Lifetime
//
// "This turn" records name the turn they were made in (TurnSeq) and go
// at the rotation seam (endDeclarationControlsForTurnLocked). "This
// combat" records (Odric, Melee) go as the cursor leaves that combat
// phase, at the seam combat-phase player control uses. A declarer who
// has left the game chooses nothing (CR 800.4a's spirit: the effect's
// controller is gone), and the players choose for themselves.

const (
	// ControlDeclaring is the projection of a DeclarationControl onto
	// a player while the record's declarer is making that player's
	// attack or block declaration. It is never stored in
	// Game.PlayerControls.
	ControlDeclaring ControlStage = "declaring"
	// ControlUntilDeclarations is a ControlDeclaring control's length:
	// one combat declaration.
	ControlUntilDeclarations ControlUntil = "declarations"
)

// DeclarationControl is one "you choose which creatures attack / block"
// effect. Plain data, so a table under one is still a restore point.
type DeclarationControl struct {
	// Declarer makes the choices.
	Declarer uuid.UUID `json:"declarer"`
	// Source is the card whose effect this is, and SourceName its name
	// for the view.
	Source     uuid.UUID `json:"source"`
	SourceName string    `json:"sourceName,omitempty"`
	// Attackers: the declarer chooses which creatures attack.
	// Blockers: the declarer chooses which creatures block and how.
	Attackers bool `json:"attackers,omitempty"`
	Blockers  bool `json:"blockers,omitempty"`
	// TurnSeq is the turn the effect is for.
	TurnSeq int `json:"turnSeq"`
	// ThisCombat narrows it to the combat phase it was made in
	// ("this combat"): it ends as the cursor leaves that phase.
	ThisCombat bool `json:"thisCombat,omitempty"`
}

// AttackHandoffStage is where an attack declaration made under a
// DeclarationControl stands.
type AttackHandoffStage string

const (
	// AttackHandoffDeclarer: the declarer is choosing (again, after the
	// active player declined a cost).
	AttackHandoffDeclarer AttackHandoffStage = ""
	// AttackHandoffTargets: the declarer has chosen, and the active
	// player is choosing targets and deciding the costs.
	AttackHandoffTargets AttackHandoffStage = "targets"
	// AttackHandoffChosen: both halves are done.
	AttackHandoffChosen AttackHandoffStage = "chosen"
)

// AttackHandoff is the state of one combat's attack declaration under a
// DeclarationControl. Cleared with the rest of combat.
type AttackHandoff struct {
	Stage      AttackHandoffStage `json:"stage,omitempty"`
	Declarer   uuid.UUID          `json:"declarer"`
	Source     uuid.UUID          `json:"source"`
	SourceName string             `json:"sourceName,omitempty"`
	// Declined lists the creatures whose cost to attack the active
	// player declined this combat. They can't be chosen again.
	Declined []uuid.UUID `json:"declined,omitempty"`
}

func cloneAttackHandoff(h *AttackHandoff) *AttackHandoff {
	if h == nil {
		return nil
	}
	out := *h
	out.Declined = slices.Clone(h.Declined)
	return &out
}

var (
	// ErrAttackersChosen refuses the active player adding an attacker,
	// or leaving one out, when another player chose them.
	ErrAttackersChosen = errors.New("game: another player chose which creatures attack this combat")
	// ErrAttackDeclined refuses choosing a creature whose cost to attack
	// the active player declined this combat.
	ErrAttackDeclined = errors.New("game: its controller declined to pay for that creature to attack this combat")
	// ErrDeclarerCantExert refuses an exert sent by the declarer:
	// exerting is a cost to attack, and the active player chooses it.
	ErrDeclarerCantExert = errors.New("game: the creature's controller decides whether to exert it as it attacks")
	// ErrChosenAttackerMustAttack refuses leaving out a chosen attacker
	// that could attack something without paying a cost.
	ErrChosenAttackerMustAttack = errors.New("game: a creature chosen to attack can attack something without paying a cost, so it attacks")
	// ErrAttackCostsUndecided refuses the active player's pass while a
	// chosen attacker's target costs something and another target
	// would not: point it elsewhere, or pay with declare_attackers.
	ErrAttackCostsUndecided = errors.New("game: a chosen attacker's target costs something to attack; point it at a free target or confirm the attacks and pay")
	// ErrAttackDeclarationOpen refuses skipping out of the declare
	// attackers step while a chosen declaration is unfinished.
	ErrAttackDeclarationOpen = errors.New("game: the attack declaration isn't finished")
)

// BeginDeclarationControlForEffect starts a "you choose which creatures
// attack / block" effect for the current turn. It refuses (false) a
// declarer who is not in the game, or one that chooses nothing.
//
// Caller must hold g.mu.
func (g *Game) BeginDeclarationControlForEffect(dc DeclarationControl) bool {
	if (!dc.Attackers && !dc.Blockers) || !g.playerInGameLocked(dc.Declarer) {
		return false
	}
	dc.TurnSeq = g.Turn.Seq
	// A fresh slice: an undo entry may share the old backing array.
	next := make([]DeclarationControl, len(g.DeclarationControls), len(g.DeclarationControls)+1)
	copy(next, g.DeclarationControls)
	g.DeclarationControls = append(next, dc)
	return true
}

// declarationRecordLocked is the newest live record that chooses
// attackers (attack) or blockers (!attack) this turn.
//
// Caller must hold g.mu.
func (g *Game) declarationRecordLocked(attack bool) (DeclarationControl, bool) {
	for i := len(g.DeclarationControls) - 1; i >= 0; i-- {
		r := g.DeclarationControls[i]
		if r.TurnSeq != g.Turn.Seq || !g.playerInGameLocked(r.Declarer) {
			continue
		}
		if (attack && r.Attackers) || (!attack && r.Blockers) {
			return r, true
		}
	}
	return DeclarationControl{}, false
}

// attackDeclarerPhaseLiveLocked reports whether the declarer's half of
// an attack declaration is open: the declare attackers step, a record
// that chooses attackers for another player than the active one, and
// neither the declarer's half nor the whole declaration done.
//
// Caller must hold g.mu.
func (g *Game) attackDeclarerPhaseLiveLocked() bool {
	if len(g.DeclarationControls) == 0 || g.State != StateActive || g.Turn.Step != StepDeclareAttackers {
		return false
	}
	if g.attacksDeclared {
		return false
	}
	if h := g.AttackHandoff; h != nil && h.Stage != AttackHandoffDeclarer {
		return false
	}
	r, ok := g.declarationRecordLocked(true)
	if !ok {
		return false
	}
	active := g.activeSeatIDLocked()
	return active != uuid.Nil && active != r.Declarer && g.playerInGameLocked(active)
}

// attackHandoffStageLocked is the stage of this combat's handoff, and
// whether there is one.
//
// Caller must hold g.mu.
func (g *Game) attackHandoffStageLocked() (AttackHandoffStage, bool) {
	if g.AttackHandoff == nil {
		return "", false
	}
	return g.AttackHandoff.Stage, true
}

// attackTargetsPhaseLocked reports whether the active player's half is
// open.
//
// Caller must hold g.mu.
func (g *Game) attackTargetsPhaseLocked() bool {
	st, ok := g.attackHandoffStageLocked()
	return ok && st == AttackHandoffTargets && g.Turn.Step == StepDeclareAttackers
}

// attackHandoffOpenLocked reports whether a chosen attack declaration is
// unfinished: either half still open. The lock-in waits for it.
//
// Caller must hold g.mu.
func (g *Game) attackHandoffOpenLocked() bool {
	return g.attackDeclarerPhaseLiveLocked() || g.attackTargetsPhaseLocked()
}

// attackChosenLocked reports whether another player chose this combat's
// attackers and both halves are done.
//
// Caller must hold g.mu.
func (g *Game) attackChosenLocked() bool {
	st, ok := g.attackHandoffStageLocked()
	return ok && st == AttackHandoffChosen
}

// attackDeclinedLocked reports whether the active player declined to
// pay for id to attack this combat.
//
// Caller must hold g.mu.
func (g *Game) attackDeclinedLocked(id uuid.UUID) bool {
	return g.AttackHandoff != nil && slices.Contains(g.AttackHandoff.Declined, id)
}

// declarationControlOverLocked is the projection: the record whose
// declarer is making p's declaration right now, if any.
//
//   - In declare attackers, p is the active player and the declarer's
//     half is open.
//   - In declare blockers, p is a defending player other than the
//     declarer whose declaration is pending.
//
// Caller must hold g.mu.
func (g *Game) declarationControlOverLocked(p uuid.UUID) (DeclarationControl, bool) {
	if len(g.DeclarationControls) == 0 || g.State != StateActive {
		return DeclarationControl{}, false
	}
	switch g.Turn.Step {
	case StepDeclareAttackers:
		if p != g.activeSeatIDLocked() || !g.attackDeclarerPhaseLiveLocked() {
			return DeclarationControl{}, false
		}
		return g.declarationRecordLocked(true)
	case StepDeclareBlockers:
		r, ok := g.declarationRecordLocked(false)
		if !ok || r.Declarer == p || !g.playerInGameLocked(p) {
			return DeclarationControl{}, false
		}
		if g.BlockDeclarationStatusLocked(p) != BlockDeclarationPending {
			return DeclarationControl{}, false
		}
		return r, true
	}
	return DeclarationControl{}, false
}

// declarationControlsLiveLocked is the projection for every seat, as
// PlayerControl records with stage ControlDeclaring, in seat order:
// what LivePlayerControlsLocked lists for them.
//
// Caller must hold g.mu.
func (g *Game) declarationControlsLiveLocked() []PlayerControl {
	if len(g.DeclarationControls) == 0 {
		return nil
	}
	var out []PlayerControl
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		r, ok := g.declarationControlOverLocked(p.ID)
		if !ok {
			continue
		}
		decider := g.playerControllerLocked(r.Declarer)
		if decider == p.ID {
			continue
		}
		out = append(out, PlayerControl{
			Controller: decider,
			Controlled: p.ID,
			Source:     r.Source,
			SourceName: r.SourceName,
			Stage:      ControlDeclaring,
			Until:      ControlUntilDeclarations,
		})
	}
	return out
}

// stagedAttackersLocked is every creature the active player has staged
// as an attacker and that has not been announced, in battlefield order.
//
// Caller must hold g.mu.
func (g *Game) stagedAttackersLocked() []*Card {
	if g.Battlefield == nil {
		return nil
	}
	ap := g.activeSeatIDLocked()
	var out []*Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil || c.Controller != ap || g.announcedAttacks[c.InstanceID] {
			continue
		}
		out = append(out, c)
	}
	return out
}

// setAttackHandoffLocked stores a fresh copy of h: the old one may be
// shared with an undo entry.
//
// Caller must hold g.mu.
func (g *Game) setAttackHandoffLocked(h AttackHandoff) {
	g.AttackHandoff = cloneAttackHandoff(&h)
}

// attackHandoffRefusalLocked is what the per-creature declaration verb
// refuses under a DeclarationControl, before anything is checked or
// paid. Nil at every table without one.
//
// Caller must hold g.mu.
func (g *Game) attackHandoffRefusalLocked(card *Card, exert bool) error {
	if g.AttackHandoff == nil && len(g.DeclarationControls) == 0 {
		return nil
	}
	switch {
	case g.attackDeclarerPhaseLiveLocked():
		if g.attackDeclinedLocked(card.InstanceID) {
			return ErrAttackDeclined
		}
		if exert {
			return ErrDeclarerCantExert
		}
	case g.attackTargetsPhaseLocked():
		// The active player points a chosen attacker elsewhere; they
		// can't add one.
		if card.AttackingTarget == uuid.Nil || g.announcedAttacks[card.InstanceID] {
			return ErrAttackersChosen
		}
	case g.attackChosenLocked():
		// Re-pointing an attacker already announced stays the
		// sandbox's hand-correction (DeclareAttackerDeclWith); a new
		// one is refused.
		if !g.announcedAttacks[card.InstanceID] {
			return ErrAttackersChosen
		}
	}
	return nil
}

// attackTaxDeferredLocked reports whether a declaration staged now
// pays no attack tax: a chosen declaration's costs are the active
// player's, and are paid when the active player confirms the attacks
// (confirmAttackTargetsLocked).
//
// Caller must hold g.mu.
func (g *Game) attackTaxDeferredLocked() bool {
	return g.attackHandoffOpenLocked()
}

// finishDeclarerAttacksLocked is the declarer's pass in their half: the
// CR 508.1d checkpoint, then the active player's half, or (nothing
// chosen) the declaration over.
//
// Caller must hold g.mu in write mode.
func (g *Game) finishDeclarerAttacksLocked() error {
	g.RecomputeLayersIfStaleLocked()
	if e := g.attackRequirementsUnmetLocked(); e != nil {
		return e
	}
	r, _ := g.declarationRecordLocked(true)
	h := AttackHandoff{Declarer: r.Declarer, Source: r.Source, SourceName: r.SourceName}
	if g.AttackHandoff != nil {
		h.Declined = g.AttackHandoff.Declined
	}
	if len(g.stagedAttackersLocked()) == 0 {
		h.Stage = AttackHandoffChosen
		g.setAttackHandoffLocked(h)
		g.markAttackDeclarationJudgedLocked()
		return nil
	}
	h.Stage = AttackHandoffTargets
	g.setAttackHandoffLocked(h)
	return nil
}

// freeAttackAvailableLocked reports whether c could be pointed at some
// target without paying anything (CR 508.1a's tax): a target it may
// attack, that no count limit or requirement refuses it, at no price.
// Exactly the free re-points the enumerator offers.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) freeAttackAvailableLocked(c *Card) bool {
	for _, t := range g.AttackTargetsForAttackerForEffect(c) {
		decl := []AttackDeclaration{{Attacker: c.InstanceID, Target: t.ID}}
		if !g.priceAttackDeclarationLocked(decl).IsFree() {
			continue
		}
		if t.ID != c.AttackingTarget &&
			(g.attackLimitRefusalLocked(decl) != nil || g.attackRequirementRefusalLocked(decl) != nil) {
			continue
		}
		return true
	}
	return false
}

// declineChosenAttackersLocked takes the creatures whose costs the
// active player declined out of the declaration and hands it back to
// the declarer (the Odric ruling).
//
// Caller must hold g.mu in write mode.
func (g *Game) declineChosenAttackersLocked(cards []*Card) {
	h := *g.AttackHandoff
	h.Declined = slices.Clone(h.Declined)
	for _, c := range cards {
		h.Declined = append(h.Declined, c.InstanceID)
		g.removeFromCombatLocked(c)
		// Declaring it tapped it (CR 508.1f); the declaration that
		// tapped it is being made again.
		if !HasKeyword(c, "vigilance") {
			c.Tapped = false
		}
	}
	h.Stage = AttackHandoffDeclarer
	g.setAttackHandoffLocked(h)
}

// AttackTargetsPassRefusalForEffect is the refusal the active player's
// pass gets in their half (ErrAttackCostsUndecided), or nil. The
// enumerator offers the pass exactly when this is nil.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) AttackTargetsPassRefusalForEffect() error {
	if !g.attackTargetsPhaseLocked() {
		return nil
	}
	for _, c := range g.stagedAttackersLocked() {
		price := g.priceAttackDeclarationLocked([]AttackDeclaration{{Attacker: c.InstanceID, Target: c.AttackingTarget}})
		if !price.IsFree() && g.freeAttackAvailableLocked(c) {
			return ErrAttackCostsUndecided
		}
	}
	return nil
}

// attackHandoffPassLocked is a pass in the declare attackers step under
// a DeclarationControl. handled reports whether it was spent on the
// handoff: the declarer finishing their half, or the active player
// declining costs (the declaration goes back to the declarer). A pass
// that accepts the active player's half returns handled false, so the
// ordinary pass runs next and locks the declaration in.
//
// Caller must hold g.mu in write mode.
func (g *Game) attackHandoffPassLocked() (handled bool, err error) {
	if g.State != StateActive || g.Turn.Step != StepDeclareAttackers || g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		return false, nil
	}
	if g.attackDeclarerPhaseLiveLocked() {
		return true, g.finishDeclarerAttacksLocked()
	}
	if !g.attackTargetsPhaseLocked() {
		return false, nil
	}
	g.RecomputeLayersIfStaleLocked()
	if err := g.AttackTargetsPassRefusalForEffect(); err != nil {
		return true, err
	}
	var owed []*Card
	for _, c := range g.stagedAttackersLocked() {
		price := g.priceAttackDeclarationLocked([]AttackDeclaration{{Attacker: c.InstanceID, Target: c.AttackingTarget}})
		if !price.IsFree() {
			owed = append(owed, c)
		}
	}
	if len(owed) > 0 {
		g.declineChosenAttackersLocked(owed)
		return true, nil
	}
	h := *g.AttackHandoff
	h.Stage = AttackHandoffChosen
	g.setAttackHandoffLocked(h)
	return false, nil
}

// confirmAttackTargetsLocked is declare_attackers in the active player's
// half: decls name the chosen attackers with the target each attacks
// (and whether to exert it), and the whole attack tax is paid at once.
// A chosen attacker left out is a declined cost, allowed only for one
// with no free target; then nothing is paid or moved, and the
// declaration goes back to the declarer without those creatures.
//
// Caller must hold g.mu in write mode, with fresh layers.
func (g *Game) confirmAttackTargetsLocked(decls []AttackDeclaration, params DeclareAttackersParams) ([]uuid.UUID, error) {
	staged := g.stagedAttackersLocked()
	byID := make(map[uuid.UUID]*Card, len(staged))
	for _, c := range staged {
		byID[c.InstanceID] = c
	}
	named := make(map[uuid.UUID]bool, len(decls))
	for _, d := range decls {
		c, ok := byID[d.Attacker]
		if !ok {
			return nil, ErrAttackersChosen
		}
		if named[d.Attacker] {
			return nil, ErrInvalidParam
		}
		named[d.Attacker] = true
		if err := g.canAttackTargetWithLocked(c, d.Target); err != nil {
			return nil, err
		}
	}
	var omitted []*Card
	for _, c := range staged {
		if named[c.InstanceID] {
			continue
		}
		if g.freeAttackAvailableLocked(c) {
			return nil, ErrChosenAttackerMustAttack
		}
		omitted = append(omitted, c)
	}
	if len(omitted) > 0 {
		g.declineChosenAttackersLocked(omitted)
		return nil, nil
	}
	if err := g.attackLimitRefusalLocked(decls); err != nil {
		return nil, err
	}
	if err := g.attackRequirementRefusalLocked(decls); err != nil {
		return nil, err
	}
	if err := g.validateExertChoicesLocked(decls); err != nil {
		return nil, err
	}
	ap := g.activeSeatIDLocked()
	price := g.priceAttackDeclarationLocked(decls)
	if err := g.attackTaxAffordableLocked(ap, price, params); err != nil {
		return nil, err
	}
	if err := g.payAttackTaxLocked(ap, price, params); err != nil {
		return nil, err
	}
	out := make([]uuid.UUID, 0, len(decls))
	for _, d := range decls {
		c := findBattlefieldCard(g, d.Attacker)
		if c == nil {
			continue
		}
		g.setAttackTargetLocked(c, d.Target)
		if d.Exert {
			c.ExertOnAttack = true
		}
		out = append(out, d.Attacker)
	}
	h := *g.AttackHandoff
	h.Stage = AttackHandoffChosen
	g.setAttackHandoffLocked(h)
	// The lock-in, as the bulk verb's own (CR 508.2).
	g.runStateChecksLocked()
	return out, nil
}

// endDeclarationControlsLocked drops the records that are over. As a
// combat phase ends (turnEnds false) that is every "this combat" record
// and any record of another turn; as the turn ends, every record.
// Caller must hold g.mu.
func (g *Game) endDeclarationControlsLocked(turnEnds bool) {
	if len(g.DeclarationControls) == 0 {
		return
	}
	var kept []DeclarationControl
	for _, r := range g.DeclarationControls {
		if turnEnds || r.TurnSeq != g.Turn.Seq || r.ThisCombat {
			continue
		}
		kept = append(kept, r)
	}
	if len(kept) == len(g.DeclarationControls) {
		return
	}
	g.DeclarationControls = kept
}

// DeclarationControlsForEffect is the records that apply this turn,
// oldest first, whose declarers are in the game: the view's
// declaration_controls. Caller must hold g.mu.
func (g *Game) DeclarationControlsForEffect() []DeclarationControl {
	var out []DeclarationControl
	for _, r := range g.DeclarationControls {
		if r.TurnSeq == g.Turn.Seq && g.playerInGameLocked(r.Declarer) {
			out = append(out, r)
		}
	}
	return out
}

// AttackDeclarerPhaseLiveForEffect is attackDeclarerPhaseLiveLocked for
// the enumerator and the view. Caller must hold g.mu.
func (g *Game) AttackDeclarerPhaseLiveForEffect() bool { return g.attackDeclarerPhaseLiveLocked() }

// AttackHandoffForEffect is this combat's handoff, or nil. The caller
// must not modify it. Caller must hold g.mu.
func (g *Game) AttackHandoffForEffect() *AttackHandoff { return g.AttackHandoff }

// AttackTargetsPhaseForEffect is attackTargetsPhaseLocked for the
// enumerator and the bots. Caller must hold g.mu.
func (g *Game) AttackTargetsPhaseForEffect() bool { return g.attackTargetsPhaseLocked() }

// AttackTargetsOwedBy reports whether seat is the active player and
// owes the targets half of a chosen attack declaration. Caller must NOT
// hold g.mu.
func (g *Game) AttackTargetsOwedBy(seat uuid.UUID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.attackTargetsPhaseLocked() && g.activeSeatIDLocked() == seat &&
		g.playerControllerLocked(seat) == seat
}

// AttackChosenForEffect is attackChosenLocked for the enumerator.
// Caller must hold g.mu.
func (g *Game) AttackChosenForEffect() bool { return g.attackChosenLocked() }

// AttackDeclinedForEffect is attackDeclinedLocked for the enumerator.
// Caller must hold g.mu.
func (g *Game) AttackDeclinedForEffect(id uuid.UUID) bool { return g.attackDeclinedLocked(id) }

// StagedAttackersForEffect is stagedAttackersLocked for the
// enumerator. Caller must hold g.mu.
func (g *Game) StagedAttackersForEffect() []*Card { return g.stagedAttackersLocked() }

// FreeAttackAvailableForEffect is freeAttackAvailableLocked for the
// enumerator. Caller must hold g.mu with fresh layers.
func (g *Game) FreeAttackAvailableForEffect(c *Card) bool { return g.freeAttackAvailableLocked(c) }
