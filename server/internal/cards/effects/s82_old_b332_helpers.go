package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b332_helpers.go — card-side helpers of set OLD, batch 332 (Dragons
// of Tarkir: mana and ramp, Auras, S82, #5551).

// RegisterBasePTAura registers "Enchant creature. Enchanted creature has base
// power and toughness P/T." (Reduce in Stature, as Illusory Wrappings): a layer
// 7b set on the host, under which counters and other pumps still apply
// (CR 613.4b, 613.4c).
func RegisterBasePTAura(oracleID, name string, power, toughness int) {
	Register(Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{SetAttachedBasePT(power, toughness)},
	})
}

// OpponentsCreatureEntered is "whenever a creature an opponent controls
// enters": a creature, as it is now, that entered under the control of a
// player other than the source's controller who is still in the game.
func OpponentsCreatureEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || c.Controller == source.Controller {
		return false
	}
	p := g.PlayerByIDForEffect(c.Controller)
	return p != nil && !p.Eliminated
}

// AttachThisAuraToTheEnteredCreature is "attach this Aura to that creature"
// (Illusory Gains): the creature whose entry triggered the ability, if it is
// still on the battlefield as the ability resolves. The Aura is moved as one of
// its own abilities resolves (AttachSourceForEffect); one that has left the
// battlefield attaches nothing, and a creature the Aura can't legally enchant
// (CR 303.4d, protection) leaves it where it is (CR 701.3b).
func AttachThisAuraToTheEnteredCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left {
		return nil
	}
	return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: ctx.Trigger().Object.Ref().ID})
}
