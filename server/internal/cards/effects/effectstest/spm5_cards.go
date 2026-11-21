package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm5Hand seeds a card into a player's hand and returns its id.
func Spm5Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	p.Hand.PushTop(c)
	return c.InstanceID
}

// spm5TapForEach taps a land for its mana ability once per wanted colour,
// answering the colour pick, and checks each pool.
func Spm5TapForEach(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, options int, want []string) {
	t.Helper()
	for _, w := range want {
		if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		pick := RiderLatestManaPick(g, me.ID)
		if pick == nil || len(pick.ColorOptions) != options {
			t.Fatalf("colour options %+v, want %d", pick, options)
		}
		RiderAnswerManaPicks(t, g, me.ID, w)
		if got := Batch01PoolColors(me); len(got) != 1 || got[0] != w {
			t.Errorf("pool %v, want [%s]", got, w)
		}
		me.ManaPool.EmptyPool()
		B08Untap(g, land)
	}
}
