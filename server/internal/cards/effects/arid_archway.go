package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arid Archway — Land — Desert (Outlaws of Thunder Junction, #4468):
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand. If another Desert was returned this way, surveil 1.
//	 {T}: Add {C}{C}."
//
// Guildless Commons with a rider. The bounce is the karoos' choice made
// on resolution (no "target"), the same own-permanents prompt as
// ReturnOneYouControl, with its own keyed continuation so the rider can
// read what was chosen: a Desert (the post-layer land subtype,
// b02IsDesert) other than this Archway that really reached its owner's
// hand earns the surveil. Returning the Archway itself is legal and earns
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3be3d7e6-7860-438a-b8c8-ef154c18c163",
		Name:         "Arid Archway",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{ReturnsLandOnEntry: true},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Arid Archway — return a land you control; if another Desert was returned, surveil 1",
				Do(Instruction(func(ctx *Context) error {
					return ChoosePermanents{
						Question: "Arid Archway — return a land you control to its owner's hand",
						Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
							var out []uuid.UUID
							for _, c := range g.BattlefieldCardsForEffect() {
								if c.Controller == of && c.IsLand() {
									out = append(out, c.InstanceID)
								}
							}
							return out, 1, 1
						},
						Keyed: aridArchwayReturned,
					}.Apply(ctx)
				}))),
		},
	})
}

// aridArchwayReturned returns the chosen land, then surveils 1 if it was
// another Desert and it reached a hand. An on-disk key: never renamed,
// never reused.
var aridArchwayReturned = RegisterPermanentsPick("prompt-run/arid-archway-return-land", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			c, ok := ctx.Game.LookupCardForEffect(id)
			desert := ok && id != ctx.Source() && b02IsDesert(ctx.Game, ctx.Controller(), c)
			if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if !desert {
				continue
			}
			if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
				return Surveil{N: 1}.Apply(ctx)
			}
		}
		return nil
	},
})
