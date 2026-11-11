package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Druid of Purification — Creature — Human Druid {3}{G}, 2/3:
//
//	"When this creature enters, starting with you, each player may choose an
//	 artifact or enchantment you don't control. Destroy each permanent chosen
//	 this way."
//
// The Legend of Yangchen's chain of choices (CR 101.4 order, earlier choices
// known): every player in turn order from the Druid's controller picks up to
// one artifact or enchantment that the Druid's controller does not control,
// not one already chosen; after the last player answers, all the chosen
// permanents are destroyed together. It does not target, so hexproof does
// nothing, and indestructible still saves a permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bb32efd-9e58-4021-bbda-4ffccfa1d601",
		Name:         "Druid of Purification",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Druid of Purification — starting with you, each player may choose an artifact or enchantment you don't control; destroy each permanent chosen this way",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return cmdrB24DruidChoose(ctx, seatsStartingWith(g, ctx.Controller()), 0, nil)
				}),
		},
	})
}

// cmdrB24DruidChoose asks players[i] for their pick from what is left of the
// pool, then the next player; after the last it destroys the lot. Only IDs
// are carried between legs.
func cmdrB24DruidChoose(ctx *Context, players []uuid.UUID, i int, chosen []uuid.UUID) error {
	if i >= len(players) {
		for _, id := range chosen {
			if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
	controller := ctx.Controller()
	taken := append([]uuid.UUID(nil), chosen...)
	return ChoosePermanents{
		Player:   players[i],
		Of:       []uuid.UUID{players[i]},
		Question: "Druid of Purification — you may choose an artifact or enchantment the Druid's controller doesn't control",
		Candidates: func(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
			var pool []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Controller == controller || slices.Contains(taken, c.InstanceID) {
					continue
				}
				if c.IsArtifact() || c.IsEnchantment() {
					pool = append(pool, c.InstanceID)
				}
			}
			return pool, 0, 1
		},
		Keyed: cmdrB24DruidChosen,
		Carry: packIDLists(players, taken),
		Nums:  []int{i},
	}.Apply(ctx)
}

// cmdrB24DruidChosen is one player's pick, then the next player's, keyed. The
// carry is the players in order, then the permanents chosen so far; the
// number is which player this was. Registered in init because the answer asks
// the next player, which names the key again. An on-disk key: never renamed,
// never reused.
var cmdrB24DruidChosen PermanentsThen

func init() {
	cmdrB24DruidChosen = RegisterPermanentsPick("prompt-run/druid-of-purification-destroy", PermanentsDef{
		ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
			players, taken := unpackIDPair(r.Carry)
			i := 0
			if len(r.Nums) > 0 {
				i = r.Nums[0]
			}
			return cmdrB24DruidChoose(ctx, players, i+1, append(taken, picked.Cards()...))
		},
	})
}
