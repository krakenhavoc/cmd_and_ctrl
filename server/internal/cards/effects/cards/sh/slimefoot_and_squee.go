package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slimefoot and Squee — Legendary Creature — Fungus Goblin {B}{R}{G}, 3/3
// (March of the Machine Commander):
//
//	"Whenever Slimefoot and Squee enters or attacks, create a 1/1 green
//	 Saproling creature token.
//	 {1}{B}{R}{G}, Sacrifice a Saproling: Return this card and up to one other
//	 target creature card from your graveyard to the battlefield. Activate
//	 only as a sorcery."
//
// The second ability works from the graveyard (Cauldron Familiar's shape), so
// the card itself is not a target: only the "up to one other" card is. Both
// come back as the ability resolves, each only if it is still in the
// graveyard (CR 608.2b for the target, CR 602.5 for the source).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a6f4181-f5ad-4ecf-9566-e52870af14d0",
		Name:         "Slimefoot and Squee",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Slimefoot and Squee — create a 1/1 green Saproling creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{B}{R}{G}, Sacrifice a Saproling: Return this card and up to one other target creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost: game.AbilityCost{
				Mana:           "{1}{B}{R}{G}",
				SacrificeOther: effects.SacrificeSpec("a Saproling", effects.HasSubtype("Saproling")),
			},
			Targets: effects.Another(effects.TargetCardInGraveyard("up to one other target creature card from your graveyard",
				effects.YouOwn(), effects.Creature()).WithCount(0, 1)),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.ReturnThisFromGraveyardToBattlefield(g, item); err != nil {
					return err
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard && t.ID != uuid.Nil {
						if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			},
		}},
	})
}
