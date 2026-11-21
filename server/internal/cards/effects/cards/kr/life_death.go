package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Life // Death — split card (Apocalypse, #5614):
//
//	Life — Sorcery {G}: "All lands you control become 1/1 creatures until
//	 end of turn. They're still lands."
//	Death — Sorcery {1}{B}: "Return target creature card from your
//	 graveyard to the battlefield. You lose life equal to its mana value."
//
// Each half is an ADR 0034 face (ADR 0103): Life under the bare oracle
// ID, Death under "#1". Only one half can be cast (CR 709.3; the card has
// no fuse).
//
// Life is Rude Awakening's land animation at 1/1: the set is the lands
// you control as it resolves (CR 611.2c), so a land played afterwards is
// not animated. Each becomes a creature with base power and toughness
// 1/1 and keeps its land types and abilities. A land you haven't
// controlled continuously since your turn began can't attack, nor use a
// {T} ability, mana abilities included (CR 302.6).
//
// Death puts the card onto the battlefield under your control, then you
// lose life equal to its mana value, read from the card before it moved
// (a card
// with {X} in its cost counts X as 0, CR 202.3e). A card gone from the
// graveyard by resolution is an illegal target and the spell does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d476f4b3-db63-4756-ab76-4b35f63c2825",
		Name:         "Life",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ScopedEffectFor{
				Match: effects.And(effects.Land(), effects.YouControl()),
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.SetBasePowerMod(1),
					game.SetBaseToughnessMod(1),
				},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Life — lands you control are 1/1 creatures",
			}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "d476f4b3-db63-4756-ab76-4b35f63c2825#1",
		Name:         "Death",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, mv, ok := effects.FirstLegalTargetCardManaValue(ctx)
			if !ok {
				return nil
			}
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: -mv}.Apply(ctx)
		},
	})
}
