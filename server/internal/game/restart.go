package game

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

// restart.go — CR 727, restarting the game (#3856, ADR 0201).
//
// One card restarts the game: Karn Liberated's −14, "Restart the game,
// leaving in exile all non-Aura permanent cards exiled with Karn
// Liberated. Then put those cards onto the battlefield under your
// control." CR 727.1 ends the game with no winner and starts a new one
// among the players still in it, through rule 103, with the player who
// controlled the ability as the starting player (727.1a). Every card in
// the old game is in the new one (727.2) except the ones the effect
// exempts (727.5), and the effect's "then" happens just before the first
// turn's untap step, after the mulligans (727.4).
//
// The process has the shape CR 724.1's end-the-turn has (end_turn.go),
// for the same reason: it may not run inside a resolving callback. The
// game the ability is resolving in is about to stop existing, and the
// resolution frame still has its own tail to run against it (the
// resolved line, the CR 704.3 hold). So it happens in three beats:
//
//  1. RestartTheGameForEffect, from the resolving effect, records the
//     request on Game.RestartPending and parks priority. Nothing else.
//  2. The resolution's CR 704.3 boundary (runStateChecksLocked, once
//     nothing holds it) consumes it in finishRestartLocked: the new game
//     is built and adopted in place, and the pregame procedure runs
//     through the same path Start uses (beginPregameLocked,
//     dealAndBeginLocked) with the restarter as the starting player.
//     The effect's "then" waits on Game.RestartTail.
//  3. The mulligan window's close (settleMulliganLocked) runs the tail:
//     the exempted cards enter together under the restarter's control,
//     and only then does the first turn's untap step begin.
//
// The new game is adopted IN PLACE (RestoreFrom), because the room, the
// hub, the bot runners and every handler hold this *Game and its mutex.
// A fresh NewGame supplies every game-scoped field at its starting
// value, so a field added to Game later resets on a restart without
// anyone remembering to list it here; what is CARRIED across is the
// short, named list in buildRestartedGameLocked.
//
// Every card in the new game is a NEW OBJECT with a new instance ID.
// The event log is kept (it is the table's history, and the replay and
// the undo stack both read it by Seq), and much of the engine walks it
// by card ID — Approach of the Second Sun's "this game", the CR 607
// exile record — so a reused ID would let the old game's history speak
// for a card in the new one.

// GameRestart is a restart in progress: the request a resolving effect
// made (Game.RestartPending) and, once the new game has begun, its
// "then" (Game.RestartTail). Plain data: carried by Clone, RestoreFrom
// and the snapshot.
type GameRestart struct {
	// Restarter controlled the spell or ability that restarted the
	// game, and is the new game's starting player (CR 727.1a). The
	// tail's cards enter under their control.
	Restarter uuid.UUID `json:"restarter"`
	// Source is the card whose ability restarted the game, for the log.
	// On the tail it names an object of the OLD game and finds nothing;
	// SourceName is what the log prints.
	Source     uuid.UUID `json:"source,omitempty"`
	SourceName string    `json:"sourceName,omitempty"`
	// Exempt is CR 727.5's cards: they stay in exile, out of their
	// owners' decks. On RestartPending they are old-game IDs; on
	// RestartTail they are the same cards' new-game IDs, and they are
	// the cards the tail puts onto the battlefield.
	Exempt []uuid.UUID `json:"exempt,omitempty"`
}

func cloneGameRestart(r *GameRestart) *GameRestart {
	if r == nil {
		return nil
	}
	out := *r
	out.Exempt = append([]uuid.UUID(nil), r.Exempt...)
	return &out
}

// RestartTheGameForEffect restarts the game (CR 727) from inside a
// resolving spell or ability. `restarter` is the player who controlled
// it, `source` its card, and `exempt` the cards CR 727.5 leaves out of
// the new game's decks — Karn Liberated's "non-Aura permanent cards
// exiled with Karn Liberated". The exempted cards are put onto the
// battlefield under the restarter's control as the new game's mulligans
// end (CR 727.4); an effect that exempts cards it does not then put
// onto the battlefield does not exist and is not modelled.
//
// It must be the LAST thing the resolving effect does: the game it is
// resolving in ends at the resolution's boundary. A no-op outside an
// active game.
//
// Caller must hold g.mu.
func (g *Game) RestartTheGameForEffect(restarter, source uuid.UUID, exempt []uuid.UUID) {
	if g.State != StateActive || len(g.Seats) == 0 {
		return
	}
	name := ""
	if z, pos := g.locateCardLocked(source); z != nil && pos >= 0 {
		name = EffectiveOf(&z.Cards[pos]).Name
	}
	g.RestartPending = &GameRestart{
		Restarter:  restarter,
		Source:     source,
		SourceName: name,
		Exempt:     append([]uuid.UUID(nil), exempt...),
	}
	// CR 727.1: the game ends immediately. Nobody gets priority in it
	// again; the boundary that follows this resolution begins the new
	// one.
	g.grantPriorityLocked(NoPriority)
}

// finishRestartLocked is CR 727's second beat: it ends this game and
// begins the new one (see the file header). Consumes RestartPending.
// Called from runStateChecksLocked once nothing holds the boundary, and
// nowhere else.
//
// Caller must hold g.mu (write).
func (g *Game) finishRestartLocked() {
	req := g.RestartPending
	g.RestartPending = nil
	if req == nil || g.State != StateActive {
		return
	}
	ng, exempt := g.buildRestartedGameLocked(req)
	g.RestoreFrom(ng)
	// RestoreFrom adopts what an undo adopts. These are the transient
	// fields it leaves alone because an undo never needs them reset;
	// a new game does.
	g.openingHands = nil
	g.libraryTopWatch = nil
	g.scopedEffectMemo = scopedEffectAdapterMemo{}
	g.promptKeys = promptKeyState{}
	g.TurnTally = TurnTally{FirstEvent: len(g.Events)}

	start := g.restartStartingSeatLocked(req.Restarter)
	g.beginPregameLocked()
	ev := Event{
		Kind:   EventGameRestarted,
		Actor:  req.Restarter,
		Source: req.Source,
		Label:  req.SourceName,
		Amount: len(exempt),
	}
	if req.Source != uuid.Nil && req.SourceName != "" {
		// The source is a card of the old game; nothing in the new one
		// answers to its ID, so the log line carries the name.
		ev.PublicNames = map[uuid.UUID]string{req.Source: req.SourceName}
	}
	g.EmitEvent(ev)
	// CR 727.1a, then rule 103 from 103.2b on (ADR 0201 §3): the
	// starting player is known, so there is no opening roll.
	g.dealAndBeginLocked(start)
	// CR 727.3: a player whose deck cannot supply seven cards loses
	// when state-based actions are first checked, whatever mulligans
	// they take. The deal stopped short; the attempt is what CR 704.5b
	// reads.
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if p.Library.Size()+p.Hand.Size() < OpeningHandSize {
			p.AttemptedEmptyDraw = true
		}
	}
	if len(exempt) > 0 {
		g.RestartTail = &GameRestart{
			Restarter:  req.Restarter,
			SourceName: req.SourceName,
			Exempt:     exempt,
		}
	}
}

// restartStartingSeatLocked is the seat of the player who restarted the
// game (CR 727.1a), or — when that player is no longer in it — the next
// seat in turn order that is.
func (g *Game) restartStartingSeatLocked(restarter uuid.UUID) int {
	n := len(g.Seats)
	from := 0
	for i, p := range g.Seats {
		if p != nil && p.ID == restarter {
			from = i
			break
		}
	}
	for k := 0; k < n; k++ {
		i := (from + k) % n
		if p := g.Seats[i]; p != nil && !p.Eliminated {
			return i
		}
	}
	return from
}

// buildRestartedGameLocked builds the new game CR 727.1 starts, in the
// lobby state with every player seated and every deck in place, ready
// for RestoreFrom and the pregame procedure. It returns the game and the
// new IDs of the exempted cards, which wait in its exile.
//
// What is CARRIED from the old game is listed here and nowhere else:
// the table's identity and settings (not restored by RestoreFrom
// anyway), the event log and its counters, the configuration the game
// was constructed with, the table's rolls, and each seat's identity and
// preferences (restartedPlayer). Everything else is NewGame's.
//
// Caller must hold g.mu.
func (g *Game) buildRestartedGameLocked(req *GameRestart) (*Game, []uuid.UUID) {
	ng := NewGame()
	ng.ID = g.ID
	ng.CreatedAt = g.CreatedAt
	ng.Settings = g.Settings
	ng.LoopThreshold = g.LoopThreshold
	// The log continues: Seq and batch numbers keep rising, so the undo
	// stack, the secret-vote check and the public log's fold, which all
	// compare Seq values, never see one go backwards.
	ng.Events = g.Events
	ng.eventSeq = g.eventSeq
	ng.eventBatch = g.eventBatch
	ng.Listeners = g.Listeners
	ng.BuiltinReplacements = g.BuiltinReplacements
	ng.tableRolls = g.tableRolls
	ng.sourceOrdinalNext = g.sourceOrdinalNext
	// A new key, derived: the new game's shuffles are not the old
	// game's streams replayed from turn 0, and an undo of the restart
	// that is redone derives the same key, as any undone shuffle does.
	ng.rngKey = restartRNGKey(g.rngKey)
	ng.layerVersion.Store(g.layerVersion.Load() + 1)
	ng.lastResolvedVersion.Store(g.lastResolvedVersion.Load())

	exempt := make(map[uuid.UUID]bool, len(req.Exempt))
	for _, id := range req.Exempt {
		exempt[id] = true
	}
	ng.Seats = make([]*Player, len(g.Seats))
	inGame := make(map[uuid.UUID]*Player, len(g.Seats))
	for i, p := range g.Seats {
		if p == nil || p.Eliminated {
			// CR 727.1: only the players in the game when it ended
			// start the new one. A player who has left keeps their
			// seat for spectating, as they did before.
			ng.Seats[i] = p
			continue
		}
		np := restartedPlayer(p)
		ng.Seats[i] = np
		inGame[p.ID] = np
	}

	// CR 727.2: every card in the game, wherever it is, phased out
	// included, goes to its owner. Tokens and copies are not cards and
	// do not (CR 111.7, 707.10). Zones are walked in a fixed order so a
	// given game builds the same pre-shuffle libraries every time.
	var exemptNew []uuid.UUID
	place := func(c Card) {
		owner := inGame[c.Owner]
		if owner == nil {
			// CR 800.4a: a player who left took their cards with them.
			return
		}
		wasExempt := exempt[c.InstanceID]
		ordinal, hasOrdinal := g.sourceOrdinals[c.InstanceID]
		for _, card := range restartCards(c) {
			if card.IsToken() || card.CardCopy || card.PrepareCopy {
				continue
			}
			card.InstanceID = uuid.New()
			if hasOrdinal {
				if ng.sourceOrdinals == nil {
					ng.sourceOrdinals = make(map[uuid.UUID]uint64)
				}
				ng.sourceOrdinals[card.InstanceID] = ordinal
			}
			switch {
			case wasExempt:
				// CR 727.5: not in the deck. It begins the new game in
				// exile, where the effect's "then" finds it; an
				// exempted commander does not begin in the command
				// zone and is still its deck's commander (727.5a).
				ng.Exile.PushTop(card)
				exemptNew = append(exemptNew, card.InstanceID)
			case card.IsCommander:
				// CR 103.2c.
				owner.Command.PushTop(card)
			case card.HasAstrotoriumBack():
				// CR 727.2 involves nontraditional cards in the new
				// game too, and CR 717.2 starts an Attraction card in
				// its owner's Attraction deck (ADR 0207 §2), from the
				// battlefield, the junkyard or the deck alike.
				seatAttractionCard(owner, card)
			default:
				owner.Library.PushTop(card)
			}
		}
	}
	walk := func(z *Zone, skip func(Card) bool) {
		if z == nil {
			return
		}
		for _, c := range z.Cards {
			if skip != nil && skip(c) {
				continue
			}
			place(c)
		}
	}
	for _, p := range g.Seats {
		if p == nil || inGame[p.ID] == nil {
			continue
		}
		walk(p.Command, nil)
		walk(p.Library, nil)
		walk(p.Hand, nil)
		walk(p.Graveyard, nil)
		walk(p.AttractionDeck, nil)
		walk(p.Junkyard, nil)
	}
	walk(g.Battlefield, nil)
	walk(g.PhasedOut, nil)
	walk(g.Stack, func(c Card) bool {
		// A copy of a spell is not a card (CR 707.10).
		item := g.StackMeta[c.InstanceID]
		return item != nil && item.IsCopy
	})
	walk(g.Exile, nil)
	return ng, exemptNew
}

// restartedPlayer is the new game's seat for a player of the old one:
// the same player — ID, name, seat, who is sitting there, the seat's
// own settings and the deck's outside-the-game containers — with every
// fact about the old game gone. TestRestartClassifiesEveryPlayerField
// holds every Player field to one side of that line.
func restartedPlayer(p *Player) *Player {
	np := newPlayerWithID(p.ID, p.Name, p.Seat)
	np.DeckImported = p.DeckImported
	np.DiscordID = p.DiscordID
	np.DiscordAvatarHash = p.DiscordAvatarHash
	np.DisplayName = p.DisplayName
	np.IsBot = p.IsBot
	np.BotTier = p.BotTier
	np.BotDeck = p.BotDeck
	np.Agent = p.Agent
	np.AgentClient = p.AgentClient
	np.TriggerOrder = p.TriggerOrder
	np.AutoAnswers = copyAutoAnswers(p.AutoAnswers)
	// Outside the game is outside every game (CR 400.11): a companion
	// still waiting there may be revealed again (CR 103.2b), and the
	// sideboard is still the pool. A card brought in from either is in
	// the game now, and CR 727.2 puts it in the deck.
	np.Companion = p.Companion
	np.Sideboard = p.Sideboard
	np.HasSideboard = p.HasSideboard
	return np
}

// restartCards is the card `c` as it begins the new game: a new object
// in a library, with nothing of the old game on it. It goes through
// MoveCard from a battlefield, the one place every per-object fact is
// stripped (CR 400.7), which also splits a melded permanent into its
// two cards and turns a double-faced card front face up; then a copy
// effect ends (the card is itself again), and the facts MoveCard leaves
// because they belong to a zone rather than to an object go too.
func restartCards(c Card) []Card {
	from := newZone(ZoneBattlefield, uuid.Nil)
	to := newZone(ZoneLibrary, c.Owner)
	from.Cards = []Card{c}
	if _, err := MoveCard(from, to, c.InstanceID); err != nil {
		return nil
	}
	out := to.Cards
	for i := range out {
		card := &out[i]
		card.restorePrintedSelf()
		card.Controller = card.Owner
		card.KnownBy = nil
		card.Counters = nil
		card.CounterStampedAt = nil
		card.ObjectEpoch = 0
		card.TurnOvers = 0
		card.RingTemptations = 0
		card.CommanderReturnDue = false
		card.PutInGraveyardFrom = ""
		card.PutInGraveyardThisTurn = false
		card.FaceDownListed = nil
		card.effective = nil
	}
	return out
}

// restartRNGKey derives the new game's RNG key from the old one.
func restartRNGKey(old [32]byte) [32]byte {
	h := sha256.New()
	h.Write(old[:])
	h.Write([]byte("cmdctrl/restart-the-game"))
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// runRestartTailLocked is CR 727's third beat: the restarting effect
// finishes resolving just before the first turn's untap step (727.4).
// The exempted cards enter together under the restarter's control,
// through the ordinary simultaneous entry, so their replacement effects
// ask what they ask and their enters triggers wait for the first player
// to receive priority. The first turn's step entry runs from the
// entry's continuation, so a card that stops to ask (a Clone's copy, a
// shockland's life) is answered before the untap step begins.
//
// Reports whether it took the step entry over; false means there was no
// tail and the caller begins the turn itself.
//
// Caller must hold g.mu (write).
func (g *Game) runRestartTailLocked() bool {
	tail := g.RestartTail
	if tail == nil {
		return false
	}
	entries := make([]BatchEntry, 0, len(tail.Exempt))
	for _, id := range tail.Exempt {
		if g.Exile != nil && g.Exile.Contains(id) {
			entries = append(entries, BatchEntry{CardID: id, From: ZoneExile})
		}
	}
	if len(entries) == 0 {
		g.RestartTail = nil
		return false
	}
	ctrl := tail.Restarter
	if p := g.playerByIDLocked(ctrl); p == nil || p.Eliminated {
		// Under its owner's control instead (ZoneEntryOptions).
		ctrl = uuid.Nil
	}
	if err := g.PutOntoBattlefieldTogetherThenForEffect(entries, ZoneEntryOptions{Controller: ctrl}, restartTailEntered); err != nil {
		g.EmitEvent(Event{Kind: EventEffectError, Actor: tail.Restarter, ErrorMsg: err.Error()})
		if g.RestartTail == tail && g.blockingChoiceLocked() == nil {
			// The entry refused without reaching its continuation;
			// the game still has to begin.
			_ = restartTailEntered(g, nil)
		}
	}
	return true
}

// restartTailEntered ends the restart and begins the first turn. A
// package-level func so a paused entry captures nothing.
//
// The permanents entered before the first turn began, so their
// controller has controlled them continuously since their most recent
// turn began (CR 302.6) and they are not summoning sick. A game's first
// turn does not begin through the rotation seam that clears the marker
// (onTurnBeganLocked), so it is cleared here.
func restartTailEntered(g *Game, entered []uuid.UUID) error {
	g.RestartTail = nil
	for _, id := range entered {
		if z, pos := g.locateCardLocked(id); z != nil && z.Kind == ZoneBattlefield && pos >= 0 {
			z.Cards[pos].SummonedThisTurn = false
		}
	}
	if g.State == StateActive && !g.MulligansOpen {
		g.runStepEntryHooksLocked()
	}
	return nil
}
