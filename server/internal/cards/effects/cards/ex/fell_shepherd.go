package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fell Shepherd — Creature — Avatar {5}{B}{B}, 8/6 (Commander 2013, S82 OLD
// b386):
//
//	"Whenever this creature deals combat damage to a player, you may return to
//	 your hand all creature cards that were put into your graveyard from the
//	 battlefield this turn.
//	 {B}, Sacrifice another creature: Target creature gets -2/-2 until end of
//	 turn."
//
// The return reads the set once as the trigger resolves, from your own
// graveyard, by each card's newest arrival there (Gerrard, Weatherlight Hero's
// read: a card that left and came back is a new object, CR 400.7); a token is
// not a card (CR 108.2). The -2/-2 is Erebos, Bleak-Hearted's row with a
// sacrifice of another creature and one more point of toughness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5fd78088-53db-453b-90a3-b8426b0a8269",
		Name:         "Fell Shepherd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverThisDealsCombatDamageToAPlayer("Fell Shepherd — you may return to your hand all creature cards put into your graveyard from the battlefield this turn",
				fellShepherdReturn), "Fell Shepherd — return the creature cards put into your graveyard from the battlefield this turn to your hand?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Sacrifice another creature: Target creature gets -2/-2 until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Fell Shepherd — -2/-2"}.Apply(ctx)
			},
		}},
	})
}

func fellShepherdReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ids := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
		return !c.IsToken() && c.IsCreature() && effects.PutIntoGraveyardFromBattlefieldThisTurn(g, c.InstanceID)
	})
	for _, id := range ids {
		if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
