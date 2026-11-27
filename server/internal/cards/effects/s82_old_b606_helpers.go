package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b606_helpers.go — Mirrodin, batch 606 (#5606): Ogre Leadfoot, Sphere
// of Purity, Loxodon Peacekeeper, Mirror Golem, Spellweaver Helix and Thought
// Prison.

// OldB606BlockedByAnArtifactCreature is Ogre Leadfoot's condition: the source
// is the attacker of a block pair (CR 509.3d: once for each creature that
// blocks it) and the blocker is an artifact creature.
func OldB606BlockedByAnArtifactCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBlock || ev.Target != source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.IsArtifact()
}

// OldB606DestroyTheBlocker destroys the creature named by the block event, as
// long as it is still on the battlefield.
func OldB606DestroyTheBlocker(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !OnBattlefield(g, id) {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(NewContext(g, item))
}

// OldB606PreventOneFromArtifactsToYou is Sphere of Purity's "If an artifact
// would deal damage to you, prevent 1 of that damage." (CR 615.1a). The source
// is read as the damage event snapshotted it (CR 608.2h), by its types.
func OldB606PreventOneFromArtifactsToYou() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageTarget != src.Controller {
				return false
			}
			ch, ok := DamageSourceCharacteristics(ev, g)
			return ok && slices.Contains(ch.Types, "Artifact")
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount--
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      "Sphere of Purity — prevent 1 of that damage",
		RowLabel:   "If an artifact would deal damage to you, prevent 1 of that damage.",
	}
}

// OldB606RustElementalUpkeep is Rust Elemental's upkeep trigger: sacrifice
// another artifact; if you can't, tap it and you lose 4 life (Xathrid Demon's
// shape).
func OldB606RustElementalUpkeep(g *game.Game, item *game.StackItem) error {
	source := item.SourceCardID
	return g.PlayerSacrificesThenKeyedForEffect(source, item.Controller,
		SacrificeSpec("another artifact", Artifact(), NotSelf(source)),
		"Rust Elemental — sacrifice another artifact", 1,
		oldB606RustElementalAfter, []uuid.UUID{item.Controller})
}

// oldB606RustElementalAfter: nothing sacrificed taps the Elemental (when it is
// still there) and costs 4 life. An on-disk key.
var oldB606RustElementalAfter = game.RegisterPromptRunThen("prompt-run/old-b606-rust-elemental", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 || len(r.Sacrifices().Cards()) > 0 {
		return nil
	}
	controller := r.Carry[0]
	ctx := KeyedContext(g, controller, r.Source)
	if err := (TapTarget{Target: r.Source}).Apply(ctx); err != nil {
		return err
	}
	return g.ChangePlayerLifeForEffect(r.Source, controller, -4)
})

// OldB606PeacekeeperUpkeep is Loxodon Peacekeeper's upkeep trigger: the player
// with the lowest life total gains control of it; with a tie, its controller
// chooses among the tied players.
func OldB606PeacekeeperUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	lowest, found := 0, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if !found || p.Life < lowest {
			lowest, found = p.Life, true
		}
	}
	if !found {
		return nil
	}
	var tied []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated && p.Life == lowest {
			tied = append(tied, p.ID)
		}
	}
	if len(tied) == 1 {
		return oldB606PeacekeeperGives(ctx, tied[0])
	}
	return ChoosePlayer{
		Only:     tied,
		Question: "Loxodon Peacekeeper — choose a player with the lowest life total to gain control of it",
		ThenKey:  oldB606PeacekeeperChosen,
	}.Apply(ctx)
}

var oldB606PeacekeeperChosen = ChoosePlayerThen("option-pick/old-b606-loxodon-peacekeeper", func(ctx *Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	return oldB606PeacekeeperGives(ctx, chosen)
})

func oldB606PeacekeeperGives(ctx *Context, to uuid.UUID) error {
	return GainControl{
		Target:     ctx.Source(),
		Controller: to,
		Duration:   game.Duration{Kind: game.Indefinite},
		Label:      "Loxodon Peacekeeper — the player with the lowest life total gains control of it",
	}.Apply(ctx)
}

// OldB606ProtectionFromExiledCardTypes is Mirror Golem's static: protection
// from each card type of the cards exiled with it (CR 607.2a, 702.16).
func OldB606ProtectionFromExiledCardTypes() game.StaticAbility {
	return game.StaticAbility{
		Label:          "This creature has protection from each of the exiled card's card types.",
		Layer:          game.Layer6Ability,
		AppliesTo:      SelfOnly,
		DependsOnExile: true,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			for _, exiled := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
				_, types, _ := game.ParseTypeLine(exiled.TypeLine)
				for _, t := range types {
					if tok := game.ProtectionFromCardType(t); tok != "" && !slices.Contains(c.Abilities, tok) {
						c.Abilities = append(c.Abilities, tok)
					}
				}
			}
		},
	}
}

// OldB606ExiledWith lists the cards exiled with this object, in exile order.
func OldB606ExiledWith(g *game.Game, source *game.Card) []game.Card {
	var out []game.Card
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
		out = append(out, *c)
	}
	return out
}

// OldB606HelixOther is Spellweaver Helix's "copy the other": the cast card
// shares its name with one of the cards exiled with the Helix, and the other
// exiled card is returned. A copy of a card is no card (CR 707.12), so a
// cast copy never counts.
func OldB606HelixOther(g *game.Game, source *game.Card, castID uuid.UUID) (uuid.UUID, bool) {
	cast, ok := g.LookupCardForEffect(castID)
	if !ok || cast.CardCopy {
		return uuid.Nil, false
	}
	exiled := OldB606ExiledWith(g, source)
	for i, c := range exiled {
		if !strings.EqualFold(c.Name, cast.Name) {
			continue
		}
		for j, other := range exiled {
			if j != i {
				return other.InstanceID, true
			}
		}
	}
	return uuid.Nil, false
}

// OldB606HelixCast is the condition of Spellweaver Helix's cast trigger.
func OldB606HelixCast(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	_, ok := OldB606HelixOther(g, source, ev.CardID)
	return ok
}

// OldB606HelixCopyTheOther copies the other exiled card and offers a free
// cast of the copy. The name match is checked again as the ability resolves
// (CR 603.4).
func OldB606HelixCopyTheOther(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Left {
		return nil
	}
	source, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	other, ok := OldB606HelixOther(g, &source, item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	return CopyAndCast{Cards: []uuid.UUID{other}, Free: true}.Apply(ctx)
}

// OldB606ExiledCardSharesWithSpell is Thought Prison's condition: a spell was
// cast that shares a color or a mana value with the card exiled with the
// Prison. A colorless card shares no color; a card with no mana cost has mana
// value 0.
func OldB606ExiledCardSharesWithSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	for _, exiled := range OldB606ExiledWith(g, source) {
		if exiled.ManaValue() == spell.ManaValue() {
			return true
		}
		for _, col := range exiled.EffectiveColors() {
			if slices.Contains(spell.EffectiveColors(), col) {
				return true
			}
		}
	}
	return false
}

// OldB606PrisonDamagesTheCaster is "this artifact deals 2 damage to that
// player": the caster named by the cast event.
func OldB606PrisonDamagesTheCaster(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Actor == uuid.Nil {
		return nil
	}
	return g.DealDamageToPlayerForEffect(item.SourceCardID, item.Trigger.Event.Actor, 2)
}
