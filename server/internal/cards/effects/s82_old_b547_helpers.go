package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b547_helpers.go — shared bodies for OLD batch 547 (#5593, Lorwyn).

// S82OldB547CountYouControl counts the permanents the resolving ability's
// controller controls that have at least one of the given creature types,
// each permanent once, read through the layered view so a changeling counts
// (CR 702.73a). The types are words of the source's text (CR 612.2).
func S82OldB547CountYouControl(ctx *Context, subtypes ...string) int {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller != ctx.Controller() {
			continue
		}
		for _, st := range subtypes {
			if c.HasSubtypeWord(st) {
				n++
				break
			}
		}
	}
	return n
}

// S82OldB547HandCardsNamedLikeASpellCastThisTurn lists the cards in the
// player's hand that have the same name as a spell cast this turn by any
// player, each name as the spell had it on the stack (Game.TurnTally.SpellNames,
// CR 201.2a: either half of a split card). Caller holds g.mu.
func S82OldB547HandCardsNamedLikeASpellCastThisTurn(g *game.Game, player uuid.UUID) []uuid.UUID {
	cast := map[string]bool{}
	for _, r := range g.TurnTally.SpellNames {
		cast[r.Name] = true
	}
	var out []uuid.UUID
	if len(cast) == 0 {
		return nil
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return nil
	}
	for _, c := range p.Hand.Cards {
		for _, n := range game.NamesOf(c) {
			if cast[n] {
				out = append(out, c.InstanceID)
				break
			}
		}
	}
	return out
}

// S82OldB547TargetPlayerShufflesTheirLibrary is "Target player shuffles their
// library." as an ability's whole effect (Boggart Forager): each player still a
// legal target as it resolves shuffles their own library (CR 701.20).
func S82OldB547TargetPlayerShufflesTheirLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (ShuffleLibrary{Player: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
