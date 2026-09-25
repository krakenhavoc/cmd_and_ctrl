package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vaultborn Tyrant — Creature — Dinosaur {5}{G}{G}, 6/6 (issue
// #1117):
//
//	"Trample
//	 Whenever this creature or another creature you control with
//	 power 4 or greater enters, you gain 3 life and draw a card.
//	 When this creature dies, if it's not a token, create a token
//	 that's a copy of it, except it's an artifact in addition to its
//	 other types."
//
// The ETB half reads the entering permanent's CURRENT power
// (enteredUnderYourControl + CurrentPower, the Elemental Bond /
// Garruk's Packleader shape) and INCLUDES Vaultborn Tyrant itself —
// unlike "another creature", there is no exclusion here, so the
// Tyrant's own ETB triggers its own ability.
//
// The dies half is the CR 603.10 last-known-information case the
// catalog convention exists for: by the time the trigger's AppliesTo
// runs, the card is already in the graveyard, so "if it's not a
// token" is answered off the harvester's LKI Characteristic (the
// third AppliesTo/Build argument) rather than off the departed card,
// which the graveyard object cannot answer (#762's token-entry
// pipeline is what makes a copy's own type line carry "Token" as a
// supertype, so the LKI read and the token check use exactly the same
// signal). The created token carries the Tyrant's oracle ID
// (CreateTokenCopy, token_copy.go), so it inherits this very dies
// trigger — but it is a token, so the LKI check refuses it when IT
// dies, and the chain stops at one copy, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8d0accc-b320-4c07-8a71-a09db860351e",
		Name:            "Vaultborn Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.CurrentPower() >= 4
			}, "Vaultborn Tyrant — gain 3 life and draw a card", Do(GainLife{Amount: 3}, DrawCards{N: 1})),
			{
				Watches: []game.EventKind{game.EventLTB},
				Key:     "Vaultborn Tyrant — create a token copy of it that's also an artifact",
				AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return cardDied(ev, source) && !hasFold(lki.Supertypes, "Token")
				},
				Effect: vaultbornTyrantArtifactCopy,
			},
		},
	})
}

// vaultbornTyrantArtifactCopy is the dies trigger's body: a token
// copy of the departed Vaultborn Tyrant, with "except it's an
// artifact in addition to its other types" applied to the template
// before the token exists (the Saheeli Rai −2 shape — reusing its
// type-line helper rather than a second copy of it).
func vaultbornTyrantArtifactCopy(g *game.Game, item *game.StackItem) error {
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       item.SourceCardID,
		N:          1,
		Except: func(t *game.Card) {
			t.TypeLine = saheeliArtifactInAddition(t.TypeLine)
		},
	}.Apply(NewContext(g, item))
}
