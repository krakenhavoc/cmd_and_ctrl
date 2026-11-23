package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reality Fracture slice fra-planeswalker-a (tracker #2795).

const (
	RfaAjaniResolute      = "0d60c095-4c6c-4618-8ef3-508f3d454efa"
	RfaAjaniUnrelenting   = "71620e92-45a4-49cb-94d5-348013ecf886"
	RfaAwakenInferno      = "e729fce5-a1cf-4e95-9fd5-fccf4f704d31"
	RfaBreakUnderPress    = "8dd23c50-cc84-4064-89cd-8d52fbfbd4ab"
	RfaChandra            = "c3dfa1e2-6785-49a0-a194-fb842a8eb63c"
	RfaCompelBrutality    = "3a9059ac-c595-4e71-b704-b04d2d47db70"
	RfaCraftworkCrusher   = "d25284b9-50e4-4dcb-9aaf-6f6478c4a90a"
	RfaDedicatedCommons   = "d4e0749a-e179-4c19-8536-bf8f1a198549"
	RfaEdgar              = "cab2bc93-f38b-4301-ac42-05765615586b"
	RfaEntrustTheSpark    = "b796ecc8-a2e8-4ab9-8736-6825b650434d"
	RfaEssenceBurn        = "f1216f36-00fd-4d03-85c1-f4910a02d4d0"
	RfaExtendedAbsence    = "973fb1e3-83d5-4763-bea8-799b426a8848"
	RfaFaceYourself       = "3c0fe71c-bee1-41d3-aeb5-1723028eedde"
	RfaFateholdAnnex      = "d35876a3-e891-43d3-a50c-69e8469497e0"
	RfaFlourishingGrapple = "6b4f1569-2025-46a8-89ba-79859208b04e"
	RfaFormidableCommons  = "3e51c060-1e86-4a0e-8fee-8b7273161468"
	RfaFulminousForte     = "2a29d028-d1a4-413f-9c01-d6842294420a"
)

func RfaColoredCreature(g *game.Game, owner uuid.UUID, name string, p, t int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Colors: colors,
		Power: p, Toughness: t, Owner: owner, Controller: owner,
	})
}
