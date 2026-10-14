package legal

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// controlled_play.go — moves for a player another player controls (CR
// 723, ADR 0164 §2 and §4).
//
//   - A controlled seat is offered nothing: its decisions are its
//     controller's (newEnumerator).
//   - A controller is offered the prompts the players it controls owe,
//     enumerated AS that player, so every choice case reads the right
//     seat's hand and permanents (controlledChoiceMoves). The moves keep
//     Player as the controlled seat; the actions layer runs them as that
//     seat.
//   - Word of Command's controlled_play prompt is answered by casting
//     (or playing) the one card as the controlled player, each move
//     naming the prompt, or — when there is no such move — by "it can't
//     be played" (controlledPlayMoves). The engine asks the same
//     question through the probe registered below, so it accepts that
//     answer exactly when no play is offered (#544).

// controlledPlayExpansion is the cap on a forced play's moves.
const controlledPlayExpansion = 200

func init() {
	game.SetControlledPlayProbe(func(g *game.Game, controlled, card uuid.UUID) bool {
		return len(controlledPlayCasts(g, controlled, card, Options{}.withDefaults())) > 0
	})
}

// enumeratorAs is an enumerator for seat that skips newEnumerator's
// "is this seat controlled" refusal: the controller's own expansion of
// the seat it controls. Nil when the seat is gone.
func enumeratorAs(g *game.Game, seat uuid.UUID, opts Options) *enumerator {
	p := playerByID(g, seat)
	if p == nil || p.Eliminated {
		return nil
	}
	e := &enumerator{g: g, p: p, seat: seat, opts: opts}
	e.leave()
	return e
}

// controlledPlayCasts is every play of card the controlled player could
// make now, as the engine would take it: the card's casts (whose timing
// the engine lifts for a forced play, CR 608.2g) and its land play.
// Mana is planned from their lands alone, because the auto-tapper asks
// the engine (Game.LandManaOnlyLocked). Caller must hold g.mu.
func controlledPlayCasts(g *game.Game, controlled, card uuid.UUID, opts Options) []Move {
	// One card, and every way of playing it is a choice the controller
	// is owed: the per-source cap that keeps a whole board's expansion
	// small would cut legal targets here.
	opts.MaxExpansionPerSource = max(opts.MaxExpansionPerSource, controlledPlayExpansion)
	e := enumeratorAs(g, controlled, opts)
	if e == nil || e.p.Hand == nil {
		return nil
	}
	for _, c := range e.p.Hand.Cards {
		if c.InstanceID != card {
			continue
		}
		speed := g.SorcerySpeedOpenLocked(controlled)
		landOwed := g.LandDropsRemainingLocked(controlled) > 0
		e.castMovesFromZone(c, game.ZoneHand, string(game.ZoneHand), speed, landOwed, g.AnyCastPermissionsForEffect(), true)
		break
	}
	return e.out
}

// controlledPlayMoves answers a controlled_play prompt c owed by e's
// seat: the controlled player's plays of the card, each naming the
// prompt, or the empty answer when there are none.
func (e *enumerator) controlledPlayMoves(c *game.PendingChoice) {
	if len(c.DiscardOptions) != 1 {
		return
	}
	reason := c.Reason
	if reason == "" {
		reason = "Word of Command"
	}
	plays := controlledPlayCasts(e.g, c.FromPlayer, c.DiscardOptions[0], e.opts)
	for _, m := range plays {
		var params map[string]any
		if err := json.Unmarshal(m.Params, &params); err != nil {
			continue
		}
		params["controlled_play"] = c.ID.String()
		m.Params = mustJSON(params)
		m.Label = reason + ": " + m.Label
		m.ControlledPlay = c.ID.String()
		e.add(m)
	}
	if len(plays) > 0 {
		return
	}
	e.add(Move{
		Type:           TypeResolveChoice,
		Player:         e.seat,
		Kind:           KindChoice,
		Label:          reason + ": it can't be played",
		Source:         c.Source,
		Params:         mustJSON(choiceParams{ChoiceID: c.ID.String(), CardIDs: []string{}}),
		ControlledPlay: c.ID.String(),
	})
}

// controlledChoiceMoves offers e's seat the prompts owed by each player
// it controls (CR 723.5), enumerated as that player. Reports whether
// any was owed: then they are the only moves, as a seat's own prompts
// are.
func (e *enumerator) controlledChoiceMoves() bool {
	owed := false
	for _, p := range e.g.Seats {
		if p == nil || p.ID == e.seat || p.Eliminated {
			continue
		}
		if e.g.ControllerOfPlayerLocked(p.ID) != e.seat {
			continue
		}
		sub := enumeratorAs(e.g, p.ID, e.opts)
		if sub == nil {
			continue
		}
		sub.report = e.report
		if sub.choiceMoves() {
			owed = true
		}
		e.out = append(e.out, sub.out...)
		e.cuts = append(e.cuts, sub.cuts...)
		e.added += sub.added
	}
	return owed
}
