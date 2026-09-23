package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kindred Discovery — Enchantment {3}{U}{U} (EDHREC rank ~360):
//
//	"As this enchantment enters, choose a creature type.
//	 Whenever a creature you control of the chosen type enters or
//	 attacks, draw a card."
//
// # A named-tribe permanent (S26) whose payoff is a trigger, not an anthem
//
// ChooseCreatureTypeAsEnters is the same CR 614.12 prompt Cavern of
// Souls and Vanquisher's Banner use, landing the answer on
// Card.NamedTribe. Every OTHER card that reads it is a static
// (TribeFilter{Chosen: true}) checked every layer rebuild; this one
// reads it from a TRIGGERED ability's AppliesTo instead — an ordinary
// field read, not a new kind of "chosen" consumer, since NamedTribe is
// per-instance state on the source card handed to every AppliesTo
// call already.
//
// # One ability, two trigger conditions, one shared type check
//
// "Enters or attacks" is OnAny of the two event kinds (Sun Titan's own
// shape), and the two conditions are set apart only by the wiring
// note EventETB / EventAttack already share for exactly this reason:
// both carry the creature in ev.CardID (attackDeclared's own
// comment). The chosen-type gate — c.IsCreature() &&
// c.HasSubtype(source.NamedTribe) — sits once, after the two branches
// pick out which creature and which controller, rather than being
// duplicated per branch.
//
// Until the enters-a-creature-type prompt is answered, NamedTribe is
// empty and the trigger's own guard refuses to fire on it — a card
// that entered before the answer landed does not retroactively count,
// which is the same "empty matches nothing" floor the static builders
// already hold.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "005ee549-1bf5-478f-bc3f-3e791bd7eecf",
		Name:         "Kindred Discovery",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Kindred Discovery"),
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack}, kindredDiscoveryOfChosenType,
				"Kindred Discovery — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// kindredDiscoveryOfChosenType is "a creature you control of the
// chosen type enters or attacks" — the ETB half reuses
// enteredUnderYourControl (source included, since Kindred Discovery
// is not itself a creature and the exclusion never matters here); the
// attack half reuses attackDeclaredByYou.
func kindredDiscoveryOfChosenType(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source.NamedTribe == "" {
		return false
	}
	switch ev.Kind {
	case game.EventETB:
		c, ok := enteredUnderYourControl(ev, source, g, false)
		return ok && c.IsCreature() && c.HasSubtype(source.NamedTribe)
	case game.EventAttack:
		if !attackDeclaredByYou(ev, source.Controller) {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.IsCreature() && c.HasSubtype(source.NamedTribe)
	default:
		return false
	}
}
