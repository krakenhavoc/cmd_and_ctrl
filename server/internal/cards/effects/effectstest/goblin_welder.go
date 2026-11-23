package effectstest

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const GoblinWelderOracle = "9fc28e1f-f4ea-4bf8-9442-a36dba3b3b29"

func WelderGraveyardArtifact(p *game.Player, name, oracle string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.Controller, c.TypeLine, c.OracleID = p.ID, "Artifact", oracle
	p.Graveyard.PushTop(c)
	return c.InstanceID
}
