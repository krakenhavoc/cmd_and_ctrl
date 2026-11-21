package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvan Offering — Sorcery {X}{G}:
//
//	"Choose an opponent. You and that player each create an X/X green
//	 Treefolk creature token.
//	 Choose an opponent. You and that player each create X 1/1 green Elf
//	 Warrior creature tokens."
//
// Intellectual Offering's two separate choices, so the second opponent may
// differ from the first; each is a real prompt that stays a restore point.
// X is announced with the cast (CR 107.3a) and read back as each half
// resolves. The Treefolk is sized as it is created, so it is a plain
// template; the Elf Warriors are the table's 1/1 green one. The caster's
// tokens come first, then the opponent's, and with no opponent left to
// choose the caster's half still happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "732aaa4b-22ea-49e2-b09f-e6169f726537",
		Name:         "Sylvan Offering",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChoosePlayer{
				Among:    effects.Opponents,
				Question: "Sylvan Offering — choose an opponent; you and that player each create an X/X green Treefolk creature token",
				ThenKey:  sylvanOfferingTreefolk,
			}.Apply(ctx)
		},
	})
}

var sylvanOfferingTreefolk = effects.ChoosePlayerThen("option-pick/sylvan-offering-treefolk", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	x := ctx.X()
	for _, who := range sylvanOfferingPlayers(ctx, chosen) {
		if err := (effects.CreateToken{Controller: who, N: 1, Template: game.Card{
			Name:           "Treefolk",
			TypeLine:       "Token Creature — Treefolk",
			Colors:         []string{"G"},
			Power:          x,
			Toughness:      x,
			PrintedPTKnown: true,
		}}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Sylvan Offering — choose an opponent; you and that player each create X 1/1 green Elf Warrior creature tokens",
		ThenKey:  sylvanOfferingElves,
	}.Apply(ctx)
})

var sylvanOfferingElves = effects.ChoosePlayerThen("option-pick/sylvan-offering-elves", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	for _, who := range sylvanOfferingPlayers(ctx, chosen) {
		if err := (effects.CreateToken{Controller: who, N: ctx.X(), Template: effects.TokenCard("1/1 green Elf Warrior")}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// sylvanOfferingPlayers is "you and that player": the controller, then the
// chosen opponent when there is one.
func sylvanOfferingPlayers(ctx *effects.Context, chosen uuid.UUID) []uuid.UUID {
	if chosen == uuid.Nil {
		return []uuid.UUID{ctx.Controller()}
	}
	return []uuid.UUID{ctx.Controller(), chosen}
}
