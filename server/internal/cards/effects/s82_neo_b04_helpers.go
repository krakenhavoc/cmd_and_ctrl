package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// neoB04ArtifactYouControlEntered is "an artifact you control enters".
func neoB04ArtifactYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsArtifact()
}
