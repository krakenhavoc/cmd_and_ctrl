package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodline Necromancer — Creature — Vampire Wizard {4}{B}, 3/2
// (EDHREC rank 4070):
//
//	"Lifelink
//	 When this creature enters, you may return target Vampire or
//	 Wizard creature card from your graveyard to the battlefield."
//
// A tribal Gravedigger that reanimates straight to the battlefield
// rather than to the hand, which is the whole difference: five mana
// buys a 3/2 lifelink body and whatever Vampire you lost, both
// attacking next turn. In an Edgar Markov or an Inalla deck the
// Necromancer is itself a legal target for the next one, and it
// blinks well for the same reason.
//
// THE FILTER IS "VAMPIRE OR WIZARD", NOT "VAMPIRE WIZARD". Either
// subtype on its own qualifies — the Necromancer is both, which is
// the joke, but a plain Wizard is a legal target. Read off the
// printed type line, because a card in a graveyard has no
// layer-applied characteristics; a changeling in the yard qualifies
// on both counts.
//
// "YOUR GRAVEYARD" is baked into the target clause with YouOwn(), so
// an opponent's dead Vampire is never offered. Owner and new
// controller are therefore the same player, and the reanimation goes
// through the shared path so the returning creature's own entry
// triggers fire.
//
// "YOU MAY" IS A REAL PROMPT (TriggeredAbility.OptionalPrompt). It
// matters more than it looks: declining leaves the graveyard intact
// for a later, bigger reanimation, and the printed card lets you.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f2a54e85-dd8b-462c-a811-c9553dc349a5",
		Name:            "Bloodline Necromancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WhenThisEnters("Bloodline Necromancer — return a Vampire or Wizard to the battlefield",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							_, _ = reanimateSingleTarget(ctx, item.Controller)
							return nil
						}),
					TargetCardInGraveyard("target Vampire or Wizard creature card in your graveyard",
						b39IsVampireOrWizardCreatureCard(), YouOwn()),
				),
				"Bloodline Necromancer — return it to the battlefield?"),
		},
	})
}
