package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nibelheim Aflame — Sorcery {2}{R}{R} (Final Fantasy, #4462):
//
//	"Choose target creature you control. It deals damage equal to its
//	 power to each other creature. If this spell was cast from a
//	 graveyard, discard your hand and draw four cards.
//	 Flashback {5}{R}{R} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Nova Flame's second instruction (targetDealsItsPowerToEachOtherCreature,
// shared): the targeted creature deals the
// damage, so its deathtouch or lifelink applies, at its power as the
// spell resolves, to every other creature in one damage event. A target
// that has left (or is no longer a creature you control) is illegal and
// the whole spell does nothing (CR 608.2b), the wheel included.
//
// "Cast from a graveyard" is the spell's CastFromZone (Tidings of War's
// reading): flashback is one way, any other permission to cast it from
// a graveyard is another. It is read once, as the spell resolves, and
// kept as a step record, because the later steps may run from a prompt.
// The wheel is Tolarian Winds' two steps: the whole hand is discarded,
// then four cards are drawn once any prompt the discard asks (a madness
// offer) is answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "eab5eef8-d180-4b08-a770-6b3a1df03745",
		Name:             "Nibelheim Aflame",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature you control", YouControl()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{R}{R}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				fromGraveyard := 0
				if ctx.Item != nil && ctx.Item.CastFromZone == game.ZoneGraveyard {
					fromGraveyard = 1
				}
				ctx.RecordStep(nibelheimAflameFromGraveyard, nil, fromGraveyard)
				return targetDealsItsPowerToEachOtherCreature(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if ctx.StepRecord(nibelheimAflameFromGraveyard).N() != 1 {
					return nil
				}
				_, err := discardWholeHand(ctx.Game, ctx.Controller())
				return err
			}),
			Instruction(func(ctx *Context) error {
				if ctx.StepRecord(nibelheimAflameFromGraveyard).N() != 1 {
					return nil
				}
				return DrawCards{N: 4}.Apply(ctx)
			}),
		),
	})
}

// nibelheimAflameFromGraveyard is 1 when the spell was cast from a
// graveyard, 0 otherwise.
const nibelheimAflameFromGraveyard = "nibelheim-aflame/cast-from-graveyard"
