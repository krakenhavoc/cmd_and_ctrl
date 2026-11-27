package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b476_helpers.go — shared pieces of the Magic 2011 batch 476
// (#5579): Auras with attack and block triggers, a Plains count, a block
// trigger and the attack restriction on an opponent's damage.

// S82OldB476PlainsYouControl is "the number of Plains you control" for an
// Aura's own count ("you" is the Aura's controller). Plains is a word of the
// Aura's text (CR 612.2). It walks the battlefield directly, because a layer
// pass may hold only the read lock (PumpAttachedPer).
func S82OldB476PlainsYouControl(g *game.Game, source *game.Card) int {
	plains := source.TextSubtype("Plains")
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == source.Controller && c.HasSubtype(plains) {
			n++
		}
	}
	return n
}

// S82OldB476EnchantedAttacksOrBlocks is the trigger "When enchanted creature
// attacks or blocks": the enchanted creature is the event's creature, an
// attack event names the attacker and a block event the blocker, once per
// block declaration (Amount <= 1, so a creature that blocks two attackers
// triggers once).
func S82OldB476EnchantedAttacksOrBlocks(label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, OldB484EnchantedAttacksOrBlocks, label, effect)
}

// S82OldB476ThisBlocked is "whenever this creature blocks": the event is a
// block declaration of the source (once per creature it blocks, CR 509.3b).
func S82OldB476ThisBlocked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return B28SelfBlocked(ev, source)
}

// S82OldB476CantAttackUnlessAnOpponentWasDealtDamage is "This creature can't
// attack unless an opponent has been dealt damage this turn" (Bloodcrazed
// Goblin): the attack half of the sentence, a fact about the attacking
// creature's controller read live at declaration (CR 508.1c).
func S82OldB476CantAttackUnlessAnOpponentWasDealtDamage() game.StaticAbility {
	s := selfAttackTargetRestriction(game.AttackTargetRestriction{OpponentMustHaveBeenDealtDamage: true})
	s.Label = "Can't attack unless an opponent has been dealt damage this turn."
	return s
}
