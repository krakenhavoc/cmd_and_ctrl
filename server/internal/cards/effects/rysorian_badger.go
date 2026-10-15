package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rysorian Badger — Creature — Badger {2}{G}, 2/2:
//
//	"Whenever this creature attacks and isn't blocked, you may exile up to
//	 two target creature cards from defending player's graveyard. If you
//	 do, you gain 1 life for each card exiled this way and this creature
//	 assigns no combat damage this turn."
//
// #3531: the (up to two) cards are targeted as the trigger goes on the
// stack, from the graveyard of the player the Badger is attacking. Saying
// yes is the "if you do", even for a trigger left with no legal target (a
// card exiled in response), which exiles nothing and gains nothing but still
// stops the Badger's damage.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cf9fc317-3184-4999-84ef-945a8842096e",
		Name:         "Rysorian Badger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(unblockedAttackerMay("Rysorian Badger", "exile up to two target creature cards from defending player's graveyard",
				func(ctx *Context, _ uuid.UUID) error {
					exiled := 0
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						exiled++
					}
					return GainLife{Amount: exiled}.Apply(ctx)
				}), targetCreatureCardsInDefendersGraveyard),
		},
	})
}

// targetCreatureCardsInDefendersGraveyard is "up to two target creature
// cards from defending player's graveyard", read off the source's attack.
func targetCreatureCardsInDefendersGraveyard(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	defender := g.DefendingPlayerForAttackForEffect(source.AttackingTarget)
	if defender == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "card_in_graveyard", Label: "up to two target creature cards from defending player's graveyard",
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature() && c.Owner == defender
		},
		Min: 0, Max: 2,
	}
}
