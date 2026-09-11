package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// karoo_lands.go — two of the ten Ravnica "bounce lands" (karoos),
// the two the roadmap's batch 02 (#295) ranks in the top 360:
//
//	"This land enters tapped."
//	"When this land enters, return a land you control to its
//	 owner's hand."
//	"{T}: Add {G}{U}."
//
// Azorius Chancery shipped first and this file is the same card
// three ways over, with ONE deliberate difference: the enters-tapped
// clause here is a real CR 614 self-replacement rather than an OnETB
// tap.
//
// That difference matters and is worth stating, because the
// Chancery's file argues the opposite. Its comment says a catalog
// replacement "cannot fire on its own source's entry at all, because
// gatherActiveReplacementsLocked collects catalog replacements by
// walking g.Battlefield.Cards, and the entering card is not on the
// battlefield yet". That was true when it was written. It stopped
// being true with the Temple cycle, which added the ENTERING card's
// own replacements as a third gathering block specifically so
// SelfEntersTapped would work — see enters_tapped.go and
// conditional_enters_tapped.go, and the three whole land cycles
// (checklands, fastlands, slowlands) that depend on it. The Chancery
// note is stale, not wrong-at-the-time.
//
// The observable difference: an OnETB tap means the land ENTERS
// UNTAPPED and is tapped a beat later, emitting EventTapCard.
// Anything watching for a tap, or for an untapped permanent
// entering, sees an event that never happened in paper.
//
// DECLARED SIMPLIFICATION, inherited from the Chancery: the printed
// bounce is a CHOICE, not a target — the oracle text has no
// "target" in it. It is modelled as a target clause so the
// controller gets the existing picker, which makes the choice
// visible to opponents at announce and makes the trigger
// respondable-to in a way the printed card is not. The trigger can
// never fizzle for want of a legal choice, because the karoo itself
// is always a land its controller controls, and bouncing itself is
// a normal (if sad) line.
func init() {
	for _, land := range []struct{ oracleID, name, produced string }{
		{"046f5783-cc7b-416a-8cf6-2bcef9c2cc1a", "Simic Growth Chamber", "{G}{U}"},
		{"1b301478-b14f-4ef8-94e6-9647d582eabe", "Golgari Rot Farm", "{B}{G}"},
	} {
		name := land.name
		Register(Spec{
			OracleID:     land.oracleID,
			Name:         name,
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
					return game.NewTriggeredItem(source, name+" — return a land you control",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 {
								return nil
							}
							return BounceToHand{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
						})
				},
			}},
		})
	}
}
