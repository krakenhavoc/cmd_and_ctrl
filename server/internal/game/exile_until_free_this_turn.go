package game

import "github.com/google/uuid"

// exile_until_free_this_turn.go — "exile cards from the top of your library
// until you exile <a card>. Until end of turn, you may cast that card without
// paying its mana cost. Put each other card exiled this way on the bottom of
// your library in a random order" (Codie, Vociferous Codex, S82 BNS).
//
// It is the exile-until walk cascade and discover share, with a different
// ending. The pile goes to the bottom at once (nothing waits on the cast), and
// the hit stays in exile under a {0}, cast-only permission that lasts until
// end of turn at the card's ordinary timing: an instant any time the caster
// has priority, a sorcery only in a main phase with an empty stack (the
// ability resolves while the triggering spell is still on the stack, so the
// sorcery waits). An uncast hit simply stays in exile when the turn ends. A
// walk that finds no hit bottoms every card it exiled.

// ExileUntilCastFreeThisTurnForEffect performs that walk for `player` and
// grants the free-cast permission over the hit. Every "nothing happened" case
// returns nil (CR 608.2c).
//
// `hit` is called during the walk only; it is never stored. Caller must hold
// g.mu.
func (g *Game) ExileUntilCastFreeThisTurnForEffect(player, source uuid.UUID, hit func(Card) bool) error {
	pile, found, err := g.exileUntilLocked(player, hit)
	if err != nil {
		return err
	}
	if err := g.PutOnBottomInRandomOrderForEffect(player, ZoneExile, pile); err != nil {
		return err
	}
	if found == uuid.Nil || !g.playerInGameLocked(player) {
		return nil
	}
	c, ok := g.cardInZoneLocked(g.Exile, found)
	if !ok {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:   player,
		Zone:     ZoneExile,
		Duration: g.UntilEndOfTurnDuration(),
		Cost:     "{0}",
		CastOnly: true,
		Source:   source,
		Label:    "Cast it without paying its mana cost this turn",
	}, []Card{c})
	return nil
}
