package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xathrid Gorgon — Creature — Gorgon {5}{B}, 3/6 (Magic 2013):
//
//	"Deathtouch
//	 {2}{B}, {T}: Put a petrification counter on target creature. It gains
//	 defender and becomes a colorless artifact in addition to its other types.
//	 Its activated abilities can't be activated."
//
// The counter has no ability of its own; the rest is one indefinite effect on
// the target as it is when the ability resolves (CR 611.2a, CR 611.2c), locked
// to that object: a layer 4 added Artifact type, a layer 5 colourless, a
// layer 6 defender, and the bar on activating its abilities, mana abilities
// included, since the text names all of them. It stays when the counter does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e65fac0-23fc-4a32-bd36-4e3f7ebfe33d",
		Name:            "Xathrid Gorgon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}: Put a petrification counter on target creature. It gains defender and becomes a colorless artifact in addition to its other types. Its activated abilities can't be activated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.AddCounter{Target: id, Kind: "petrification", N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.ScopedEffectFor{
					Target: id,
					Mods: []game.Mod{
						game.AddTypesMod("Artifact"),
						game.SetColorsMod(),
						game.AddKeywordsMod("defender"),
						game.AddRestrictionsMod(game.CantActivate | game.CantActivateMana),
					},
					Duration: game.IndefiniteDuration(),
					Label:    "Xathrid Gorgon — defender, a colorless artifact, its abilities can't be activated",
				}.Apply(ctx)
			},
		}},
	})
}
