package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// drawn_to_death_1_test.go — S58 PR 8, the Drawn To Death deck's
// hand-size punishers and Swamp counts. #2063.

const (
	IronMaidenOracle          = "e0351638-9224-4779-8f7c-f2b11d134a5f"
	ViselingOracle            = "05b9dccb-2364-4bca-a12e-8cb54a3f3a0b"
	TheRackOracle             = "3d873e1d-4fac-42c4-bb31-77e76099e1ef"
	DarkSuspicionsOracle      = "cf5cf174-d976-4cd8-9721-c30fe784fb70"
	MasterOfTheFeastOracle    = "01dc86f3-5ebb-4b10-bf68-8fc9f232c724"
	IndulgentTormentorOracle  = "8b202c63-c961-4590-958f-d17e76610ab5"
	EntropicSpecterOracle     = "a9c44443-85c7-4d0a-82ab-36137151e4e5"
	GlassesOfUrzaOracle       = "af7fabf4-8d55-4b06-9c21-472f4a5775b4"
	DefileOracle              = "49dddec0-d958-4810-8c6e-225fc8118c8f"
	ConsumingCorruptionOracle = "3bc0bd36-70be-4180-9fb3-b10ce054107b"
	MagusOfTheCoffersOracle   = "23dd895c-92bd-4af0-8b1a-7d76ca49178f"
	CryptGhastOracle          = "a3c8d817-7949-4dae-b9f5-f9d952479270"
	NirkanaRevenantOracle     = "6dff1def-5b94-40c8-a942-dea8743ab47c"
	BogWitchOracle            = "85573cba-07ae-4421-a167-a8569f85c0f7"
)

// d2dSetHand makes a seat's hand exactly n filler cards.
func D2dSetHand(p *game.Player, n int) {
	p.Hand.Cards = nil
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "filler", Owner: p.ID, Controller: p.ID})
	}
}
