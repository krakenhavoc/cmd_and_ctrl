package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b11_helpers.go — shared bodies for Kamigawa: Neon Dynasty batch 11
// (S82, #5493). "Modified" is b07IsModified (CR 700.9).

// s82NeoB11ModifiedCreaturesYouControl is the number of modified creatures
// player controls. Caller must hold g.mu.
func s82NeoB11ModifiedCreaturesYouControl(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && b07IsModified(g, c) {
			n++
		}
	}
	return n
}

// s82NeoB11CounterOnEachModified is "put a +1/+1 counter on each modified
// creature you control": the set is fixed as the ability resolves (CR 608.2),
// then each creature still on the battlefield gets one counter.
func s82NeoB11CounterOnEachModified(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && b07IsModified(g, c) {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// s82NeoB11UntapModified is "untap all modified creatures you control"
// (Akki Battle Squad).
type s82NeoB11UntapModified struct{}

func (s82NeoB11UntapModified) Apply(ctx *Context) error {
	g := ctx.Game
	return b16UntapAllYouControlMatching(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.IsCreature() && b07IsModified(g, c)
	})
}

// s82NeoB11ModifiedCreatureYouControlAttacked is the condition of "whenever a
// modified creature you control attacks" (Kami of Celebration, Akki Battle
// Squad): one of your creatures was declared as an attacker and is modified
// (CR 700.9) as the attack is declared.
func s82NeoB11ModifiedCreatureYouControlAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && b07IsModified(g, c)
}

// s82NeoB11ControlsCreatureWithAttached reports whether player controls a
// creature that has an attached permanent matching attached: "an enchanted
// creature" (any Aura) or "an equipped creature" (any Equipment).
func s82NeoB11ControlsCreatureWithAttached(g *game.Game, player uuid.UUID, attached func(game.Card) bool) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsCreature() {
			continue
		}
		for _, a := range g.BattlefieldCardsForEffect() {
			if a.IsAttachedTo(c.InstanceID) && attached(a) {
				return true
			}
		}
	}
	return false
}
