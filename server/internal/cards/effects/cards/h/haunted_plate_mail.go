package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haunted Plate Mail — Artifact — Equipment {4} (Magic 2015, #5556):
//
//	"Equipped creature gets +4/+4.
//	 {0}: Until end of turn, this permanent becomes a 4/4 Spirit artifact
//	 creature that's no longer an Equipment. Activate only if you control
//	 no creatures.
//	 Equip {4}"
//
// One continuous effect from the resolving ability, pinned to this permanent
// (CR 611.2c, 400.7): Creature added to its types (layer 4), Spirit added
// and Equipment lost, base power and toughness 4/4 (layer 7b). It stays an
// artifact. Once it is a creature and no longer an Equipment it falls off
// what it was attached to (CR 301.5c, 704.5n). The gate is read when the
// ability is activated, as printed: a creature you gain with the ability on
// the stack doesn't stop it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84594f86-724b-4f6b-b374-2265a00db566",
		Name:         "Haunted Plate Mail",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(4, 4)},
		Activated: []effects.ActivatedAbility{
			{
				Label:     "{0}: Until end of turn, this permanent becomes a 4/4 Spirit artifact creature that's no longer an Equipment. Activate only if you control no creatures.",
				Purpose:   game.Purpose{Answers: game.AnswerAnimate},
				Cost:      effects.ManaCost("{0}"),
				Condition: effects.S82OldB357YouControlNoCreatures,
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: append([]game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Spirit")...),
							game.LoseSubtypesMod("Equipment"),
						}, game.SetBasePTMods(4, 4)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Haunted Plate Mail — a 4/4 Spirit artifact creature, no longer an Equipment",
					}.Apply(ctx)
				},
			},
			effects.EquipAbility("{4}"),
		},
	})
}
