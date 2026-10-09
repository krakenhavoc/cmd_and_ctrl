package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk, Veiled Butcher — Legendary Planeswalker — Garruk {3}{B}{B},
// loyalty 5 (Reality Fracture):
//
//		"If a creature an opponent controls would die, exile it instead.
//		 +2: Up to one target creature gets -4/-1 until your next turn.
//		 −2: Each player sacrifices a creature of their choice. If you
//		     sacrificed a creature this way, create a 4/4 green Beast
//		     creature token with trample.
//		 −3: Each opponent discards two cards. For each opponent who didn't
//		     discard two nonland cards this way, you draw a card."
//
//	  - The static is a CR 614 replacement on a creature moving from the
//	    battlefield to the graveyard, Liesa's shape: the creature is exiled,
//	    so it never died and its dies-triggers do not fire. "An opponent
//	    controls" is read as the creature stands on the battlefield.
//	  - The −2 is the Rise of the Witch-king edict: one prompt per seat,
//	    the token only if the controller's own run named a creature.
//	  - The −3 asks every opponent for two cards with the engine's discard
//	    fan-out, then draws once per opponent whose answer was not two
//	    nonland cards. An opponent with fewer than two cards (or an empty
//	    hand, who is never asked) did not discard two nonland cards.
func init() {
	Register(Spec{
		OracleID:     "4a6972fe-348a-4a19-a1d0-bdf8f1f9c792",
		Name:         "Garruk, Veiled Butcher",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: rfOpponentCreatureWouldDie,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Garruk, Veiled Butcher: exile instead of dying",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Up to one target creature gets -4/-1 until your next turn.",
				Cost:    LoyaltyCost(2),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (ScopedEffectFor{
							Target:   t.ID,
							Mods:     []game.Mod{game.ModifyPTMod(-4, -1)},
							Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
							Label:    "Garruk, Veiled Butcher — -4/-1 until your next turn",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label: "−2: Each player sacrifices a creature of their choice. If you sacrificed a creature this way, create a 4/4 green Beast creature token with trample.",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return EachPlayerSacrifices{
						Match: Creature(),
						Label: "a creature",
						Then: func(ctx *Context, sacrificed game.PromptedSacrifices) error {
							if !sacrificed.Sacrificed(ctx.Controller()) {
								return nil
							}
							return CreateToken{Template: TokenCard("4/4 green Beast with trample"), N: 1}.Apply(ctx)
						},
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−3: Each opponent discards two cards. For each opponent who didn't discard two nonland cards this way, you draw a card.",
				Cost:   LoyaltyCost(-3),
				Effect: garrukVeiledButcherMinusThree,
			},
		},
	})
}

func garrukVeiledButcherMinusThree(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	source := item.SourceCardID
	return g.EachPlayerDiscardsThenForEffect(controller,
		game.DiscardPrompt{
			Source:   source,
			N:        2,
			Question: "Garruk, Veiled Butcher — discard two cards",
		},
		func(g *game.Game, discarded game.PromptedDiscards) error {
			draws := 0
			for _, p := range g.Seats {
				if p == nil || p.Eliminated || p.ID == controller {
					continue
				}
				cards := discarded.By(p.ID)
				ok := len(cards) == 2
				for _, id := range cards {
					if c, found := g.LookupCardForEffect(id); !found || c.IsLand() {
						ok = false
					}
				}
				if !ok {
					draws++
				}
			}
			if draws == 0 {
				return nil
			}
			return g.DrawNForEffect(controller, draws)
		})
}
