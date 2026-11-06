package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Teo, Spirited Glider — Legendary Creature — Human Pilot Ally {3}{U},
// 1/4 (Avatar: The Last Airbender, #4459):
//
//	"Flying
//	 Whenever one or more creatures you control with flying attack, draw
//	 a card, then discard a card. When you discard a nonland card this
//	 way, put a +1/+1 counter on target creature you control."
//
// "One or more ... attack" is one trigger per declaration (OncePerBatch,
// Clandestine Meddler's shape), the attacker's flying read through the
// layered view as it is declared. The loot is Seven of Nine's
// draw-then-discard with a mandatory discard, and only a NONLAND discard
// creates the CR 603.12 reflexive trigger, whose target is chosen as it
// goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ea74940f-01de-4804-92ed-b273fddf6d62",
		Name:            "Teo, Spirited Glider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && attacker.Controller == source.Controller && game.HasKeyword(&attacker, "flying")
			}, "Teo, Spirited Glider — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: teoSpiritedGliderDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				})),
		},
	})
}

var (
	// The loot's discard half, run once the draw has landed. The key is
	// an on-disk identity.
	teoSpiritedGliderDiscard = game.RegisterDrawThen("teo-spirited-glider/discard", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			Question: "Teo, Spirited Glider — discard a card",
		}, teoSpiritedGliderDiscarded, nil)
	})

	// The discard's answer, keyed: a nonland discard creates the
	// reflexive trigger. The key is an on-disk identity.
	teoSpiritedGliderDiscarded = game.RegisterDiscardThen("choose-cards/teo-spirited-glider-discarded", func(g *game.Game, r game.DiscardsDone) error {
		for _, id := range r.Discarded.Cards() {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Teo, Spirited Glider"}
				return WhenYouDo("Teo, Spirited Glider — put a +1/+1 counter on target creature you control",
					teoSpiritedGliderCounterBody).Apply(NewContext(g, parent))
			}
		}
		return nil
	})

	// "Put a +1/+1 counter on target creature you control."
	teoSpiritedGliderCounterBody = game.ReflexiveBody("teo-spirited-glider/counter", simpleBody(PlusOneCounterOnTheTarget),
		constTargets(func() *game.TargetSpec {
			return TargetCreature("target creature you control", YouControl())
		}))
)
