package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Extraction — Sorcery {2}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Target player reveals their hand and discards all cards of each of
//	 the sacrificed creature's colors."
//
// The colours are the sacrificed creature's as it last existed on the
// battlefield (CR 608.2h), read off the payment record (ADR 0113 §1). A
// card that shares at least one of them is discarded; a colourless
// sacrifice reveals the hand and discards nothing (the 2004-10-04
// rulings).
//
// "Discards all" names the cards, so there is nothing to choose: the
// discard prompt asks for exactly the matching cards and accepts only
// that set.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:       "0077740a-528b-4ee3-b331-fac321b95302",
		Name:           "Mind Extraction",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			player := item.Targets[0].ID
			p := ctx.Game.PlayerByIDForEffect(player)
			if p == nil {
				return nil
			}
			ctx.Game.RevealHandForEffect(player)
			info, ok := ctx.SacrificedPermanent()
			if !ok || len(info.Characteristic.Colors) == 0 {
				return nil
			}
			matches := func(c game.Card) bool {
				for _, col := range info.Characteristic.Colors {
					if c.HasColor(col) {
						return true
					}
				}
				return false
			}
			n := 0
			for _, c := range p.Hand.Cards {
				if matches(c) {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player:   player,
				Source:   item.SourceCardID,
				N:        n,
				Question: "Mind Extraction — discard every card that shares a color with the sacrificed creature",
				Validate: func(picked []game.Card) bool {
					for _, c := range picked {
						if !matches(c) {
							return false
						}
					}
					return true
				},
			})
			return nil
		},
	})
}
