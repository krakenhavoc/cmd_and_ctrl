package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b21_cards_test.go — Standard set leftovers, batches 21 and 22 (#4452).

const (
	StdB21TeachOracle          = "d85ba214-fa05-44ca-a7f9-ab2c5dc07962"
	StdB21GateColossusOracle   = "386c5895-9713-4c28-9ba8-8ca0caa06c13"
	StdB21SpyglassOracle       = "b2187f45-80ae-4ac4-9f83-5eb7a00978e2"
	StdB21HellkiteOracle       = "6e64f1e0-ddd6-4ee7-8797-086e234b02e8"
	StdB21BarrensOracle        = "dfe1a112-97aa-4e81-8431-81552ba2cdcf"
	StdB21MazesEndOracle       = "49479778-c4c0-43ba-a7b7-45f00d067462"
	StdB21RuneSealedWallOracle = "74d31e99-ab48-483a-a62e-5b424b27d017"
	StdB21DiamondMareOracle    = "9a440122-a015-4af2-b270-37f019884458"
	StdB21AbductionOracle      = "8f5495d7-ac44-47cd-aa39-48a35034b390"
	StdB21OrphansOracle        = "f758bad0-8f36-4985-ad1d-66b695f3bbb6"
	StdB21DontMakeASoundOracle = "f28a373b-4a8d-40a5-ace1-3bbd8489ad36"
	StdB21DominationOracle     = "ebdace8f-8e8e-411b-b1d5-5b27cca7d9fb"
	StdB21FallingOracle        = "00ba96e0-42a7-432a-a500-5e607d75a358"
	StdB21ImpostorsOracle      = "628ae0e4-5d54-48a1-a24c-b9c1d862df9d"
	StdB21DrownerOracle        = "43c69d19-c06c-470a-bdba-a0335ff2d655"
	StdB21GetOutOracle         = "b2f3c27e-b646-4449-901a-8b809cdfa84e"
	StdB21LeylineOracle        = "cfaa13d6-9992-4223-8bd7-c44046505555"
	StdB21GrimoireOracle       = "096eac52-00c2-4439-b2f3-ff555d15ec88"
	StdB21TaleOracle           = "98a13f7b-ebe7-4a90-a0b9-95dccfff7fa6"
	StdB21CrackedSkullOracle   = "4415b65d-5473-4b19-94f7-e5cf4ce64d5e"
)

// stdB21Library pushes `n` plain cards on top of the player's library.
func StdB21Library(p *game.Player, cards ...game.Card) {
	for _, c := range cards {
		c.Owner, c.Controller = p.ID, p.ID
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		p.Library.PushTop(c)
	}
}
