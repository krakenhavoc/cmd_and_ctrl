package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// shocklands.go — the Ravnica "shockland" cycle:
//
//	"({T}: Add {X} or {Y}.)"
//	"As this land enters, you may pay 2 life. If you don't, it
//	 enters tapped."
//
// Three of the ten, the three this deck plays. The mana ability is
// declared rather than left to the synthetic basic-land shape: a
// shockland is a NONBASIC land that happens to carry two basic land
// types, and the reminder-text ability comes from those types.
// Declaring it is also what lets the pipe syntax offer one picker
// instead of two menu entries.
//
// # Declared sandbox simplification: enters tapped, then untaps
//
// "As this land enters, you may pay 2 life" is a REPLACEMENT with a
// choice in it, and the replacement pipeline is synchronous — it has
// no way to stop and ask a player anything. So this ships as the
// honest half-measure rather than a fixed guess:
//
//	enters tapped (real CR 614 self-replacement)
//	+ an optional ETB trigger "pay 2 life to untap it"
//
// The CHOICE survives, which is the part that matters — a player at
// 3 life can decline, and one who wants the untapped land pays for
// it. What is observably wrong:
//
//   - The land is tapped for the window between entering and the
//     trigger resolving, and it emits an untap event when it
//     resolves. Nothing watches for that today.
//   - The trigger uses the stack, so opponents get priority in
//     between. In paper nobody does; "as this enters" is not a
//     trigger at all.
//
// The alternative — picking one branch at build time and always
// taking it — would be strictly wrong in one of the two cases and
// silent about it, which is worse than being slow and correct about
// the choice.
//
// The trigger is offered only when the controller has at least 2
// life, so it cannot be used to pay a cost the player cannot afford.
func init() {
	for _, t := range []struct {
		oracleID string
		name     string
		a, b     string
	}{
		{"73864fcc-1bde-4bc0-831e-2b93e546e417", "Godless Shrine", "W", "B"},
		{"f1750962-a87c-49f6-b731-02ae971ac6ea", "Hallowed Fountain", "W", "U"},
		{"fc9ec820-4245-4a96-b009-5308a818ca58", "Watery Grave", "U", "B"},
	} {
		// Capture per iteration: the closures below outlive the loop
		// body, and a shared loop variable would give every shockland
		// the last entry's name.
		name := t.name
		Register(Spec{
			OracleID:      t.oracleID,
			Name:          name,
			Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{dualManaAbility(t.a, t.b)},
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID != source.InstanceID {
						return false
					}
					p := g.PlayerByIDForEffect(source.Controller)
					return p != nil && p.Life >= 2
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, name+" — pay 2 life to untap it",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							if err := g.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -2); err != nil {
								return err
							}
							return UntapTarget{Target: item.SourceCardID}.Apply(ctx)
						})
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{
					Question: name + " — pay 2 life so it is untapped?",
				},
			}},
		})
	}
}
