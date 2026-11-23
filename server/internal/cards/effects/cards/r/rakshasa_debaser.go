package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakshasa Debaser — Creature — Demon {4}{B}{B}, 6/6 (Commander Starter
// Decks):
//
//	"Whenever this creature attacks, put target creature card from defending
//	 player's graveyard onto the battlefield under your control.
//	 Encore {6}{B}{B} ({6}{B}{B}, Exile this card from your graveyard: For
//	 each opponent, create a token copy that attacks that opponent this turn
//	 if able. They gain haste. Sacrifice them at the beginning of the next end
//	 step. Activate only as a sorcery.)"
//
// The attack trigger targets, so its clause is a fact about what happened
// (TargetsFrom, Mjölnir's shape): the defending player is the one this
// attacker named, a planeswalker's or battle's controller included (CR 506.3),
// and only creature cards in THAT player's graveyard may be picked. It is
// re-judged as the trigger resolves (CR 608.2b) and the card enters under the
// Debaser's controller. Encore is the shared keyword (encore.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db286eef-f0df-4f7c-8e06-957447166950",
		Name:         "Rakshasa Debaser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclared(ev, source)
			},
			TargetsFrom: rakshasaDebaserTarget,
			Key:         "Rakshasa Debaser — put target creature card from defending player's graveyard onto the battlefield under your control",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				ids := effects.LegalTargetCardIDs(ctx)
				if len(ids) == 0 {
					return nil
				}
				return effects.ReturnFromGraveyardTogether{Targets: ids[:1], Controller: ctx.Controller()}.Apply(ctx)
			},
		}},
		Activated: []effects.ActivatedAbility{effects.Encore("{6}{B}{B}")},
	})
}

// rakshasaDebaserTarget is "target creature card from defending player's
// graveyard": the defending player of the attack declaration that triggered it.
func rakshasaDebaserTarget(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	defender := effects.DefendingPlayerOfAttack(g, tc.Event)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode:  "card_in_graveyard",
		Label: "target creature card from defending player's graveyard",
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Owner == defender
		},
		Min: 1, Max: 1,
	}
}
