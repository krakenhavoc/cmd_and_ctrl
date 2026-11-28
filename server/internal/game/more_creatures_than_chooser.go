package game

import "github.com/google/uuid"

// more_creatures_than_chooser.go — #6100: a PLAYER target judged against the
// player who chooses it, on a triggered ability whose target the triggering
// player chooses (ChosenByTriggerActor, ADR 0019 amendment 2026-10-23).
//
// Oath of Druids: "At the beginning of each player's upkeep, that player
// chooses target player who controls more creatures than they do and is
// their opponent." The controller of the Oath controls the ability, but
// the upkeep player chooses the target (CR 603.3d with CR 601.2c), and
// both halves of the clause are relative to the upkeep player, not to the
// controller: "their opponent" is an opponent of the upkeep player, and
// "more creatures than they do" compares with the upkeep player's
// creatures. PlayerOK cannot say it, since it is handed the controller,
// never the event.
//
// So the clause carries it as DATA (TargetSpec.MoreCreaturesThanChooser),
// judged beside PlayerOK in the one player walk (playerChooserAdmits): the
// legal set the upkeep player is offered, the announce check (CR 601.2c),
// the resolution re-check (CR 608.2b, so a target who no longer controls
// more creatures by then makes the ability not resolve) and the bot's
// moves are one answer. The chooser is triggerChooserOf, the same reading
// ChosenByTriggerActor's own walk asks. A walk with no chooser admits
// nobody, never everybody. In a free-for-all game every other player
// still in the game is an opponent (GiftOpponentsLocked's rule).

// playerChooserAdmits is MoreCreaturesThanChooser's gate on one player
// candidate: true without the flag; with it, a player other than the
// clause's chooser who controls more creatures than the chooser does.
//
// Caller must hold g.mu.
func (s *TargetSpec) playerChooserAdmits(g *Game, src TargetSource, p *Player) bool {
	if !s.moreCardsInHandThanChooserAdmits(g, src, p) {
		return false
	}
	if !s.moreLandsThanChooserAdmits(g, src, p) {
		return false
	}
	if !s.MoreCreaturesThanChooser && !s.FewerCreatureCardsInGraveyardThanChooser {
		return true
	}
	who := triggerChooserOf(src)
	if who == uuid.Nil || p == nil || p.ID == who || p.Eliminated {
		return false
	}
	if s.MoreCreaturesThanChooser && g.creaturesControlledByLocked(p.ID) <= g.creaturesControlledByLocked(who) {
		return false
	}
	if s.FewerCreatureCardsInGraveyardThanChooser {
		return g.creatureCardsInGraveyardLocked(p.ID) < g.creatureCardsInGraveyardLocked(who)
	}
	return true
}

// creatureCardsInGraveyardLocked is how many creature cards are in
// `player`'s graveyard (Oath of Ghouls, CR 700.2 types read off the card).
//
// Caller must hold g.mu.
func (g *Game) creatureCardsInGraveyardLocked(player uuid.UUID) int {
	p := g.playerByIDLocked(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for i := range p.Graveyard.Cards {
		if p.Graveyard.Cards[i].IsCreature() {
			n++
		}
	}
	return n
}

// creaturesControlledByLocked is how many creatures `player` controls, by
// the types after the layers.
//
// Caller must hold g.mu.
func (g *Game) creaturesControlledByLocked(player uuid.UUID) int {
	if g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsCreature() {
			n++
		}
	}
	return n
}
