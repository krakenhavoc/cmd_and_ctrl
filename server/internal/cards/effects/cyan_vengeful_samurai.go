package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyan, Vengeful Samurai — Legendary Creature — Human Samurai {6}{W},
// 3/3 (deck request #3655):
//
//	"This spell costs {1} less to cast for each creature card in your
//	 graveyard.
//	 Double strike
//	 Whenever one or more creature cards leave your graveyard, put a
//	 +1/+1 counter on Cyan."
//
// The reduction counts creature cards in the caster's graveyard (printed
// type lines: a card there has no layers) and is a self cost modifier,
// so it applies only while Cyan is a spell. The trigger is one ability
// per batch of cards leaving (CR 603.2c), so a mass return is one
// counter, not one per card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a2fa420-c27f-4c14-b285-e0bb2828e053",
		Name:            "Cyan, Vengeful Samurai",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(creatureCardsInYourGraveyard,
				"This spell costs {1} less to cast for each creature card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b02CreatureCardLeftYourGraveyard(ev, source, g)
			}, "Cyan, Vengeful Samurai — put a +1/+1 counter on Cyan",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

// creatureCardsInYourGraveyard is the amount behind "for each creature
// card in your graveyard", never counting the card being cast.
func creatureCardsInYourGraveyard(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID != q.Card.InstanceID && c.IsCreature() {
			n++
		}
	}
	return n
}
