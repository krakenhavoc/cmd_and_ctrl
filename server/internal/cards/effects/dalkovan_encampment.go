package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dalkovan Encampment — Land (EDHREC rank 3484):
//
//	"This land enters tapped unless you control a Swamp or a
//	 Mountain.
//	 {T}: Add {W}.
//	 {2}{W}, {T}: Whenever you attack this turn, create two 1/1 red
//	 Warrior creature tokens that are tapped and attacking. Sacrifice
//	 them at the beginning of the next end step."
//
// The Mardu land that brings two Warriors to every attack it was
// primed for. The entry is the checkland shape
// (SelfEntersTappedUnless over a Swamp or a Mountain, either
// controlled by the land's controller); the mana ability is a plain
// {W}.
//
// The activated ability creates a delayed triggered ability — "this
// turn" — and the engine has no registry for one, so the land
// carries the trigger itself and gates it on the activation: the
// attack trigger watches EventAttack for an attack by the
// controller's creatures, once per combat ("whenever you attack" is
// one trigger however many creatures were declared — the per-label
// "one or more" dedup), and fires only when the ability has resolved
// this turn, read off the per-turn tally's resolution count.
// Two activations are two delayed triggers and four Warriors, which
// is what the count buys. The Warriors enter tapped and attacking
// the player the first declared attacker was declared against
// (captured from the event); the tokens' IDs are read back off the
// token-created events and handed to a delayed trigger that
// sacrifices them at the beginning of the next end step.
//
// Two shape notes, weaker than printed and never stronger:
//
//   - The trigger lives on the land, so an Encampment that leaves
//     the battlefield after being activated makes no Warriors; the
//     printed delayed trigger would still fire.
//   - The Warriors attack the player the first attacker was declared
//     against rather than a player of the controller's choice (CR
//     508.4). No prompt exists for that choice; the tokens' target is
//     the attack the controller just made.
func init() {
	Register(Spec{
		OracleID:     "33a90122-7280-4481-9b97-5879194cae40",
		Name:         "Dalkovan Encampment",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Warrior tokens attack the same player as your first declared attacker rather than a player of your choice, and the ability does nothing if the land has left the battlefield by the time you attack.",
		},
		Replacements: []game.ReplacementEffect{
			SelfEntersTappedUnless(youControlLandTyped("Swamp", "Mountain")),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
		Activated: []ActivatedAbility{{
			Label: b33DalkovanEncampmentLabel,
			Cost:  Plus(ManaCost("{2}{W}"), TapCost()),
			// The ability's whole effect is the delayed trigger the
			// land's own attack trigger stands in for; resolving it
			// only records the resolution the trigger counts.
			Effect: func(_ *game.Game, _ *game.StackItem) error { return nil },
		}},
		Triggered: []game.TriggeredAbility{{
			OncePerBatch: true,
			Watches:      []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) &&
					b33ResolutionsThisTurn(g, source.InstanceID, b33DalkovanEncampmentLabel) > 0
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, b33DalkovanAttackLabel, b33DalkovanWarriors(ev.Target))
			},
		}},
	})
}
