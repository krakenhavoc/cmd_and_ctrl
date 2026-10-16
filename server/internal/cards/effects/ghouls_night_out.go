package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghouls' Night Out — Sorcery {3}{B}{B}:
//
//	"For each player, choose a creature card in that player's graveyard.
//	 Put those cards onto the battlefield under your control. They're
//	 black Zombies in addition to their other colors and types and they
//	 gain decayed. (A creature with decayed can't block. When it
//	 attacks, sacrifice it at end of combat.)"
//
// Not targeted: the caster is asked once per player with a creature
// card in their graveyard (one choose-cards leg each, the caster's own
// graveyard included), and only when every pick is in do the chosen
// cards enter together (ReturnFromGraveyardTogether), so they see each
// other enter. Each one then gets Liliana, Death's Majesty's indefinite
// "black Zombie in addition" effect plus decayed, pinned to that object,
// so it ends when the permanent leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc9f9f14-a11c-4e33-9e11-8bf6ef13cf2b",
		Name:         "Ghouls' Night Out",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var legs []game.ChooseCardsPrompt
			for _, seat := range apnapLivingSeats(ctx.Game) {
				p := ctx.Game.PlayerByIDForEffect(seat)
				if p == nil {
					continue
				}
				var creatures []uuid.UUID
				for _, c := range p.Graveyard.Cards {
					if c.IsCreature() {
						creatures = append(creatures, c.InstanceID)
					}
				}
				if len(creatures) == 0 {
					continue
				}
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    item.Controller,
					FromPlayer: seat,
					Source:     item.SourceCardID,
					Question:   "Ghouls' Night Out — choose a creature card in " + p.Name + "'s graveyard",
					Cards:      creatures,
					Min:        1,
					Max:        1,
					Zone:       game.ZoneGraveyard,
				})
			}
			if len(legs) == 0 {
				return nil
			}
			_, err := ctx.Game.ChooseCardsRunThenForEffect(legs, func(g *game.Game, picks game.PromptedPicks) error {
				return ReturnFromGraveyardTogether{
					Targets:    picks.Cards(),
					Controller: item.Controller,
					Then:       ghoulsNightOutRiders,
				}.Apply(NewContext(g, item))
			})
			return err
		},
	})
}

// ghoulsNightOutRiders makes each returned creature a black Zombie in
// addition to its other colours and types, with decayed, for as long as
// it stays on the battlefield.
func ghoulsNightOutRiders(ctx *Context, entered []uuid.UUID) error {
	for _, id := range entered {
		if err := (ScopedEffectFor{
			Target: id,
			Mods: []game.Mod{
				game.AddSubtypesMod(ctx.TextSubtypes("Zombie")...), game.AddColorsMod(ctx.TextColors("B")...),
				game.AddKeywordsMod(game.KeywordDecayed),
			},
			Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
			Label:    "Ghouls' Night Out — a black Zombie with decayed in addition to its other colors and types",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
