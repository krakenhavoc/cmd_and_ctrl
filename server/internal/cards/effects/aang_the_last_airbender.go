package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aang, the Last Airbender — 3/2 Legendary Creature — Human Avatar
// Ally for {3}{W}:
//
//	"Flying
//	When Aang enters, airbend up to one other target nonland
//	permanent. (Exile it. While it's exiled, its owner may cast it
//	for {2} rather than its mana cost.)
//	Whenever you cast a Lesson spell, Aang gains lifelink until end
//	of turn."
//
// The first card in the catalog to airbend anything, and the reason
// the mechanic's three pieces were built: an alternative cost
// (#257), a permission with no expiry, and an exile-a-targeted-
// permanent primitive. See airbend.go and game/exile_play.go.
//
// Shape notes:
//
//   - "Up to one … target" is modelled as an optional trigger with
//     exactly one target, the same way Deputy of Acquittals and Sun
//     Titan model theirs. Declining the prompt is how the caster
//     chooses zero. The observable difference from a true "up to
//     one" is that the ability never goes on the stack with no
//     targets, so nothing can respond to an Aang that airbends
//     nothing — a distinction with no card in the catalog able to
//     notice it. The trigger is likewise dropped outright when the
//     board holds no legal target (CR 603.3d, engine-wide).
//
//   - "OTHER target nonland permanent" cannot be expressed in the
//     target clause: TargetSpec is built at init() and NotSelf needs
//     an InstanceID that doesn't exist until Aang is on the
//     battlefield. Same sandbox gap Deputy of Acquittals carries,
//     and the same remedy — the picker offers Aang, and the Effect
//     declines to airbend the source, so the printed restriction
//     holds at resolution.
//
// SIMPLIFICATION — the Lesson clause is not implemented. "Aang
// gains lifelink until end of turn" needs a continuous effect with
// a turn-scoped duration, and the layer engine has no such thing:
// StaticAbility is recomputed from the battlefield every pass, so
// there is nowhere for a floating until-end-of-turn grant to live.
// Granting lifelink permanently would be stronger than printed,
// which is the one direction that isn't allowed, so the clause is
// omitted entirely. Aang is weaker than the real card by one
// conditional lifelink.
func init() {
	Register(Spec{
		OracleID:        "70564c3a-858f-498e-8b92-acb3ca54ae7e",
		Name:            "Aang, the Last Airbender",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("another target nonland permanent", Nonland()),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Aang, the Last Airbender — airbend a nonland permanent",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						target := item.Targets[0]
						// "Another": Aang does not airbend himself.
						if target.ID == item.SourceCardID {
							return nil
						}
						return Airbend{Target: target.ID}.Apply(NewContext(g, item))
					})
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Aang, the Last Airbender — airbend a nonland permanent? (Exile it; its owner may cast it for {2}.)",
			},
		}},
	})
}
