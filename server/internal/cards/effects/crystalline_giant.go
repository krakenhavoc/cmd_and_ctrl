package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crystalline Giant — Artifact Creature — Giant {3}, 3/3:
//
//	"At the beginning of combat on your turn, choose a kind of counter
//	 at random that this creature doesn't have on it from among flying,
//	 first strike, deathtouch, hexproof, lifelink, menace, reach,
//	 trample, vigilance, and +1/+1. Put a counter of that kind on this
//	 creature."
//
// The candidates are the listed kinds the Giant has no counter of as
// the ability resolves; one is drawn through the engine's random door
// (each kind stands for a stable id so ChooseAtRandomForEffect picks
// among them). With all ten on it nothing happens. The Giant must still
// be the same permanent for the counter to land. The keyword kinds are
// keyword counters (CR 122.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49a46cfa-f2e9-4718-a6cb-47dbb0792b56",
		Name:         "Crystalline Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Crystalline Giant — put a random counter it doesn't have on this creature", crystallineGiantCounter),
		},
	})
}

// crystallineGiantKinds is the printed list, in printed order.
var crystallineGiantKinds = []string{
	game.CounterFlying, game.CounterFirstStrike, game.CounterDeathtouch, game.CounterHexproof,
	game.CounterLifelink, game.CounterMenace, game.CounterReach, game.CounterTrample,
	game.CounterVigilance, game.CounterPlusOne,
}

func crystallineGiantKindID(kind string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("crystalline-giant/"+kind))
}

func crystallineGiantCounter(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	byID := map[uuid.UUID]string{}
	var ids []uuid.UUID
	for _, kind := range crystallineGiantKinds {
		if counterCountOn(g, item.SourceCardID, kind) > 0 {
			continue
		}
		id := crystallineGiantKindID(kind)
		byID[id] = kind
		ids = append(ids, id)
	}
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), ids, 1)
	if len(pick) == 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: byID[pick[0]], N: 1}.Apply(ctx)
}
