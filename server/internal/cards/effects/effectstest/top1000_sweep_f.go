package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top1000_sweep_f_test.go — the sixth slice of the missing top-1,000
// cards (#3654).

const (
	SweepFAetherfluxOracle   = "5f15125e-ae89-471e-b055-226a783bc7c1"
	SweepFKutzilOracle       = "3233af43-7826-4895-b67e-03c6102c2cd5"
	SweepFWishclawOracle     = "81c70ae7-3c18-4c9b-8505-e4db9e0e6518"
	SweepFMindbreakOracle    = "48712d36-5c13-4ded-87a4-68106da4a5cd"
	SweepFCaretakerOracle    = "c97957a2-8310-4cff-8aad-871b7901d124"
	SweepFInnkeeperOracle    = "915b75d1-b5e3-4aca-a54c-beaccef0a2e4"
	SweepFSpringheartOracle  = "8a3ad2ef-8bcb-40c0-85de-f03328c2b644"
	SweepFVeyranOracle       = "d57a6b19-bd06-4b07-8ee1-0b45612ba9e4"
	SweepFIsochronOracle     = "ddd649f4-dbaf-48ec-8ff7-95581257772d"
	SweepFHighFaeOracle      = "28c1a4a0-f58c-4d07-a918-51ff8a74f050"
	SweepFExpressiveOracle   = "c7aecca5-2f67-4245-ab2d-e723d8b23a67"
	SweepFTaintedPactOracle  = "1a85ba2b-ae10-4917-954a-7709b75a9740"
	SweepFBladeOracle        = "9c329f3d-4d0a-4012-8aa6-92afdd120a52"
	SweepFEvendoOracle       = "83161d59-2520-4741-9328-e2a4a8b5d5bc"
	SweepFConsultationOracle = "9a1412db-45ad-46ea-8f12-a85d203113d8"
	SweepFApproachOracle     = "e4125377-34c0-4b54-bdf8-4e88f5d24565"
	SweepFCyberdriveOracle   = "126a5d0b-e7d6-458e-87af-47d845678014"
	SweepFArchmageOracle     = "8305d576-21d8-4ce7-8eda-a7cd9793aca5"
)

// sweepFTop puts a card on TOP of p's library.
func SweepFTop(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func SweepFAddMana(g *game.Game, who uuid.UUID, mana string) {
	g.WithWriteLock(func() { _ = g.AddManaForEffect(who, uuid.Nil, mana) })
}

// --- Springheart Nantuko -----------------------------------------------

// sweepFBestowed puts the Nantuko onto the battlefield as an Aura
// attached to host.
func SweepFBestowed(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Springheart Nantuko", TypeLine: "Enchantment — Aura",
		OracleID: SweepFSpringheartOracle, Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

// --- Etali, Primal Conqueror -------------------------------------------

func EtaliRow() cards.Card {
	return TransformRow(cardse.EtaliOracleID,
		"Etali, Primal Conqueror", "Legendary Creature — Elder Dinosaur", "{5}{R}{R}",
		"Etali, Primal Sickness", "Legendary Creature — Phyrexian Elder Dinosaur", "11", "11", []string{"R"})
}
