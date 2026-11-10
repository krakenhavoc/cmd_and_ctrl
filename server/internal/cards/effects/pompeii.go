package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pompeii — Plane — Earth (Doctor Who, #4870):
//
//	"When you planeswalk to Pompeii and at the beginning of your upkeep,
//	 put an eruption counter on it.
//	 Whenever you roll a blank on the planar die, scry 2. Then put an
//	 eruption counter on Pompeii.
//	 When chaos ensues, Pompeii deals damage to each creature and each
//	 player equal to the number of eruption counters on Pompeii. Each
//	 player sacrifices a land of their choice. Then planeswalk."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// counters sit on the face-up card (game.PutCountersOnFaceUpPlanarCardForEffect)
// and a planeswalk away makes it a new object with none (CR 901.7a). The blank
// trigger watches the planar die (CR 901.9a, ev.PlanarFace) for a roll by the
// planar controller. The chaos ability (CR 311.7) deals the damage as one
// instance, then every player chooses a land at once (CR 101.4) and, once the
// sacrifices have landed, the planar controller planeswalks (CR 701.31).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5dc1deab-ce04-45c0-a4ff-c526a948e500",
		Name:         "Pompeii",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Pompeii — put an eruption counter on it", pompeiiEruption),
			On(game.EventRollDie, pompeiiYouRolledABlank, "Pompeii — scry 2, then put an eruption counter on it",
				Do(Scry{N: 2}, Instruction(func(ctx *Context) error { return pompeiiEruption(ctx.Game, ctx.Item) }))),
			WheneverChaosEnsues("Pompeii — it deals damage to each creature and each player equal to its eruption counters; each player sacrifices a land; then planeswalk",
				pompeiiChaos),
		},
	})
}

// pompeiiEruptionCounter is the counter kind on the plane. An on-disk key.
const pompeiiEruptionCounter = "eruption"

func pompeiiYouRolledABlank(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.PlanarFace == game.PlanarFaceBlank && ev.Actor == source.Controller
}

func pompeiiEruption(g *game.Game, item *game.StackItem) error {
	g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, pompeiiEruptionCounter, 1)
	return nil
}

func pompeiiChaos(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n, ok := g.PutCountersOnFaceUpPlanarCardForEffect(item.SourceCardID, pompeiiEruptionCounter, 0)
	if ok && n > 0 {
		if err := b23DamageEachCreatureAndEachPlayer(ctx, n); err != nil {
			return err
		}
	}
	players := apnapPlayers(g)
	return ctx.PlayersSacrificeThen(players, sacrificeSpec("a land of your choice", Land()),
		"Pompeii — sacrifice a land", pompeiiThenPlaneswalk, ctx.RunData([]uuid.UUID{item.Controller}))
}

// pompeiiThenPlaneswalk is "Then planeswalk", once every sacrifice has
// settled. An on-disk key: never renamed, never reused.
var pompeiiThenPlaneswalk = RunThen("prompt-run/pompeii-planeswalk", func(ctx *Context, r game.RunDone) error {
	if ctx.Game.Planechase == nil || len(r.Carry) == 0 {
		return nil
	}
	return ctx.Game.PlaneswalkForEffect(r.Carry[0])
})
