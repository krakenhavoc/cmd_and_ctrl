package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b471_helpers.go — Scars of Mirrodin, S82 older sets batch 471
// (#5577): small bodies the batch's cards share. Append-only.

// OldB471RevealUntilPutOntoBattlefield is the "reveals cards from the top of
// their library until they reveal a <match> card. That player puts that card
// onto the battlefield, then shuffles all other cards revealed this way into
// their library" half of Shape Anew (Transmogrify's reveal, for an artifact).
// The player puts it there, so it enters under their control; a library with
// no match reveals itself entirely and shuffles. The rest never left the
// library, so shuffling it is a shuffle of the library.
func OldB471RevealUntilPutOntoBattlefield(ctx *Context, player uuid.UUID, match func(game.Card) bool, reason string) error {
	run, hit := RevealUntil(ctx, player, match, reason)
	return ctx.Game.PutByEnteringPlayerForEffect(func() error {
		return PutFromLibraryOntoBattlefield{
			Player:   player,
			Cards:    run,
			Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
			All:      true,
			ThenTail: S82OldB109ShuffleRestTail,
		}.Apply(ctx)
	})
}

// OldB471UnattachEquipmentFrom is "unattach all Equipment from <hosts>":
// every Equipment permanent attached to one of the hosts as the instruction
// is read.
func OldB471UnattachEquipmentFrom(ctx *Context, hosts []uuid.UUID) error {
	var equipment []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.AttachedTo.Kind != game.TargetCard || !c.HasSubtype("Equipment") {
			continue
		}
		for _, h := range hosts {
			if c.AttachedTo.ID == h {
				equipment = append(equipment, c.InstanceID)
				break
			}
		}
	}
	for _, id := range equipment {
		if err := ctx.Game.UnattachForEffect(id); err != nil {
			return err
		}
	}
	return nil
}

// OldB471ReturnThisSpellIfLandAmong is the clause after Psychic Miasma's
// discard: "If a land card is discarded this way, return Psychic Miasma to
// its owner's hand." Read off what really left the hand.
func OldB471ReturnThisSpellIfLandAmong(g *game.Game, spell uuid.UUID, cards []uuid.UUID) {
	for _, id := range cards {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			g.ReturnResolvingSpellToHandForEffect(spell)
			return
		}
	}
}
