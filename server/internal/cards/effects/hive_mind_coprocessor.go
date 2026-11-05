package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hive Mind Coprocessor — Artifact Creature — Borg {4}{B}, 4/4
// (Star Trek, #4973 batch 12):
//
//	"Artifact creatures you control have "Whenever this creature deals
//	 combat damage to a player, you may pay 1 life. If you do, draw a
//	 card.""
//
// A layer-6 grant of a catalog bundle (ADR 0093), so the trigger is
// each recipient's own: "this creature" is the artifact creature that
// connected, and "you" is its controller. The Coprocessor is an
// artifact creature itself, so it has the ability too. A creature that
// stops being an artifact (or changes controller) loses it with the
// layer pass, and two Coprocessors grant two instances, which trigger
// separately (CR 603.2).
//
// The payment is the shared "pay N life, draw M" prompt
// (mayPayLifeDraw), re-checked when the answer arrives (CR 119.4).
//
// No simplification.
const hiveMindCoprocessorGrant = "hive-mind-coprocessor/combat-pay-life-draw"

func init() {
	const text = "Whenever this creature deals combat damage to a player, you may pay 1 life. If you do, draw a card."
	Register(Spec{
		OracleID:     "8e6ae6ed-065c-47bd-a3c6-69b2608e7ddc",
		Name:         "Hive Mind Coprocessor",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: hiveMindCoprocessorGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Hive Mind Coprocessor — you may pay 1 life to draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return MayChoice{
							Question: "Hive Mind Coprocessor — pay 1 life to draw a card?",
							YesLabel: "Pay 1 life",
							NoLabel:  "Decline",
							LifeCost: 1,
							Then:     mayPayLifeDraw,
							Word:     "1",
						}.Apply(NewContext(g, item))
					}),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.IsCreature() && target.Controller == source.Controller
		}, hiveMindCoprocessorGrant), "Artifact creatures you control have \""+text+"\"")},
	})
}
