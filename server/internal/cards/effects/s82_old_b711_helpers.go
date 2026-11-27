package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b711_helpers.go — shared helpers for OLD batch 711 (Portal Three
// Kingdoms, #5622): the two edicts in which the targeted opponent chooses a
// creature they control.

// OldB711DestroyChosen is Imperial Edict's continuation: destroy the creature
// the opponent chose. An on-disk key, never renamed.
var OldB711DestroyChosen = game.RegisterChooseCardsThen("choose-cards/s82-old-b711-imperial-edict-destroy",
	func(g *game.Game, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if !g.Battlefield.Contains(id) {
				continue
			}
			if err := g.DestroyPermanentForEffect(id); err != nil {
				return err
			}
		}
		return nil
	})

// OldB711OthersCantBlock is Eunuchs' Intrigues' continuation: every creature
// the choosing player controls except the one they chose can't block this
// turn. The effect doesn't change characteristics, so it is read live and
// covers a creature that arrives later (CR 611.2c). An on-disk key, never
// renamed.
var OldB711OthersCantBlock = game.RegisterChooseCardsThen("choose-cards/s82-old-b711-eunuchs-intrigues-others-cant-block",
	func(g *game.Game, r game.CardsChosen) error {
		return oldB711OthersCantBlock(g, r.Source, r.Chooser, r.Picked)
	})

func oldB711OthersCantBlock(g *game.Game, source, defender uuid.UUID, chosen []uuid.UUID) error {
	mod := game.AddRestrictionsMod(game.CantBlock)
	mod.Objects = OldB665PileRefs(g, chosen)
	g.RegisterScopedRuleEffectForEffect(source, game.ScopeYourCreaturesExceptObjects, defender,
		[]game.Mod{mod}, g.UntilEndOfTurnDuration(), "Eunuchs' Intrigues — other creatures can't block this turn")
	return nil
}

// OldB711ChooseACreatureThen is "target opponent chooses a creature they
// control" and then `then` runs on the pick. With no creature to choose, the
// choice is empty and `none` (nil for nothing) runs instead.
func OldB711ChooseACreatureThen(ctx *Context, victim uuid.UUID, question string, then game.ChooseCardsThen, none func() error) error {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == victim && c.IsCreature() {
			return Msc17ExileEdict(ctx, victim, question, Creature(), then)
		}
	}
	if none != nil {
		return none()
	}
	return nil
}

// OldB711NoCreatureCantBlock is Eunuchs' Intrigues when the opponent has no
// creature to choose: every creature they control can't block this turn.
func OldB711NoCreatureCantBlock(ctx *Context, defender uuid.UUID) error {
	return oldB711OthersCantBlock(ctx.Game, ctx.Source(), defender, nil)
}

// OldB711TapTheTargets is "tap <the targets>": every card target still legal
// as the spell resolves is tapped (CR 608.2b), in announce order; a target
// that has gone is skipped (Broken Dam, Stall for Time).
func OldB711TapTheTargets(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
