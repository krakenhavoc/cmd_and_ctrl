package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coward // Killer — split Sorcery (Doctor Who):
//
//	Coward {1}{R}: "Target creature can't block this turn and becomes a
//	 Coward in addition to its other types until end of turn.
//	 Time travel."
//	Killer {2}{R}{R}: "Killer deals 3 damage to target creature and each
//	 other creature that shares a creature type with it."
//
// A split card is two Specs: the card's oracle id for the left half and "#1"
// for the right (Fire // Ice). Coward pins both the can't-block restriction and
// the added creature type to the target until end of turn, then time travels
// (CR 701.56a). Killer reads the target's creature types as it resolves
// (game.SharesCreatureType: a changeling shares with every creature that has a
// type, CR 702.73a) and deals its damage to the target and every other such
// creature in one damage instance.
//
// No simplification.
func init() {
	const oracle = "21591447-c045-4317-8d0e-705627aa157b"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Coward",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock,
					Label: "Coward — can't block this turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtype("Coward"))},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Coward — becomes a Coward in addition to its other types until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.TimeTravel{}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Killer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				victim, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				ids := []uuid.UUID{t.ID}
				for _, other := range effects.MatchingBattlefield(ctx, effects.Creature()) {
					if other.InstanceID != t.ID && game.SharesCreatureType(&victim, &other) {
						ids = append(ids, other.InstanceID)
					}
				}
				return effects.DealDamageToEachThen(ctx, ids, 3, effects.RecipientThen{})
			}
			return nil
		},
	})
}
