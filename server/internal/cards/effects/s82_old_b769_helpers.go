package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b769_helpers.go — shared shapes of S82 OLD batch 769 (Tempest, life
// gain and loss, instants and sorceries, token makers; #5628): the Tempest
// enchantments that watch a spell colour or an upkeep board, Corpse Dance, and
// Orim's Prayer.

// S82OldB769OpponentLosesLife is "they lose N life" for the opponent whose cast
// set the trigger off (Havoc): a life change the enchantment names, no damage.
func S82OldB769OpponentLosesLife(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		player := TriggeringActor(item)
		if player == uuid.Nil {
			return nil
		}
		return g.ChangePlayerLifeForEffect(ctx.Source(), player, -n)
	}
}

// S82OldB769NoPermanentOfSubtype is "there are no <subtype> [tokens] on the
// battlefield" (Spirit Mirror's Reflection tokens, Sarcomancy's Zombies): any
// controller's, read through the layers.
func S82OldB769NoPermanentOfSubtype(subtype string, tokenOnly bool) func(g *game.Game, you uuid.UUID) bool {
	return func(g *game.Game, _ uuid.UUID) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if tokenOnly && !c.IsToken() {
				continue
			}
			if c.HasSubtypeWord(subtype) {
				return false
			}
		}
		return true
	}
}

// S82OldB769GainLifePerAttacker is "you gain 1 life for each attacking
// creature" (Orim's Prayer): every creature attacking as the ability resolves,
// whichever player or planeswalker it attacks, counted live (a creature that has
// left combat is no longer an attacking creature, CR 506.4).
func S82OldB769GainLifePerAttacker(g *game.Game, item *game.StackItem) error {
	n := len(AttackingCreatures(g))
	if n == 0 {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
}

// S82OldB769ReturnTopCreatureHasteExile is Shallow Grave's body under another
// name: "Return the top creature card of your graveyard to the battlefield. That
// creature gains haste until end of turn. Exile it at the beginning of the next
// end step" (Corpse Dance). The topmost creature card need not be the top card
// of the graveyard; no target, so it is found as the spell resolves.
func S82OldB769ReturnTopCreatureHasteExile(name string) func(item *game.StackItem, ctx *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
		if p == nil {
			return nil
		}
		var id uuid.UUID
		for i := len(p.Graveyard.Cards) - 1; i >= 0; i-- {
			if p.Graveyard.Cards[i].IsCreature() {
				id = p.Graveyard.Cards[i].InstanceID
				break
			}
		}
		if id == uuid.Nil {
			return nil
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: name + " — the returned creature has haste"}).Apply(ctx); err != nil {
			return err
		}
		return ScheduleDelayedTrigger{
			Label: name + " — exile the returned creature",
			Cards: []uuid.UUID{id},
			Body:  ExileListedCardsBody,
		}.Apply(ctx)
	}
}

// S82OldB769PayGoal is what a bot pays for Minion of the Wastes: half its life,
// keeping at least 10 (the floor the heuristic holds every life payment to).
func S82OldB769PayGoal(life int) int {
	return max(0, min(life/2, life-10))
}

// S82OldB769LifePaidPT is "power and toughness are each equal to the life paid
// as it entered" (CR 604.3, layer 7a): the number stored on the permanent as it
// entered, read as the layers are recomputed.
func S82OldB769LifePaidPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		Label:     label,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			c.Power, c.Toughness = source.ChosenNumber, source.ChosenNumber
		},
	}
}
