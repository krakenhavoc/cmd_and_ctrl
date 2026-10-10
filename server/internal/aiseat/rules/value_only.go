package rules

import (
	"encoding/json"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// valueOnly is RuleValueOnly's test (ADR 0142 decision 6): something is
// on the stack, and every move other than the pass is floating mana or
// an untargeted activation of a row its card declares `value`, with
// nothing the move spends about to be lost.
//
// "About to be lost" is read wider here than the heuristic reads it,
// because Layer A must never take a window the heuristic would answer
// differently: the heuristic lets a value row spend a permanent it is
// about to lose (ADR 0126 §7), and this rule escalates every window in
// which that could be so. That is a window where the move's source, or
// a permanent it sacrifices or exiles, is a target of anything on the
// stack or is attacking or blocking; where a spell on the stack declares
// a sweep; or where the enumerator marks the move as interacting (ADR
// 0142 owner answer 4).
func valueOnly(in aiseat.Input) bool {
	v := &in.View
	if len(v.StackItems) == 0 {
		return false
	}
	targeted := map[string]bool{}
	for i := range v.StackItems {
		it := &v.StackItems[i]
		for _, t := range it.Targets {
			targeted[t.ID] = true
		}
		if it.Kind == "spell" && declaresSweep(stackCard(v, it.SourceCardID)) {
			return false
		}
	}
	bf := make(map[string]*protocol.CardView, len(v.Battlefield.Cards))
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		bf[c.InstanceID] = c
	}
	spent := func(id string) bool {
		c := bf[id]
		return targeted[id] || c == nil || c.AttackingTarget != "" || c.BlockingTarget != "" || len(c.BlockingTargets) > 0
	}
	activations := 0
	for i := range in.Moves {
		m := &in.Moves[i]
		switch m.Kind {
		case legal.KindPass, legal.KindMana:
			continue
		case legal.KindActivate:
		default:
			return false
		}
		if m.Interacts || m.CombatInteracts || m.HasTargets || m.TargetsStack {
			return false
		}
		var p struct {
			Source    string   `json:"source_card_id"`
			Index     int      `json:"ability_index"`
			Sacrifice []string `json:"sacrifice_ids"`
			Exile     []string `json:"exile_permanent_ids"`
		}
		if json.Unmarshal(m.Params, &p) != nil {
			return false
		}
		src := bf[p.Source]
		if src == nil || src.Controller != in.Seat.String() || spent(p.Source) {
			return false
		}
		row := rowOf(src, p.Index)
		if row == nil || !row.Purpose.AnswersNothing() {
			return false
		}
		for _, ids := range [][]string{p.Sacrifice, p.Exile} {
			for _, id := range ids {
				if spent(id) {
					return false
				}
			}
		}
		activations++
	}
	return activations > 0
}

// rowOf is the activated row with this index on c.
func rowOf(c *protocol.CardView, index int) *protocol.ActivatedAbilityView {
	for i := range c.ActivatedAbilities {
		if c.ActivatedAbilities[i].Index == index {
			return &c.ActivatedAbilities[i]
		}
	}
	return nil
}

// stackCard is the card in the stack zone with this instance ID, nil
// when the view does not show it.
func stackCard(v *protocol.GameView, id string) *protocol.CardView {
	for i := range v.Stack.Cards {
		if v.Stack.Cards[i].InstanceID == id {
			return &v.Stack.Cards[i]
		}
	}
	return nil
}

// declaresSweep reports whether a spell declares a sweep anywhere: on
// the card, a mode or an alternative cost. A card the view does not
// show is read as one.
func declaresSweep(c *protocol.CardView) bool {
	if c == nil {
		return true
	}
	if c.Purpose != nil && c.Purpose.Sweep != nil {
		return true
	}
	if c.Modes != nil {
		for _, o := range c.Modes.Options {
			if o.Purpose != nil && o.Purpose.Sweep != nil {
				return true
			}
		}
	}
	for _, ac := range c.AlternativeCosts {
		if ac.Purpose != nil && ac.Purpose.Sweep != nil {
			return true
		}
	}
	return false
}
