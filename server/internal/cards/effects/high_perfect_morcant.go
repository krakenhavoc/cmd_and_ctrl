package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High Perfect Morcant — Legendary Creature — Elf Noble {2}{B}{G}, 4/4:
//
//	"Whenever High Perfect Morcant or another Elf you control enters, each
//	 opponent blights 1. (They each put a -1/-1 counter on a creature they
//	 control.)
//	 Tap three untapped Elves you control: Proliferate. Activate only as a
//	 sorcery. (Choose any number of permanents and/or players, then give
//	 each another counter of each kind already there.)"
//
// "Each opponent blights 1" is Blight with every opponent named (#4683):
// each chooses a creature THEY control, in APNAP order and seeing the
// earlier choices (CR 101.4), and an opponent with one creature has no
// choice to make. An opponent with no creature that can take the counter
// does not blight (CR 701.68b). The choice is not a target, so hexproof
// does not stop it. The counters are an effect's, placed by the
// blighting opponent.
//
// The proliferate taps three untapped Elves you control, Morcant included
// when it is untapped; it is not the {T} symbol, so an Elf that arrived
// this turn may be tapped for it (CR 302.6).
func init() {
	const tapThree = "three untapped Elves you control"
	Register(Spec{
		OracleID:     "dd27c5fa-4355-4e5c-8219-b74cfe209da0",
		Name:         "High Perfect Morcant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, morcantOrAnotherElfEntered,
				"High Perfect Morcant — each opponent blights 1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return Blight{N: 1, Players: ctx.Opponents()}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "Tap three untapped Elves you control: Proliferate.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: TargetPermanent(tapThree, Creature(), OfCreatureType("Elf")),
				Label:  tapThree,
			}},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Proliferate{}.Apply(NewContext(g, item))
			},
		}},
	})
}

// morcantOrAnotherElfEntered is "whenever High Perfect Morcant or another
// Elf you control enters". "Elf" is a word of the card's text (CR 612.2).
func morcantOrAnotherElfEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok {
		return false
	}
	return c.InstanceID == source.InstanceID || c.HasSubtype(source.TextSubtype("Elf"))
}
