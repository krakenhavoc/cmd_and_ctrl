package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaito, Bane of Nightmares — Legendary Planeswalker — Kaito {2}{U}{B},
// starting loyalty 4 (Duskmourn: House of Horror, #4466):
//
//	"Ninjutsu {1}{U}{B}
//	 During your turn, as long as Kaito has one or more loyalty counters on
//	 him, he's a 3/4 Ninja creature and has hexproof.
//	 +1: You get an emblem with "Ninjas you control get +1/+1."
//	 0: Surveil 2. Then draw a card for each opponent who lost life this
//	 turn.
//	 −2: Tap target creature. Put two stun counters on it."
//
// The second ability is four layered effects on one condition (it is his
// controller's turn and he holds a loyalty counter), read at every layer
// pass: layer 4 SETS his types to Creature (he stops being a planeswalker,
// so the Kaito subtype goes with it, CR 205.1a) and his subtype to Ninja,
// layer 6 adds hexproof, layer 7b sets 3/4. His loyalty abilities stay
// activatable (CR 606.3 says permanent, not planeswalker). The Ninjutsu
// entry is the shared keyword (ninjutsu.go): he arrives tapped and
// attacking, never declared as an attacker. The emblem is a layer-7c
// anthem on Ninjas, owned by the player who got it (CR 114.2). The 0 is a
// Surveil whose registered tail counts the opponents with life lost this
// turn once the cards are placed. The −2 taps and stuns.
func init() {
	Register(Spec{
		OracleID:        "fa43077b-5d8b-4fb1-8dda-2a9185e6715e",
		Name:            "Kaito, Bane of Nightmares",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Kaito, Bane of Nightmares emblem",
			Text:  "Ninjas you control get +1/+1.",
			Static: []game.StaticAbility{{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && target.HasSubtypeWord("Ninja")
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			}},
		},
		Static: []game.StaticAbility{
			{
				Label:     "During your turn, as long as this has one or more loyalty counters on him, he's a Ninja creature.",
				Layer:     game.Layer4Type,
				AppliesTo: kaitoIsACreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.Types = []string{"Creature"}
					c.Subtypes = source.TextSubtypes("Ninja")
				},
			},
			Labelled(KeywordGrant(kaitoIsACreature, "hexproof"), "During your turn, as long as this has one or more loyalty counters on him, he has hexproof."),
			{
				Label:     "During your turn, as long as this has one or more loyalty counters on him, he's a 3/4.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: kaitoIsACreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 3, 4
				},
			},
		},
		Activated: []ActivatedAbility{
			Ninjutsu("{1}{U}{B}"),
			{
				Label: "+1: You get an emblem with \"Ninjas you control get +1/+1.\"",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "0: Surveil 2. Then draw a card for each opponent who lost life this turn.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return Surveil{N: 2, ThenTail: ctx.ItemTail(kaitoDrawPerOpponentWhoLostLife)}.Apply(ctx)
				},
			},
			{
				Label:   "−2: Tap target creature. Put two stun counters on it.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
						return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}

// kaitoIsACreature is the condition of his second ability: it is his
// controller's turn and he has a loyalty counter.
func kaitoIsACreature(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller) &&
		source.Counters[game.CounterLoyalty] >= 1
}

// kaitoDrawPerOpponentWhoLostLife draws a card for each opponent who lost
// life this turn. An on-disk key.
var kaitoDrawPerOpponentWhoLostLife = CardTail("kaito-bane-of-nightmares/draw-per-opponent-who-lost-life",
	func(ctx *Context, _ game.TailOutcome) error {
		n := 0
		for _, seat := range ctx.Game.Seats {
			if seat != nil && seat.ID != ctx.Controller() && ctx.Game.TurnTallyFor(seat.ID).LifeLost > 0 {
				n++
			}
		}
		return DrawCards{N: n}.Apply(ctx)
	})
