package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bounce_lands.go — the Ravnica "bounce lands" / karoos:
//
//	"This land enters tapped.
//	 When this land enters, return a land you control to its owner's
//	 hand.
//	 {T}: Add {G}{U}."
//
// Azorius Chancery's shape, as a table: the tempo cost (tapped, and a
// land drop set back) buys one land that taps for two coloured mana.
// Roadmap batch 02 (#295) ranks Simic Growth Chamber and Golgari Rot
// Farm; batch 11 (#304) adds Guildless Commons, the colourless one;
// the rest of the cycle belongs here when it arrives.
//
// Two differences from the Chancery file, both deliberate:
//
//   - The tapped entry is a real CR 614 self-replacement. The
//     Chancery's comment says a catalog replacement cannot fire on
//     its own entry; that was true when it was written and has not
//     been since the Temple cycle (gatherActiveReplacementsLocked's
//     third block). The Chancery file itself is left to its owner.
//   - Same declared simplification as the Chancery: the printed
//     bounce is a CHOICE, not a target — modelled as a target clause
//     so the controller gets the existing picker, which locks the
//     choice at announce rather than resolution. It can never fizzle
//     for want of a choice, because the land itself is always a
//     legal one, and bouncing itself is sometimes the right line.
func init() {
	for _, land := range []struct{ oracleID, name, produced string }{
		{"046f5783-cc7b-416a-8cf6-2bcef9c2cc1a", "Simic Growth Chamber", "{G}{U}"},
		{"1b301478-b14f-4ef8-94e6-9647d582eabe", "Golgari Rot Farm", "{B}{G}"},
		// Roadmap batch 11 (#304): the colourless karoo — same three
		// clauses, "{T}: Add {C}{C}".
		{"ee723c7c-ec9f-4ffb-8f36-cd7637eb1fae", "Guildless Commons", "{C}{C}"},
	} {
		name := land.name
		Register(Spec{
			OracleID:     land.oracleID,
			Name:         land.name,
			Completeness: CompletenessCaveats,
			Caveats:      []string{"You pick the land to return when the trigger goes on the stack, not on resolution, and it's a target that can be removed in response."},
			Replacements: []game.ReplacementEffect{SelfEntersTapped()},
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: land.produced,
				Label:    "Add " + land.produced,
			}},
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Targets: TargetPermanent("a land you control", And(Land(), YouControl())),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, name+" — return a land you control to hand",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
								return nil
							}
							return BounceToHand{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
						})
				},
			}},
		})
	}
}
