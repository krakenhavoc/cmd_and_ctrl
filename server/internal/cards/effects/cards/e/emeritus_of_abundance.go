package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emeritus of Abundance // Regrowth — Creature — Elf Druid {2}{G}, 3/4 //
// Sorcery {1}{G} (Secrets of Strixhaven, #4456; preparation card, CR 722):
//
//	"Vigilance
//	 This creature enters prepared.
//	 Whenever this creature attacks, if you control eight or more lands,
//	 this creature becomes prepared."
//
//	Regrowth — "Return target card from your graveyard to your hand."
//
// A preparation card (ADR 0090): the creature registers under the bare
// oracle ID and its prepare spell under "<oracle>#1". "Enters prepared" is
// SelfEntersPrepared. The attack trigger's "if you control eight or more
// lands" is an intervening if (CR 603.4): asked as the attack is declared,
// so with seven lands nothing triggers, and asked again as the ability
// resolves. A creature that is already prepared gains nothing (CR 722.3a).
// The prepare spell is Regrowth word for word, so it shares that card's
// body (regrowth.go).
//
// No simplification.
const EmeritusOfAbundanceOracleID = "e5152e8d-9f55-493d-a0fc-753cc72ee299"

func init() {
	effects.Register(effects.Spec{
		OracleID:        EmeritusOfAbundanceOracleID,
		Name:            "Emeritus of Abundance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ThisAttacked(ev, source, lki, g) && emeritusOfAbundanceEightLands(g, source.Controller)
			}, "Emeritus of Abundance — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					if !emeritusOfAbundanceEightLands(g, item.Controller) {
						return nil
					}
					return effects.BecomePrepared{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     EmeritusOfAbundanceOracleID + "#1",
		Name:         "Regrowth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target card in your graveyard", effects.YouOwn()),
		OnResolve:    effects.ReturnTargetFromGraveyardToHand,
	})
}

// emeritusOfAbundanceEightLands is "if you control eight or more lands",
// counted on the battlefield by what each permanent is now, so a land
// creature counts.
func emeritusOfAbundanceEightLands(g *game.Game, controller uuid.UUID) bool {
	return effects.CountControlled(g, controller, func(c game.Card) bool { return c.IsLand() }) >= 8
}
