package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const magneticTheftOracle = "9b22cc97-003f-4227-acdc-7a0857674b67"

// TestMagneticTheftMovesAnOpponentsEquipmentOntoYourCreatureWithoutChangingControl
// is the whole card: neither slot says "you control", so it can grab
// an opponent's Equipment and park it on your own creature, and the
// Equipment's CONTROLLER never changes — only the AttachedTo link
// does.
func TestMagneticTheftMovesAnOpponentsEquipmentOntoYourCreatureWithoutChangingControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirSword := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bonesplitter", TypeLine: equipTypeLine,
		OracleID: bonesplitterOracle, Owner: opp.ID, Controller: opp.ID,
	})
	myBear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Magnetic Theft", "Instant", magneticTheftOracle,
		[]game.TargetRef{
			{Kind: game.TargetCard, ID: theirSword},
			{Kind: game.TargetCard, ID: myBear},
		})
	passPriorityAroundTable(t, g)

	if host := attachmentHostOf(t, g, theirSword); host.Kind != game.TargetCard || host.ID != myBear {
		t.Fatalf("AttachedTo = %+v, want card %s", host, myBear)
	}
	if got := effectivePower(t, g, myBear); got != 4 {
		t.Errorf("my Bear should get the Bonesplitter's +2/+0: power %d, want 4", got)
	}
	if c, ok := g.LookupCardForEffect(theirSword); !ok || c.Controller != opp.ID {
		t.Errorf("the Equipment's controller must not change: %+v", c)
	}
}
