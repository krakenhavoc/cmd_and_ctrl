package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB377ShuffleTargetIntoOwnersLibrary is "Its owner shuffles it into
// their library" for the one target of a spell (Unravel the Aether): the
// permanent goes into its OWNER's library and that library is shuffled.
// The shuffle is the tuck's continuation, not the next line, because the
// tuck can pause (a commander's owner is asked about the command zone,
// CR 903.9). A target that is no longer legal is left alone (CR 608.2b).
func S82OldB377ShuffleTargetIntoOwnersLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 {
		return nil
	}
	t := item.Targets[0]
	if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	c, ok := g.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{t.ID}, game.TuckOptions{}, game.KeyedTail{Ref: ThenShuffle, Player: c.Owner})
}

// S82OldB377PumpSelfOrEnchanted is "This creature gets +P/+T until end of
// turn. If it's an Aura, enchanted creature gets +P/+T until end of turn
// instead" (Everflame Eidolon): a bestowed Aura boosts the creature it
// enchants, anything else boosts the source itself. Read as the ability
// resolves (CR 608.2h): an Aura that has fallen off is a creature again.
func S82OldB377PumpSelfOrEnchanted(label string, power, toughness int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := item.SourceCardID
		if src, ok := g.LookupCardForEffect(item.SourceCardID); ok && src.IsAura() {
			host := AttachedHostFor(g, item.SourceCardID)
			if host == nil {
				return nil
			}
			target = host.InstanceID
		}
		if !OnBattlefield(g, target) {
			return nil
		}
		return BoostUntilEOT{Target: target, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
	}
}

// S82OldB377ThisOrEnchantedAttacked is the condition of "Whenever this
// creature or enchanted creature attacks" (Spiteful Returned).
func S82OldB377ThisOrEnchantedAttacked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventAttack && (ev.CardID == source.InstanceID || source.IsAttachedTo(ev.CardID))
}

// S82OldB377DefenderLoses is "defending player loses N life" for an attack
// trigger: the defending player is read off the attack event, planeswalker
// and battle attacks included.
func S82OldB377DefenderLoses(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if item.Trigger == nil {
			return nil
		}
		defender := g.DefendingPlayerForAttackForEffect(item.Trigger.Event.Target)
		if defender == uuid.Nil || !DefendingPlayerStillIn(g, defender) {
			return nil
		}
		return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -n)
	}
}
