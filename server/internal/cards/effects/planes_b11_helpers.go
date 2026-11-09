package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// planes_b11_helpers.go — shapes the gold-bordered planes of batches 11 and
// 12 share (ADR 0216 slice 4, #4870). Append-only.

// printedB11BlackLotusToken is the Black Lotus token Black Lotus Lounge makes:
// a colourless artifact token named Black Lotus with "{T}, Sacrifice this
// artifact: Add three mana of any one color." The three mana are one colour
// pick (the Lotus token's shape).
func printedB11BlackLotusToken() tokenTemplate {
	return tokenTemplate{
		Slug: "black-lotus",
		Card: game.Card{
			Name:     "Black Lotus",
			TypeLine: "Token Artifact",
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:       true,
			SacrificeCost: true,
			Produced:      OneColorOfAmount(3),
			Label:         "{T}, Sacrifice: Add three mana of any one color",
		}},
		Text: "{T}, Sacrifice this artifact: Add three mana of any one color.",
	}
}

// B11BlackLotusToken is the Black Lotus artifact token.
func B11BlackLotusToken() game.Card { return tokenFromCatalog(printedB11BlackLotusToken) }

// planesB11YouPlaneswalkedHereOrACreatureEnteredUnderYourControl is
// Asmoranomardicadaistinaculdacar's Kitchen's first trigger condition,
// "whenever you planeswalk here and whenever a creature enters under your
// control": the plane itself coming face up under the planar controller
// (CR 901.11b), or any creature entering under the planar controller's
// control (the source is a plane, so "another" never matters).
func planesB11YouPlaneswalkedHereOrACreatureEnteredUnderYourControl(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventPlaneswalkTo:
		return ev.CardID == source.InstanceID && ev.Actor == source.Controller
	case game.EventETB:
		return CreatureEnteredUnderYourControl(ev, source, lki, g)
	}
	return false
}

// planesB11PlaneswalkedHereOrChaos is "whenever you planeswalk here or chaos
// ensues" (Jalira's Show): this plane coming face up under the planar
// controller (CR 901.11b), or the chaos face of the planar die (CR 311.7).
func planesB11PlaneswalkedHereOrChaos(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	switch ev.Kind {
	case game.EventPlaneswalkTo:
		return ev.CardID == source.InstanceID && ev.Actor == source.Controller
	case game.EventRollDie:
		return ev.PlanarFace == game.PlanarFaceChaos
	}
	return false
}
