package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moraug, Fury of Akoum — Legendary Creature — Minotaur Warrior
// {4}{R}{R}, 6/6 (EDHREC top 1,000, #3654):
//
//	"Each creature you control gets +1/+0 for each time it has attacked
//	 this turn.
//	 Landfall — Whenever a land you control enters, if it's your main
//	 phase, there's an additional combat phase after this phase. At the
//	 beginning of that combat, untap all creatures you control."
//
// The pump is a layer 7c static reading the per-object attack tally
// (Aurelia's "first time each turn" memory), so a creature that attacks
// in the extra combat reads +2/+0 and one that has not attacked reads
// nothing. The landfall half is the extra-combat machinery Aurelia and
// Combat Celebrant already use, gated on a main phase of its
// controller's turn (the intervening-if of CR 603.4).
//
// DECLARED SIMPLIFICATION (untap timing): the creatures untap as the
// trigger resolves, as Aurelia's do, rather than at the beginning of
// the additional combat. The two differ only if you tap a creature
// between the land and that combat.
func init() {
	Register(Spec{
		OracleID:     "6288c0cb-cf32-4ae9-abf0-104ec71654fa",
		Name:         "Moraug, Fury of Akoum",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Your creatures untap as the extra combat is set up, not at the beginning of that combat."},
		Static: []game.StaticAbility{{
			Label: "Each creature you control gets +1/+0 for each time it has attacked this turn.",
			// The count changes at each declaration, which is the event
			// that re-evaluates a static flagged this way.
			DependsOnAttackingStatus: true,
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				c.Power += g.TimesAttackedThisTurn(target.InstanceID)
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(LandEnteredUnderYourControl, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return source != nil && b14InYourMainPhase(g, source.Controller)
			}),
				"Moraug, Fury of Akoum — additional combat phase, untap all creatures you control",
				untapAllYouControlThenExtraCombat),
		},
	})
}
