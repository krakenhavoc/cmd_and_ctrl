package game

import "github.com/google/uuid"

// ring.go — CR 701.54, "the Ring tempts you" (ADR 0114 §1–§4, #2076).
//
//	701.54a Each time the Ring tempts you, choose a creature you
//	        control. That creature becomes your Ring-bearer until
//	        another creature becomes your Ring-bearer or another player
//	        gains control of it.
//	701.54b Ring-bearer is a designation a permanent can have. Being a
//	        Ring-bearer is not a copiable value.
//	701.54c If a player doesn't have an emblem named The Ring at the
//	        time the Ring tempts them, they get an emblem named The
//	        Ring before choosing a creature to be their Ring-bearer.
//	        [Its four abilities, by how many times it has tempted.]
//	701.54d The Ring tempts a player whenever they complete the
//	        actions in 701.54a, even if some or all of those actions
//	        were impossible.
//	701.54e [A creature] "is your Ring-bearer" if that creature is on
//	        the battlefield under your control and has the Ring-bearer
//	        designation.
//
// Four pieces, each with one home:
//
//   - THE EMBLEM is an ordinary ADR 0064 emblem in Player.Emblems with a
//     fixed key, RingEmblemKey, because no card creates it (§2). Its
//     abilities are declared once, in cards/effects/the_ring.go, and
//     reach the layer pass, the trigger harvester and the block-rule
//     walk exactly as every other emblem's do.
//   - THE COUNT lives on the emblem object, Card.RingTemptations. The
//     emblem exists exactly when the count is at least one and never
//     leaves while its owner is in the game, so the count on it IS the
//     player's count — and it is what DesignationRingTempted reads,
//     which keeps ADR 0071's gate an object-only question.
//   - THE DESIGNATION is Card.RingBearer on the permanent (§3). It is
//     set in ONE place, finishRingTemptLocked, which first clears every
//     other permanent the chooser controls, so "only one Ring-bearer at
//     a time" is held by that one write site. It is cleared by CR 400.7
//     (zone.go, entry_tail.go) and by a control change
//     (materialiseControlLocked), the two ways CR 701.54a ends it.
//   - THE CHOICE is PendingChoiceRingBearer, a card-set pick of its own
//     (§4). With exactly one candidate it is not asked: the creature is
//     chosen automatically and the log says so (owner decision 2, as
//     amass does).

// RingEmblemKey is the Ring emblem's catalog key and OracleID — an
// on-disk identity, like a token slug: never renamed, never reused.
// ADR 0064 derives an emblem's key from the card that creates it; no
// card creates this one (CR 701.54c makes it a rules object), so its
// key is fixed.
const RingEmblemKey = EmblemKeyPrefix + "the-ring"

// ringEmblemName is the emblem's name (CR 701.54c: "an emblem named
// The Ring"). Unlike most emblems this one has a name, and it is the
// label the board and the log use when the catalog is not loaded.
const ringEmblemName = "The Ring"

// PendingChoiceRingBearer is "choose a creature you control" as the
// Ring tempts you (CR 701.54a), answered with {card_ids: [id]} —
// exactly one. It carries the card-set payload (ChooseCards, ChooseMin,
// ChooseMax), so the projection, the enumerator's acceptance check and
// the client's card grid are the shared ones.
//
// A kind of its own rather than own_permanents or choose_cards because
// a bot orders the answer by what the question is for, and a
// Ring-bearer is the creature a seat MOST wants to be legendary,
// evasive and looting — the opposite of own_permanents' "the permanent
// you would miss least" (ADR 0114 §4).
//
// It is not targeting: hexproof, shroud, protection and ward do not
// apply, and nothing can respond to the choice. The candidates are
// permanents on the battlefield, which every seat can see, so nothing
// about the prompt is redacted.
const PendingChoiceRingBearer PendingChoiceKind = "ring_bearer"

// EventRingTempted — the Ring tempted a player (CR 701.54d): they
// completed the actions in CR 701.54a, even if some or all of them
// were impossible. Actor = the player, CardID = the creature chosen as
// their Ring-bearer (uuid.Nil when they controlled none), Amount = how
// many times the Ring has now tempted them, Source = the card whose
// effect tempted them. Label is RingTemptedForced when the choice was
// made for the player because they controlled exactly one creature
// (owner decision 2).
//
// Emitted once per temptation, AFTER the emblem, the count and the
// designation are in place, so a "whenever the Ring tempts you" trigger
// reads the new state, and before the rest of the effect's sentence
// runs. Re-choosing the creature that already is the Ring-bearer emits
// it again with that creature: "that still counts as choosing that
// creature" (2023-06-16 ruling). Bumps the layer version for the reason
// EventBecameMonstrous does — a gated emblem static may have switched
// on, and the legendary grant has moved.
const EventRingTempted EventKind = "ring_tempted"

// RingTemptedForced is EventRingTempted's Label when the player
// controlled exactly one creature and it was chosen for them.
const RingTemptedForced = "forced"

// ringBearerQuestion is the prompt's reason — and so its dialog name in
// the action dock (ADR 0111 §10, ADR 0114 §9). Part of the labels
// contract once the client ships its display.
const ringBearerQuestion = "choose your Ring-bearer"

// RingTempted builds the gate on one of the Ring emblem's abilities:
// it exists while the Ring has tempted its owner n or more times
// (CR 701.54c). Read off the emblem object (Card.RingTemptations).
func RingTempted(n int) Designation { return Designation{Kind: DesignationRingTempted, N: n} }

// IsRingBearerOf reports whether `c` is `player`'s Ring-bearer, given
// that `c` is a permanent on the battlefield: it has the designation
// and that player controls it (CR 701.54e). A phased-out permanent is
// not on the battlefield (it sits in Game.PhasedOut, CR 702.26b), so a
// reader walking the battlefield never sees one.
//
// Pure: no *Game, no lock, so a layer-pass AppliesTo and a block rule
// can both call it.
func IsRingBearerOf(c Card, player uuid.UUID) bool {
	return player != uuid.Nil && c.RingBearer && c.Controller == player
}

// RingBearerOf is `player`'s Ring-bearer (CR 701.54e): the permanent
// on the battlefield, phased in, under their control, with the
// designation. uuid.Nil when they have none.
//
// Caller must hold g.mu (read or write).
func RingBearerOf(g *Game, player uuid.UUID) uuid.UUID {
	if g == nil || g.Battlefield == nil {
		return uuid.Nil
	}
	for i := range g.Battlefield.Cards {
		if IsRingBearerOf(g.Battlefield.Cards[i], player) {
			return g.Battlefield.Cards[i].InstanceID
		}
	}
	return uuid.Nil
}

// RingTemptCount is how many times the Ring has tempted `player` this
// game — "if the Ring has tempted you N or more times this game". Read
// off the Ring emblem; 0 for a player who has none.
//
// Caller must hold g.mu (read or write).
func RingTemptCount(g *Game, player uuid.UUID) int {
	if g == nil {
		return 0
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return 0
	}
	if e := ringEmblemOf(p); e != nil {
		return e.RingTemptations
	}
	return 0
}

// IsRingEmblem reports whether an emblem is the Ring.
func (c Card) IsRingEmblem() bool { return c.OracleID == RingEmblemKey }

// ringEmblemOf is the player's Ring emblem, or nil. Each player has at
// most one (2023-06-16 ruling), because the one write site below only
// creates it when there is none.
func ringEmblemOf(p *Player) *Card {
	if p == nil || p.Emblems == nil {
		return nil
	}
	for i := range p.Emblems.Cards {
		if p.Emblems.Cards[i].IsRingEmblem() {
			return &p.Emblems.Cards[i]
		}
	}
	return nil
}

// RingTemptsForEffect is the keyword action: the Ring tempts `player`
// (CR 701.54a and 701.54c, in the reminder card's order).
//
//  1. If the player has no Ring emblem, they get one.
//  2. The emblem's count goes up by one.
//  3. They choose a creature they control: none, one (chosen for them
//     and logged, owner decision 2), or two or more (asked, with the
//     ring_bearer prompt).
//  4. The chosen creature becomes their Ring-bearer, and their previous
//     one stops being one, even if it is phased out.
//  5. EventRingTempted is emitted — even when nothing could be chosen
//     (CR 701.54d).
//  6. `then` runs with the chosen creature (uuid.Nil when none): the
//     rest of the effect's sentence ("then search your library …").
//
// `source` is the card whose effect is tempting — the prompt's Source
// and the event's. It can PAUSE at step 3; the rest runs from the
// prompt's answer.
//
// A player who has left the game is not tempted (CR 800.4a), but
// `then` still runs, with nothing chosen, so the rest of the card is
// never stranded (#544).
//
// Caller must hold g.mu in write mode (it is an effect-time helper).
func (g *Game) RingTemptsForEffect(player, source uuid.UUID, then func(g *Game, ringBearer uuid.UUID) error) error {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		if then == nil {
			return nil
		}
		return then(g, uuid.Nil)
	}
	// Steps 1 and 2.
	emblem := ringEmblemOf(p)
	if emblem == nil {
		if p.Emblems == nil {
			p.Emblems = newZone(ZoneCommand, p.ID)
		}
		name := ringEmblemName
		if def := catalogDef(RingEmblemKey); def != nil && def.Emblem != nil && def.Emblem.Label != "" {
			name = def.Emblem.Label
		}
		// CR 114.1 / 114.3: no characteristics but its abilities. The
		// timestamp is the emblem's as it enters the command zone (CR
		// 613.7d), which its statics take (CR 613.7a).
		p.Emblems.PushTop(Card{
			InstanceID:           uuid.New(),
			Name:                 name,
			OracleID:             RingEmblemKey,
			Owner:                p.ID,
			Controller:           p.ID,
			EnteredBattlefieldAt: timeNowUnixNano(),
		})
		emblem = ringEmblemOf(p)
	}
	emblem.RingTemptations++
	// A gated ability of the emblem may just have switched on.
	g.layerVersion.Add(1)

	// Step 3. Fresh layers: "a creature" is the effective type.
	g.RecomputeLayersIfStaleLocked()
	cands := g.ringBearerCandidatesLocked(player)
	switch len(cands) {
	case 0:
		return g.finishRingTemptLocked(player, source, uuid.Nil, false, then)
	case 1:
		return g.finishRingTemptLocked(player, source, cands[0], true, then)
	}
	g.QueueChoiceForEffect(PendingChoice{
		Kind:        PendingChoiceRingBearer,
		Chooser:     player,
		FromPlayer:  player,
		Count:       1,
		Source:      source,
		Reason:      ringBearerQuestion,
		ChooseCards: cands,
		ChooseMin:   1,
		ChooseMax:   1,
		chooseCardsResume: &chooseCardsFrame{
			// Re-checked against the live battlefield on submit, and
			// pruned off the open prompt as the board moves: a
			// creature can leave between the question and the answer.
			zone: ZoneBattlefield,
			then: ringBearerThen(player, source, then),
		},
	})
	return nil
}

// ringBearerCandidatesLocked is "a creature you control": every
// creature `player` controls on the battlefield, in battlefield order.
// Phased-out permanents are not on the battlefield (CR 702.26b), so
// they are never offered.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) ringBearerCandidatesLocked(player uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// ringBearerThen is the prompt's continuation. A package-level
// constructor closing over scalars and the card's own continuation,
// the StackItem.Effect contract: it resolves against whichever *Game
// it is handed, which after an undo is the restored one.
//
// An empty pick is the drop path (the chooser left, CR 800.4a, or every
// candidate left the battlefield): the tempt still completes with
// nothing chosen (CR 701.54d).
func ringBearerThen(player, source uuid.UUID, then func(*Game, uuid.UUID) error) func(*Game, []uuid.UUID) error {
	return func(g *Game, picked []uuid.UUID) error {
		chosen := uuid.Nil
		if len(picked) == 1 {
			chosen = picked[0]
		}
		return g.finishRingTemptLocked(player, source, chosen, false, then)
	}
}

// finishRingTemptLocked is steps 4 to 6: the designation, the event and
// the rest of the sentence. The one write site of Card.RingBearer.
//
// Caller must hold g.mu in write mode.
func (g *Game) finishRingTemptLocked(player, source, chosen uuid.UUID, forced bool, then func(*Game, uuid.UUID) error) error {
	if chosen != uuid.Nil {
		// The zone re-check already ran on a prompted answer; this
		// also holds the controller, because "a creature you control"
		// is the question and the board may have moved since.
		if c := findBattlefieldCard(g, chosen); c == nil || c.Controller != player {
			chosen = uuid.Nil
		}
	}
	if chosen != uuid.Nil {
		// "Until another creature becomes your Ring-bearer" (CR
		// 701.54a): the previous one stops being one, wherever it is —
		// a phased-out Ring-bearer keeps the designation (CR 702.26d)
		// until this moment.
		unmark := func(z *Zone) {
			if z == nil {
				return
			}
			for i := range z.Cards {
				if z.Cards[i].Controller == player {
					z.Cards[i].RingBearer = false
				}
			}
		}
		unmark(g.Battlefield)
		unmark(g.PhasedOut)
		findBattlefieldCard(g, chosen).RingBearer = true
	}
	// A player who left while the prompt was open (the drop path) took
	// their emblem with them (CR 800.4a) and controls nothing a
	// "whenever the Ring tempts you" trigger could go on the stack for
	// (CR 800.4d), so there is no temptation left to announce. The rest
	// of the sentence still runs.
	if p := g.playerByIDLocked(player); p != nil && !p.Eliminated {
		ev := Event{
			Kind:   EventRingTempted,
			Actor:  player,
			Source: source,
			CardID: chosen,
			Amount: RingTemptCount(g, player),
		}
		if forced && chosen != uuid.Nil {
			ev.Label = RingTemptedForced
		}
		g.EmitEvent(ev)
		// The designation moved and the emblem's count rose: the
		// legendary grant and any gated emblem ability are stale.
		g.layerVersion.Add(1)
	}
	if then == nil {
		return nil
	}
	return then(g, chosen)
}

// ResolveRingBearer answers a PendingChoiceRingBearer.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveRingBearer(choiceID, chooserID uuid.UUID, picks []uuid.UUID) error {
	return g.resolveCardSetPick(PendingChoiceRingBearer, choiceID, chooserID, picks)
}
