package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ziatora, the Incinerator — Legendary Creature — Demon Dragon
// {3}{B}{R}{G}, 6/6 (EDHREC rank 3098):
//
//	"Flying
//	 At the beginning of your end step, you may sacrifice another
//	 creature. When you do, Ziatora deals damage equal to that
//	 creature's power to any target and you create three Treasure
//	 tokens."
//
// The Riveteers' Fling engine. Flying rides PrintedKeywords. The
// end-step trigger is Ruthless Technomancer's shape: the "you may"
// is the trigger's optional prompt, the creature is chosen as the
// trigger's target (any other creature you control — the picker
// the engine has for a choice among your own permanents), and on
// resolution it is sacrificed (b29SacrificeChosenCreature).
//
// The printed "when you do" is a REFLEXIVE trigger, and it is one
// here: a second ability watching for a sacrifice by the
// controller that happens while the end-step ability is resolving
// (b29SacrificedByYouDuring, keyed on the ability's label in the
// event log). It goes on the stack with its own "any target" clause
// — chosen after the sacrifice, as printed — carrying the
// sacrificed creature's power as it last stood, and on resolution
// deals that much damage and makes the three Treasures whether or
// not the damage lands (b29DamageChosenTargetAndThreeTreasures).
// The power is read when the target is picked, by which time the
// creature is in the graveyard, so it is its printed power plus the
// counters it had when it left (b29PowerAsItLastStood).
//
// Three sandbox simplifications, declared, all weaker than printed:
//
//   - The creature to sacrifice is chosen when the end-step trigger
//     goes on the stack, not on resolution (the Technomancer's
//     caveat): an opponent who removes it in response fizzles the
//     trigger, where printed you would pick another. A second
//     Ziatora cannot be chosen — "another" is by name.
//   - A static power bonus from another permanent (an anthem, an
//     Equipment) is not in the sacrificed creature's last-known
//     power; counters are.
//   - The reflexive trigger is harvested from Ziatora on the
//     battlefield, so if Ziatora herself has left before the
//     end-step ability resolves, the sacrifice still happens but
//     nothing follows it, where printed the damage and Treasures
//     would still come.
func init() {
	Register(Spec{
		OracleID:     "d46c3fb6-f1c7-4a96-ae42-5bce17fc7c1d",
		Name:         "Ziatora, the Incinerator",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the creature to sacrifice when the end-step trigger goes on the stack rather than on resolution, so opponents can respond to the choice.",
			"The sacrificed creature's power is read from its printed value and its counters only — a bonus from another permanent isn't counted in the damage.",
			"If Ziatora leaves the battlefield before the end-step trigger resolves, the sacrifice still happens but no damage is dealt and no Treasures are made.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBeginEndStep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Ziatora, the Incinerator — sacrifice another creature to deal damage equal to its power and make three Treasures?"},
				Targets:        TargetCreature("another creature you control", YouControl(), b03NotNamed("Ziatora, the Incinerator")),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b29ZiatoraSacrificeLabel, b29SacrificeChosenCreature)
				},
			},
			{
				Watches: []game.EventKind{game.EventSacrifice},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := b29SacrificedByYouDuring(ev, source, g, b29ZiatoraSacrificeLabel)
					return ok
				},
				Targets: TargetAny(),
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					power := b29PowerAsItLastStood(g, ev.CardID)
					return game.NewTriggeredItem(source, "Ziatora, the Incinerator — damage equal to the sacrificed creature's power to any target, and three Treasures",
						func(g *game.Game, item *game.StackItem) error {
							return b29DamageChosenTargetAndThreeTreasures(g, item, power)
						})
				},
			},
		},
	})
}
