package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valgavoth's Faithful — Creature — Human Cleric {B}, 1/1 (Duskmourn:
// House of Horror, #4466):
//
//	"{3}{B}, Sacrifice this creature: Return target creature card from
//	 your graveyard to the battlefield. Activate only as a sorcery."
//
// Priest of Fell Rites's reanimation at a mana price and with no {T}, so
// it can be activated the turn the Faithful arrives. Targets are chosen
// before costs are paid (CR 601.2c, CR 602.2b), so the Faithful can't
// return itself. Sorcery timing is CR 602.5d. The target is checked
// again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a871dfbf-1fb9-4d1c-81bb-e4f0b131a11b",
		Name:         "Valgavoth's Faithful",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B}, Sacrifice this creature: Return target creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.Reanimates(0)),
			Cost:         effects.Plus(effects.ManaCost("{3}{B}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets:      effects.TargetCreatureInYourGraveyard(),
			Effect:       effects.ReturnFirstLegalGraveyardTargetToBattlefield,
		}},
	})
}
