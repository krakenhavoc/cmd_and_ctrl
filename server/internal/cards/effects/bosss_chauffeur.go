package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boss's Chauffeur — Creature — Elf Citizen {4}{W}, 0/0 (#5368):
//
//	"This creature enters with a number of +1/+1 counters on it equal to
//	 one plus the number of other creatures you control.
//	 Alliance — Whenever another creature you control enters, put a +1/+1
//	 counter on this creature.
//	 When this creature dies, create a 1/1 green and white Citizen creature
//	 token for each +1/+1 counter on it."
//
// The entry counters are a CR 614.1c replacement counted as the creature
// enters (Dawn of a New Age's shape), the other creatures its controller
// controls plus one. Alliance is Good-Fortune Unicorn's "another creature
// you control enters". The death trigger reads the counters the creature
// had as it last existed on the battlefield (CR 603.10a), so one with none
// makes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a6176cc-98bf-43ac-815a-1f19c4e9329c",
		Name:         "Boss's Chauffeur",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, src *game.Card) int {
				n := 1
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.InstanceID != src.InstanceID && c.IsCreature() && c.Controller == src.Controller {
						n++
					}
				}
				return n
			}, "Boss's Chauffeur: enters with one plus the number of other creatures you control +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			}, "Boss's Chauffeur — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			WhenThisDies("Boss's Chauffeur — create a 1/1 Citizen for each +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					left, ok := ctx.TriggeringPermanent()
					if !ok || left.Counters[game.CounterPlusOne] <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: left.Counters[game.CounterPlusOne]}.Apply(ctx)
				}),
		},
	})
}
