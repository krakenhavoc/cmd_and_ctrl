package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ursine Champion — Creature — Human Berserker {1}{G}, 2/2 (Core Set 2019):
//
//	"{5}{G}: This creature gets +3/+3 and becomes a Bear Berserker until end
//	 of turn. Activate only once each turn."
//
// One resolution, two layered pieces pinned to this object (CR 400.7): the
// layer 7c +3/+3 and a layer 4 "becomes a Bear Berserker" that replaces its
// creature types (CR 205.1a). The once-per-turn limit is checked when the
// ability is announced (CR 602.1b).
//
// No simplification.
func init() {
	const label = "{5}{G}: This creature gets +3/+3 and becomes a Bear Berserker until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "4f142d99-8146-4b16-a2cc-ac03da19b960",
		Name:         "Ursine Champion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:    effects.ManaCost("{5}{G}"),
			Condition: effects.AllConditions(effects.OncePerTurnActivation(label), func(g *game.Game, _, source uuid.UUID) bool {
				_, ok := g.LookupCardForEffect(source)
				return ok
			}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 3, Toughness: 3,
					Label: "Ursine Champion — +3/+3 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return effects.ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     []game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes("Bear", "Berserker")...)},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Ursine Champion — becomes a Bear Berserker",
				}.Apply(ctx)
			},
		}},
	})
}
