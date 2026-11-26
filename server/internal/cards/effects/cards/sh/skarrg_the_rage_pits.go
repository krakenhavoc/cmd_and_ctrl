package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skarrg, the Rage Pits — Land (Duel Decks: Heroes vs. Monsters, S82 old
// sets):
//
//	"{T}: Add {C}.
//	 {R}{G}, {T}: Target creature gets +1/+1 and gains trample until end of
//	 turn."
//
// A colourless mana ability and Unnatural Predation's body as an activated
// ability: both halves land on the target if it is still legal and end at
// cleanup (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92bac34e-2045-4331-842f-185711c1ac56",
		Name:         "Skarrg, the Rage Pits",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}{G}, {T}: Target creature gets +1/+1 and gains trample until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{R}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "trample")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SocB27TargetGetsAndGains(1, 1, "trample", "Skarrg, the Rage Pits")(item, effects.NewContext(g, item))
			},
		}},
	})
}
