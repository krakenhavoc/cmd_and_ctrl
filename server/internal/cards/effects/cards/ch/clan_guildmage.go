package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clan Guildmage — Creature — Human Shaman {R}{G}, 2/2 (Ravnica Allegiance):
//
//	"{1}{R}, {T}: Target creature can't block this turn.
//	 {2}{G}, {T}: Target land you control becomes a 4/4 Elemental creature
//	 with haste until end of turn. It's still a land."
//
// Two tap abilities: the first is a "can't block this turn" restriction on the
// target creature; the second is Elemental Uprising's animation (without its
// must-be-blocked rider).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "838eeca8-02f9-44fa-8170-1426622adf30",
		Name:         "Clan Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}, {T}: Target creature can't block this turn.",
				Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					targets := ctx.LegalTargets()
					if len(targets) == 0 || targets[0].Kind != game.TargetCard {
						return nil
					}
					return effects.RestrictUntilEOT{Target: targets[0].ID, Restrictions: game.CantBlock,
						Label: "Clan Guildmage — can't block this turn"}.Apply(ctx)
				},
			},
			{
				Label:   "{2}{G}, {T}: Target land you control becomes a 4/4 Elemental creature with haste until end of turn. It's still a land.",
				Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost()),
				Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					land := effects.FirstLegalBattlefieldTarget(ctx)
					if land == uuid.Nil {
						return nil
					}
					const label = "Clan Guildmage — the land becomes a 4/4 Elemental creature with haste until end of turn"
					if err := (effects.BecomeCreatureUntilEOT{Target: land, Types: []string{"Creature"}, Subtypes: []string{"Elemental"}, SetPower: 4, SetToughness: 4, Label: label}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: land, Keywords: []string{"haste"}, Label: label}.Apply(ctx)
				},
			},
		},
	})
}
