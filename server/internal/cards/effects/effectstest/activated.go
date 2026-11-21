package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activated_test.go — S21 sub-PR 2: catalog activated abilities.

const (
	GoblinBombardmentOracle = "edad60c6-80de-4033-af1b-a703ac332983"
	CarrionFeederOracle     = "a1cc5e37-b09a-4b7f-afd5-77c1c35aa425"
	KrenkoOracle            = "68418069-f615-40ef-ae0d-764192acae00"
)

// pushCatalogPermanent seeds a catalog card onto the battlefield,
// already free of summoning sickness unless `sick` says otherwise.
//
// Pushes the card, then fires EventZoneMove the same way
// pushBattlefieldCardWithTimestamp (anthem_test.go) does, so the layer
// listener stamps EnteredBattlefieldAt and bumps layerVersion. Without
// that stamp/event a layer-engine test using this helper only passed
// when something else in the test happened to trigger a recompute.
// The listener unconditionally sets SummonedThisTurn = true on entry
// (S18), so `sick` is re-applied after the emit to keep this helper's
// existing contract — a non-sick creature the caller can act with
// immediately.
func PushCatalogPermanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, sick bool) uuid.UUID {
	return PushCatalogPermanentWithID(g, uuid.New(), owner, name, typeLine, oracle, sick)
}

// pushCatalogPermanentWithID is pushCatalogPermanent with a caller-chosen
// InstanceID. A keyed random stream (ADR 0054) is seeded from the SOURCE
// card's id as well as the game key, so a test that pins the key but lets
// the source id be uuid.New() still draws a different sequence every run.
func PushCatalogPermanentWithID(g *game.Game, id, owner uuid.UUID, name, typeLine, oracle string, sick bool) uuid.UUID {
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
		SummonedThisTurn: sick,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:    game.EventZoneMove,
			CardID:  id,
			OldZone: game.ZoneHand,
			NewZone: game.ZoneBattlefield,
		})
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = sick
				break
			}
		}
	})
	return id
}
