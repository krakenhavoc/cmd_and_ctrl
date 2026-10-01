package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unlicensed Hearse — Artifact — Vehicle {2}, */*:
//
//	"{T}: Exile up to two target cards from a single graveyard.
//	 Unlicensed Hearse's power and toughness are each equal to the
//	 number of cards exiled with it.
//	 Crew 2"
//
// #1807, ADR 0106 §5. The power and toughness are a characteristic-
// defining ability (CR 604.3, applied in layer 7a), linked to the tap
// ability (CR 607.2a): "cards exiled with it" are the ones that
// ability exiled and that are still in exile, which b27ExiledWith
// reads off the event log by the ability's label. A Hearse that leaves
// and comes back is a new object and counts from zero (CR 400.7).
//
// The count changes when a card arrives in exile from a graveyard,
// which already refreshes the layers, and when one leaves exile for a
// hand or a library, which DependsOnExile covers.
//
// Crewed with nothing exiled, it is a 0/0 and dies (the 2022-04-29
// ruling). Crew's own effect sets no power or toughness, so the CDA is
// what the crewed creature has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c640654c-487e-4a2c-aced-126ed835b78f",
		Name:         "Unlicensed Hearse",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			ExileFromASingleGraveyardAbility("{T}", TapCost(), 2),
			{
				Label:  "Crew 2",
				Cost:   CrewCost(2),
				Effect: CrewEffect("Unlicensed Hearse"),
			},
		},
		Static: []game.StaticAbility{{
			Layer:          game.Layer7PT,
			SubLayer:       game.SubLayer7A_CDA,
			DependsOnExile: true,
			AppliesTo:      selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := len(b27ExiledWith(g, source.InstanceID, unlicensedHearseExileLabel))
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}

// unlicensedHearseExileLabel is the tap ability's label, the key its
// exiled-with record is read by. It is spelled exactly as
// ExileFromASingleGraveyardAbility builds it.
const unlicensedHearseExileLabel = "{T}: Exile up to two target cards from a single graveyard."
