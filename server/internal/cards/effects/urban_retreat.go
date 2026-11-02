package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urban Retreat — Land (Marvel's Spider-Man, #4460):
//
//	"This land enters tapped.
//	 {T}: Add {G}, {W}, or {U}.
//	 {2}, Return a tapped creature you control to its owner's hand: Put this
//	 card from your hand onto the battlefield. Activate only as a sorcery."
//
// The third ability works from the hand (CR 113.6b is declared by Zones). Its
// cost returns a tapped creature you control, chosen as the ability is
// announced. The land enters through the ordinary put-from-hand path, so its
// own "enters tapped" replacement applies. If the card has left the hand (or is
// a new object there) by the time the ability resolves, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18290c4b-cff6-4ee5-a487-fe0d5706d3de",
		Name:         "Urban Retreat",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|W|U}",
			Label:    "Add {G}, {W}, or {U}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, Return a tapped creature you control to its owner's hand: Put this card from your hand onto the battlefield. Activate only as a sorcery.",
			Cost: Plus(ManaCost("{2}"),
				ReturnAPermanentToHand("a tapped creature you control", hobB3TappedCreature(), YouControl())),
			Zones:        []game.ZoneKind{game.ZoneHand},
			SorcerySpeed: true,
			Effect:       urbanRetreatPutOntoBattlefield,
		}},
	})
}

// urbanRetreatPutOntoBattlefield puts the land from its owner's hand onto the
// battlefield under its activator's control, if it is still that hand card.
func urbanRetreatPutOntoBattlefield(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneHand {
		return nil
	}
	ref := item.SourceObject
	if ref.ID == item.SourceCardID && ref.ID != uuid.Nil {
		for i := range z.Cards {
			if z.Cards[i].InstanceID == ref.ID && z.Cards[i].ObjectEpoch != ref.Epoch {
				return nil
			}
		}
	}
	_, err := g.PutFromHandOntoBattlefieldForEffect(item.SourceCardID, game.HandEntryOptions{Controller: item.Controller})
	return err
}
