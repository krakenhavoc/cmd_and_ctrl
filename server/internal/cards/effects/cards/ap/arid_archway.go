package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "3be3d7e6-7860-438a-b8c8-ef154c18c163",
		Name:         "Arid Archway",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ReturnsLandOnEntry: true},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arid Archway — return a land you control; if another Desert was returned, surveil 1",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.ChoosePermanents{
						Question:   "Arid Archway — return a land you control to its owner's hand",
						Candidates: effects.S82OldB089LandCandidates,
						Keyed:      aridArchwayReturned,
					}.Apply(ctx)
				}))),
		},
	})
}

// aridArchwayReturned returns the chosen land. Whether it is "another
// Desert" is read while it is still on the battlefield; the surveil
// then hangs off the bounce's own continuation, so it is paid only once
// the land has really reached a hand (CR 400.7, #993): a bounce the
// CR 614 window cancelled or redirected earns nothing. An on-disk key:
// never renamed, never reused.
var aridArchwayReturned = effects.RegisterPermanentsPick("prompt-run/arid-archway-return-land", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			c, ok := ctx.Game.LookupCardForEffect(id)
			bounce := effects.BounceToHand{Target: id}
			if ok && id != ctx.Source() && effects.B02IsDesert(ctx.Game, ctx.Controller(), c) {
				bounce.ThenTail = ctx.Tail(aridArchwaySurveil)
			}
			if err := bounce.Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

// aridArchwaySurveil is "If another Desert was returned this way,
// surveil 1": attached only when the chosen land was another Desert, and
// paid only when it reached a hand. An on-disk key: never renamed, never
// reused.
var aridArchwaySurveil = effects.CardTail("arid-archway/surveil-if-desert-returned", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.Surveil{N: 1}.Apply(ctx)
})
