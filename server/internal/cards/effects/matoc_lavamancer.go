package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Matoc, Lavamancer — Legendary Creature — Human Barbarian Wizard {3}{R},
// 2/3 (Mystery Booster Commander Edition):
//
//	"Whenever a source you control deals damage, if at least one permanent
//	 or player was dealt exactly 2 of that damage, exile the top card of
//	 your library. You may play it until the end of your next turn.
//	 {R}, {T}, Exile two cards from your graveyard: Matoc deals 2 damage to
//	 any other target."
//
// THE TRIGGER. One per source's damage, not one per recipient: the engine
// emits an event per recipient, so the ability is OncePerBatch keyed on the
// damage source (BatchKey). The first event of a batch from a source that
// dealt exactly 2 to its recipient fires it and the source's later events
// in the same batch are declined, so a spell that deals 2 to each of three
// creatures is one trigger, and a creature dealing 2 to one blocker and 3
// to another (split damage) is one too. Damage to a recipient that is not
// exactly 2 (a source's 3 damage to one creature) does not count. Combat
// damage and noncombat damage both count, and so does damage from a spell
// on the stack. The exile and the window are Reckless Impulse's.
//
// THE ACTIVATION is Grim Lavamancer's, at any target but Matoc itself.
//
// No simplification.
func init() {
	const label = "Matoc, Lavamancer — exile the top card of your library; you may play it until the end of your next turn"
	t := On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventDealDamage || ev.Amount != 2 {
			return false
		}
		controller, ok := b34DamageSourceController(ev, g)
		return ok && controller == source.Controller
	}, label, func(g *game.Game, item *game.StackItem) error {
		return ExileTopNUntilYourNextTurn(NewContext(g, item), 1)
	})
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Source.String() }
	Register(Spec{
		OracleID:     "bbe79ca2-8f05-4f1b-b5d3-84724d3eff3a",
		Name:         "Matoc, Lavamancer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}, Exile two cards from your graveyard: This creature deals 2 damage to any other target.",
			Cost:    Plus(ManaCost("{R}"), TapCost(), ExileFromGraveyard(2, "two cards", nil)),
			Targets: Another(TargetAny()),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  b33DamageChosenTargetFromSource(2),
		}},
	})
}
