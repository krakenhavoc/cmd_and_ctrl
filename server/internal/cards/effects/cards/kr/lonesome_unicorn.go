package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lonesome Unicorn // Rider in Need — Creature — Unicorn {4}{W}, 3/3 // Sorcery — Adventure {2}{W} (Throne of Eldraine, #5516):
//
//	Lonesome Unicorn — "Vigilance"
//	Rider in Need — "Create a 2/2 white Knight creature token with vigilance."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const lonesomeUnicornOracleID = "c8f6cc1f-b7e4-470d-8d48-ef848d7c1116"

func init() {
	effects.Register(effects.Spec{
		OracleID:        lonesomeUnicornOracleID,
		Name:            "Lonesome Unicorn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
	effects.Register(effects.Spec{
		OracleID:     lonesomeUnicornOracleID + "#1",
		Name:         "Rider in Need",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("2/2 white Knight with vigilance"), N: 1}),
	})
}
