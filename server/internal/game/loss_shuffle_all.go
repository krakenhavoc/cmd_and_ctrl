package game

import "github.com/google/uuid"

// loss_shuffle_all.go — Lich's Mirror (S82 OLD batch 509, #5588):
//
//	"If you would lose the game, instead shuffle your hand, your graveyard,
//	 and all permanents you own into your library, then draw seven cards and
//	 your life total becomes 20."
//
// The kind is read where every other loss replacement is (replaceLossLocked).
// CR 614.6: the loss never happens. The cards leave as one simultaneous
// exit (a tuck can pause on a commander's CR 903.9 offer), and the tail
// shuffles and draws seven. The life total is set first so a state-based
// pass that runs while a prompt is open does not see life 0 again; the
// result is the same total (CR 119.5).

// lossShuffleAllTail is the continuation after the tuck: shuffle the
// player's library, then draw seven. Key is an on-disk identity: never
// renamed, never reused.
var lossShuffleAllTail = RegisterTail("then/loss-shuffle-all-draw-seven", func(g *Game, o TailOutcome) error {
	if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	return g.DrawNForEffect(o.Tail.Player, 7)
})

// replaceLossShuffleAllLocked applies the Lich's Mirror replacement for p,
// reporting whether it did. Caller must hold g.mu (write).
func (g *Game) replaceLossShuffleAllLocked(p *Player, cause LossCause) bool {
	ids, rows := g.gameActionReplacementSourcesLocked(p, ReplaceLoseGameShuffleAllDrawSevenLifeBecomes)
	if len(ids) == 0 {
		return false
	}
	source, row := ids[0], rows[0]
	g.EmitEvent(Event{
		Kind:   EventGameLossReplaced,
		Actor:  p.ID,
		Source: source,
		CardID: source,
		Label:  string(cause),
	})
	if delta := row.Life - p.Life; delta != 0 {
		_ = g.ChangePlayerLifeForEffect(source, p.ID, delta)
	}
	var cards []uuid.UUID
	if p.Hand != nil {
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
	}
	if p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			cards = append(cards, c.InstanceID)
		}
	}
	if g.Battlefield != nil {
		for _, c := range g.Battlefield.Cards {
			if c.Owner == p.ID {
				cards = append(cards, c.InstanceID)
			}
		}
	}
	_ = g.TuckCardsToLibraryKeyedForEffect(cards, TuckOptions{}, KeyedTail{Ref: lossShuffleAllTail, Player: p.ID, Source: source})
	return true
}
