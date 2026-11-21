package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sbondPairCard casts a soulbond card and pairs it with `partner`
// through its enters trigger.
func SbondPairCard(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, power, toughness int, partner uuid.UUID, keywords ...string) uuid.UUID {
	t.Helper()
	id := SbondHand(me, name, typeLine, oracle, power, toughness, append(keywords, game.KeywordSoulbond)...)
	SbondCast(t, g, me, id)
	NghcPassOnce(t, g)
	AnswerOwnPermanents(t, g, me.ID, partner)
	PassPriorityAroundTable(t, g)
	if SbondPartner(g, id) != partner || SbondPartner(g, partner) != id {
		t.Fatalf("%s is not paired with the chosen creature", name)
	}
	return id
}

// sbondDie puts a permanent into its owner's graveyard and settles the
// stack.
func SbondDie(t *testing.T, g *game.Game, owner, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
		t.Fatalf("die: %v", err)
	}
	B21SettleStack(t, g)
}
