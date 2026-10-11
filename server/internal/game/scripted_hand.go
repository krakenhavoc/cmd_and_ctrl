package game

import "github.com/google/uuid"

// scripted_hand.go is the tutorial's scripted opening hand (ADR 0076
// §2.2, amended 2026-10-11, #2992). The practice table deals its human
// seat a fixed seven, so every lesson of the tutorial's first turn has
// a card to teach it with, and the rest of the library stays shuffled.
//
// It is a table-setup knob, not a rule: nothing in the comprehensive
// rules deals a chosen hand, and no table but the practice table sets
// one. The deal (dealAndBeginLocked) still shuffles first, so the
// library's order is exactly as random as at any other table; the
// scripted cards are then taken out of it by name, and the hand is
// filled to seven from the top.
//
// A mulligan deals a normal random seven. Mulligan shuffles the hand
// back and draws, and does not read the script: a player who sends the
// scripted hand back has asked for a different one, and the tutorial's
// steps wait for a turn in which they can still happen.

// ScriptOpeningHand names the cards playerID's opening hand is dealt,
// by card name, at most OpeningHandSize of them; a name may repeat (3
// Forests). It must be called before the deal: in the lobby, or while
// the opening roll is open. A name the library has no card left for at
// the deal is skipped, and the hand is filled to seven from the top of
// the shuffled library as usual, so a script never deals a short hand.
// An empty list clears the script.
//
// Not persisted: only the practice table sets it, and a practice table
// writes no restore point.
func (g *Game) ScriptOpeningHand(playerID uuid.UUID, names []string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateLobby && g.OpeningRoll == nil {
		return ErrGameAlreadyStarted
	}
	if len(names) > OpeningHandSize {
		return ErrInvalidParam
	}
	if g.playerByIDLocked(playerID) == nil {
		return ErrPlayerNotFound
	}
	if len(names) == 0 {
		delete(g.openingHands, playerID)
		return nil
	}
	if g.openingHands == nil {
		g.openingHands = map[uuid.UUID][]string{}
	}
	g.openingHands[playerID] = append([]string(nil), names...)
	return nil
}

// dealScriptedLocked moves p's scripted opening cards from its
// (already shuffled) library to its hand, and returns how many it
// moved. Each name takes the topmost library card of that name, so the
// rest of the library keeps its shuffled order.
func (g *Game) dealScriptedLocked(p *Player) int {
	dealt := 0
	for _, name := range g.openingHands[p.ID] {
		for i := len(p.Library.Cards) - 1; i >= 0; i-- {
			if p.Library.Cards[i].Name != name {
				continue
			}
			c, err := p.Library.Remove(p.Library.Cards[i].InstanceID)
			if err != nil {
				break
			}
			p.Hand.PushTop(c)
			dealt++
			break
		}
	}
	return dealt
}
