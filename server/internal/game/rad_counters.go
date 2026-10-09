package game

import "github.com/google/uuid"

// rad_counters.go is the inherent triggered ability of rad counters
// (#2042). CR 728, reproduced because the file is a transcription of
// it:
//
//	728.1. Rad counters are a kind of counter a player can have (see
//	rule 122, "Counters"). There is an inherent triggered ability
//	associated with rad counters. This ability has no source and is
//	controlled by the active player. This is an exception to rule
//	113.8. The full text of this ability is "At the beginning of each
//	player's precombat main phase, if that player has one or more rad
//	counters, that player mills a number of cards equal to the number
//	of rad counters they have. For each nonland card milled this way,
//	that player loses 1 life and removes one rad counter from
//	themselves."
//
//	728.1a A card that refers to life loss "from radiation" refers to
//	life lost as a result of the triggered ability associated with rad
//	counters.
//
// (Quoted from the September 25, 2026 edition, the one AGENTS.md §6
// pins.)
//
// SHAPE. The monarch's two abilities (CR 725.2, monarch.go) and speed's
// (CR 702.179d, speed.go) are the precedent: a sourceless inherent
// trigger has no card for the harvester to find it on, so a Listener
// watches the event and queues a keyed StackItem through
// queueHarvestedTriggerLocked. It uses the stack like any other
// triggered ability, so it can be responded to (and Stifled), and a
// restore point taken while it waits names its body, "rad/mill", and
// rebuilds it in the running binary.
//
// WHOSE PHASE. "Each player's precombat main phase" is only ever the
// active player's: a player's main phase happens on their own turn. So
// it is one trigger a turn at most, for the active player, who is also
// its controller (CR 728.1). EventBeginPrecombatMain is emitted only for
// the first main phase of a turn (CR 505.1a), so an added main phase
// never fires it again.
//
// THE INTERVENING IF (CR 603.4). "If that player has one or more rad
// counters" is read as the phase begins (no counters, no trigger) and
// again on resolution (all removed in response, nothing happens). The
// number milled is the count AT RESOLUTION: "a number of cards equal to
// the number of rad counters they have" is read when the effect does
// it, so a proliferate in response mills one more card.
//
// "MILLED THIS WAY". The cards counted are the ones that ARRIVED in the
// graveyard (MillToZoneThenForEffect, CR 400.7): a card a replacement
// sent elsewhere ("if a card would be put into a graveyard from
// anywhere, exile it instead") was not milled, and costs no life and
// removes no counter. A library smaller than the count mills what it
// has (CR 701.17b) and nobody loses for it. Each card's type is read in
// the graveyard, the public zone it moved to (CR 701.17c).
//
// ONE LIFE LOSS. The life is lost as one event of N, not N events of 1:
// the ability is one instruction resolved once, and a "whenever you
// lose life" trigger sees one loss. The counters come off after it, in
// the order the text gives, and the number removed is the number of
// nonland cards milled — not the life that actually moved, so a player
// whose loss is replaced ("you gain life rather than lose life from
// radiation", Strong, the Brutish Thespian) still sheds the counters.

// radMillBody is "rad/mill": CR 728.1's effect, for Params.Player. It is
// keyed (ADR 0041 P9) because the ability has no catalog row and no
// source card; a restore point with it on the stack rebuilds it from
// the key alone.
var radMillBody = DelayedBody("rad/mill", func(g *Game, _ *StackItem, p EffectParams) error {
	return g.resolveRadTriggerLocked(p.Player)
})

// radTriggers is the Listener half of CR 728.1, registered from NewGame
// beside the monarch's and speed's.
type radTriggers struct{}

// OnEvent runs under g.mu in write mode from notifyListenersLocked.
func (radTriggers) OnEvent(g *Game, ev Event) {
	if ev.Kind == EventBeginPrecombatMain {
		g.radTriggerLocked(ev.Actor)
	}
}

// RadCountersOf is how many rad counters player has. Caller must hold
// g.mu.
func (g *Game) RadCountersOf(player uuid.UUID) int {
	p := g.playerByIDLocked(player)
	if p == nil {
		return 0
	}
	return p.Counters[CounterRad]
}

// radTriggerLocked queues CR 728.1's ability for `player`, whose
// precombat main phase has just begun, if they have a rad counter.
//
// Caller must hold g.mu in write mode.
func (g *Game) radTriggerLocked(player uuid.UUID) {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated || p.Counters[CounterRad] <= 0 {
		return
	}
	params := EffectParams{Player: player}
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind: StackItemTriggered,
		// CR 728.1: no source, controlled by the active player — who
		// is the player whose phase it is.
		Controller: player,
		Owner:      player,
		Label:      "Rad counters — " + p.Name + " mills",
		Body:       radMillBody.Key(),
		Params:     params,
		Effect:     bodyEffect(radMillBody.Key(), params),
	})
}

// resolveRadTriggerLocked is the resolution of CR 728.1's ability for
// `player`. See the file comment for each choice.
//
// Caller must hold g.mu in write mode.
func (g *Game) resolveRadTriggerLocked(player uuid.UUID) error {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return nil
	}
	// CR 603.4: the condition is checked again as the ability resolves.
	n := p.Counters[CounterRad]
	if n <= 0 {
		return nil
	}
	return g.MillToZoneThenForEffect(player, n, ZoneGraveyard, nil, func(g *Game, milled []uuid.UUID) error {
		nonland := 0
		for _, id := range milled {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				nonland++
			}
		}
		if nonland == 0 {
			return nil
		}
		return g.loseLifeFromRadiationLocked(player, nonland, func(g *Game, _ int) error {
			if pl := g.playerByIDLocked(player); pl == nil || pl.Eliminated {
				return nil
			}
			return g.AddPlayerCounterByForEffect(player, player, CounterRad, -nonland)
		})
	})
}

// loseLifeFromRadiationLocked is `player` losing `amount` life from
// radiation (CR 728.1a): the ordinary CR 614 life window, with the event
// marked so a replacement can tell it apart (ReplacementEvent.
// LifeFromRadiation) and the EventChangeLife it lands as says so too
// (Event.FromRadiation). `then` runs with the delta that moved, as for
// ChangePlayerLifeThenForEffect.
//
// The ability has no source (CR 728.1), so the event's Source is
// uuid.Nil.
//
// Caller must hold g.mu in write mode.
func (g *Game) loseLifeFromRadiationLocked(player uuid.UUID, amount int, then func(g *Game, applied int) error) error {
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		if then != nil {
			return then(g, 0)
		}
		return nil
	}
	ev := &ReplacementEvent{
		Kind:              RepEventLife,
		LifePlayer:        player,
		LifeDelta:         -amount,
		LifeFromRadiation: true,
	}
	if then != nil {
		ev.lifeTail = &lifeTail{then: then}
	}
	_, err := g.changeLifeThroughReplacementsLocked(ev)
	return err
}
