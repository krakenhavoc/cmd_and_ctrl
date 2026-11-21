package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aminatou, the Fateshifter — Legendary Planeswalker — Aminatou {W}{U}{B}, loyalty 3 (S82 GAP, #5948):
//
//	"+1: Draw a card, then put a card from your hand on top of your library.
//	 −1: Exile another target permanent you own, then return it to the
//	 battlefield under your control.
//	 −6: Choose left or right. Each player gains control of all nonland
//	 permanents other than Aminatou controlled by the next player in the chosen
//	 direction.
//	 Aminatou, the Fateshifter can be your commander."
//
// The +1 is Enter the Infinite's draw-then-put-back; the −1 is a flicker under
// the controller's control (Cloudshift's Flicker). The −6 reads who controls what
// before anything changes (the whole exchange is simultaneous, CR 608.2c): for
// each player still in the game it gives them the nonland permanents (other than
// Aminatou) that the next player in the chosen direction controls, as an
// indefinite layer-2 effect (CR 613.1b). "Can be your commander" is deck
// construction, read from the oracle text.
func init() {
	aminatouPutBack = game.RegisterDrawThen("aminatou-the-fateshifter-put-back", func(g *game.Game, d game.DrawThen) error {
		ctx := effects.NewContext(g, &game.StackItem{Controller: d.Player, SourceCardID: d.Source})
		return effects.PutFromHandOnTopInAnyOrder{
			Player: d.Player,
			N:      d.N,
			Label:  "Aminatou, the Fateshifter — put a card from your hand on top of your library",
		}.Apply(ctx)
	})
	aminatouRotate = effects.OptionPickThen("option-pick/aminatou-the-fateshifter-rotate", func(ctx *effects.Context, r game.OptionPicked) error {
		if r.Option == nil {
			return nil
		}
		dir := game.AttackDirectionOfWord(r.Option.Label)
		if dir == "" {
			return nil
		}
		g := ctx.Game
		type move struct{ perm, to uuid.UUID }
		var moves []move
		for _, p := range g.Seats {
			if p == nil || p.Eliminated {
				continue
			}
			from := g.NearestOpponentForEffect(p.ID, dir)
			if from == uuid.Nil {
				continue
			}
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == from && !c.IsLand() && c.InstanceID != r.Source {
					moves = append(moves, move{perm: c.InstanceID, to: p.ID})
				}
			}
		}
		for _, m := range moves {
			if err := (effects.GainControl{
				Target:     m.perm,
				Controller: m.to,
				Duration:   game.IndefiniteDuration(),
				Label:      "Aminatou, the Fateshifter — gains control of the next player's nonland permanents",
			}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	})
	effects.Register(effects.Spec{
		OracleID:        "3a30089d-cd2d-49be-9b06-7a2454117692",
		Name:            "Aminatou, the Fateshifter",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Draw a card, then put a card from your hand on top of your library.",
				Purpose: game.Purpose{Draws: 1},
				Cost:    effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{Ref: aminatouPutBack, Player: item.Controller, Source: item.SourceCardID, N: 1})
				},
			},
			{
				Label:   "−1: Exile another target permanent you own, then return it to the battlefield under your control.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.Another(effects.TargetPermanent("another target permanent you own", effects.YouOwn())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
						return nil
					}
					return effects.Flicker{Target: item.Targets[0].ID, Controller: ctx.Controller()}.Apply(ctx)
				},
			},
			{
				Label: "−6: Choose left or right. Each player gains control of all nonland permanents other than Aminatou controlled by the next player in the chosen direction.",
				Cost:  effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					options := []game.ChoiceOption{{Label: game.AttackDirectionLeftWord}, {Label: game.AttackDirectionRightWord}}
					return effects.PickOption{
						Question: "Aminatou, the Fateshifter — choose left or right (left is the next player in turn order, right the player before)",
						Options:  options,
						ThenKey:  aminatouRotate,
					}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

var (
	aminatouPutBack game.DrawThenRef
	aminatouRotate  game.OptionPickThen
)
