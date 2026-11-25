package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b267_helpers.go — S82 OLD batch 267 (Conspiracy: Take the Crown,
// #5540): the Goblin Hold the Perimeter makes, and the readers its upkeep
// triggers and Echoing Boon share.

// printedS82OldB267GoblinToken is Conspiracy: Take the Crown's 1/1 red Goblin
// creature token with "This creature can't block." Its slug is an on-disk
// identity: never renamed, never reused.
func printedS82OldB267GoblinToken() TokenTemplate {
	return TokenTemplate{
		Slug: "goblin-cant-block",
		Card: game.Card{
			Name:      "Goblin",
			TypeLine:  "Token Creature — Goblin",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This creature can't block.",
	}
}

// S82OldB267GoblinToken is the Goblin's card value for CreateToken.
func S82OldB267GoblinToken() game.Card { return TokenFromCatalog(printedS82OldB267GoblinToken) }

// S82OldB267FirstUpkeep is the When of "at the beginning of your first
// upkeep" (mine true) or "each other player's first upkeep" (mine false): the
// upkeep step that is beginning is that player's first (Player.UpkeepsBegun
// counts it as the step begins, so a trigger reads the upkeep it is in).
func S82OldB267FirstUpkeep(mine bool) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventBeginUpkeep || (ev.Actor == source.Controller) != mine {
			return false
		}
		p := g.PlayerByIDForEffect(ev.Actor)
		return p != nil && !p.Eliminated && p.UpkeepsBegun == 1
	}
}

// S82OldB267SpellTargetsCreatureWithTheChosenName is "if it targets a creature
// you control with the chosen name" for the spell `spell` as it stands on the
// stack now: still there, and at least one of its targets is a creature on the
// battlefield under conspiracy's controller with a name chosen for it.
func S82OldB267SpellTargetsCreatureWithTheChosenName(g *game.Game, spell uuid.UUID, conspiracy *game.Card) bool {
	item := g.StackItemForEffect(spell)
	if item == nil || item.Kind != game.StackItemSpell || conspiracy == nil {
		return false
	}
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !OnBattlefield(g, t.ID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(t.ID); ok && c.IsCreature() && c.Controller == conspiracy.Controller &&
			HasTheChosenName(c, conspiracy) {
			return true
		}
	}
	return false
}

// S82OldB267CastInstantOrSorceryTargetingYourNamedCreature is Echoing Boon's
// When: the conspiracy's controller cast an instant or sorcery spell that
// targets a creature they control with the chosen name.
func S82OldB267CastInstantOrSorceryTargetingYourNamedCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !(c.IsInstant() || c.IsSorcery()) {
		return false
	}
	return S82OldB267SpellTargetsCreatureWithTheChosenName(g, ev.CardID, source)
}

// S82OldB267CopyTheSpellIfItStillTargetsANamedCreature is Echoing Boon's
// whole effect: "if it targets a creature you control with the chosen name"
// is an intervening if (CR 603.4), so the spell is read again as the ability
// resolves; a spell that has been countered or lost that target is not
// copied. The copy's new targets are the controller's to choose (CR 707.10c).
func S82OldB267CopyTheSpellIfItStillTargetsANamedCreature(g *game.Game, item *game.StackItem) error {
	var boon *game.Card
	for _, c := range g.ConspiraciesOf(item.Controller) {
		if c.InstanceID == item.SourceCardID {
			c := c
			boon = &c
		}
	}
	if !S82OldB267SpellTargetsCreatureWithTheChosenName(g, item.Trigger.Event.CardID, boon) {
		return nil
	}
	return CopyTheTriggeringSpellWithNewTargets(g, item)
}
