package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alirios, Enraptured — Legendary Creature — Human {2}{U}, 2/3 (Theros Beyond
// Death, #5514):
//
//	"Alirios enters tapped.
//	 Alirios doesn't untap during your untap step if you control a Reflection.
//	 When Alirios enters, create a 3/2 blue Reflection creature token."
//
// The entry replacement is SelfEntersTapped (CR 614.1d); the untap clause is an
// untap-step restriction (CR 502.3) judged as the step begins, over the
// controller's Reflections as they stand then, "Reflection" read through Alirios's
// own text (CR 612.2); the token comes from the entry trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7dbf9e96-e61f-43db-8767-63a351cf4ef8",
		Name:         "Alirios, Enraptured",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label: "doesn't untap during your untap step if you control a Reflection",
			Restricts: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if target == nil || source == nil || target.InstanceID != source.InstanceID {
					return false
				}
				reflection := source.TextSubtype("Reflection")
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == source.Controller && c.HasSubtype(reflection) {
						return true
					}
				}
				return false
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Alirios, Enraptured — create a 3/2 blue Reflection creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/2 blue Reflection"), N: 1})),
		},
	})
}
