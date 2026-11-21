package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0109_pr8_cards_test.go — the cards ADR 0109 §7 (#1902) and owner
// decision 3 unblock: costs that put a card from a hand on top of a
// library, that exile the top of a library, and that discard at
// random. The engine half is library_costs_test.go.

const (
	P8StormscaleAnarch = "819a7235-b61a-49d7-a0ad-3d8bbf83091e"
	P8MageIlVec        = "adea7db6-634c-4c5d-be40-264b4acffc53"
	P8FreneticOgre     = "6eb4b649-c578-4b29-8da1-7c757412bec0"
	P8CanyonDrake      = "b5b46c99-1cc1-465a-a831-3fd6665ef560"
	P8CoralHelm        = "aa2970c8-f2ea-4e06-8b8f-ec89af0012a0"
	P8Pyromancy        = "71a176ce-cbe7-4cec-a2ca-58abf59ad964"
	P8MeteorStorm      = "03f96c23-de0b-4b85-81f8-febc850aa621"
	P8PardicSwordsmith = "946d770f-0a59-415b-b334-8f7353b96046"
	P8Stormbind        = "78f50668-36fc-4911-84f7-93667436b0c7"
	P8Amok             = "2ddbbe65-f928-4b8e-8c4c-a9ce82e2594d"
	P8Pyromania        = "2c89cf19-8dc6-422e-bbf8-4dd5a399c630"
	P8HellBentRaider   = "ee8ab29b-0749-463a-abbd-eb0b9013aec8"
	P8DwarvenStrike    = "66dd0e68-f5ec-45e5-991f-d588aa726387"
	P8DraconianCylix   = "0880b36a-6141-4955-b532-cf88daca0869"
	P8OgreShaman       = "04c7bb20-6e40-4e79-a0ec-ced920d3491e"
	P8PardicLancer     = "1bb2b090-19a4-40d9-823a-671f6eaca9f9"
	P8WhirlingCatapult = "0987b5e9-0012-4189-8745-45f10e9557f3"
	P8ArcSlogger       = "a28fc506-a98c-4b4d-a0f4-7d971489ccbb"
	P8SeasonedTactic   = "d7c2f27c-8c82-4283-a17d-f979ee636542"
	P8RoyalHerbalist   = "b3baf498-3be7-4a22-8fed-806d8d8ac748"
	P8StormElemental   = "f1fee486-660e-4356-896e-671e8b675ad8"
	P8PhyrexianDevour  = "28b1ed69-784c-4642-86c7-104792d5afc2"
	P8Leashling        = "8460abd9-ab5d-4a7e-8132-185b4a310190"
	P8Penance          = "7818d6e0-74d8-46ae-94f6-d2a349c9507b"
)

func P8LibraryTop(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Library.PushTop(c) })
	return c.InstanceID
}

func P8Player(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: id}}
}
