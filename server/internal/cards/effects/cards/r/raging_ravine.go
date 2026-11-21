package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Ravine — Land:
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}.
//	 {2}{R}{G}: Until end of turn, this land becomes a 3/3 red and green
//	 Elemental creature with 'Whenever this creature attacks, put a
//	 +1/+1 counter on it.' It's still a land."
//
// The catalog's first manland. The animation is ONE continuous effect
// from a resolving ability (CR 611.2a, until end of turn), pinned to
// this object (CR 611.2c, CR 400.7), applied in each layer it touches
// at one timestamp (CR 613.6):
//
//   - layer 4: it gains the creature type and the Elemental subtype.
//     "It's still a land" (CR 205.1b): its land type and its mana
//     ability stay.
//   - layer 5: its colours BECOME red and green (a land is colourless).
//   - layer 6: it gains the quoted trigger, an ability bundle granted
//     with the effect (GrantAbilitiesFor, ADR 0093).
//   - layer 7b: its base power and toughness are set to 3/3, so a
//     +1/+1 counter, an anthem or a pump still applies on top (7c).
//
// The rest is the engine's: it is a creature that came under its
// controller's control this turn if it entered this turn, so it can't
// attack, and its own {T} mana ability can't be activated while it is
// a creature, until its controller has controlled it continuously
// since the turn began (CR 302.6, the 2025-07-25 ruling). The +1/+1
// counters stay on it after it stops being a creature and count again
// the next time it animates (the ruling). Activating it twice gives it
// the granted trigger twice, each instance functioning independently
// (CR 113.2c; the engine never merges identical grants), so it gets
// two counters when it attacks (the ruling). An
// attacking Ravine that stops being a creature leaves combat (CR
// 506.4).
//
// The trigger belongs to the land (ADR 0093 Decision 4), so "it" is
// the permanent it is granted to.
//
// No simplification.
const RagingRavineAttackCounter = "raging-ravine/attack-counter"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d38194e-b607-4ff4-9c19-0e8636d463bf",
		Name:         "Raging Ravine",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Grants: []effects.AbilityGrant{{
			Key: RagingRavineAttackCounter,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Raging Ravine — put a +1/+1 counter on it",
					effects.PlusOneCountersOnThis(1)),
			},
			Text: "Whenever this creature attacks, put a +1/+1 counter on it.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}{G}: Until end of turn, this land becomes a 3/3 red and green Elemental creature with \"Whenever this creature attacks, put a +1/+1 counter on it.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: item.SourceCardID,
					Keys:   []string{RagingRavineAttackCounter},
					Also: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Elemental")...),
						game.SetColorsMod(g.TextReaderForEffect(item).TextColors("R", "G")...),
						game.SetBasePowerMod(3),
						game.SetBaseToughnessMod(3),
					},
					Label: "Raging Ravine — a 3/3 red and green Elemental creature until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
