package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sawhorn Nemesis — Creature — Dinosaur, {3}{R}, 2/4:
//
//	"As this creature enters, choose a player.
//	 If a source would deal damage to the chosen player or a permanent
//	 they control, it deals double that damage instead."
//
// The second card on #980's as-enters player choice, and the reason it
// is worth having: it proves game.Card.ChosenPlayer is a PERMANENT'S
// STORED ANSWER and not a protection back door. Nothing here touches
// protection.go. The same AsEnters hook stores the same field, and a
// replacement effect reads it back — exactly as ChosenColor is read
// back by a mana ability on Coldsteel Heart and by an anthem on
// Heraldic Banner.
//
// The replacement is Fiendish Duo's shape (b34DoubleDamageToOpponents)
// pointed at one stored seat instead of at every opponent, and the
// wider target set is the printed difference: "the chosen player OR A
// PERMANENT THEY CONTROL", so the Duo's player-only test grows a
// battlefield arm. Every source doubles — combat and noncombat, this
// creature's controller's own Bolt included, and the chosen player's
// own creatures damaging their own planeswalker.
//
// Choosing yourself is legal and is occasionally the point (a
// Nemesis plus a damage payoff of your own). CR 616 with a second
// doubler: the affected player orders them and x2 twice is x4 either
// way, so nothing here has to care.
//
// Until the controller answers the prompt the field is uuid.Nil and
// the replacement applies to nobody — the weaker direction, and
// unobservable because the open prompt holds priority.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f9107c5-991a-4c20-9b77-2e2fb4b9dc53",
		Name:         "Sawhorn Nemesis",
		Completeness: CompletenessFull,
		AsEnters:     ChoosePlayerAsEnters("Sawhorn Nemesis", Players),
		Replacements: []game.ReplacementEffect{
			doubleDamageToTheChosenPlayer("Sawhorn Nemesis: double damage to the chosen player"),
		},
	})
}

// doubleDamageToTheChosenPlayer is Sawhorn Nemesis's replacement: any
// source's damage to the seat stored on `src` — or to a permanent that
// seat controls — is doubled.
//
// The chosen player is read LIVE off the source permanent on every
// event rather than captured when the Nemesis entered, for the reason
// every stored-answer reader in the catalog does: the answer arrives
// after the permanent does, and the permanent can leave and be replayed
// with a different answer.
func doubleDamageToTheChosenPlayer(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
				return false
			}
			chosen := ChosenPlayerOf(g, src.InstanceID)
			if chosen == uuid.Nil {
				return false
			}
			if ev.DamageTarget == chosen {
				return true
			}
			// "…or a permanent they control". Control, not ownership,
			// and read now rather than when the damage was announced.
			perm, ok := g.LookupCardForEffect(ev.DamageTarget)
			return ok && perm.Controller == chosen
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
