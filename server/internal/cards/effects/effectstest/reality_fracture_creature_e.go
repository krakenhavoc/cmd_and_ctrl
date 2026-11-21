package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_e_test.go — slice fra-creature-e (tracker
// #2795): Reality Fracture creatures, one or more tests each.

const (
	OracleSamutHazoret    = "add02a36-8811-438b-a8e1-f5c8a8146b4e"
	OracleSoulbreaker     = "47e33aa1-2f54-471c-a484-de13bcd2dceb"
	OracleCryomancer      = "c3936db3-d3b6-4a92-9773-6f3c52dc419d"
	OraclePegasus         = "7d6641e7-4d14-4a06-b588-0c752af529bd"
	OracleSimulacrum      = "9e45e2da-7064-4ec0-8d09-55fc7d1aeaa8"
	OracleBattlecarver    = "4fc7c7a0-bfff-4833-a0d0-a38dc29e9dfb"
	OracleSolariumSentry  = "afddba56-f9a0-4917-8167-c9fd17ca9954"
	OracleSphinxFalse     = "9f004382-c57b-4fde-af97-0c85516c3cf4"
	OracleStingcaster     = "056b651e-e0e2-4333-9235-d1ffe8fcca29"
	OracleSureshot        = "5bbc9d7d-91bd-4a58-8d19-f7be650720b7"
	OracleTamiyoUpriser   = "8e0f2e26-1e14-422a-aafe-e1da38ab10a3"
	OracleTethermage      = "571cb06d-55b5-46e4-9678-1f74aca73068"
	OracleTetherTech      = "a9d0657f-2c5e-4b6f-b8ae-d657cc0f7aef"
	OracleTetsukoPursuer  = "708d4567-a5a3-437f-bae4-ae10f1537aa9"
	OracleThaliaSurvivor  = "55ef129e-698e-424f-be3c-3fbba6c2cc3e"
	OracleTheoreticalNecr = "6c5ed756-b999-49f7-9a52-fb8196202f0f"
	OracleTinybonesPocket = "97dcf9fb-3f2f-4aa6-959a-e3c4889c1673"
	OracleTitanbones      = "510bccb1-61cd-49f1-a54a-80352653d1e3"
	OracleTomikSparkmage  = "16604fee-cd8d-41e4-8269-3c7ff756613d"
)

func RfCrEFoe(g *game.Game) *game.Player { return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)] }
