package effects

import (
	"github.com/google/uuid"

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
const emeritusOfAbundanceOracleID = "e5152e8d-9f55-493d-a0fc-753cc72ee299"

func init() {
	Register(Spec{
		OracleID:        emeritusOfAbundanceOracleID,
		Name:            "Emeritus of Abundance",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ThisAttacked(ev, source, lki, g) && emeritusOfAbundanceEightLands(g, source.Controller)
			}, "Emeritus of Abundance — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					if !emeritusOfAbundanceEightLands(g, item.Controller) {
						return nil
					}
					return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     emeritusOfAbundanceOracleID + "#1",
		Name:         "Regrowth",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card in your graveyard", YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}

// emeritusOfAbundanceEightLands is "if you control eight or more lands",
// counted on the battlefield by what each permanent is now, so a land
// creature counts.
func emeritusOfAbundanceEightLands(g *game.Game, controller uuid.UUID) bool {
	return countControlled(g, controller, func(c game.Card) bool { return c.IsLand() }) >= 8
}
