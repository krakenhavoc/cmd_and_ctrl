package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b54_helpers.go — shared pieces of the Wilds of Eldraine standard
// batch 54 (#4471).

// std54FoodFightDamage is Food Fight's granted ability: the sacrificed
// artifact deals 1 plus the number of permanents named Food Fight the
// ability's controller controls to the announced target, if still legal.
func std54FoodFightDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	amount := 1
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.Name == "Food Fight" {
			amount++
		}
	}
	for _, t := range ctx.LegalTargets() {
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
	}
	return nil
}

// std54SpellDamagedItsOnlyTarget is Imodane's condition: ev is damage dealt
// by an instant or sorcery spell the source's controller controls, a spell
// whose announced targets are all one creature, to that creature.
func std54SpellDamagedItsOnlyTarget(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil || ev.Target == uuid.Nil {
		return false
	}
	// The spell is mid-resolution: off the stack, but parked as it was.
	spell, item, ok := g.ResolvingSpellForEffect(ev.Source)
	if !ok || item.Kind != game.StackItemSpell || item.Controller != source.Controller || len(item.Targets) == 0 {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID != ev.Target {
			return false
		}
	}
	if !(spell.IsInstant() || spell.IsSorcery()) {
		return false
	}
	victim, ok := g.LookupCardForEffect(ev.Target)
	return ok && victim.IsCreature()
}

// std54HasAdventure reports whether a card "has an Adventure": an
// adventure-layout card whose second half is an Adventure (CR 715). An Omen
// or Vengeance half shares the layout but is no Adventure.
func std54HasAdventure(c game.Card) bool {
	if c.Layout != game.LayoutAdventure {
		return false
	}
	if len(c.Faces) > 1 {
		return strings.Contains(c.Faces[1].TypeLine, "Adventure")
	}
	return true
}

// "Whenever a nontoken creature you control dies this turn" (Gnawing
// Crescendo): a creature that left the battlefield for a graveyard under
// the delayed ability's controller's control, a non-token as it last
// existed (CR 603.10a). The keys are on-disk identities.
var (
	std54NontokenCreatureYouControlDiedCondition = game.DelayedCondition("dies/nontoken-creature-you-control-this-turn", std54NontokenCreatureYouControlDied)
	std54CantBlockRatBody                        = game.SimpleDelayedBody("gnawing-crescendo/cant-block-rat", func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Template: CantBlockRatToken(), N: 1}.Apply(NewContext(g, item))
	})
)

func std54NontokenCreatureYouControlDied(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	dead, ok := diedCreature(ev, g)
	return ok && leftUnderControlOf(ev, dead) == dt.Controller && !IsToken(dead)
}

// std54InstantSorceryOrAdventureInGraveyard counts the cards in a player's
// graveyard that are instant cards, sorcery cards, and/or have an Adventure
// (Frantic Firebolt, Hearth Elemental): each card is counted once.
func std54InstantSorceryOrAdventureInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() || std54HasAdventure(c) {
			n++
		}
	}
	return n
}
