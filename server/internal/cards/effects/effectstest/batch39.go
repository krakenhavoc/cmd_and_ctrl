package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch39_test.go — card-level coverage for the card-coverage roadmap's
// batch 39 (#402, `edhrec_rank` 4053–4153). One test per observable
// behaviour, driven through a real cast, activation, attack or combat
// damage step rather than by calling primitives.

const (
	B39GlacialFloodplainOracle  = "5d3563dd-a2c1-463c-a0ed-5ac22388bdbe"
	B39SnowfieldSinkholeOracle  = "749c2c8e-9588-4e83-b07f-3c37eb63338b"
	B39DragonSniperOracle       = "913dbce2-6805-453f-a9c3-64f1cc658627"
	B39ZoZuOracle               = "0e412b68-9179-4094-960a-95692428855b"
	B39MaskOfAvacynOracle       = "ab66f8a8-eb3d-4c2d-95e9-26a53c66b237"
	B39MaulOfTheSkyclavesOracle = "088a1f56-198c-41a1-b244-ec1c7a771041"
	B39RoguesGlovesOracle       = "308d7868-49f9-47a3-a7b1-4b0332d610f1"
	B39FeatherOfFlightOracle    = "c61b05b8-00f1-4371-9856-c65852b4ca02"
	B39KorCartographerOracle    = "a5f537f6-3254-45a3-a4e7-d4d4fd7a972d"
	B39BloodlineNecroOracle     = "f2a54e85-dd8b-462c-a811-c9553dc349a5"
	B39SpawningKrakenOracle     = "f0357833-80b5-40ba-874a-bd22ea6e4e46"
	B39NecropolisRegentOracle   = "33c3a6be-8cae-4de8-9d9b-43216d727e93"
	B39MerfolkSovereignOracle   = "16aba9af-7a12-4c3d-87eb-06278921cb7c"
	B39LuxCannonOracle          = "ebdf26a4-77ee-4635-8d52-926bdca623f7"
	B39NestedShamblerOracle     = "bfa882ee-de18-4ef8-8957-3e04ed6e3c1e"
	B39ThrillingDiscoveryOracle = "c5f5a234-c751-4976-a3e1-fbd52b3255c9"
	B39JanJansenOracle          = "cb491d8a-2e9f-46fe-9590-44c3b4a25f1b"
	B39BezaOracle               = "020de6d7-f5a2-4036-ad25-451e5977b4d4"
	B39ShalaiAndHallarOracle    = "e7604cd9-d00d-4957-82c9-46a7cdb88209"
	B39AutogeneratorOracle      = "6832f9e2-294a-44a2-8af9-eb16ccfdfc36"
	B39AnaraOracle              = "a59ff932-f758-476e-ba31-0623bd748231"
	B39AirbendingLessonOracle   = "e101b744-9175-4aa5-bf80-d6944359e538"
	B39InquisitorsFlailOracle   = "a89ae357-b5aa-4256-beb3-a2e5e7f43200"
	B39DawnOfANewAgeOracle      = "37df5ade-6a21-4b0c-89f8-e2e917589a8d"
	B39MarketbackWalkerOracle   = "0405e0a9-6d02-4691-bdb8-59c72b824dab"
	B39AjaniOracle              = "f5d9be71-91d0-4166-ba58-cbbf5d490c40"
	B39ObNixilisOracle          = "88cd12e5-ca89-4176-96fd-320fe2ccd087"
	B39SorinMarkovOracle        = "c151e8b7-4b28-4f03-8a81-9bf623f893d7"
	B39CatharticPyreOracle      = "f7e55117-69fb-412e-b16e-d92e0f95f629"
	B39BanefireOracle           = "5eff8a06-e0d6-435a-a7c0-db9f9d98636a"
	B39ErebosOracle             = "a5690dc9-f12e-4ecc-b528-f2f83c79c3c8"
)

// --- helpers -------------------------------------------------------

// b39Creature seeds a creature under owner, free of summoning
// sickness, with a real timestamp so the layer cache sees it.
func B39Creature(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b39CurrentPower is the combat-relevant power: the post-layer value
// PLUS the +1/+1 counters, which Effective() deliberately leaves out
// (see game.Card.CurrentPower). Every counters card in this batch is
// asserted through it.
func B39CurrentPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	out, seen := 0, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			seen, out = true, c.CurrentPower()
			return
		}
	})
	if !seen {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// b39NextTurnOf advances a whole turn cycle back to `seat`'s precombat
// main phase — a planeswalker's loyalty abilities are once per turn,
// so the second activation needs a new turn and not merely a new step.
func B39NextTurnOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for g.Turn.ActiveSeat == seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	AdvanceToMainOf(t, g, seat)
}

// b39IsCreature reports whether the post-layer view of a permanent is
// a creature — Erebos's devotion clause is a layer 4 type change.
func B39IsCreature(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	out, seen := false, false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != id {
				continue
			}
			seen, out = true, c.IsCreature()
			return
		}
	})
	if !seen {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// b39HandIDs returns the first n card IDs in a player's hand.
func B39HandIDs(p *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if len(out) == n {
			break
		}
		out = append(out, c.InstanceID)
	}
	return out
}
