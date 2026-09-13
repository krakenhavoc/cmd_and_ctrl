package game

import "github.com/google/uuid"

// legend_rule.go is the CR 704.5j state-based action: if a player
// controls two or more legendary permanents with the same NAME, that
// player chooses one of them and the rest go to their owners'
// graveyards.
//
// The issue that asked for this called it "planeswalker uniqueness",
// which is what the rule used to be — a separate CR 704.5k keyed on
// the planeswalker's SUBTYPE, so two different Jaces could not
// coexist. Dominaria (2018) deleted that rule and made every
// planeswalker legendary instead, so what is actually needed is the
// legend rule, and implementing it covers every legendary permanent
// rather than only the walkers. Implementing the old rule would have
// been both more work and wrong: it would stop a Teferi, Temporal
// Pilgrim and a Teferi, Hero of Dominaria sharing a battlefield,
// which they legally do.
//
// THE CHOICE IS THE HARD HALF. The rule does not say "keep the
// newest" — it says the CONTROLLER chooses, and the choice is real
// (keep the one with counters on it, keep the one that isn't
// summoning-sick). So this queues a prompt rather than picking, and
// it reuses the pick_target payload shape the trigger picker and the
// battle protector already use: one chooser, a server-computed set,
// one answer back.
//
// WHILE THE PROMPT IS OPEN the duplicates all stay on the
// battlefield. A pending choice stops priority, so nothing can happen
// in that window — and the alternative, provisionally removing them
// and putting one back, would fire leave-the-battlefield triggers for
// permanents that never left.

// PendingChoiceLegendRule is the "you control two or more legendary
// permanents with the same name — choose one to keep" prompt
// (CR 704.5j). Answered with the same `{target: {kind: "card", id}}`
// payload the pick_target prompts use; the id is the one the
// controller KEEPS.
//
// Declared here rather than in pending_choice.go's const block for
// the reason PendingChoiceEntryPayLife and
// PendingChoiceChooseProtector are: the kind, the state-based action
// that queues it and the resume that settles it are one mechanism.
const PendingChoiceLegendRule PendingChoiceKind = "legend_rule"

// IsLegendary reports whether the card has the Legendary supertype
// right now, after continuous effects.
//
// Effective, not printed, so a Mirror Box or an effect that removes
// the supertype is visible here the day one is written — and so a
// TOKEN copy of a legendary permanent is legendary, which is exactly
// the case the rule most often bites on.
func (c Card) IsLegendary() bool {
	if c.effective == nil {
		super, _, _ := ParseTypeLine(c.TypeLine)
		return typeListHas(super, "legendary")
	}
	return typeListHas(c.effective.Supertypes, "legendary")
}

// legendRuleChoicesLocked finds the legend-rule violations on the
// battlefield: for each (controller, name) pair with two or more
// legendary permanents, the full set of their instance IDs.
//
// Keyed on the EFFECTIVE name, because a copy effect changes a
// permanent's name and the rule reads the name it has now (CR 201.2).
// Grouped per controller, because two players may each control a
// Sol Ring's worth of the same legend and neither has to sacrifice
// anything — the rule is about one player's own board.
//
// Returns nothing when a prompt for that group is already open, so
// the SBA loop does not queue a second copy of a question nobody has
// answered yet.
//
// Caller must hold g.mu.
func (g *Game) legendRuleChoicesLocked() map[uuid.UUID][]uuid.UUID {
	type key struct {
		controller uuid.UUID
		name       string
	}
	groups := map[key][]uuid.UUID{}
	for _, c := range g.Battlefield.Cards {
		if !c.IsLegendary() {
			continue
		}
		name := c.Effective().Name
		if name == "" {
			name = c.Name
		}
		if name == "" {
			// A nameless permanent — a fixture, a half-built token —
			// cannot violate a rule about names. Skipping is what
			// keeps the demo seed from collapsing onto itself.
			continue
		}
		k := key{controller: c.Controller, name: name}
		groups[k] = append(groups[k], c.InstanceID)
	}
	out := map[uuid.UUID][]uuid.UUID{}
	for k, ids := range groups {
		if len(ids) < 2 {
			continue
		}
		if g.legendRulePromptOpenLocked(k.controller, ids) {
			continue
		}
		out[k.controller] = append(out[k.controller], ids...)
	}
	return out
}

// legendRulePromptOpenLocked reports whether a legend-rule prompt is
// already waiting for this controller over any of these permanents.
// Caller must hold g.mu.
func (g *Game) legendRulePromptOpenLocked(controller uuid.UUID, ids []uuid.UUID) bool {
	for _, ch := range g.PendingChoices {
		if ch == nil || ch.Kind != PendingChoiceLegendRule || ch.Chooser != controller {
			continue
		}
		for _, offered := range ch.PickTargetCards {
			for _, id := range ids {
				if offered == id {
					return true
				}
			}
		}
	}
	return false
}

// queueLegendRuleChoicesLocked queues one prompt per violating
// (controller, name) group. Returns true when it queued anything,
// which the SBA loop reports as "something fired" so the check runs
// again once the answer lands.
//
// Caller must hold g.mu.
func (g *Game) queueLegendRuleChoicesLocked() bool {
	choices := g.legendRuleChoicesLocked()
	if len(choices) == 0 {
		return false
	}
	queued := false
	for controller, ids := range choices {
		name := "a legendary permanent"
		if c, ok := g.LookupCardForEffect(ids[0]); ok && c.Name != "" {
			name = c.Name
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:            PendingChoiceLegendRule,
			Chooser:         controller,
			Count:           1,
			Source:          ids[0],
			Reason:          "Legend rule — keep one " + name,
			PickTargetCards: append([]uuid.UUID(nil), ids...),
			PickTargetMin:   1,
			PickTargetMax:   1,
		})
		queued = true
	}
	return queued
}

// ResolveLegendRule settles a PendingChoiceLegendRule: `keepID` stays,
// every other permanent the prompt offered goes to its owner's
// graveyard (CR 704.5j).
//
// The losers go through the ordinary battlefield-leave path, so they
// are not "destroyed" — the legend rule is a state-based action that
// PUTS them into the graveyard, which indestructible does not stop
// and which still fires dies-triggers and the CR 903.9 commander-zone
// replacement. Routing them through the destroy helper would have
// made an indestructible legend immortal against its own rule.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveLegendRule(choiceID, chooserID, keepID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceLegendRule {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	offered := append([]uuid.UUID(nil), choice.PickTargetCards...)
	legal := false
	for _, id := range offered {
		if id == keepID {
			legal = true
			break
		}
	}
	if !legal {
		return ErrIllegalTarget
	}
	g.dequeueChoiceLocked(idx)
	for _, id := range offered {
		if id == keepID {
			continue
		}
		// A permanent that has already left in the meantime — a
		// removal spell that resolved while the prompt sat open —
		// simply is not there, and the move is a no-op.
		if findBattlefieldCard(g, id) == nil {
			continue
		}
		if err := g.routeBattlefieldCardToOwnerGraveyardLocked(id); err != nil {
			g.EmitEvent(Event{Kind: EventEffectError, Source: id, ErrorMsg: err.Error()})
		}
	}
	g.runStateChecksLocked()
	return nil
}
