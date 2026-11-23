package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// commander_ninjutsu_test.go — #1278: commander ninjutsu (CR 702.49c)
// on Yuriko, the Tiger's Shadow. Plain ninjutsu is ninjutsu_test.go;
// the enumerator half is legal/ninjutsu_test.go and the wire half
// protocol/ninjutsu_view_test.go.

const YurikoOracle = "a7043fbd-1dfd-42cf-be4b-cc343d0949e5"

// yurikoCard is Yuriko with the body and the designation the tests
// need, for whichever zone the caller pushes her into.
func YurikoCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(),
		Name:       "Yuriko, the Tiger's Shadow",
		TypeLine:   "Legendary Creature — Human Ninja",
		OracleID:   YurikoOracle,
		ManaCost:   "{1}{U}{B}",
		Power:      1, Toughness: 3,
		Owner: owner, Controller: owner,
		IsCommander: true,
	}
}

// yurikoInCommandZone seeds Yuriko into her owner's command zone as a
// commander that has already been cast `casts` times, so a tax that
// wrongly applied would have something to charge.
func YurikoInCommandZone(p *game.Player, casts int) uuid.UUID {
	c := YurikoCard(p.ID)
	p.Command.PushTop(c)
	if casts > 0 {
		p.CommanderCasts[c.InstanceID] = casts
	}
	return c.InstanceID
}
