package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b594_helpers.go — shared shapes of the Betrayers of Kamigawa Auras
// (S82 OLD batch 594, #5602): the genju cycle, Heart of Light's two-way
// prevention and Phantom Wings' bounce.

// GenjuForm is what a Genju's "{2}: Enchanted <land> becomes ..." makes the land.
type GenjuForm struct {
	// Colors is the colour the land becomes ("a 4/4 green Spirit creature"); nil keeps
	// its colour (Genju of the Realm names none).
	Colors []string
	// Power and Toughness are the base size (CR 613.4b).
	Power, Toughness int
	// Keywords are the keywords it has ("with flying").
	Keywords []string
	// GrantKeys are catalog bundles of abilities it has ("with '{B}: ...'"), declared
	// in the card's Spec.Grants.
	GrantKeys []string
	// Legendary makes it a legendary creature (Genju of the Realm).
	Legendary bool
}

// GenjuBecomesCreature is the effect of a Genju's activated ability: the land the Aura
// is attached to becomes a Spirit creature with the form's characteristics until end of
// turn, and is still a land (the types are ADDED to its own, CR 205.1b). It is the
// permanent the Aura is attached to as the ability resolves (CR 608.2h); if the Aura
// has no host the ability does nothing. The creature type, colour and supertype are
// words of the Aura's text (CR 612.2).
func GenjuBecomesCreature(name string, f GenjuForm) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		host, ok := EnchantedCreatureAsItResolves(ctx)
		if !ok {
			return nil
		}
		mods := []game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Spirit")...),
		}
		if len(f.Colors) > 0 {
			mods = append(mods, game.SetColorsMod(ctx.TextColors(f.Colors...)...))
		}
		if f.Legendary {
			mods = append(mods, game.AddSupertypesMod("Legendary"))
		}
		if len(f.Keywords) > 0 {
			mods = append(mods, game.AddKeywordsMod(f.Keywords...))
		}
		if len(f.GrantKeys) > 0 {
			mods = append(mods, game.GrantAbilitiesMod(f.GrantKeys...))
		}
		mods = append(mods, game.SetBasePTMods(f.Power, f.Toughness)...)
		return ScopedEffectFor{
			Target:   host,
			Mods:     mods,
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    name + " — the enchanted land becomes a Spirit creature until end of turn",
		}.Apply(ctx)
	}
}

// GenjuAbility is "{2}: Enchanted <land> becomes ... until end of turn. It's still a
// land." as the Aura's own activated ability.
func GenjuAbility(name, label string, f GenjuForm) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerAnimate},
		Cost:    ManaCost("{2}"),
		Effect:  GenjuBecomesCreature(name, f),
	}
}

// GenjuReturnTrigger is "When enchanted <land> is put into a graveyard, you may return
// this card from your graveyard to your hand." The Aura is still attached when the land
// is put into the graveyard (the state-based action that puts the Aura there has not run),
// so the trigger is harvested from it (CR 603.10a); the card is returned only if it is
// still in a graveyard as the ability resolves.
func GenjuReturnTrigger(name string) game.TriggeredAbility {
	return Optional(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return EquippedCreatureDied(ev, source)
	}, name+" — you may return this card from your graveyard to your hand", ReturnThisCardFromYourGraveyardToHand),
		name+" — return this card from your graveyard to your hand?")
}

// PreventAllDamageDealtToAndByAttached is "Prevent all damage that would be dealt to and
// dealt by enchanted creature" (Heart of Light): one CR 615.1a prevention that covers
// damage whose recipient or whose source is the permanent the Aura is attached to. The
// attachment is read as the damage would be dealt.
func PreventAllDamageDealtToAndByAttached(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return src != nil && ev.Kind == game.RepEventDamage && ev.DamageAmount > 0 &&
				(src.IsAttachedTo(ev.DamageSource) || src.IsAttachedTo(ev.DamageTarget))
		},
		Replace:    preventThatDamage,
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID { return src.Controller },
		Label:      label,
		RowLabel:   rowLabelFromHeader(label),
	}
}

// ReturnEnchantedCreatureToHand is "Return enchanted creature to its owner's hand" in an
// Aura's own ability whose cost sacrificed the Aura: the creature is the one the Aura was
// attached to as it left (CR 608.2h).
func ReturnEnchantedCreatureToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := EnchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return BounceToHand{Target: host}.Apply(ctx)
}

// MarkOfSakikoMana is the granted trigger's effect: add that much {G}, kept until end of
// turn.
func MarkOfSakikoMana(g *game.Game, item *game.StackItem) error {
	n := item.Trigger.Event.Amount
	if n <= 0 {
		return nil
	}
	return AddMana{Amount: ManaN(n, "G"), Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}}.Apply(NewContext(g, item))
}

// ShizukoMana is "that player adds {G}{G}{G}, kept until end of turn" for the player
// whose upkeep it is.
func ShizukoMana(g *game.Game, item *game.StackItem) error {
	player := TriggeringActor(item)
	if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
		return nil
	}
	return AddMana{Player: player, Produced: "{G}{G}{G}", Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}}.Apply(NewContext(g, item))
}
