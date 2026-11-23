package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a6_cards_test.go — Marvel Super Heroes, half A, sixth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	MshA6MoonstoneOracle   = "aeeb50f2-8e02-4d40-a1dd-28b1c5c85c7e"
	MshA6IncursionOracle   = "d3678fce-6897-4f2c-bf50-243facb6581b"
	MshA6NickFuryOracle    = "62916107-14c0-4516-9fed-a032b1a80e5d"
	MshA6NinjaOracle       = "c38f7447-e646-438d-81bd-fe2a71c8a79c"
	MshA6OriginOracle      = "d7f3178d-7b7c-4346-af50-a7b0206bee22"
	MshA6PatriotOracle     = "1df7a09f-fbc8-4c2a-a903-bab9e680eda7"
	MshA6BrokerOracle      = "2b3d76e7-faaa-4e6f-9e21-a88dd719247e"
	MshA6PunchOracle       = "454c4e8b-e417-4381-a931-6a11356ef64e"
	MshA6QuakeOracle       = "3966c94b-e597-4b3f-a4f5-001d52fff9fe"
	MshA6RaftOracle        = "60754d4b-2add-4928-8d30-8ce9c0f54f91"
	MshA6RedGuardianOracle = "e85567e8-1102-4d44-a15b-d5f1fdef8465"
	MshA6RecruitOracle     = "4e60c38c-cbf5-4ff7-8fff-634f7e3086cc"
)

func MshA6HandCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID, Power: 1, Toughness: 1})
	return id
}
