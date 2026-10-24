package game

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/google/uuid"
)

// attack_split.go — "create N tokens that are tapped and attacking",
// each attacking what its controller chooses (CR 508.4), as data (#4294).
//
// CR 508.4: when a creature is put onto the battlefield attacking, its
// controller chooses which defending player or planeswalker or battle it
// is attacking. A group of tokens made together is split among the things
// they could attack: the controller is asked how many attack each in turn,
// the first thing being the one the creature whose trigger this is
// attacks, and the last takes the rest, so there is at most one question
// per attackable thing but the last, and none once every token is placed.
//
// The questions are number prompts (number_then.go). The state between
// them is an AttackSplit, plain data the prompt carries
// (PendingChoice.AttackSplit): the template, how many, the things to
// attack in asking order, the shares fixed so far and the continuation as
// a keyed tail, so a table waiting on any of the questions is a restore
// point. The tokens are made by one creation once the last share is
// known, so a token doubler doubles each share and a "whenever one or more
// tokens are created" trigger fires once.

// AttackSplit is where a tokens-attacking split is.
type AttackSplit struct {
	// Controller answers the questions and controls the tokens; Source is
	// the card that makes them.
	Controller, Source uuid.UUID
	// Template is the token as it is made; Count how many; Tapped whether
	// they enter tapped.
	Template Card
	Count    int
	Tapped   bool
	// Targets are the things the tokens could attack, in asking order, and
	// Shares how many attack each of Targets[:len(Shares)].
	Targets []uuid.UUID
	Shares  []int
	// Label names the card in the questions.
	Label string
	// Tail is what the card does once the tokens are made, told the tokens
	// that landed (RegisterTail). The zero tail does nothing.
	Tail KeyedTail
}

// attackSplitKey is the split's number branch, an on-disk identity: never
// renamed, never reused.
const attackSplitKey = "token/attack-split"

func init() {
	RegisterNumberThen(attackSplitKey, func(g *Game, r NumberAnswer) error {
		if r.split == nil {
			return fmt.Errorf("%w: attack split with no state", ErrInvalidParam)
		}
		s := r.split.clone()
		left := s.Count - sumInts(s.Shares)
		s.Shares = append(s.Shares, min(max(r.Number, 0), left))
		return g.advanceAttackSplitLocked(s)
	})
}

// StartAttackSplitForEffect makes the tokens `s` describes, asking the
// controller how many attack each thing when there is more than one
// (CR 508.4). With nothing or one thing to attack nothing is asked. The
// tail runs with every token made, also when there is nothing to attack
// (CR 508.4a: they are made, not attacking).
//
// Caller must hold g.mu.
func (g *Game) StartAttackSplitForEffect(s AttackSplit) error {
	if s.Count <= 0 {
		return nil
	}
	return g.advanceAttackSplitLocked(s.clone())
}

// advanceAttackSplitLocked asks the next question, or makes the tokens
// when there is none.
//
// Caller must hold g.mu.
func (g *Game) advanceAttackSplitLocked(s AttackSplit) error {
	left := s.Count - sumInts(s.Shares)
	i := len(s.Shares)
	if len(s.Targets) <= 1 || left <= 0 || i >= len(s.Targets)-1 {
		if i < max(len(s.Targets), 1) {
			s.Shares = append(s.Shares, left)
		}
		return g.createAttackSplitLocked(s)
	}
	goal := 0
	if i == 0 {
		goal = left
	}
	return g.QueueChooseNumberForEffect(ChooseNumber{
		Chooser:  s.Controller,
		Source:   s.Source,
		Resource: PayResourceNone,
		Max:      left,
		Goal:     goal,
		Question: s.Label + " — how many of the " + strconv.Itoa(left) + " tokens attack " + g.attackTargetNameLocked(s.Targets[i]) + "?",
		Branch:   NumberBranch{Then: NumberThen{key: attackSplitKey}},
		split:    &s,
	})
}

// createAttackSplitLocked is the creation: shares[i] of the tokens
// attacking targets[i], or none attacking when there is nothing to attack.
//
// Caller must hold g.mu.
func (g *Game) createAttackSplitLocked(s AttackSplit) error {
	var groups []TokenGroup
	for i, n := range s.Shares {
		if n <= 0 {
			continue
		}
		grp := TokenGroup{Template: s.Template, Count: n, Entry: TokenEntryOptions{Tapped: s.Tapped}}
		if i < len(s.Targets) {
			grp.Attacking = s.Targets[i]
		}
		groups = append(groups, grp)
	}
	return g.CreateTokensKeyedForEffect(TokenCreation{
		Controller: s.Controller,
		Groups:     groups,
		Source:     s.Source,
	}, s.Tail)
}

// attackTargetNameLocked is a player's or a permanent's name for a
// question.
//
// Caller must hold g.mu.
func (g *Game) attackTargetNameLocked(id uuid.UUID) string {
	if p := g.playerByIDLocked(id); p != nil && p.Name != "" {
		return p.Name
	}
	if c := findBattlefieldCard(g, id); c != nil && c.Name != "" {
		return c.Name
	}
	return "that player"
}

func sumInts(xs []int) int {
	n := 0
	for _, x := range xs {
		n += x
	}
	return n
}

// clone is a deep copy.
func (s AttackSplit) clone() AttackSplit {
	out := s
	out.Template = cloneCard(s.Template)
	out.Targets = slices.Clone(s.Targets)
	out.Shares = slices.Clone(s.Shares)
	out.Tail = copyKeyedTail(s.Tail)
	return out
}

// attackSplitSnapshot is an AttackSplit on disk: the template through the
// card mirror.
type attackSplitSnapshot struct {
	Controller uuid.UUID          `json:"controller"`
	Source     uuid.UUID          `json:"source"`
	Template   cardSnapshot       `json:"template"`
	Count      int                `json:"count"`
	Tapped     bool               `json:"tapped,omitempty"`
	Targets    []uuid.UUID        `json:"targets,omitempty"`
	Shares     []int              `json:"shares,omitempty"`
	Label      string             `json:"label,omitempty"`
	Tail       *keyedTailSnapshot `json:"tail,omitempty"`
}

func snapshotAttackSplit(s *AttackSplit, cen *ContinuationCensus) *attackSplitSnapshot {
	if s == nil {
		return nil
	}
	return &attackSplitSnapshot{
		Controller: s.Controller, Source: s.Source, Template: snapshotCard(s.Template, cen),
		Count: s.Count, Tapped: s.Tapped, Targets: copyUUIDs(s.Targets), Shares: copyInts(s.Shares),
		Label: s.Label, Tail: snapshotKeyedTail(s.Tail),
	}
}

func restoreAttackSplit(s *attackSplitSnapshot) *AttackSplit {
	if s == nil {
		return nil
	}
	return &AttackSplit{
		Controller: s.Controller, Source: s.Source, Template: restoreCard(&s.Template),
		Count: s.Count, Tapped: s.Tapped, Targets: copyUUIDs(s.Targets), Shares: copyInts(s.Shares),
		Label: s.Label, Tail: restoreKeyedTail(s.Tail),
	}
}

// cloneAttackSplit is a deep copy, nil for nil.
func cloneAttackSplit(s *AttackSplit) *AttackSplit {
	if s == nil {
		return nil
	}
	out := s.clone()
	return &out
}
