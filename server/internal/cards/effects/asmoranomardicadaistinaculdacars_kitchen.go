package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Asmoranomardicadaistinaculdacar's Kitchen — Plane — Hell (punk, #4870):
//
//	"Whenever you planeswalk here and whenever a creature enters under
//	 your control, create a Food token.
//	 Whenever you gain life, target opponent loses that much life.
//	 Whenever Chaos ensues, you may sacrifice any number of artifacts or
//	 creatures. Draw that many cards and lose that much life."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// first is one ability with two trigger conditions. The second is the shared
// "gain life, target opponent loses that much" row. The chaos ability (CR
// 311.7) asks the planar controller at resolution for any number of their
// artifacts and creatures (a floor of zero is the "may"), sacrifices them
// together, then draws and loses life equal to the number chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d2bad25-dbe3-4b51-a78b-4a95dd5669fe",
		Name:         "Asmoranomardicadaistinaculdacar's Kitchen",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventPlaneswalkTo, game.EventETB}, planesB11YouPlaneswalkedHereOrACreatureEnteredUnderYourControl,
				"Asmoranomardicadaistinaculdacar's Kitchen — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			WheneverYouGainLifeTargetOpponentLosesThatMuch("Asmoranomardicadaistinaculdacar's Kitchen — target opponent loses that much life"),
			WheneverChaosEnsues("Asmoranomardicadaistinaculdacar's Kitchen — you may sacrifice any number of artifacts or creatures; draw that many cards and lose that much life",
				func(g *game.Game, item *game.StackItem) error {
					return ChoosePermanents{
						Question: "Asmoranomardicadaistinaculdacar's Kitchen — sacrifice any number of artifacts or creatures",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							out := permanentsControlledByMatching(g, of, Or(Artifact(), Creature()))
							return out, 0, len(out)
						},
						Keyed: asmoranomardicadaistinaculdacarsKitchenSacrificed,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// asmoranomardicadaistinaculdacarsKitchenSacrificed draws and loses life for
// the permanents chosen, once they have been sacrificed. An on-disk key:
// never renamed, never reused.
var asmoranomardicadaistinaculdacarsKitchenSacrificed = RegisterPermanentsPick("prompt-run/asmoranomardicadaistinaculdacars-kitchen-sacrifice", PermanentsDef{
	Sacrifice: true,
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n == 0 {
			return nil
		}
		if err := (DrawCards{N: n}).Apply(ctx); err != nil {
			return err
		}
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Controller(), -n)
	},
})
