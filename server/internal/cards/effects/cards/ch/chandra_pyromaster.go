package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Pyromaster — Legendary Planeswalker — Chandra {2}{R}{R}, loyalty 4
// (Magic 2015, #5556):
//
//	"+1: Chandra deals 1 damage to target player or planeswalker and 1 damage
//	     to up to one target creature that player or that planeswalker's
//	     controller controls. That creature can't block this turn.
//	 0:  Exile the top card of your library. You may play it this turn.
//	 −7: Exile the top ten cards of your library. Choose an instant or sorcery
//	     card exiled this way and copy it three times. You may cast the copies
//	     without paying their mana costs."
//
// The +1 is Soul of Shandalar's two clauses with 1 damage, then the creature
// (if it is still a legal target) can't block until end of turn. The 0 is an
// impulse exile whose permission lasts until end of turn. The −7 exiles the
// ten as one instruction; its keyed tail picks the instant or sorcery (asked
// only when there is more than one) and copies it three times with
// CopyAndCast (CR 707.12, free casts the controller takes in any order during
// the resolution); the copies not cast cease to exist.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68e12748-9a7b-4e99-a6de-be097a269342",
		Name:            "Chandra, Pyromaster",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Chandra deals 1 damage to target player or planeswalker and 1 damage to up to one target creature that player or that planeswalker's controller controls. That creature can't block this turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.S82OldB358PlayerOrPlaneswalkerAndTheirCreature(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := effects.S82OldB358DamageBothTargets(1)(g, item); err != nil {
						return err
					}
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(1)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Chandra, Pyromaster — can't block this turn"}.Apply(ctx)
				},
			},
			{
				Label: "0: Exile the top card of your library. You may play it this turn.",
				Cost:  effects.LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, game.CastPermission{
						Duration:   effects.DurationUntilEndOfTurn(ctx),
						Source:     item.SourceCardID,
						SourceName: "Chandra, Pyromaster",
						Label:      "Play the card Chandra, Pyromaster exiled",
					})
					return err
				},
			},
			{
				Label: "−7: Exile the top ten cards of your library. Choose an instant or sorcery card exiled this way and copy it three times. You may cast the copies without paying their mana costs.",
				Cost:  effects.LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return g.ExileTopKeyedForEffect(item.Controller, item.Controller, 10, ctx.ItemTail(chandraPyromasterExiled))
				},
			},
		},
	})
}

// chandraPyromasterExiled is the −7 once the ten are exiled: the instant and
// sorcery cards among the ones that landed are the candidates; one is
// chosen (asked only when there is a choice) and copied three times. An
// on-disk key.
var chandraPyromasterExiled = effects.CardTail("then/chandra-pyromaster-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	var candidates []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && (c.IsInstant() || c.IsSorcery()) {
			candidates = append(candidates, id)
		}
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return chandraPyromasterCopy(ctx.Game, ctx.Controller(), ctx.Source(), candidates[0])
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Chandra, Pyromaster — choose an instant or sorcery card exiled this way",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  chandraPyromasterChosen,
		LiveZone: game.ZoneExile,
	})
	return nil
})

// chandraPyromasterChosen copies the chosen card three times. An on-disk key.
var chandraPyromasterChosen = game.RegisterChooseCardsThen("choose-cards/chandra-pyromaster-copy-three-times", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return chandraPyromasterCopy(g, r.Chooser, r.Source, r.Picked[0])
})

// chandraPyromasterCopy is "copy it three times. You may cast the copies
// without paying their mana costs".
func chandraPyromasterCopy(g *game.Game, player, source, card uuid.UUID) error {
	return effects.CopyAndCast{Cards: []uuid.UUID{card, card, card}, Free: true}.Apply(effects.KeyedContext(g, player, source))
}
