package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Narfi, Betrayer King — Legendary Snow Creature — Zombie Wizard {3}{U}{B},
// 4/3 (Kaldheim):
//
//	"Other snow and Zombie creatures you control get +1/+1.
//	 {S}{S}{S}: Return this card from your graveyard to the battlefield
//	 tapped. ({S} can be paid with one mana from a snow source.)"
//
// The anthem is a layer 7c static over the other creatures its controller
// controls that are snow or Zombies (one +1/+1 even for a snow Zombie); it
// reads the layered view on every pass. The return is an activated ability
// of the card in the graveyard (CR 602.2) paid with three snow mana ({S}
// pays only with mana from a snow source, CR 107.4h), re-checked as it
// resolves that the card is still there (CR 608.2b); it enters tapped.
//
// No simplification.
func init() {
	anthem := effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.InstanceID != source.InstanceID &&
			(target.HasSupertype("snow") || target.HasSubtype(source.TextSubtype("Zombie")))
	}, 1, 1)
	anthem.Label = "Other snow and Zombie creatures you control get +1/+1."
	effects.Register(effects.Spec{
		OracleID:     "18beb8eb-6f67-482f-9a7d-282f1e7ecbc1",
		Name:         "Narfi, Betrayer King",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{anthem},
		Activated: []effects.ActivatedAbility{{
			Label:   "{S}{S}{S}: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{S}{S}{S}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				id := item.SourceCardID
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return effects.ReturnFromGraveyard{
					Target:     id,
					Dest:       game.ZoneBattlefield,
					Controller: item.Controller,
					Tapped:     true,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
