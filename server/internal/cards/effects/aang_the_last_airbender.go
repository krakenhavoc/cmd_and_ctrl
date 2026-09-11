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
//   - The Lesson clause is LIVE as of S32. It shipped omitted in
//     S23 because "Aang gains lifelink until end of turn" needed a
//     continuous effect with a turn-scoped duration and the layer
//     engine had nowhere to put one — every static was recomputed
//     from the battlefield, so a floating grant could not exist.
//     `GrantKeywordUntilEOT` (until_end_of_turn.go) is that place
//     now: the grant lands in `Game.TurnScopedStatics`, the
//     recompute picks it up, and the cleanup step sweeps it
//     (CR 514.2). Lifelink is one of the twelve keywords the
//     combat code honours, so the life gain is real, not cosmetic.
//     See ADR 0035.
//
//   - "Lesson spell" is matched on the cast card's type line, so it
//     fires for any Lesson — a catalog one or a non-catalog card
//     put into a hand by the dev spawner. The trigger reads the
//     spell off the stack the way Beast Whisperer's does; Aang's
//     own cast can't trigger it (he isn't on the battlefield yet,
//     and he isn't a Lesson).
//
//   - The grant is pinned to Aang's battlefield instance at the
//     moment the trigger resolves. An Aang who has left by then
//     gains nothing, and an Aang flickered out and back after the
//     grant loses it (CR 400.7 — he returns a new object).
//
// No remaining simplifications on this clause.
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
		}, {
			// "Whenever you cast a Lesson spell, Aang gains
			// lifelink until end of turn." Mandatory, untargeted.
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && containsFoldASCII(spell.TypeLine, "lesson")
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Aang, the Last Airbender — lifelink until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return GrantKeywordUntilEOT{
							Target:   item.SourceCardID,
							Keywords: []string{"lifelink"},
							Label:    "Aang, the Last Airbender — lifelink",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
